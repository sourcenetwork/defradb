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
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/graphql-go/graphql/language/ast"
	cid "github.com/ipfs/go-cid"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/db/fetcher"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
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

// wraps a selectNode and all the logic of a plan
// graph into a single struct for proper plan
// expansion
// Executes the top level plan node.
type selectTopNode struct {
	source     planNode
	group      *groupNode
	sort       *sortNode
	limit      planNode
	render     *renderNode
	aggregates []aggregateNode

	// top of the plan graph
	plan planNode

	// plan -> limit -> sort -> sort.plan = (values -> container | SORT_STRATEGY) -> render -> source

	// ... source -> MultiNode -> TypeJoinNode.plan = (typeJoinOne | typeJoinMany) -> scanNode
}

func (n *selectTopNode) Init() error                   { return n.plan.Init() }
func (n *selectTopNode) Start() error                  { return n.plan.Start() }
func (n *selectTopNode) Next() (bool, error)           { return n.plan.Next() }
func (n *selectTopNode) Spans(spans core.Spans)        { n.plan.Spans(spans) }
func (n *selectTopNode) Value() map[string]interface{} { return n.plan.Value() }
func (n *selectTopNode) Source() planNode              { return n.source }
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

	groupSelect *parser.Select
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
		parsed.QueryType == parser.VersionedScanQuery,
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
		if parsed.QueryType == parser.VersionedScanQuery {
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
			if f.Statement.Name.Value == parser.CountFieldName {
				plan, aggregateError = n.p.Count(f)
			} else if f.Statement.Name.Value == parser.SumFieldName {
				plan, aggregateError = n.p.Sum(&n.sourceInfo, f, parsed)
			} else if f.Statement.Name.Value == parser.AverageFieldName {
				// Average utilises count and sum in order to calculate it's return value,
				// so we have to add those nodes here if they do not exist else the generated
				// field names could collide.  Value is currently 3 as if each field was an
				// _avg field - then the final number of fields would be 3N (count+sum+average)
				const fieldLenMultiplier = 3
				countField, countExists := tryGetAggregateField(
					n.p.ctx,
					parsed.Fields,
					parser.CountFieldName,
					f.Statement.Arguments,
				)

				if !countExists {
					const countFieldIndexOffset = 1
					astField := ast.Field{
						Name:      &ast.Name{Value: parser.CountFieldName},
						Arguments: f.Statement.Arguments,
					}
					// We need to make sure the new aggregate index does not clash with any existing aggregate fields
					countFieldIndex := (len(parsed.Fields) * fieldLenMultiplier) + countFieldIndexOffset
					countField, aggregateError = parser.ParseSelect(f.Root, &astField, countFieldIndex)
					if aggregateError != nil {
						return nil, aggregateError
					}

					countPlan, err := n.p.Count(countField)
					if err != nil {
						return nil, err
					}
					// We must not count nil values else they will corrupt the average
					averageSource, err := f.GetAggregateSource()
					if err != nil {
						return nil, err
					}
					childField := n.p.getSourceProperty(averageSource, parsed)
					countPlan.filter = &parser.Filter{
						Conditions: map[string]interface{}{
							childField: map[string]interface{}{
								"$ne": nil,
							},
						},
					}
					aggregates = append(aggregates, countPlan)
				}

				sumField, sumExists := tryGetAggregateField(n.p.ctx, parsed.Fields, parser.SumFieldName, f.Statement.Arguments)
				if !sumExists {
					const sumFieldIndexOffset = 2
					astField := ast.Field{
						Name:      &ast.Name{Value: parser.SumFieldName},
						Arguments: f.Statement.Arguments,
					}
					// We need to make sure the new aggregate index does not clash with any existing aggregate fields
					sumFieldIndex := (len(parsed.Fields) * fieldLenMultiplier) + sumFieldIndexOffset
					sumField, aggregateError = parser.ParseSelect(f.Root, &astField, sumFieldIndex)
					if aggregateError != nil {
						return nil, aggregateError
					}
					sumPlan, err := n.p.Sum(&n.sourceInfo, sumField, parsed)
					if err != nil {
						return nil, err
					}
					aggregates = append(aggregates, sumPlan)
				}

				plan, aggregateError = n.p.Average(sumField, countField, f)
			} else if f.Name == parser.VersionFieldName { // reserved sub type for object queries
				commitSlct := &parser.CommitSelect{
					Name:  f.Name,
					Alias: f.Alias,
					// Type:   parser.LatestCommits,
					Fields: f.Fields,
				}
				// handle _version sub selection query differently
				// if we are executing a regular Scan query
				// or a TimeTravel query.
				if parsed.QueryType == parser.VersionedScanQuery {
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
			} else if f.Root == parser.ObjectSelection {
				if f.Name == parser.GroupFieldName {
					n.groupSelect = f
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

				aggregateError = n.joinAggregatedChild(parsed, f)
				if aggregateError != nil {
					return nil, aggregateError
				}
			}
		}
	}

	return aggregates, nil
}

// tryGetAggregateField attempts to find an existing aggregate field that matches the given
// name and arguements.  Will return the match field and true if one is found, false otherwise.
func tryGetAggregateField(
	ctx context.Context,
	fields []parser.Selection,
	name string,
	arguements []*ast.Argument,
) (*parser.Select, bool) {
	for _, field := range fields {
		f, isSelect := field.(*parser.Select)
		if !isSelect {
			continue
		}

		// compare the name on the statement vs the given name.
		// the field name should not be compared here as it may be
		// different.
		if f.Statement.Name.Value == name {
			allArguementsMatch := true

			for _, possibleMatchingArguement := range f.Statement.Arguments {
				for _, targetArguement := range arguements {
					if possibleMatchingArguement.Name.Value != targetArguement.Name.Value {
						allArguementsMatch = false
						break
					}

					if !areASTValuesEqual(ctx, possibleMatchingArguement.Value, targetArguement.Value) {
						allArguementsMatch = false
						break
					}
				}
				if !allArguementsMatch {
					break
				}
			}

			if allArguementsMatch {
				return f, true
			}
		}
	}
	return nil, false
}

func areASTValuesEqual(ctx context.Context, thisValue ast.Value, otherValue ast.Value) bool {
	if thisValue.GetKind() != otherValue.GetKind() {
		return false
	}

	switch thisTypedValue := thisValue.GetValue().(type) {
	case *ast.Variable, *ast.IntValue, *ast.FloatValue, *ast.StringValue, *ast.EnumValue, *ast.BooleanValue:
		// For these primative types we just have to compare the inner values
		if thisTypedValue != otherValue.GetValue() {
			return false
		}
	case *ast.ObjectValue:
		return areASTValuesEqual(ctx, thisTypedValue, otherValue.GetValue().(*ast.ObjectValue))
	case *ast.ListValue:
		otherTypedValue := otherValue.GetValue().(*ast.ListValue)
		if len(thisTypedValue.Values) != len(otherTypedValue.Values) {
			return false
		}
		for i, innerValue := range thisTypedValue.Values {
			if !areASTValuesEqual(ctx, innerValue, otherTypedValue.Values[i]) {
				return false
			}
		}
	case []*ast.ObjectField:
		for i, field := range thisTypedValue {
			otherTypedValue := otherValue.GetValue().([]*ast.ObjectField)
			if len(thisTypedValue) != len(otherTypedValue) {
				return false
			}
			if !areASTValuesEqual(ctx, field, otherTypedValue[i]) {
				return false
			}
		}
	default:
		// If we do not recognise the type, we should state that they do not equal and continue
		log.Error(
			ctx,
			"Could not evaluate arguement equality, unknown type.",
			logging.NewKV("Type", fmt.Sprintf("%T", thisValue.GetValue())),
		)
		return false
	}

	return true
}

// Join any child collections required by the given transformation if the child
//  collections have not been requested for render by the consumer
func (n *selectNode) joinAggregatedChild(
	parsed *parser.Select,
	field *parser.Select,
) error {
	source, err := field.GetAggregateSource()
	if err != nil {
		return err
	}

	hasChildProperty := false
	for _, field := range parsed.Fields {
		if source.HostProperty == field.GetName() {
			hasChildProperty = true
			break
		}
	}

	// If the child item is not requested, then we have add in the necessary components
	//  to force the child records to be scanned through (they wont be rendered)
	if !hasChildProperty {
		if source.HostProperty == parser.GroupFieldName {
			// It doesn't really matter at the moment if multiple counts are requested
			//  and we overwrite the n.groupSelect property
			n.groupSelect = &parser.Select{
				Name: parser.GroupFieldName,
			}
		} else if parsed.Root != parser.CommitSelection {
			fieldDescription, _ := n.sourceInfo.collectionDescription.GetField(source.HostProperty)
			if fieldDescription.Kind == client.FieldKind_FOREIGN_OBJECT_ARRAY {
				subtype := &parser.Select{
					Name: source.HostProperty,
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

	groupPlan, err := p.GroupBy(groupBy, s.groupSelect)
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

	aggregates, err := s.initSource(parsed)
	if err != nil {
		return nil, err
	}

	groupPlan, err := p.GroupBy(groupBy, s.groupSelect)
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
