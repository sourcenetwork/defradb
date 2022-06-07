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
	"fmt"
	"sort"
	"strings"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/db/fetcher"
	"github.com/sourcenetwork/defradb/query/graphql/parser"

	cid "github.com/ipfs/go-cid"
	parserTypes "github.com/sourcenetwork/defradb/query/graphql/parser/types"
)

/*

SELECT * From TableA as A JOIN TableB as B ON a.id = b.friend_id

{
	query {
		user {
			age

			friend {
				name
			}

			address {
				street
			}
		}
	}
}

*/

// Wraps a selectNode and all the logic of a plan graph into a single struct for proper plan expansion.
// Executes the top level plan node.
type selectTopNode struct {
	group      *groupNode
	sort       *sortNode
	limit      planNode
	render     *renderNode
	aggregates []aggregateNode

	// source is used pre-wiring of the plan (before expansion and all).
	source planNode

	// plan is the top of the plan graph (the wired and finalized plan graph).
	plan planNode
}

func (n *selectTopNode) Kind() string { return "selectTopNode" }

func (n *selectTopNode) Init() error { return n.plan.Init() }

func (n *selectTopNode) Start() error { return n.plan.Start() }

func (n *selectTopNode) Next() (bool, error) { return n.plan.Next() }

func (n *selectTopNode) Spans(spans core.Spans) { n.plan.Spans(spans) }

func (n *selectTopNode) Value() map[string]interface{} { return n.plan.Value() }

func (n *selectTopNode) Source() planNode { return n.plan.Source() }

// Explain method for selectTopNode returns no attributes but is used to
// subscribe / opt-into being an explainablePlanNode.
func (n *selectTopNode) Explain() (map[string]interface{}, error) {
	// No attributes are returned for selectTopNode.
	return nil, nil
}

func (n *selectTopNode) Close() error {
	if n.plan == nil {
		return nil
	}
	return n.plan.Close()
}

type selectNode struct {
	documentIterator

	p *Planner

	// main data source for the select node.
	source planNode

	// original source that was first given when the select node
	// was created
	origSource planNode

	// cache information about the original data source
	// collection name, meta-data, etc.
	sourceInfo sourceInfo

	// data related to rendering
	renderInfo *renderInfo

	// top level filter expression
	// filter is split between select, scan, and typeIndexJoin.
	// The filters which only apply to the main collection
	// are stored in the root scanNode.
	// The filters that are defined on the root query, but apply
	// to the sub type are defined here in the select.
	// The filters that are defined on the subtype query
	// are defined in the subtype scan node.
	filter *parser.Filter

	groupSelects []*parser.Select
}

func (n *selectNode) Kind() string {
	return "selectNode"
}

func (n *selectNode) Init() error {
	return n.source.Init()
}

func (n *selectNode) Start() error {
	return n.source.Start()
}

// Next iterates through the source plan
// until a doc is returned, applies any
// remaining top level filtering, and
// renders the doc.
func (n *selectNode) Next() (bool, error) {
	for {
		if next, err := n.source.Next(); !next {
			return false, err
		}

		n.currentValue = n.source.Value()
		passes, err := parser.RunFilter(n.currentValue, n.filter, n.p.evalCtx)
		if err != nil {
			return false, err
		}

		if passes {
			return true, err
		}
		// didn't pass, keep looping
	}
}

func (n *selectNode) Spans(spans core.Spans) {
	n.source.Spans(spans)
}

func (n *selectNode) Close() error {
	return n.source.Close()
}

// Explain method returns a map containing all attributes of this node that
// are to be explained, subscribes / opts-in this node to be an explainablePlanNode.
func (n *selectNode) Explain() (map[string]interface{}, error) {
	explainerMap := map[string]interface{}{}

	// Add the filter attribute if it exists.
	if n.filter == nil || n.filter.Conditions == nil {
		explainerMap[filterLabel] = nil
	} else {
		explainerMap[filterLabel] = n.filter.Conditions
	}

	return explainerMap, nil
}

// initSource is the main workhorse for recursively constructing
// all the necessary data source objects. This includes
// creating scanNodes, typeIndexJoinNodes, and splitting
// the necessary filters. Its designed to work with the
// planner.Select construction call.
func (n *selectNode) initSource(parsed *parser.Select) ([]aggregateNode, error) {
	if parsed.CollectionName == "" {
		parsed.CollectionName = parsed.Name
	}
	sourcePlan, err := n.p.getSource(
		parsed.CollectionName,
		parsed.QueryType == parserTypes.VersionedScanQuery,
	)
	if err != nil {
		return nil, err
	}
	n.source = sourcePlan.plan
	n.origSource = sourcePlan.plan
	n.sourceInfo = sourcePlan.info

	// split filter
	// apply the root filter to the source
	// and rootSubType filters to the selectNode
	// @todo: simulate splitting for now
	origScan, ok := n.source.(*scanNode)
	if ok {
		origScan.filter = n.filter
		n.filter = nil

		// If we have both a DocKey and a CID, then we need to run
		// a TimeTravel (History-Traversing Versioned) query, which means
		// we need to propagate the values to the underlying VersionedFetcher
		if parsed.QueryType == parserTypes.VersionedScanQuery {
			c, err := cid.Decode(parsed.CID)
			if err != nil {
				return nil, fmt.Errorf(
					"Failed to propagate VersionFetcher span, invalid CID: %w",
					err,
				)
			}
			spans := fetcher.NewVersionedSpan(
				core.DataStoreKey{DocKey: parsed.DocKeys[0]},
				c,
			) // @todo check len
			origScan.Spans(spans)
		} else if parsed.DocKeys != nil {
			// If we *just* have a DocKey(s), run a FindByDocKey(s) optimization
			// if we have a FindByDockey filter, create a span for it
			// and propagate it to the scanNode
			// @todo: When running the optimizer, check if the filter object
			// contains a _key equality condition, and upgrade it to a point lookup
			// instead of a prefix scan + filter via the Primary Index (0), like here:
			spans := make(core.Spans, len(parsed.DocKeys))
			for i, docKey := range parsed.DocKeys {
				dockeyIndexKey := base.MakeDocKey(sourcePlan.info.collectionDescription, docKey)
				spans[i] = core.NewSpan(dockeyIndexKey, dockeyIndexKey.PrefixEnd())
			}
			origScan.Spans(spans)
		}
	}

	return n.initFields(parsed)
}

func (n *selectNode) initFields(parsed *parser.Select) ([]aggregateNode, error) {
	// re-organize the fields slice into reverse-alphabetical
	// this makes sure the reserved database fields that start with
	// a "_" end up at the end. So if/when we build our MultiNode
	// all the AppendPlans end up at the end.
	sort.Slice(parsed.Fields, func(i, j int) bool {
		return !(strings.Compare(parsed.Fields[i].GetName(), parsed.Fields[j].GetName()) < 0)
	})

	aggregates := []aggregateNode{}
	// loop over the sub type
	// at the moment, we're only testing a single sub selection
	for _, field := range parsed.Fields {
		switch f := field.(type) {
		case *parser.Select:
			var plan aggregateNode
			var aggregateError error
			// @todo: check select type:
			// - TypeJoin
			// - commitScan
			if f.Statement.Name.Value == parserTypes.CountFieldName {
				aggregateError = n.joinAggregatedChild(parsed, f)
				if aggregateError != nil {
					return nil, aggregateError
				}
				plan, aggregateError = n.p.Count(f, parsed)
			} else if f.Statement.Name.Value == parserTypes.SumFieldName {
				aggregateError = n.joinAggregatedChild(parsed, f)
				if aggregateError != nil {
					return nil, aggregateError
				}
				plan, aggregateError = n.p.Sum(&n.sourceInfo, f, parsed)
			} else if f.Statement.Name.Value == parserTypes.AverageFieldName {
				averageSource, err := f.GetAggregateSource(parsed)
				if err != nil {
					return nil, err
				}
				childField := n.p.getSourceProperty(averageSource, parsed)
				// We must not count nil values else they will corrupt the average (they would be counted otherwise)
				// so here we append the nil filter to the average (and child nodes) before joining any children.
				// The nil clause is appended to average and sum as well as count in order to make it much easier
				// to find them and safely identify existing nodes.
				appendNotNilFilter(f, childField)

				// then we join the potentially missing child using the dummy field (will be used by sum+count)
				aggregateError = n.joinAggregatedChild(parsed, f)
				if aggregateError != nil {
					return nil, aggregateError
				}

				// value of the suffix is unimportant here, just needs to be unique
				dummyCountField := f.Clone(fmt.Sprintf("%s_internalCount", f.Name), parserTypes.CountFieldName)
				countField, countExists := tryGetField(parsed.Fields, dummyCountField)
				// Note: sumExists will always be false until we support filtering by nil in the query
				if !countExists {
					countField = dummyCountField
					countPlan, err := n.p.Count(countField, parsed)
					if err != nil {
						return nil, err
					}
					aggregates = append(aggregates, countPlan)
				}

				// value of the suffix is unimportant here, just needs to be unique
				dummySumField := f.Clone(fmt.Sprintf("%s_internalSum", f.Name), parserTypes.SumFieldName)
				sumField, sumExists := tryGetField(parsed.Fields, dummySumField)
				// Note: sumExists will always be false until we support filtering by nil in the query
				if !sumExists {
					sumField = dummySumField
					sumPlan, err := n.p.Sum(&n.sourceInfo, sumField, parsed)
					if err != nil {
						return nil, err
					}
					aggregates = append(aggregates, sumPlan)
				}

				plan, aggregateError = n.p.Average(sumField, countField, f)
			} else if f.Name == parserTypes.VersionFieldName { // reserved sub type for object queries
				commitSlct := &parser.CommitSelect{
					Name:  f.Name,
					Alias: f.Alias,
					// Type:   parser.LatestCommits,
					Fields: f.Fields,
				}
				// handle _version sub selection query differently
				// if we are executing a regular Scan query
				// or a TimeTravel query.
				if parsed.QueryType == parserTypes.VersionedScanQuery {
					// for a TimeTravel query, we don't need the Latest
					// commit. Instead, _version references the CID
					// of that Target version we are querying.
					// So instead of a LatestCommit subquery, we need
					// a OneCommit subquery, with the supplied parameters.
					commitSlct.DocKey = parsed.DocKeys[0] // @todo check length
					commitSlct.Cid = parsed.CID
					commitSlct.Type = parser.OneCommit
				} else {
					commitSlct.Type = parser.LatestCommits
				}
				commitPlan, err := n.p.CommitSelect(commitSlct)
				if err != nil {
					return nil, err
				}

				if err := n.addSubPlan(field.GetName(), commitPlan); err != nil {
					return nil, err
				}
			} else if f.Root == parserTypes.ObjectSelection {
				if f.Statement.Name.Value == parserTypes.GroupFieldName {
					n.groupSelects = append(n.groupSelects, f)
				} else {
					// nolint:errcheck
					n.addTypeIndexJoin(f) // @TODO: ISSUE#158
				}
			}

			if aggregateError != nil {
				return nil, aggregateError
			}

			if plan != nil {
				aggregates = append(aggregates, plan)
			}
		}
	}

	return aggregates, nil
}

// appendNotNilFilter appends a not nil filter for the given child field
// to the given Select.
func appendNotNilFilter(field *parser.Select, childField string) {
	if field.Filter == nil {
		field.Filter = &parser.Filter{}
	}

	if field.Filter.Conditions == nil {
		field.Filter.Conditions = map[string]interface{}{}
	}

	childBlock, hasChildBlock := field.Filter.Conditions[childField]
	if !hasChildBlock {
		childBlock = map[string]interface{}{}
		field.Filter.Conditions[childField] = childBlock
	}

	typedChildBlock := childBlock.(map[string]interface{})
	typedChildBlock["$ne"] = nil
}

// tryGetField scans the given list of fields for an item matching the given searchTerm.
// Will return the matched value and true if one is found, else will return nil and false.
func tryGetField(fields []parser.Selection, searchTerm *parser.Select) (*parser.Select, bool) {
	for _, field := range fields {
		f, isSelect := field.(*parser.Select)
		if !isSelect {
			continue
		}

		if f.Equal(*searchTerm) {
			return f, true
		}
	}

	return nil, false
}

// Join any child collections required by the given transformation if the child
//  collections have not been requested for render by the consumer
func (n *selectNode) joinAggregatedChild(
	parsed *parser.Select,
	field *parser.Select,
) error {
	source, err := field.GetAggregateSource(parsed)
	if err != nil {
		return err
	}

	targetField := field.Clone(source.HostProperty, source.ExternalHostName)

	hasChildProperty := false
	for _, siblingField := range parsed.Fields {
		siblingSelect, isSelect := siblingField.(*parser.Select)
		if isSelect && siblingSelect.Equal(*targetField) {
			hasChildProperty = true
			break
		}
	}

	// If the child item is not requested, then we have add in the necessary components
	//  to force the child records to be scanned through (they wont be rendered)
	if !hasChildProperty {
		if source.ExternalHostName == parserTypes.GroupFieldName {
			hasGroupSelect := false
			for _, childSelect := range n.groupSelects {
				if childSelect.Equal(*targetField) {
					hasGroupSelect = true
					break
				}

				// if the child filter is nil then we can use it as source with no meaningful overhead
				//
				// todo - this might be incorrect when the groupby contains a filter - test
				// consider adding fancy inclusive logic
				if childSelect.ExternalName == parserTypes.GroupFieldName && childSelect.Filter == nil {
					hasGroupSelect = true
					break
				}
			}
			if !hasGroupSelect {
				newGroup := &parser.Select{
					Alias:        source.HostProperty,
					Name:         fmt.Sprintf("_agg%v", len(parsed.Fields)),
					ExternalName: parserTypes.GroupFieldName,
					Hidden:       true,
				}
				parsed.Fields = append(parsed.Fields, newGroup)
				n.groupSelects = append(n.groupSelects, newGroup)
			}
		} else if parsed.Root != parserTypes.CommitSelection {
			fieldDescription, _ := n.sourceInfo.collectionDescription.GetField(source.HostProperty)
			if fieldDescription.Kind == client.FieldKind_FOREIGN_OBJECT_ARRAY {
				subtype := &parser.Select{
					Name:         source.HostProperty,
					ExternalName: parserTypes.GroupFieldName,
				}
				return n.addTypeIndexJoin(subtype)
			}
		}
	}

	return nil
}

func (n *selectNode) addTypeIndexJoin(subSelect *parser.Select) error {
	typeIndexJoin, err := n.p.makeTypeIndexJoin(n, n.origSource, subSelect)
	if err != nil {
		return err
	}

	if err := n.addSubPlan(subSelect.Name, typeIndexJoin); err != nil {
		return err
	}

	return nil
}

func (n *selectNode) Source() planNode { return n.source }

// func appendSource() {}

// func (n *selectNode) initRender(
//     fields []*client.FieldDescription,
//     aliases []string,
//) error {
// 	return n.p.render(fields, aliases)
// }

// SubSelect is used for creating Select nodes used on sub selections,
// not to be used on the top level selection node.
// This allows us to disable rendering on all sub Select nodes
// and only run it at the end on the top level select node.
func (p *Planner) SubSelect(parsed *parser.Select) (planNode, error) {
	plan, err := p.Select(parsed)
	if err != nil {
		return nil, err
	}

	// if this is a sub select plan, we need to remove the render node
	// as the final top level selectTopNode will handle all sub renders
	top := plan.(*selectTopNode)
	top.render = nil
	return top, nil
}

func (p *Planner) SelectFromSource(
	parsed *parser.Select,
	source planNode,
	fromCollection bool,
	providedSourceInfo *sourceInfo,
) (planNode, error) {
	s := &selectNode{
		p:          p,
		source:     source,
		origSource: source,
	}
	s.filter = parsed.Filter
	limit := parsed.Limit
	sort := parsed.OrderBy
	groupBy := parsed.GroupBy
	s.renderInfo = &renderInfo{}

	if providedSourceInfo != nil {
		s.sourceInfo = *providedSourceInfo
	}

	if fromCollection {
		desc, err := p.getCollectionDesc(parsed.Name)
		if err != nil {
			return nil, err
		}

		s.sourceInfo = sourceInfo{desc}
	}

	aggregates, err := s.initFields(parsed)
	if err != nil {
		return nil, err
	}

	groupPlan, err := p.GroupBy(groupBy, s.groupSelects)
	if err != nil {
		return nil, err
	}

	limitPlan, err := p.HardLimit(limit)
	if err != nil {
		return nil, err
	}

	sortPlan, err := p.OrderBy(sort)
	if err != nil {
		return nil, err
	}

	top := &selectTopNode{
		source:     s,
		render:     p.render(parsed),
		limit:      limitPlan,
		sort:       sortPlan,
		group:      groupPlan,
		aggregates: aggregates,
	}
	return top, nil
}

// Select constructs a SelectPlan
func (p *Planner) Select(parsed *parser.Select) (planNode, error) {
	s := &selectNode{p: p}
	s.filter = parsed.Filter
	limit := parsed.Limit
	sort := parsed.OrderBy
	groupBy := parsed.GroupBy
	s.renderInfo = &renderInfo{}
	s.groupSelects = []*parser.Select{}

	aggregates, err := s.initSource(parsed)
	if err != nil {
		return nil, err
	}

	groupPlan, err := p.GroupBy(groupBy, s.groupSelects)
	if err != nil {
		return nil, err
	}

	limitPlan, err := p.HardLimit(limit)
	if err != nil {
		return nil, err
	}

	sortPlan, err := p.OrderBy(sort)
	if err != nil {
		return nil, err
	}

	top := &selectTopNode{
		source:     s,
		render:     p.render(parsed),
		limit:      limitPlan,
		sort:       sortPlan,
		group:      groupPlan,
		aggregates: aggregates,
	}
	return top, nil
}
