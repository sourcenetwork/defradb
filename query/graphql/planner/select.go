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

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/db/fetcher"
	"github.com/sourcenetwork/defradb/query/graphql/mapper"

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
	docMapper

	group      *groupNode
	order      *orderNode
	limit      planNode
	aggregates []aggregateNode

	// selectnode is used pre-wiring of the plan (before expansion and all).
	selectnode *selectNode

	// plan is the top of the plan graph (the wired and finalized plan graph).
	plan planNode
}

func (n *selectTopNode) Kind() string { return "selectTopNode" }

func (n *selectTopNode) Init() error { return n.plan.Init() }

func (n *selectTopNode) Start() error { return n.plan.Start() }

func (n *selectTopNode) Next() (bool, error) { return n.plan.Next() }

func (n *selectTopNode) Spans(spans core.Spans) { n.plan.Spans(spans) }

func (n *selectTopNode) Value() core.Doc { return n.plan.Value() }

func (n *selectTopNode) Source() planNode { return n.plan }

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
	docMapper

	p *Planner

	// main data source for the select node.
	source planNode

	// original source that was first given when the select node
	// was created
	origSource planNode

	// cache information about the original data source
	// collection name, meta-data, etc.
	sourceInfo sourceInfo

	// top level filter expression
	// filter is split between select, scan, and typeIndexJoin.
	// The filters which only apply to the main collection
	// are stored in the root scanNode.
	// The filters that are defined on the root query, but apply
	// to the sub type are defined here in the select.
	// The filters that are defined on the subtype query
	// are defined in the subtype scan node.
	filter *mapper.Filter

	parsed       *mapper.Select
	groupSelects []*mapper.Select
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
		passes, err := mapper.RunFilter(n.currentValue, n.filter)
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
	if n.filter == nil || n.filter.ExternalConditions == nil {
		explainerMap[filterLabel] = nil
	} else {
		explainerMap[filterLabel] = n.filter.ExternalConditions
	}

	return explainerMap, nil
}

// initSource is the main workhorse for recursively constructing
// all the necessary data source objects. This includes
// creating scanNodes, typeIndexJoinNodes, and splitting
// the necessary filters. Its designed to work with the
// planner.Select construction call.
func (n *selectNode) initSource() ([]aggregateNode, error) {
	if n.parsed.CollectionName == "" {
		n.parsed.CollectionName = n.parsed.Name
	}

	sourcePlan, err := n.p.getSource(n.parsed)
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
		if n.parsed.Cid != "" {
			c, err := cid.Decode(n.parsed.Cid)
			if err != nil {
				return nil, fmt.Errorf(
					"Failed to propagate VersionFetcher span, invalid CID: %w",
					err,
				)
			}
			spans := fetcher.NewVersionedSpan(
				core.DataStoreKey{DocKey: n.parsed.DocKeys.Value[0]},
				c,
			) // @todo check len
			origScan.Spans(spans)
		} else if n.parsed.DocKeys.HasValue {
			// If we *just* have a DocKey(s), run a FindByDocKey(s) optimization
			// if we have a FindByDockey filter, create a span for it
			// and propagate it to the scanNode
			// @todo: When running the optimizer, check if the filter object
			// contains a _key equality condition, and upgrade it to a point lookup
			// instead of a prefix scan + filter via the Primary Index (0), like here:
			spans := make([]core.Span, len(n.parsed.DocKeys.Value))
			for i, docKey := range n.parsed.DocKeys.Value {
				dockeyIndexKey := base.MakeDocKey(sourcePlan.info.collectionDescription, docKey)
				spans[i] = core.NewSpan(dockeyIndexKey, dockeyIndexKey.PrefixEnd())
			}
			origScan.Spans(core.NewSpans(spans...))
		}
	}

	return n.initFields(n.parsed)
}

func (n *selectNode) initFields(parsed *mapper.Select) ([]aggregateNode, error) {
	aggregates := []aggregateNode{}
	// loop over the sub type
	// at the moment, we're only testing a single sub selection
	for _, field := range parsed.Fields {
		switch f := field.(type) {
		case *mapper.Aggregate:
			var plan aggregateNode
			var aggregateError error

			switch f.Name {
			case parserTypes.CountFieldName:
				plan, aggregateError = n.p.Count(f, parsed)
			case parserTypes.SumFieldName:
				plan, aggregateError = n.p.Sum(f, parsed)
			case parserTypes.AverageFieldName:
				plan, aggregateError = n.p.Average(f)
			}

			if aggregateError != nil {
				return nil, aggregateError
			}

			if plan != nil {
				aggregates = append(aggregates, plan)
			}
		case *mapper.Select:
			if f.Name == parserTypes.VersionFieldName { // reserved sub type for object queries
				commitSlct := &mapper.CommitSelect{
					Select: *f,
				}
				// handle _version sub selection query differently
				// if we are executing a regular Scan query
				// or a TimeTravel query.
				if parsed.Cid != "" {
					// for a TimeTravel query, we don't need the Latest
					// commit. Instead, _version references the CID
					// of that Target version we are querying.
					// So instead of a LatestCommit subquery, we need
					// a OneCommit subquery, with the supplied parameters.
					commitSlct.DocKey = parsed.DocKeys.Value[0] // @todo check length
					commitSlct.Cid = parsed.Cid
					commitSlct.Type = mapper.OneCommit
				} else {
					commitSlct.Type = mapper.LatestCommits
				}
				commitPlan, err := n.p.CommitSelect(commitSlct)
				if err != nil {
					return nil, err
				}

				if err := n.addSubPlan(f.Index, commitPlan); err != nil {
					return nil, err
				}
			} else if f.Name == parserTypes.GroupFieldName {
				n.groupSelects = append(n.groupSelects, f)
			} else {
				//nolint:errcheck
				n.addTypeIndexJoin(f) // @TODO: ISSUE#158
			}
		}
	}

	return aggregates, nil
}

func (n *selectNode) addTypeIndexJoin(subSelect *mapper.Select) error {
	typeIndexJoin, err := n.p.makeTypeIndexJoin(n, n.origSource, subSelect)
	if err != nil {
		return err
	}

	if err := n.addSubPlan(subSelect.Index, typeIndexJoin); err != nil {
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
func (p *Planner) SubSelect(parsed *mapper.Select) (planNode, error) {
	plan, err := p.Select(parsed)
	if err != nil {
		return nil, err
	}

	// if this is a sub select plan, we need to remove the render node
	// as the final top level selectTopNode will handle all sub renders
	top := plan.(*selectTopNode)
	return top, nil
}

func (p *Planner) SelectFromSource(
	parsed *mapper.Select,
	source planNode,
	fromCollection bool,
	providedSourceInfo *sourceInfo,
) (planNode, error) {
	s := &selectNode{
		p:          p,
		source:     source,
		origSource: source,
		parsed:     parsed,
		docMapper:  docMapper{&parsed.DocumentMapping},
	}
	s.filter = parsed.Filter
	limit := parsed.Limit
	orderBy := parsed.OrderBy
	groupBy := parsed.GroupBy

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

	groupPlan, err := p.GroupBy(groupBy, parsed, s.groupSelects)
	if err != nil {
		return nil, err
	}

	limitPlan, err := p.HardLimit(parsed, limit)
	if err != nil {
		return nil, err
	}

	orderPlan, err := p.OrderBy(parsed, orderBy)
	if err != nil {
		return nil, err
	}

	top := &selectTopNode{
		selectnode: s,
		limit:      limitPlan,
		order:      orderPlan,
		group:      groupPlan,
		aggregates: aggregates,
		docMapper:  docMapper{&parsed.DocumentMapping},
	}
	return top, nil
}

// Select constructs a SelectPlan
func (p *Planner) Select(parsed *mapper.Select) (planNode, error) {
	s := &selectNode{
		p:         p,
		filter:    parsed.Filter,
		parsed:    parsed,
		docMapper: docMapper{&parsed.DocumentMapping},
	}
	limit := parsed.Limit
	order := parsed.OrderBy
	groupBy := parsed.GroupBy

	aggregates, err := s.initSource()
	if err != nil {
		return nil, err
	}

	groupPlan, err := p.GroupBy(groupBy, parsed, s.groupSelects)
	if err != nil {
		return nil, err
	}

	limitPlan, err := p.HardLimit(parsed, limit)
	if err != nil {
		return nil, err
	}

	orderPlan, err := p.OrderBy(parsed, order)
	if err != nil {
		return nil, err
	}

	top := &selectTopNode{
		selectnode: s,
		limit:      limitPlan,
		order:      orderPlan,
		group:      groupPlan,
		aggregates: aggregates,
		docMapper:  docMapper{&parsed.DocumentMapping},
	}
	return top, nil
}
