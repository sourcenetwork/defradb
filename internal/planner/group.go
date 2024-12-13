// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package planner

import (
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// A node responsible for the grouping of documents by a given selection of fields.
type groupNode struct {
	documentIterator
	docMapper

	p *Planner

	// The child select information.  Will be empty if there are no child `_group` items requested.
	childSelects []*mapper.Select

	// The fields to group by - this must be an ordered collection and
	// will include any parent group-by fields (if any)
	groupByFields []mapper.Field

	// The data sources that this node will draw data from.
	dataSources []*dataSource

	values       []core.Doc
	currentIndex int

	execInfo groupExecInfo
}

type groupExecInfo struct {
	// Total number of times groupNode was executed.
	iterations uint64

	// Total number of groups.
	groups uint64

	// Total number of child selections (hidden and visible).
	childSelections uint64

	// Total number of child selections hidden before offset.
	hiddenBeforeOffset uint64

	// Total number of child selections hidden after offset and limit.
	hiddenAfterLimit uint64
}

// Creates a new group node.

// The function is recursive and will construct the node-chain for any child (`_group`) collections.
// `groupSelect` is optional and will typically be nil if the child `_group` is not requested.
func (p *Planner) GroupBy(n *mapper.GroupBy, parsed *mapper.Select, childSelects []*mapper.Select) (*groupNode, error) {
	if n == nil {
		return nil, nil
	}

	dataSources := []*dataSource{}
	// GroupBy must always have at least one data source, for example
	// childSelects may be empty if no group members are requested
	if len(childSelects) == 0 {
		dataSources = append(
			dataSources,
			// If there are no child selects, then we just take the first field index of name _group
			newDataSource(parsed.DocumentMapping.FirstIndexOfName(request.GroupFieldName)),
		)
	}

	for _, childSelect := range childSelects {
		if childSelect.GroupBy != nil {
			// group by fields have to be propagated downwards to ensure correct sub-grouping, otherwise child
			// groups will only group on the fields they explicitly reference
			childSelect.GroupBy.Fields = append(childSelect.GroupBy.Fields, n.Fields...)
		}
		dataSources = append(dataSources, newDataSource(childSelect.Index))
	}

	groupNodeObj := groupNode{
		p:             p,
		childSelects:  childSelects,
		groupByFields: n.Fields,
		dataSources:   dataSources,
		docMapper:     docMapper{parsed.DocumentMapping},
	}
	return &groupNodeObj, nil
}

func (n *groupNode) Kind() string {
	return "groupNode"
}

func (n *groupNode) Init() error {
	// We need to make sure state is cleared down on Init,
	// this function may be called multiple times per instance (for example during a join)
	n.values = nil
	n.currentValue = core.Doc{}
	n.currentIndex = 0

	for _, dataSource := range n.dataSources {
		err := dataSource.Init()
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *groupNode) Start() error {
	for _, dataSource := range n.dataSources {
		err := dataSource.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *groupNode) Prefixes(prefixes []keys.Walkable) {
	for _, dataSource := range n.dataSources {
		dataSource.Prefixes(prefixes)
	}
}

func (n *groupNode) Close() error {
	for _, dataSource := range n.dataSources {
		err := dataSource.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *groupNode) Source() planNode { return n.dataSources[0].Source() }

func (n *groupNode) Next() (bool, error) {
	n.execInfo.iterations++

	if n.values == nil {
		values, err := join(n.dataSources, n.groupByFields, n.documentMapping)
		if err != nil {
			return false, err
		}

		n.values = values.values

		for _, group := range n.values {
			n.execInfo.groups++

			for _, childSelect := range n.childSelects {
				n.execInfo.childSelections++

				subSelect := group.Fields[childSelect.Index]
				if subSelect == nil {
					// If the sub-select is nil we need to set it to an empty array and continue
					group.Fields[childSelect.Index] = []core.Doc{}
					continue
				}

				childDocs := subSelect.([]core.Doc)
				if childSelect.Limit != nil {
					l := uint64(len(childDocs))

					// We must hide all child documents before the offset
					for i := uint64(0); i < childSelect.Limit.Offset && i < l; i++ {
						childDocs[i].Hidden = true

						n.execInfo.hiddenBeforeOffset++
					}

					// We must hide all child documents after the offset plus limit
					for i := childSelect.Limit.Limit + childSelect.Limit.Offset; i < l; i++ {
						childDocs[i].Hidden = true

						n.execInfo.hiddenAfterLimit++
					}
				}
			}
		}
	}

	if n.currentIndex < len(n.values) {
		n.currentValue = n.values[n.currentIndex]
		n.currentIndex++
		return true, nil
	}

	return false, nil
}

func (n *groupNode) simpleExplain() (map[string]any, error) {
	simpleExplainMap := map[string]any{}

	// Get the parent level groupBy attribute(s).
	groupByFields := []string{}
	for _, field := range n.groupByFields {
		groupByFields = append(
			groupByFields,
			field.Name,
		)
	}
	simpleExplainMap["groupByFields"] = groupByFields

	// Get the inner group (child) selection attribute(s).
	if len(n.childSelects) == 0 {
		simpleExplainMap["childSelects"] = nil
	} else {
		childSelects := make([]map[string]any, 0, len(n.childSelects))
		for _, child := range n.childSelects {
			if child == nil {
				continue
			}

			childExplainGraph := map[string]any{}

			childExplainGraph[collectionNameLabel] = child.CollectionName

			c := child.Targetable

			// Get targetable attribute(s) of this child.
			if c.DocIDs.HasValue() {
				childExplainGraph[request.DocIDArgName] = c.DocIDs.Value()
			} else {
				childExplainGraph[request.DocIDArgName] = nil
			}

			if c.Filter == nil {
				childExplainGraph[filterLabel] = nil
			} else {
				childExplainGraph[filterLabel] = c.Filter.ToMap(n.documentMapping)
			}

			if c.Limit != nil {
				childExplainGraph[limitLabel] = map[string]any{
					limitLabel:  c.Limit.Limit,
					offsetLabel: c.Limit.Offset,
				}
			} else {
				childExplainGraph[limitLabel] = nil
			}

			if c.OrderBy != nil {
				innerOrderings := []map[string]any{}
				for _, condition := range c.OrderBy.Conditions {
					orderFieldNames := []string{}

					for _, orderFieldIndex := range condition.FieldIndexes {
						// Try to find the name of this index.
						fieldName, found := n.documentMapping.TryToFindNameFromIndex(orderFieldIndex)
						if !found {
							return nil, client.NewErrFieldIndexNotExist(orderFieldIndex)
						}

						orderFieldNames = append(orderFieldNames, fieldName)
					}
					// Put it all together for this order element.
					innerOrderings = append(innerOrderings,
						map[string]any{
							"fields":    orderFieldNames,
							"direction": string(condition.Direction),
						},
					)
				}

				childExplainGraph["orderBy"] = innerOrderings
			} else {
				childExplainGraph["orderBy"] = nil
			}

			if c.GroupBy != nil {
				innerGroupByFields := []string{}
				for _, fieldOfChildGroupBy := range c.GroupBy.Fields {
					innerGroupByFields = append(
						innerGroupByFields,
						fieldOfChildGroupBy.Name,
					)
				}
				childExplainGraph["groupBy"] = innerGroupByFields
			} else {
				childExplainGraph["groupBy"] = nil
			}

			childSelects = append(childSelects, childExplainGraph)
		}

		simpleExplainMap["childSelects"] = childSelects
	}

	return simpleExplainMap, nil
}

func (n *groupNode) excuteExplain() map[string]any {
	return map[string]any{
		"iterations":            n.execInfo.iterations,
		"groups":                n.execInfo.groups,
		"childSelections":       n.execInfo.childSelections,
		"hiddenBeforeOffset":    n.execInfo.hiddenBeforeOffset,
		"hiddenAfterLimit":      n.execInfo.hiddenAfterLimit,
		"hiddenChildSelections": n.execInfo.hiddenBeforeOffset + n.execInfo.hiddenAfterLimit,
	}
}

// Explain method returns a map containing all attributes of this node that
// are to be explained, subscribes / opts-in this node to be an explainablePlanNode.
func (n *groupNode) Explain(explainType request.ExplainType) (map[string]any, error) {
	switch explainType {
	case request.SimpleExplain:
		return n.simpleExplain()

	case request.ExecuteExplain:
		return n.excuteExplain(), nil

	default:
		return nil, ErrUnknownExplainRequestType
	}
}
