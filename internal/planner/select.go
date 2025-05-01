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
	"math"
	"slices"
	"strings"

	cid "github.com/ipfs/go-cid"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/filter"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
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
	limit      *limitNode
	aggregates []aggregateNode

	// selectNode is used pre-wiring of the plan (before expansion and all).
	selectNode *selectNode

	// This is added temporarity until Planner is refactored
	// https://github.com/sourcenetwork/defradb/issues/3467
	similarity []*similarityNode

	// plan is the top of the plan graph (the wired and finalized plan graph).
	planNode planNode
}

func (n *selectTopNode) Kind() string { return "selectTopNode" }

func (n *selectTopNode) Init() error { return n.planNode.Init() }

func (n *selectTopNode) Start() error { return n.planNode.Start() }

func (n *selectTopNode) Next() (bool, error) { return n.planNode.Next() }

func (n *selectTopNode) Prefixes(prefixes []keys.Walkable) { n.planNode.Prefixes(prefixes) }

func (n *selectTopNode) Value() core.Doc { return n.planNode.Value() }

func (n *selectTopNode) Source() planNode { return n.planNode }

// Explain method for selectTopNode returns no attributes but is used to
// subscribe / opt-into being an explainablePlanNode.
func (n *selectTopNode) Explain(explainType request.ExplainType) (map[string]any, error) {
	// No attributes are returned for selectTopNode.
	return nil, nil
}

func (n *selectTopNode) Close() error {
	if n.planNode == nil {
		return nil
	}
	return n.planNode.Close()
}

type selectNode struct {
	documentIterator
	docMapper

	planner *Planner

	// main data source for the select node.
	source planNode

	// original source that was first given when the select node
	// was created
	origSource planNode

	collection client.Collection

	// top level filter expression
	// filter is split between select, scan, and typeIndexJoin.
	// The filters which only apply to the main collection
	// are stored in the root scanNode.
	// The filters that are defined on the root request, but apply
	// to the sub type are defined here in the select.
	// The filters that are defined on the subtype request
	// are defined in the subtype scan node.
	filter *mapper.Filter

	docIDs immutable.Option[[]string]

	selectReq    *mapper.Select
	groupSelects []*mapper.Select

	execInfo selectExecInfo
}

type selectExecInfo struct {
	// Total number of times selectNode was executed.
	iterations uint64

	// Total number of times top level select filter passed / matched.
	filterMatches uint64
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
	n.execInfo.iterations++

	for {
		if hasNext, err := n.source.Next(); !hasNext {
			return false, err
		}

		n.currentValue = n.source.Value()
		passes, err := mapper.RunFilter(n.currentValue, n.filter)
		if err != nil {
			return false, err
		}

		if !passes {
			continue
		}

		n.execInfo.filterMatches++

		if n.docIDs.HasValue() {
			docID := n.currentValue.GetID()
			for _, docIDValue := range n.docIDs.Value() {
				if docID == docIDValue {
					return true, nil
				}
			}

			continue
		}

		return true, err
	}
}

func (n *selectNode) Prefixes(prefixes []keys.Walkable) {
	n.source.Prefixes(prefixes)
}

func (n *selectNode) Close() error {
	return n.source.Close()
}

func (n *selectNode) simpleExplain() (map[string]any, error) {
	simpleExplainMap := map[string]any{}

	// Add the filter attribute if it exists.
	if n.filter == nil {
		simpleExplainMap[filterLabel] = nil
	} else {
		simpleExplainMap[filterLabel] = n.filter.ToMap(n.documentMapping)
	}

	// Add the docIDs attribute if it exists.
	if !n.docIDs.HasValue() {
		simpleExplainMap[request.DocIDArgName] = nil
	} else {
		simpleExplainMap[request.DocIDArgName] = n.docIDs.Value()
	}

	return simpleExplainMap, nil
}

// Explain method returns a map containing all attributes of this node that
// are to be explained, subscribes / opts-in this node to be an explainablePlanNode.
func (n *selectNode) Explain(explainType request.ExplainType) (map[string]any, error) {
	switch explainType {
	case request.SimpleExplain:
		return n.simpleExplain()

	case request.ExecuteExplain:
		return map[string]any{
			"iterations":    n.execInfo.iterations,
			"filterMatches": n.execInfo.filterMatches,
		}, nil

	default:
		return nil, ErrUnknownExplainRequestType
	}
}

// initSource is the main workhorse for recursively constructing
// all the necessary data source objects. This includes
// creating scanNodes, typeIndexJoinNodes, and splitting
// the necessary filters. Its designed to work with the
// planner.Select construction call.
func (n *selectNode) initSource() ([]aggregateNode, []*similarityNode, error) {
	if n.selectReq.CollectionName == "" {
		n.selectReq.CollectionName = n.selectReq.Name
	}

	sourcePlan, err := n.planner.getSource(n.selectReq)
	if err != nil {
		return nil, nil, err
	}
	n.source = sourcePlan.plan
	n.origSource = sourcePlan.plan
	n.collection = sourcePlan.collection

	// split filter
	// apply the root filter to the source
	// and rootSubType filters to the selectNode
	// @todo: simulate splitting for now
	origScan, isScanNode := n.source.(*scanNode)
	if isScanNode {
		origScan.showDeleted = n.selectReq.ShowDeleted
		origScan.filter = n.filter
		n.filter = nil

		// If we have a CID, then we need to run a TimeTravel (History-Traversing Versioned)
		// query, which means we need to propagate the values to the underlying VersionedFetcher
		if n.selectReq.Cid.HasValue() {
			c, err := cid.Decode(n.selectReq.Cid.Value())
			if err != nil {
				return nil, nil, err
			}

			// This exists because the fetcher interface demands a []Prefixes, yet the versioned
			// fetcher type (that will be the only one consuming this []Prefixes) does not use it
			// as a prefix. And with this design limitation this is
			// currently the least bad way of passing the cid in to the fetcher.
			origScan.Prefixes(
				[]keys.Walkable{
					keys.HeadstoreDocKey{
						Cid: c,
					},
				},
			)
		} else if n.selectReq.DocIDs.HasValue() {
			shortID, err := id.GetShortCollectionID(
				n.planner.ctx,
				n.planner.txn,
				sourcePlan.collection.Description().CollectionID,
			)
			if err != nil {
				return nil, nil, err
			}

			// If we *just* have a DocID(s), run a FindByDocID(s) optimization
			// if we have a FindByDocID filter, create a prefix for it
			// and propagate it to the scanNode
			// @todo: When running the optimizer, check if the filter object
			// contains a _docID equality condition, and upgrade it to a point lookup
			// instead of a prefix scan + filter via the Primary Index (0), like here:
			prefixes := make([]keys.Walkable, len(n.selectReq.DocIDs.Value()))

			for i, docID := range n.selectReq.DocIDs.Value() {
				prefixes[i] = keys.DataStoreKey{
					CollectionShortID: shortID,
					DocID:             docID,
				}
			}
			origScan.Prefixes(prefixes)
		}
	}

	aggregates, similarity, err := n.initFields(n.selectReq)
	if err != nil {
		return nil, nil, err
	}

	if isScanNode {
		origScan.index = findIndexByFilteringField(origScan)
		origScan.initFetcher(n.selectReq.Cid)
	}

	return aggregates, similarity, nil
}

func findIndexByFilteringField(scanNode *scanNode) immutable.Option[client.IndexDescription] {
	if scanNode.filter == nil {
		return immutable.None[client.IndexDescription]()
	}
	colDesc := scanNode.col.Description()

	conditions := scanNode.filter.ExternalConditions
	var indexCandidates []client.IndexDescription
	filter.TraverseFields(conditions, func(path []string, val any) bool {
		for _, field := range scanNode.col.Schema().Fields {
			if field.Name != path[0] {
				continue
			}
			indexes := colDesc.GetIndexesOnField(field.Name)
			if len(indexes) > 0 {
				indexCandidates = append(indexCandidates, indexes...)
				return true
			}
		}
		return true
	})
	if len(indexCandidates) == 0 {
		return immutable.None[client.IndexDescription]()
	}

	slices.SortFunc(indexCandidates, func(a, b client.IndexDescription) int {
		return strings.Compare(a.Name, b.Name)
	})
	// we return the first found index. We will optimize it later.
	// https://github.com/sourcenetwork/defradb/issues/2680
	return immutable.Some(indexCandidates[0])
}

func findIndexByFieldName(col client.Collection, fieldName string) immutable.Option[client.IndexDescription] {
	for _, field := range col.Schema().Fields {
		if field.Name != fieldName {
			continue
		}
		indexes := col.Description().GetIndexesOnField(field.Name)
		if len(indexes) > 0 {
			// At the moment we just take the first index, but later we want to run some kind of analysis to
			// determine which index is best to use. https://github.com/sourcenetwork/defradb/issues/2680
			return immutable.Some(indexes[0])
		}
	}
	return immutable.None[client.IndexDescription]()
}

func (n *selectNode) initFields(selectReq *mapper.Select) ([]aggregateNode, []*similarityNode, error) {
	aggregates := []aggregateNode{}
	similarity := []*similarityNode{}
	// loop over the sub type
	// at the moment, we're only testing a single sub selection
	for _, field := range selectReq.Fields {
		switch f := field.(type) {
		case *mapper.Aggregate:
			var plan aggregateNode
			var aggregateError error
			var aggregateFilter *mapper.Filter

			// extract aggregate filters from the select
			selectReq.Filter, aggregateFilter = filter.SplitByFields(selectReq.Filter, f.Field)
			switch f.Name {
			case request.CountFieldName:
				plan, aggregateError = n.planner.Count(f, selectReq, aggregateFilter)
			case request.SumFieldName:
				plan, aggregateError = n.planner.Sum(f, selectReq, aggregateFilter)
			case request.AverageFieldName:
				plan, aggregateError = n.planner.Average(f, aggregateFilter)
			case request.MaxFieldName:
				plan, aggregateError = n.planner.Max(f, selectReq, aggregateFilter)
			case request.MinFieldName:
				plan, aggregateError = n.planner.Min(f, selectReq, aggregateFilter)
			}

			if aggregateError != nil {
				return nil, nil, aggregateError
			}

			if plan != nil {
				aggregates = append(aggregates, plan)
			}
		case *mapper.Select:
			if f.Name == request.VersionFieldName { // reserved sub type for object queries
				commitSlct := &mapper.CommitSelect{
					Select: *f,
				}
				// handle _version sub selection query differently
				// if we are executing a regular Scan query
				// or a TimeTravel query.
				if selectReq.Cid.HasValue() {
					// for a TimeTravel query, we don't need the Latest
					// commit. Instead, _version references the CID
					// of that Target version we are querying.
					// So instead of a LatestCommit subquery, we need
					// a commits query with max depth starting from the
					// target CID version
					commitSlct.DocID = immutable.Some(selectReq.DocIDs.Value()[0]) // @todo check length
					commitSlct.Cid = selectReq.Cid
					commitSlct.Depth = immutable.Some(uint64(math.MaxUint64))
				}

				commitPlan := n.planner.DAGScan(commitSlct)

				if err := n.addSubPlan(f.Index, commitPlan); err != nil {
					return nil, nil, err
				}
			} else if f.Name == request.GroupFieldName {
				if selectReq.GroupBy == nil {
					return nil, nil, ErrGroupOutsideOfGroupBy
				}
				n.groupSelects = append(n.groupSelects, f)
			} else if isSpecialNoOpField(f, selectReq) {
				// no-op
			} else if !(n.collection != nil && len(n.collection.Description().QuerySources()) > 0) {
				// Collections sourcing data from queries only contain embedded objects and don't require
				// a traditional join here
				err := n.addTypeIndexJoin(f)
				if err != nil {
					return nil, nil, err
				}
			}
		case *mapper.Similarity:
			var simFilter *mapper.Filter
			selectReq.Filter, simFilter = filter.SplitByFields(selectReq.Filter, f.Field)
			similarity = append(similarity, n.planner.Similarity(f, simFilter))
		}
	}

	return aggregates, similarity, nil
}

func isSpecialNoOpField(field *mapper.Select, parentField *mapper.Select) bool {
	// commit query link and signature fields are always added and need no special treatment here
	// WARNING: It is important to check collection name is nil and the parent select name
	// here else we risk falsely identifying user defined fields with the name `links` as a commit links field
	if field.CollectionName != "" {
		return false
	}
	isCommit := parentField.Name == request.CommitsName || parentField.Name == request.LatestCommitsName
	return isCommit && (field.Name == request.LinksFieldName || field.Name == request.SignatureFieldName)
}

func (n *selectNode) addTypeIndexJoin(subSelect *mapper.Select) error {
	typeIndexJoin, err := n.planner.makeTypeIndexJoin(n, n.origSource, subSelect)
	if err != nil {
		return err
	}
	if err := n.addSubPlan(subSelect.Index, typeIndexJoin); err != nil {
		return err
	}

	return nil
}

func (n *selectNode) Source() planNode { return n.source }

func (p *Planner) SelectFromSource(
	selectReq *mapper.Select,
	source planNode,
	fromCollection bool,
	collection client.Collection,
) (planNode, error) {
	s := &selectNode{
		planner:    p,
		source:     source,
		origSource: source,
		selectReq:  selectReq,
		docMapper:  docMapper{selectReq.DocumentMapping},
		filter:     selectReq.Filter,
		docIDs:     selectReq.DocIDs,
	}
	limit := selectReq.Limit
	orderBy := selectReq.OrderBy
	groupBy := selectReq.GroupBy

	if collection != nil {
		s.collection = collection
	}

	if fromCollection {
		col, err := p.db.GetCollectionByName(p.ctx, selectReq.Name)
		if err != nil {
			return nil, err
		}

		s.collection = col
	}

	aggregates, similarity, err := s.initFields(selectReq)
	if err != nil {
		return nil, err
	}

	groupPlan, err := p.GroupBy(groupBy, selectReq, s.groupSelects)
	if err != nil {
		return nil, err
	}

	limitPlan, err := p.Limit(selectReq, limit)
	if err != nil {
		return nil, err
	}

	orderPlan, err := p.OrderBy(selectReq, orderBy)
	if err != nil {
		return nil, err
	}

	top := &selectTopNode{
		selectNode: s,
		limit:      limitPlan,
		order:      orderPlan,
		group:      groupPlan,
		aggregates: aggregates,
		similarity: similarity,
		docMapper:  docMapper{selectReq.DocumentMapping},
	}
	return top, nil
}

// Select constructs a SelectPlan
func (p *Planner) Select(selectReq *mapper.Select) (planNode, error) {
	s := &selectNode{
		planner:   p,
		filter:    selectReq.Filter,
		docIDs:    selectReq.DocIDs,
		selectReq: selectReq,
		docMapper: docMapper{selectReq.DocumentMapping},
	}
	limit := selectReq.Limit
	orderBy := selectReq.OrderBy
	groupBy := selectReq.GroupBy

	aggregates, similarity, err := s.initSource()
	if err != nil {
		return nil, err
	}

	groupPlan, err := p.GroupBy(groupBy, selectReq, s.groupSelects)
	if err != nil {
		return nil, err
	}

	limitPlan, err := p.Limit(selectReq, limit)
	if err != nil {
		return nil, err
	}

	orderPlan, err := p.OrderBy(selectReq, orderBy)
	if err != nil {
		return nil, err
	}

	top := &selectTopNode{
		selectNode: s,
		limit:      limitPlan,
		order:      orderPlan,
		group:      groupPlan,
		aggregates: aggregates,
		similarity: similarity,
		docMapper:  docMapper{selectReq.DocumentMapping},
	}
	return top, nil
}
