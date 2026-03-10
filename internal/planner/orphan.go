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
	"errors"
	"maps"

	"github.com/sourcenetwork/immutable/enumerable"

	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/connor"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/filter"
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
	// usePointLookup is true when the parent side does NOT store the FK,
	// requiring orphan identification via point lookups on the child's FK index.
	usePointLookup bool
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
// Fetches all orphans on the first Next() call via scanNode clone, then yields
// them sequentially.
//
// Wrapper mode (source != nil): Wraps a source planNode for secondary-side orphans.
// Two distinct phases concatenated via enumerable.Concat:
//
//	ASC: orphans (via point lookup) → source docs (from ordered join)
//	DESC: source docs (from ordered join) → orphans (via point lookup)
type orphanNode struct {
	docMapper

	// join provides access to the join internals for orphan fetching
	join *invertibleTypeJoin

	// Optional source for wrapper mode.
	// When set, the node wraps source and interleaves orphans.
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
	isSubQuery               bool

	// standalone iteration state (source == nil)
	docs    []core.Doc
	current int
	fetched bool

	// wrapper iteration state (source != nil) — phases is the concatenation of
	// orphan and source enumerables, ordered by ASC/DESC.
	phases enumerable.Enumerable[core.Doc]

	// Streaming point-lookup state — lazily iterates parents and yields one orphan per Next().
	parentClone        *scanNode
	childFKFieldName   string
	childFKFieldMapIdx int
	pointLookupDone    bool

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
}

// setSubQueryFilter configures the orphanNode (in wrapper mode) for nested join use.
// Called by retrievePrimaryDocs before each Init() cycle with the parent filter
// constrained to the current target doc. The orphan phase uses this filter to scope
// the parent iteration to the relevant subset.
func (n *orphanNode) setSubQueryFilter(filter *mapper.Filter) {
	n.subQueryFilter = filter
}

func (n *orphanNode) Kind() string {
	return "orphanNode"
}

func (n *orphanNode) Init() error {
	n.docs = nil
	n.current = 0
	n.fetched = false
	n.phases = nil
	n.pointLookupDone = false
	if n.parentClone != nil {
		_ = n.parentClone.Close()
		n.parentClone = nil
	}
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
	var cloneErr error
	if n.parentClone != nil {
		cloneErr = n.parentClone.Close()
	}
	if n.source != nil {
		return errors.Join(n.source.Close(), cloneErr)
	}
	return cloneErr
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

// nextWrapped delegates to the concatenated phases enumerable.
func (n *orphanNode) nextWrapped() (bool, error) {
	if n.phases == nil {
		orphanEnum := &orphanEnumerable{node: n}
		sourceEnum := &sourceEnumerable{source: n.source}

		if n.orderDirection == mapper.ASC {
			n.phases = enumerable.Concat(orphanEnum, sourceEnum)
		} else {
			n.phases = enumerable.Concat(sourceEnum, orphanEnum)
		}
	}
	return n.phases.Next()
}

func (n *orphanNode) Value() core.Doc {
	if n.source != nil {
		if n.phases == nil {
			return core.Doc{}
		}
		doc, _ := n.phases.Value()
		return doc
	}
	if n.current > 0 && n.current <= len(n.docs) {
		return n.docs[n.current-1]
	}
	return core.Doc{}
}

// orphanEnumerable wraps the point-lookup orphan iterator as an Enumerable[core.Doc].
type orphanEnumerable struct {
	node    *orphanNode
	current core.Doc
}

func (e *orphanEnumerable) Next() (bool, error) {
	doc, found, err := e.node.nextOrphanByPointLookup()
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}
	e.current = doc
	return true, nil
}

func (e *orphanEnumerable) Value() (core.Doc, error) {
	return e.current, nil
}

func (e *orphanEnumerable) Reset() {}

// sourceEnumerable wraps a planNode as an Enumerable[core.Doc].
type sourceEnumerable struct {
	source planNode
}

func (e *sourceEnumerable) Next() (bool, error) {
	return e.source.Next()
}

func (e *sourceEnumerable) Value() (core.Doc, error) {
	return e.source.Value(), nil
}

func (e *sourceEnumerable) Reset() {}

// fetchOrphans fetches and returns parent documents that have no related children.
// Used only in standalone mode (FK IS NULL path for primary-side parents) and
// subquery mode (nested joins with primary ordering).
func (n *orphanNode) fetchOrphans() (_ []core.Doc, err error) {
	n.fetched = true

	parentScan := getNode[*scanNode](n.join.parentSide.plan)
	if parentScan == nil {
		return nil, nil
	}

	if !n.join.parentSide.relFieldDef.HasValue() {
		return nil, nil
	}

	var orphanFilter *mapper.Filter
	var relationIDFieldName string

	if n.isSubQuery {
		orphanFilter = addNullFilterOnField(n.subQueryFilter, n.subQueryRelIDFieldMapIdx)
		relationIDFieldName = n.subQueryRelIDFieldName
	} else if n.join.parentSide.isPrimary() {
		if !n.join.parentSide.relIDFieldMapIndex.HasValue() {
			return nil, nil
		}
		relIDFieldMapIndex := n.join.parentSide.relIDFieldMapIndex.Value()
		orphanFilter = addNullFilterOnField(n.join.subFilter, relIDFieldMapIndex)
		relationIDFieldName = request.ToFieldID(n.join.parentSide.relFieldDef.Value().Name)
	} else {
		// Secondary parent — should use point-lookup wrapper mode, not standalone.
		return nil, nil
	}

	result := selectIndex(selectIndexOptions{
		collection:          n.join.parentSide.col,
		filter:              orphanFilter,
		relationIDFieldName: relationIDFieldName,
		docMapping:          n.documentMapping,
	})

	scan := parentScan.cloneWithFilter(orphanFilter, result.index)
	defer func() {
		err = errors.Join(err, scan.Close())
	}()

	if err := scan.Init(); err != nil {
		return nil, err
	}

	var docs []core.Doc
	for {
		hasNext, err := scan.Next()
		if err != nil {
			return nil, err
		}
		if !hasNext {
			break
		}

		docs = append(docs, scan.Value())
	}

	n.execInfo.fetches.Add(scan.execInfo.fetches)

	return docs, nil
}

// initPointLookupState initializes the parent iterator clone and child lookup info
// for streaming orphan detection via point lookups. Called once on first need.
func (n *orphanNode) initPointLookupState() error {
	if !n.join.childSide.relFieldDef.HasValue() || !n.join.childSide.relIDFieldMapIndex.HasValue() {
		n.pointLookupDone = true
		return nil
	}

	childRelFieldName := n.join.childSide.relFieldDef.Value().Name
	n.childFKFieldName = request.ToFieldID(childRelFieldName)
	n.childFKFieldMapIdx = n.join.childSide.relIDFieldMapIndex.Value()

	parentScan := getNode[*scanNode](n.join.parentSide.plan)
	if parentScan == nil {
		n.pointLookupDone = true
		return nil
	}

	// Use subQueryFilter when set (nested join scoped to one target doc),
	// otherwise use the top-level subFilter.
	parentFilter := n.join.subFilter
	if n.subQueryFilter != nil {
		parentFilter = n.subQueryFilter
	}

	// Select the best index for the parent filter.
	parentResult := selectIndex(selectIndexOptions{
		collection: n.join.parentSide.col,
		filter:     parentFilter,
		docMapping: n.documentMapping,
	})

	n.parentClone = parentScan.cloneWithFilter(parentFilter, parentResult.index)
	if err := n.parentClone.Init(); err != nil {
		return err
	}

	return nil
}

// nextOrphanByPointLookup returns the next orphan parent by iterating parents one at a time
// and checking each via a point lookup on the child's unique FK index.
// Returns (doc, true, nil) for an orphan, (empty, false, nil) when exhausted.
func (n *orphanNode) nextOrphanByPointLookup() (_ core.Doc, _ bool, err error) {
	if n.pointLookupDone {
		return core.Doc{}, false, nil
	}

	if n.parentClone == nil {
		if err := n.initPointLookupState(); err != nil {
			return core.Doc{}, false, err
		}
		if n.pointLookupDone {
			return core.Doc{}, false, nil
		}
	}

	childScan := getNode[*scanNode](n.join.childSide.plan)
	if childScan == nil {
		n.pointLookupDone = true
		return core.Doc{}, false, nil
	}

	for {
		hasNext, err := n.parentClone.Next()
		if err != nil {
			return core.Doc{}, false, err
		}
		if !hasNext {
			n.pointLookupDone = true
			n.execInfo.fetches.Add(n.parentClone.execInfo.fetches)
			_ = n.parentClone.Close()
			n.parentClone = nil
			return core.Doc{}, false, nil
		}

		doc := n.parentClone.Value()

		// Build a lookup filter for the child's FK field = parent doc ID.
		lookupFilter := addFilterOnField(nil, n.childFKFieldMapIdx, doc.GetID())
		lookupFilter.ExternalConditions = map[string]any{
			n.childFKFieldName: map[string]any{
				"_eq": doc.GetID(),
			},
		}

		// Select the best index for the child lookup filter.
		childResult := selectIndex(selectIndexOptions{
			collection:          n.join.childSide.col,
			filter:              lookupFilter,
			relationIDFieldName: n.childFKFieldName,
			docMapping:          childScan.documentMapping,
		})

		lookupClone := childScan.cloneWithFilter(lookupFilter, childResult.index)
		initErr := lookupClone.Init()
		if initErr != nil {
			return core.Doc{}, false, initErr
		}

		hasChild, nextErr := lookupClone.Next()
		n.execInfo.fetches.Add(lookupClone.execInfo.fetches)
		closeErr := lookupClone.Close()
		if nextErr != nil {
			return core.Doc{}, false, errors.Join(nextErr, closeErr)
		}
		if closeErr != nil {
			return core.Doc{}, false, closeErr
		}

		if !hasChild {
			return doc, true, nil
		}
	}
}

// addFilterOnField returns a new filter with a condition that checks if the field equals the given value.
// It does not mutate the input filter.
func addFilterOnField(f *mapper.Filter, propIndex int, value any) *mapper.Filter {
	result := mapper.NewFilter()
	if f != nil {
		maps.Copy(result.Conditions, f.Conditions)
		result.ExternalConditions = make(map[string]any, len(f.ExternalConditions))
		maps.Copy(result.ExternalConditions, f.ExternalConditions)
	}

	propertyIndex := &mapper.PropertyIndex{Index: propIndex}
	filterConditions := map[connor.FilterKey]any{
		propertyIndex: map[connor.FilterKey]any{
			mapper.FilterEqOp: value,
		},
	}

	filter.RemoveField(result, mapper.Field{Index: propIndex})
	result.Conditions = filter.MergeConditions(result.Conditions, filterConditions)
	return result
}

// addNullFilterOnField adds a filter condition that checks if the field is NULL.
func addNullFilterOnField(f *mapper.Filter, propIndex int) *mapper.Filter {
	return addFilterOnField(f, propIndex, nil)
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
