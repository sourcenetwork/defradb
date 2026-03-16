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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/connor"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/filter"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// joinExpandState is transient state for plan expansion, consumed during expandPlan.
type joinExpandState struct {
	// Cached from mapper.Operation.Exhaustive during makePlan for use during plan expansion.
	exhaustive bool

	// Nested joins handle orphans via retrievePrimaryDocs, so we skip orphanNode wiring.
	inNestedJoin bool

	// Deferred until after order/limit nodes are wired in expandSelectTopNodePlan.
	pendingOrphanWiring *orphanWiringRequest
}

// orphanWiringRequest captures what's needed to wire orphan nodes after the plan chain is built.
type orphanWiringRequest struct {
	join      *invertibleTypeJoin
	direction mapper.SortDirection
	// True when the parent doesn't store the FK (secondary side), so we need
	// point lookups on the child's FK index to identify orphans.
	usePointLookup bool
}

type orphanExecInfo struct {
	iterations uint64
	fetches    fetcher.ExecInfo
}

// orphanNode yields parent documents that have no related children.
//
// Two modes:
//   - Standalone (source == nil): inside a sequenceNode, fetches orphans via FK IS NULL.
//     Streams results from a scanNode clone one at a time.
//   - Wrapper (source != nil): wraps an ordered join, concatenating orphans and source
//     results. ASC puts orphans first, DESC puts them last.
type orphanNode struct {
	docMapper

	join *invertibleTypeJoin

	// nil in standalone mode; set in wrapper mode to the underlying ordered join.
	source         planNode
	orderDirection mapper.SortDirection

	// Set by retrievePrimaryDocs for nested join context (per-iteration filter scope).
	subQueryFilter           *mapper.Filter
	subQueryRelIDFieldName   string
	subQueryRelIDFieldMapIdx int
	isSubQuery               bool

	// Standalone state — streams from a scanNode clone with FK IS NULL filter.
	standaloneScan *scanNode

	// Wrapper state — concatenated orphan + source enumerables.
	phases enumerable.Enumerable[core.Doc]

	// Point-lookup state for streaming orphan detection (wrapper mode).
	// parentClone iterates parent docs one at a time; for each, we check
	// whether a child with FK = parentDocID exists via a direct Has() call
	// on the child's unique FK index, avoiding a full scanNode clone per doc.
	parentClone     *scanNode
	pointLookupDone bool

	// Initialized once in initPointLookupState, reused for every parent doc.
	childFKIndex client.IndexDescription
	childShortID uint32
	planner      *Planner

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
	n.phases = nil
	n.pointLookupDone = false

	if n.standaloneScan != nil {
		if err := n.standaloneScan.Close(); err != nil {
			return err
		}
		n.standaloneScan = nil
	}
	if n.parentClone != nil {
		if err := n.parentClone.Close(); err != nil {
			return err
		}
		n.parentClone = nil
	}

	if n.source != nil {
		if err := n.source.Init(); err != nil {
			return err
		}
		return n.initPointLookupState()
	}
	return n.initStandaloneScan()
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
	var errs []error
	if n.standaloneScan != nil {
		errs = append(errs, n.standaloneScan.Close())
	}
	if n.parentClone != nil {
		errs = append(errs, n.parentClone.Close())
	}
	if n.source != nil {
		errs = append(errs, n.source.Close())
	}
	return errors.Join(errs...)
}

func (n *orphanNode) Next() (bool, error) {
	n.execInfo.iterations++

	if n.source != nil {
		return n.nextWrapped()
	}
	return n.nextStandalone()
}

// nextStandalone yields orphans one at a time from the FK IS NULL scan clone.
func (n *orphanNode) nextStandalone() (bool, error) {
	if n.standaloneScan == nil {
		return false, nil
	}
	return n.standaloneScan.Next()
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
	if n.standaloneScan != nil {
		return n.standaloneScan.Value()
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

// initStandaloneScan creates and initializes a scanNode clone with FK IS NULL filter
// for streaming orphan detection. Used in standalone mode (primary-side parents)
// and subquery mode (nested joins with primary ordering).
func (n *orphanNode) initStandaloneScan() error {
	var orphanFilter *mapper.Filter
	var relationIDFieldName string

	if n.isSubQuery {
		orphanFilter = addNullFilterOnField(n.subQueryFilter, n.subQueryRelIDFieldMapIdx)
		relationIDFieldName = n.subQueryRelIDFieldName
	} else if n.join.parentSide.isPrimary() {
		relIDFieldMapIndex := n.join.parentSide.relIDFieldMapIndex.Value()
		orphanFilter = addNullFilterOnField(n.join.subFilter, relIDFieldMapIndex)
		relationIDFieldName = request.ToFieldID(n.join.parentSide.relFieldDef.Value().Name)
	} else {
		// Secondary parent — should use point-lookup wrapper mode, not standalone.
		return nil
	}

	result := selectIndex(selectIndexOptions{
		collection:          n.join.parentSide.col,
		filter:              orphanFilter,
		relationIDFieldName: relationIDFieldName,
		docMapping:          n.documentMapping,
	})

	parentScan := getNode[*scanNode](n.join.parentSide.plan)
	n.standaloneScan = parentScan.cloneWithFilter(orphanFilter, result.index)

	return n.standaloneScan.Init()
}

// initPointLookupState initializes the parent iterator clone and child index info
// for streaming orphan detection via point lookups. Called once on first need.
//
// For each parent doc, we need to check if a child with FK = parentDocID exists.
// Instead of cloning a full scanNode per doc, we find the child's unique FK index
// once here and then do a direct datastore.Has() per doc in nextOrphanByPointLookup.
func (n *orphanNode) initPointLookupState() error {
	parentScan := getNode[*scanNode](n.join.parentSide.plan)
	n.planner = parentScan.p

	childFKFieldName := request.ToFieldID(n.join.childSide.relFieldDef.Value().Name)
	childIdx := findIndexByFieldName(n.join.childSide.col, childFKFieldName)
	if !childIdx.HasValue() {
		n.pointLookupDone = true
		return nil
	}
	n.childFKIndex = childIdx.Value()

	shortID, err := id.GetShortCollectionID(n.planner.ctx, n.join.childSide.col.Version().CollectionID)
	if err != nil {
		return err
	}
	n.childShortID = shortID

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
// and checking each via a Has() call on the child's unique FK index.
// Returns (doc, true, nil) for an orphan, (empty, false, nil) when exhausted.
func (n *orphanNode) nextOrphanByPointLookup() (_ core.Doc, _ bool, err error) {
	if n.pointLookupDone || n.parentClone == nil {
		return core.Doc{}, false, nil
	}

	txn := datastore.CtxMustGetTxn(n.planner.ctx)
	ds := txn.Datastore()

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

		// Check if a child with FK = parentDocID exists via a direct index lookup.
		// The FK index on 1-to-1 relations is unique, so the key format is:
		//   /collectionShortID/indexID/fkValue
		// and Has() is an exact match.
		indexKey := keys.NewIndexDataStoreKey(n.childShortID, n.childFKIndex.ID, []keys.IndexedField{
			{Value: client.NewNormalString(doc.GetID()), Descending: n.childFKIndex.Fields[0].Descending},
		})

		hasChild, err := ds.Has(n.planner.ctx, &indexKey)
		if err != nil {
			return core.Doc{}, false, err
		}
		n.execInfo.fetches.IndexesFetched++

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
		fetches := n.execInfo.fetches
		if n.standaloneScan != nil {
			fetches.Add(n.standaloneScan.execInfo.fetches)
		}
		if n.parentClone != nil {
			fetches.Add(n.parentClone.execInfo.fetches)
		}
		return map[string]any{
			"iterations":   n.execInfo.iterations,
			"docFetches":   fetches.DocsFetched,
			"fieldFetches": fetches.FieldsFetched,
			"indexFetches": fetches.IndexesFetched,
		}, nil

	default:
		return nil, ErrUnknownExplainRequestType
	}
}
