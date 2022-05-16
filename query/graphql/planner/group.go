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
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
)

// A node responsible for the grouping of documents by a given selection of fields.
type groupNode struct {
	documentIterator

	p *Planner

	// The child select information.  Will be empty if there are no child `_group` items requested.
	childSelects []*parser.Select

	// The fields to group by - this must be an ordered collection and
	// will include any parent group-by fields (if any)
	groupByFields []string

	// The data sources that this node will draw data from.
	dataSources []*dataSource

	values       []core.Doc
	currentIndex int
}

// Creates a new group node.  The function is recursive and will construct the node-chain for any
//  child (`_group`) collections. `groupSelect` is optional and will typically be nil if the
//  child `_group` is not requested.
func (p *Planner) GroupBy(n *parser.GroupBy, childSelects []*parser.Select) (*groupNode, error) {
	if n == nil {
		return nil, nil
	}

	dataSources := []*dataSource{}
	// GroupBy must always have at least one data source, for example
	// childSelects may be empty if no group members are requested
	if len(childSelects) == 0 {
		dataSources = append(dataSources, newDataSource(parser.GroupFieldName))
	}

	for _, childSelect := range childSelects {
		if childSelect.GroupBy != nil {
			// group by fields have to be propagated downwards to ensure correct sub-grouping, otherwise child
			// groups will only group on the fields they explicitly reference
			childSelect.GroupBy.Fields = append(childSelect.GroupBy.Fields, n.Fields...)
		}
		dataSources = append(dataSources, newDataSource(childSelect.Name))
	}

	groupNodeObj := groupNode{
		p:             p,
		childSelects:  childSelects,
		groupByFields: n.Fields,
		dataSources:   dataSources,
	}
	return &groupNodeObj, nil
}

func (n *groupNode) Init() error {
	// We need to make sure state is cleared down on Init,
	// this function may be called multiple times per instance (for example during a join)
	n.values = nil
	n.currentValue = nil
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

func (n *groupNode) Spans(spans core.Spans) {
	for _, dataSource := range n.dataSources {
		dataSource.Spans(spans)
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
	if n.values == nil {
		values, err := join(n.dataSources, n.groupByFields)
		if err != nil {
			return false, err
		}

		n.values = values.values

		for _, group := range n.values {
			for _, childSelect := range n.childSelects {
				subSelect, hasSubSelect := group[childSelect.Name]
				if !hasSubSelect {
					continue
				}

				childDocs := subSelect.([]core.Doc)
				if childSelect.Limit != nil {
					l := int64(len(childDocs))

					// We must hide all child documents before the offset
					for i := int64(0); i < childSelect.Limit.Offset && i < l; i++ {
						childDocs[i][parser.HiddenFieldName] = struct{}{}
					}

					// We must hide all child documents after the offset plus limit
					for i := childSelect.Limit.Limit + childSelect.Limit.Offset; i < l; i++ {
						childDocs[i][parser.HiddenFieldName] = struct{}{}
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
