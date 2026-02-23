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

// joinExpandState holds transient state used only during plan expansion for
// join optimization and orphan wiring. These fields are set at the start of
// plan expansion and consumed during the recursive expandPlan walk.
type joinExpandState struct {
	// exhaustive is set when the @exhaustive directive is present on the query.
	// When true, orphan parent documents will be included when ordering by relation
	// fields with indexes. When false (default), orphans are excluded for performance.
	exhaustive bool

	// inNestedJoin tracks whether we're expanding a join that is nested inside another join.
	// When true, orphanNode should not be added because nested joins are iterated via
	// retrievePrimaryDocs which handles orphans correctly with parent context.
	inNestedJoin bool

	// pendingOrphanWiring is set during expandTypeIndexJoinPlan to defer orphan node
	// wiring until after the full plan chain (order, limit) is built.
	pendingOrphanWiring *orphanWiringRequest
}

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

	// Subquery context fields. These are set by retrievePrimaryDocs before each
	// Init()/Next() cycle when the orphanNode is part of a nested join (not top-level).
	// In nested joins, retrievePrimaryDocs iterates over secondary-side docs and calls
	// the primary-side plan once per doc with a constrained filter.
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
		orphans, err := n.fetchOrphans()
		if err != nil {
			return false, err
		}
		n.docs = orphans
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
//
// On the first call, it drains the entire source into a local buffer while
// collecting doc IDs for exclusion. Then it fetches orphans and populates
// docsToYield with orphans first (ASC: nulls before non-nulls), followed
// by the buffered source docs. Subsequent calls yield from docsToYield.
func (n *orphanNode) nextASC() (bool, error) {
	if n.orphansFetched {
		return false, nil
	}
	n.orphansFetched = true

	var sourceDocs []core.Doc
	for {
		hasNext, err := n.source.Next()
		if err != nil {
			return false, err
		}
		if !hasNext {
			break
		}
		doc := n.source.Value()
		sourceDocs = append(sourceDocs, doc)
		n.yieldedDocIDs = append(n.yieldedDocIDs, doc.GetID())
	}

	n.prepareOrphanContext()
	orphans, err := n.fetchOrphans()
	if err != nil {
		return false, err
	}

	n.docsToYield = append(orphans, sourceDocs...)
	return len(n.docsToYield) > 0, nil
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
		orphans, err := n.fetchOrphans()
		if err != nil {
			return false, err
		}
		if len(orphans) > 0 {
			n.docsToYield = append(n.docsToYield, orphans...)
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

// fetchOrphans fetches and returns parent documents that have no related children.
//
// Two strategies are used depending on which side of the relation the parent is on:
//
// Primary parent (stores the FK, e.g. Book._publisherID): Uses a "FK IS NULL" filter
// to directly identify orphans via the datastore. This is efficient and precise.
//
// Secondary parent (no FK stored, e.g. Publisher with Book._publisherID pointing to it):
// The parent has no FK field to check, so orphans are identified by exclusion — fetch all
// parent docs matching the filter, then exclude those already encountered by the join.
// This is necessary because the join scans the child's index and only finds parents that
// have at least one child; parents with zero children never appear in the join results.
func (n *orphanNode) fetchOrphans() ([]core.Doc, error) {
	n.fetched = true

	parentScan := getNode[*scanNode](n.join.parentSide.plan)
	if parentScan == nil {
		return nil, nil
	}

	if !n.join.parentSide.relFieldDef.HasValue() {
		return nil, nil
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

	if n.isSubQueryExclusion {
		return fetcher.fetchAllExcluding(n.subQueryFilter, n.subQueryExcludeIDs)
	} else if n.isSubQuery {
		return fetcher.fetchOrphans(n.subQueryRelIDFieldName, n.subQueryRelIDFieldMapIdx, n.subQueryFilter)
	} else if n.join.parentSide.isPrimary() {
		relIDFieldName := request.ToFieldID(n.join.parentSide.relFieldDef.Value().Name)

		if !n.join.parentSide.relIDFieldMapIndex.HasValue() {
			return nil, nil
		}
		relIDFieldMapIndex := n.join.parentSide.relIDFieldMapIndex.Value()

		return fetcher.fetchOrphans(relIDFieldName, relIDFieldMapIndex, n.join.subFilter)
	}

	excludeIDs := make([]string, 0, len(n.join.state.encounteredDocIDs))
	for id := range n.join.state.encounteredDocIDs {
		excludeIDs = append(excludeIDs, id)
	}
	return fetcher.fetchAllExcluding(n.join.subFilter, excludeIDs)
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
