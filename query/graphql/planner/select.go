// Copyright 2020 Source Inc.
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
	"sort"
	"strings"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
)

// wraps a selectNode and all the logic of a plan
// graph into a single struct for proper plan
// expansion
// Executes the top level plan node.
type selectTopNode struct {
	source planNode
	group  *groupNode
	sort   *sortNode
	limit  *limitNode
	render *renderNode

	// top of the plan graph
	plan planNode
}

func (n *selectTopNode) Init() error                    { return n.plan.Init() }
func (n *selectTopNode) Start() error                   { return n.plan.Start() }
func (n *selectTopNode) Next() (bool, error)            { return n.plan.Next() }
func (n *selectTopNode) Spans(spans core.Spans)         { n.plan.Spans(spans) }
func (n *selectTopNode) Values() map[string]interface{} { return n.plan.Values() }
func (n *selectTopNode) Source() planNode               { return n.source }
func (n *selectTopNode) Close() error {
	if n.plan == nil {
		return nil
	}
	return n.plan.Close()
}

type renderInfo struct {
	numResults int
	fields     []string
	aliases    []string
}

type selectNode struct {
	p *Planner

	// main data source for the select node.
	source planNode

	// origal source that was first given when the select node
	// was created
	origSource planNode

	// cache information about the original data source
	// collection name, meta-data, etc.
	sourceInfo sourceInfo

	// data related to rendering
	renderInfo *renderInfo

	// internal doc pointer
	// produced when Values()
	// is called.
	doc map[string]interface{}

	// top level filter expression
	// filter is split between select, scan, and typeIndexJoin.
	// The filters which only apply to the main collection
	// are stored in the root scanNode.
	// The filters that are defined on the root query, but apply
	// to the sub type are defined here in the select.
	// The filters that are defined on the subtype query
	// are defined in the subtype scan node.
	filter *parser.Filter

	// Commenting out because unused code (structcheck) according to linter.
	// // If the select query is using a FindByDocKey filter
	// docKey string

	groupSelect *parser.Select

	// @todo restructure renderNode -> render, which is its own
	// object, and not a planNode.
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

		n.doc = n.source.Values()
		passes, err := parser.RunFilter(n.doc, n.filter, n.p.evalCtx)
		if err != nil {
			return false, err
		}

		if passes {
			n.renderDoc()
			return true, err
			// err :=
			// return err == nil, err
		}
		// didn't pass, keep looping
	}
}

// applies all the necessary rendering to doc
// as defined by the query statement. This includes
// aliases, and any transformations.
// Takes a doc map, and applies the necessary rendering.
// It also holds all the necessary render meta-data
// and ast parser data.
func (n *selectNode) renderDoc() error {
	renderData := map[string]interface{}{
		"numResults": n.renderInfo.numResults,
		"fields":     n.renderInfo.fields,
		"aliases":    n.renderInfo.aliases,
	}
	n.doc["__render"] = renderData
	return nil
}

func (n *selectNode) Spans(spans core.Spans) {
	n.source.Spans(spans)
}

func (n *selectNode) Values() map[string]interface{} {
	return n.doc
}

func (n *selectNode) Close() error {
	return n.source.Close()
}

// initSource is the main workhorse for recursively constructing
// all the necessary data source objects. This includes
// creating scanNodes, typeIndexJoinNodes, and splitting
// the necessary filters. Its designed to work with the
// planner.Select construction call.
func (n *selectNode) initSource(parsed *parser.Select) error {
	collectionName := parsed.Name
	sourcePlan, err := n.p.getSource(collectionName)
	if err != nil {
		return err
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

		// if we have a FindByDockey filter, create a span for it
		// and propogate it to the scanNode
		// @todo: When running the optimizer, check if the filter object
		// contains a _key equality condition, and upgrade it to a point lookup
		// instead of a prefix scan + filter via the Primary Index (0), like here:
		if parsed.DocKey != "" {
			dockeyIndexKey := base.MakeIndexKey(&sourcePlan.info.collectionDescription,
				&sourcePlan.info.collectionDescription.Indexes[0], core.NewKey(parsed.DocKey))
			spans := core.Spans{core.NewSpan(dockeyIndexKey, core.Key{})}
			origScan.Spans(spans)
		}
	}

	return n.initFields(parsed)
}

func (n *selectNode) initFields(parsed *parser.Select) error {
	n.renderInfo.numResults = 0
	// subTypes := make([]*parser.Select, 0)

	// iterate to build the render info
	for _, field := range parsed.Fields {
		switch node := field.(type) {
		case *parser.Select:
			// continue //ignore for now
			// future:
			// plan := n.p.Select(node)
			// n.source := p.SubTypeIndexJoin(origScan, plan)
			// f, found := n.sourceInfo.collectionDescription.GetField(node.GetName())
			// if found {
			// 	n.renderInfo.fields = append(n.renderInfo.fields, f.Name)
			// }
			n.renderInfo.fields = append(n.renderInfo.fields, node.GetName())
			// subTypes = append(subTypes, node)
		case *parser.Field, parser.Field:
			// f, found := n.sourceInfo.collectionDescription.GetField(node.GetName())
			// if found {
			// 	n.renderInfo.fields = append(n.renderInfo.fields, f.Name)
			// }
			n.renderInfo.fields = append(n.renderInfo.fields, node.GetName())
		}
		n.renderInfo.aliases = append(n.renderInfo.aliases, field.GetAlias())
		n.renderInfo.numResults++
	}

	// iterate to build sub plans
	// for _, field := range parsed.Fields {

	// }

	// re-organize the fields slice into reverse-alphabetical
	// this makes sure the reserved database fields that start with
	// a "_" end up at the end. So if/when we build our MultiNode
	// all the AppendPlans end up at the end.
	sort.Slice(parsed.Fields, func(i, j int) bool {
		return !(strings.Compare(parsed.Fields[i].GetName(), parsed.Fields[j].GetName()) < 0)
	})

	// loop over the sub type
	// at the moment, we're only testing a single sub selection
	for _, field := range parsed.Fields {
		if subtype, ok := field.(*parser.Select); ok {
			// @todo: check select type:
			// - TypeJoin
			// - commitScan
			if subtype.Name == "_version" { // reserved sub type for object queries
				commitSlct := &parser.CommitSelect{
					Name:   subtype.Name,
					Alias:  subtype.Alias,
					Type:   parser.LatestCommits,
					Fields: subtype.Fields,
				}
				commitPlan, err := n.p.CommitSelect(commitSlct)
				if err != nil {
					return err
				}

				if err := n.addSubPlan(field.GetName(), commitPlan); err != nil {
					return err
				}
			} else if subtype.Root == parser.ObjectSelection {
				if subtype.Name == parser.GroupFieldName {
					n.groupSelect = subtype
				} else {
					typeIndexJoin, err := n.p.makeTypeIndexJoin(n, n.origSource, subtype)
					if err != nil {
						return err
					}

					// n.source = typeIndexJoin
					if err := n.addSubPlan(field.GetName(), typeIndexJoin); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (n *selectNode) Source() planNode { return n.source }

// func appendSource() {}

// func (n *selectNode) initRender(fields []*base.FieldDescription, aliases []string) error {
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

func (p *Planner) SelectFromSource(parsed *parser.Select, source planNode, fromCollection bool, providedSourceInfo *sourceInfo) (planNode, error) {
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

	if err := s.initFields(parsed); err != nil {
		return nil, err
	}

	groupPlan, err := p.GroupBy(groupBy, s.groupSelect)
	if err != nil {
		return nil, err
	}

	limitPlan, err := p.Limit(limit)
	if err != nil {
		return nil, err
	}

	sortPlan, err := p.OrderBy(sort)
	if err != nil {
		return nil, err
	}

	top := &selectTopNode{
		source: s,
		render: p.render(),
		limit:  limitPlan,
		sort:   sortPlan,
		group:  groupPlan,
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

	if err := s.initSource(parsed); err != nil {
		return nil, err
	}

	groupPlan, err := p.GroupBy(groupBy, s.groupSelect)
	if err != nil {
		return nil, err
	}

	limitPlan, err := p.Limit(limit)
	if err != nil {
		return nil, err
	}

	sortPlan, err := p.OrderBy(sort)
	if err != nil {
		return nil, err
	}

	top := &selectTopNode{
		source: s,
		render: p.render(),
		limit:  limitPlan,
		sort:   sortPlan,
		group:  groupPlan,
	}
	return top, nil
}
