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

	"github.com/sourcenetwork/immutable/enumerable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// orphanPointLookupNode wraps an ordered join, concatenating orphan results and source
// results. ASC puts orphans first, DESC puts them last.
//
// Orphans are identified via point lookups on the child's unique FK index:
// for each parent doc, a Has() call checks whether a child with FK = parentDocID
// exists, avoiding a full scanNode clone per doc.
type orphanPointLookupNode struct {
	docMapper

	join   *invertibleTypeJoin
	source planNode

	orderDirection mapper.SortDirection

	// Set by retrievePrimaryDocs for nested join context (per-iteration filter scope).
	subQueryFilter *mapper.Filter

	// Concatenated orphan + source enumerables.
	phases enumerable.Enumerable[core.Doc]

	// Point-lookup state for streaming orphan detection.
	// parentClone iterates parent docs one at a time; for each, we check
	// whether a child with FK = parentDocID exists via a direct Has() call
	// on the child's unique FK index.
	parentClone     *scanNode
	pointLookupDone bool

	// Initialized once in initPointLookupState, reused for every parent doc.
	childFKIndex client.IndexDescription
	childShortID uint32
	planner      *Planner

	execInfo orphanExecInfo
}

func newOrphanPointLookupNode(
	join *invertibleTypeJoin,
	source planNode,
	direction mapper.SortDirection,
) *orphanPointLookupNode {
	return &orphanPointLookupNode{
		docMapper:      join.docMapper,
		join:           join,
		source:         source,
		orderDirection: direction,
	}
}

// setSubQueryFilter configures the orphanPointLookupNode for nested join use.
// Called by retrievePrimaryDocs before each Init() cycle with the parent filter
// constrained to the current target doc. The orphan phase uses this filter to scope
// the parent iteration to the relevant subset.
func (n *orphanPointLookupNode) setSubQueryFilter(filter *mapper.Filter) {
	n.subQueryFilter = filter
}

func (n *orphanPointLookupNode) Kind() string {
	return orphanNodeKind
}

func (n *orphanPointLookupNode) Init() error {
	n.phases = nil
	n.pointLookupDone = false

	if n.parentClone != nil {
		if err := n.parentClone.Close(); err != nil {
			return err
		}
		n.parentClone = nil
	}

	if err := n.source.Init(); err != nil {
		return err
	}
	return n.initPointLookupState()
}

func (n *orphanPointLookupNode) Start() error {
	return n.source.Start()
}

func (n *orphanPointLookupNode) Prefixes(prefixes []keys.Walkable) {
	n.source.Prefixes(prefixes)
}

func (n *orphanPointLookupNode) Source() planNode {
	return n.source
}

func (n *orphanPointLookupNode) Close() error {
	var errs []error
	if n.parentClone != nil {
		errs = append(errs, n.parentClone.Close())
	}
	errs = append(errs, n.source.Close())
	return errors.Join(errs...)
}

func (n *orphanPointLookupNode) Next() (bool, error) {
	n.execInfo.iterations++

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

func (n *orphanPointLookupNode) Value() core.Doc {
	if n.phases == nil {
		return core.Doc{}
	}
	doc, _ := n.phases.Value()
	return doc
}

// orphanEnumerable wraps the point-lookup orphan iterator as an Enumerable[core.Doc].
type orphanEnumerable struct {
	node    *orphanPointLookupNode
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

// initPointLookupState initializes the parent iterator clone and child index info
// for streaming orphan detection via point lookups.
//
// For each parent doc, we need to check if a child with FK = parentDocID exists.
// We find the child's unique FK index once here and then do a direct datastore.Has()
// per doc in nextOrphanByPointLookup.
func (n *orphanPointLookupNode) initPointLookupState() error {
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

	n.parentClone = parentScan.cloneWithFilter(parentFilter, parentResult.index, nil)
	if err := n.parentClone.Init(); err != nil {
		return err
	}

	return nil
}

// nextOrphanByPointLookup returns the next orphan parent by iterating parents one at a time
// and checking each via a Has() call on the child's unique FK index.
// Returns (doc, true, nil) for an orphan, (empty, false, nil) when exhausted.
func (n *orphanPointLookupNode) nextOrphanByPointLookup() (_ core.Doc, _ bool, err error) {
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

		indexKey := keys.NewIndexDataStoreKey(n.childShortID, n.childFKIndex.ID, []keys.IndexedField{
			{Value: client.NewNormalString(doc.GetID()), Descending: n.childFKIndex.Fields[0].Descending},
		})

		hasChild, err := ds.Has(n.planner.ctx, &indexKey)
		if err != nil {
			return core.Doc{}, false, NewErrCheckOrphanPointLookup(err)
		}
		n.execInfo.fetches.IndexesFetched++

		if !hasChild {
			return doc, true, nil
		}
	}
}

func (n *orphanPointLookupNode) simpleExplain() (map[string]any, error) {
	return map[string]any{}, nil
}

func (n *orphanPointLookupNode) Explain(explainType request.ExplainType) (map[string]any, error) {
	switch explainType {
	case request.SimpleExplain:
		return n.simpleExplain()

	case request.ExecuteExplain:
		fetches := n.execInfo.fetches
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
