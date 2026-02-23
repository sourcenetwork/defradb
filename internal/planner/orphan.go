// Copyright 2026 Democratized Data Foundation
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
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// orphanWiringRequest stores information needed to wire orphan nodes into a selectTopNode.
// This is set during expandTypeIndexJoinPlan and processed at the end of expandSelectTopNodePlan,
// after the full plan chain (order, limit) is built.
type orphanWiringRequest struct {
	join      *invertibleTypeJoin
	direction mapper.SortDirection
	// useExclusion is true when the parent side does NOT store the FK,
	// requiring orphan identification by exclusion from join results.
	useExclusion bool
}

// orphanExecInfo contains execution information for the orphanNode.
type orphanExecInfo struct {
	// Total number of times orphanNode.Next was executed.
	iterations uint64

	// Information about fetches performed when fetching orphan documents.
	fetches fetcher.ExecInfo
}

// orphanNode fetches orphan parent documents (parents without children) and yields
// them one at a time.
//
// It operates in two modes based on whether source is set:
//
// Standalone mode (source == nil): Used inside a sequenceNode for FK IS NULL path.
// Fetches all orphans on the first Next() call via subQueryFetcher, then yields
// them sequentially.
//
// Wrapper mode (source != nil): Wraps a source planNode for exclusion path.
// Buffers source docs (for ASC) or yields them first (for DESC), collects their
// IDs, then fetches orphans by exclusion.
type orphanNode struct {
	docMapper

	// join provides access to the join internals for orphan fetching
	join *invertibleTypeJoin

	// Optional source for wrapper mode (exclusion path).
	// When set, the node wraps source and buffers docs for exclusion.
	// When nil, the node fetches orphans independently (FK IS NULL path).
	source         planNode
	orderDirection mapper.SortDirection

	// subquery context — set per iteration via setSubQueryContext / setSubQueryExclusionContext / setSubQueryFilter
	subQueryFilter           *mapper.Filter
	subQueryRelIDFieldName   string
	subQueryRelIDFieldMapIdx int
	subQueryExcludeIDs       []string
	isSubQuery               bool
	isSubQueryExclusion      bool

	// standalone iteration state (source == nil)
	docs    []core.Doc
	current int
	fetched bool

	// wrapper iteration state (source != nil)
	docsToYield     []core.Doc
	bufferedDocs    []core.Doc
	yieldedDocIDs   []string
	sourceExhausted bool
	orphansFetched  bool

	execInfo orphanExecInfo
}

func newOrphanNode(join *invertibleTypeJoin) *orphanNode {
	return &orphanNode{
		docMapper: join.docMapper,
		join:      join,
	}
}

func newOrphanNodeWithSource(join *invertibleTypeJoin, source planNode, direction mapper.SortDirection) *orphanNode {
	return &orphanNode{
		docMapper:      join.docMapper,
		join:           join,
		source:         source,
		orderDirection: direction,
	}
}

// setSubQueryContext configures the orphanNode for subquery use (FK IS NULL path).
// Called by retrievePrimaryDocs before each Init() cycle with per-iteration params.
func (n *orphanNode) setSubQueryContext(filter *mapper.Filter, relIDFieldName string, relIDFieldMapIdx int) {
	n.subQueryFilter = filter
	n.subQueryRelIDFieldName = relIDFieldName
	n.subQueryRelIDFieldMapIdx = relIDFieldMapIdx
	n.isSubQuery = true
	n.isSubQueryExclusion = false
}

// setSubQueryExclusionContext configures the orphanNode for subquery exclusion use.
// Called with the collected source doc IDs as exclusion set.
func (n *orphanNode) setSubQueryExclusionContext(filter *mapper.Filter, excludeIDs []string) {
	n.subQueryFilter = filter
	n.subQueryExcludeIDs = excludeIDs
	n.isSubQueryExclusion = true
	n.isSubQuery = false
}

// setSubQueryFilter configures the orphanNode (in wrapper mode) for subquery exclusion use.
// Called by retrievePrimaryDocs before each Init() cycle with the parent filter
// constrained to the current target doc. The node collects source doc IDs
// and passes them as the exclusion set when fetching orphans.
func (n *orphanNode) setSubQueryFilter(filter *mapper.Filter) {
	n.subQueryFilter = filter
	n.isSubQueryExclusion = true
}

func (n *orphanNode) Kind() string {
	return "orphanNode"
}

func (n *orphanNode) Init() error {
	n.docs = nil
	n.current = 0
	n.fetched = false
	n.docsToYield = nil
	n.bufferedDocs = nil
	n.yieldedDocIDs = nil
	n.sourceExhausted = false
	n.orphansFetched = false
	if n.source != nil {
		return n.source.Init()
	}
	return nil
}

func (n *orphanNode) Start() error {
	if n.source != nil {
		return n.source.Start()
	}
	return nil
}

func (n *orphanNode) Prefixes(prefixes []keys.Walkable) {
	if n.source != nil {
		n.source.Prefixes(prefixes)
	}
}

func (n *orphanNode) Source() planNode {
	return n.source
}

func (n *orphanNode) Close() error {
	if n.source != nil {
		return n.source.Close()
	}
	return nil
}

func (n *orphanNode) Next() (bool, error) {
	n.execInfo.iterations++

	if n.source != nil {
		return n.nextWrapped()
	}
	return n.nextStandalone()
}

// nextStandalone fetches all orphans on first call, then yields them sequentially.
func (n *orphanNode) nextStandalone() (bool, error) {
	if !n.fetched {
		if err := n.fetchOrphans(); err != nil {
			return false, err
		}
	}

	if n.current >= len(n.docs) {
		return false, nil
	}

	n.current++
	return true, nil
}

// nextWrapped dispatches to ASC or DESC wrapper iteration.
func (n *orphanNode) nextWrapped() (bool, error) {
	if len(n.docsToYield) > 0 {
		n.docsToYield = n.docsToYield[1:]
		if len(n.docsToYield) > 0 {
			return true, nil
		}
	}

	if n.orderDirection == mapper.ASC {
		return n.nextASC()
	}
	return n.nextDESC()
}

func (n *orphanNode) Value() core.Doc {
	if n.source != nil {
		if len(n.docsToYield) == 0 {
			return core.Doc{}
		}
		return n.docsToYield[0]
	}
	if n.current > 0 && n.current <= len(n.docs) {
		return n.docs[n.current-1]
	}
	return core.Doc{}
}

// nextASC buffers all source docs, then fetches orphans by exclusion,
// then yields orphans followed by buffered docs.
func (n *orphanNode) nextASC() (bool, error) {
	if n.sourceExhausted {
		if !n.orphansFetched {
			n.orphansFetched = true
			n.prepareOrphanContext()
			if err := n.fetchOrphans(); err != nil {
				return false, err
			}
			n.docsToYield = append(n.docs, n.bufferedDocs...)
			n.docs = nil
			n.bufferedDocs = nil
			if len(n.docsToYield) > 0 {
				return true, nil
			}
		}
		return false, nil
	}

	for {
		hasNext, err := n.source.Next()
		if err != nil {
			return false, err
		}
		if !hasNext {
			n.sourceExhausted = true
			return n.nextASC()
		}
		doc := n.source.Value()
		n.bufferedDocs = append(n.bufferedDocs, doc)
		n.yieldedDocIDs = append(n.yieldedDocIDs, doc.GetID())
	}
}

// nextDESC yields source docs first, then fetches and yields orphans.
func (n *orphanNode) nextDESC() (bool, error) {
	if !n.sourceExhausted {
		hasNext, err := n.source.Next()
		if err != nil {
			return false, err
		}
		if hasNext {
			doc := n.source.Value()
			n.docsToYield = append(n.docsToYield, doc)
			n.yieldedDocIDs = append(n.yieldedDocIDs, doc.GetID())
			return true, nil
		}
		n.sourceExhausted = true
	}

	if !n.orphansFetched {
		n.orphansFetched = true
		n.prepareOrphanContext()
		if err := n.fetchOrphans(); err != nil {
			return false, err
		}
		if len(n.docs) > 0 {
			n.docsToYield = append(n.docsToYield, n.docs...)
			n.docs = nil
			return true, nil
		}
	}

	return false, nil
}

// prepareOrphanContext sets the exclusion context before fetching orphans in wrapper mode.
// For subquery exclusion, it passes the parent filter and collected source doc IDs.
// For top-level, the node already has the context it needs from join state.
func (n *orphanNode) prepareOrphanContext() {
	if n.isSubQueryExclusion {
		n.setSubQueryExclusionContext(n.subQueryFilter, n.yieldedDocIDs)
	}
}

// fetchOrphans fetches parent documents that have no related children.
func (n *orphanNode) fetchOrphans() error {
	n.fetched = true

	parentScan := getNode[*scanNode](n.join.parentSide.plan)
	if parentScan == nil {
		return nil
	}

	if !n.join.parentSide.relFieldDef.HasValue() {
		return nil
	}

	fetcher := newSubQueryFetcher(
		parentScan.p.ctx,
		parentScan.p.identity,
		parentScan.p.nodeACP,
		parentScan.p.documentACP,
		n.join.parentSide.col,
		n.documentMapping,
		parentScan.p.lensStore,
		parentScan.fields,
		&n.execInfo.fetches,
	)

	var orphans []core.Doc
	var err error

	if n.isSubQueryExclusion {
		orphans, err = fetcher.fetchAllExcluding(n.subQueryFilter, n.subQueryExcludeIDs)
	} else if n.isSubQuery {
		orphans, err = fetcher.fetchOrphans(n.subQueryRelIDFieldName, n.subQueryRelIDFieldMapIdx, n.subQueryFilter)
	} else if n.join.parentSide.isPrimary() {
		relIDFieldName := request.ToFieldID(n.join.parentSide.relFieldDef.Value().Name)

		if !n.join.parentSide.relIDFieldMapIndex.HasValue() {
			return nil
		}
		relIDFieldMapIndex := n.join.parentSide.relIDFieldMapIndex.Value()

		orphans, err = fetcher.fetchOrphans(relIDFieldName, relIDFieldMapIndex, n.join.subFilter)
	} else {
		excludeIDs := make([]string, 0, len(n.join.state.encounteredDocIDs))
		for id := range n.join.state.encounteredDocIDs {
			excludeIDs = append(excludeIDs, id)
		}
		orphans, err = fetcher.fetchAllExcluding(n.join.subFilter, excludeIDs)
	}

	if err != nil {
		return err
	}

	n.docs = orphans
	return nil
}

func (n *orphanNode) simpleExplain() (map[string]any, error) {
	return map[string]any{}, nil
}

func (n *orphanNode) Explain(explainType request.ExplainType) (map[string]any, error) {
	switch explainType {
	case request.SimpleExplain:
		return n.simpleExplain()

	case request.ExecuteExplain:
		return map[string]any{
			"iterations":   n.execInfo.iterations,
			"docFetches":   n.execInfo.fetches.DocsFetched,
			"fieldFetches": n.execInfo.fetches.FieldsFetched,
			"indexFetches": n.execInfo.fetches.IndexesFetched,
		}, nil

	default:
		return nil, ErrUnknownExplainRequestType
	}
}
