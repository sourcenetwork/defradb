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

// orphanExecInfo contains execution information for the orphanNode.
type orphanExecInfo struct {
	// Total number of times orphanNode.Next was executed.
	iterations uint64

	// Information about fetches performed when fetching orphan documents.
	fetches fetcher.ExecInfo
}

// orphanNode fetches orphan parent documents (parents without children) and yields
// them one at a time. It uses subQueryFetcher to fetch all orphans on the first
// Next() call, then yields them sequentially.
//
// This node is used inside a sequenceNode to compose orphan results with regular
// join results. The sequenceNode handles ordering (orphans first for ASC, last for DESC).
type orphanNode struct {
	docMapper

	// join provides access to the join internals for orphan fetching
	join *invertibleTypeJoin

	// iteration state
	docs    []core.Doc
	current int
	fetched bool

	execInfo orphanExecInfo
}

func newOrphanNode(join *invertibleTypeJoin) *orphanNode {
	return &orphanNode{
		docMapper: join.docMapper,
		join:      join,
	}
}

func (n *orphanNode) Kind() string {
	return "orphanNode"
}

func (n *orphanNode) Init() error {
	n.docs = nil
	n.current = 0
	n.fetched = false
	return nil
}

func (n *orphanNode) Start() error {
	return nil
}

func (n *orphanNode) Prefixes(prefixes []keys.Walkable) {
	// orphanNode fetches independently via subQueryFetcher, no prefixes needed
}

func (n *orphanNode) Source() planNode {
	return nil
}

func (n *orphanNode) Close() error {
	return nil
}

func (n *orphanNode) Next() (bool, error) {
	n.execInfo.iterations++

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

func (n *orphanNode) Value() core.Doc {
	if n.current > 0 && n.current <= len(n.docs) {
		return n.docs[n.current-1]
	}
	return core.Doc{}
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

	if n.join.parentSide.isPrimary() {
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

// orphanWrapperNode wraps a join source and an orphanNode for secondary parent queries
// where orphans can only be identified by exclusion after the join runs. It buffers
// source docs (for ASC) or yields them first (for DESC), then fetches orphans.
//
// For ASC: buffers all source docs, fetches orphans, yields orphans then source docs.
// For DESC: yields source docs first, then fetches and yields orphans.
type orphanWrapperNode struct {
	docMapper

	source         planNode
	orphan         *orphanNode
	orderDirection mapper.SortDirection

	// state for iteration
	docsToYield     []core.Doc
	bufferedDocs    []core.Doc
	sourceExhausted bool
	orphansFetched  bool

	execInfo orphanExecInfo
}

func newOrphanWrapperNode(source planNode, orphan *orphanNode, orderDirection mapper.SortDirection) *orphanWrapperNode {
	return &orphanWrapperNode{
		docMapper:      orphan.docMapper,
		source:         source,
		orphan:         orphan,
		orderDirection: orderDirection,
	}
}

func (n *orphanWrapperNode) Kind() string {
	return "orphanNode"
}

func (n *orphanWrapperNode) Init() error {
	n.docsToYield = nil
	n.bufferedDocs = nil
	n.sourceExhausted = false
	n.orphansFetched = false
	return n.source.Init()
}

func (n *orphanWrapperNode) Start() error {
	return n.source.Start()
}

func (n *orphanWrapperNode) Prefixes(prefixes []keys.Walkable) {
	n.source.Prefixes(prefixes)
}

func (n *orphanWrapperNode) Source() planNode {
	return n.source
}

func (n *orphanWrapperNode) Close() error {
	return n.source.Close()
}

func (n *orphanWrapperNode) Value() core.Doc {
	if len(n.docsToYield) == 0 {
		return core.Doc{}
	}
	return n.docsToYield[0]
}

func (n *orphanWrapperNode) Next() (bool, error) {
	n.execInfo.iterations++

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

// nextASC buffers all source docs, then fetches orphans by exclusion,
// then yields orphans followed by buffered docs.
func (n *orphanWrapperNode) nextASC() (bool, error) {
	if n.sourceExhausted {
		if !n.orphansFetched {
			n.orphansFetched = true
			if err := n.orphan.fetchOrphans(); err != nil {
				return false, err
			}
			n.docsToYield = append(n.orphan.docs, n.bufferedDocs...)
			n.orphan.docs = nil
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
		n.bufferedDocs = append(n.bufferedDocs, n.source.Value())
	}
}

// nextDESC yields source docs first, then fetches and yields orphans.
func (n *orphanWrapperNode) nextDESC() (bool, error) {
	if !n.sourceExhausted {
		hasNext, err := n.source.Next()
		if err != nil {
			return false, err
		}
		if hasNext {
			n.docsToYield = append(n.docsToYield, n.source.Value())
			return true, nil
		}
		n.sourceExhausted = true
	}

	if !n.orphansFetched {
		n.orphansFetched = true
		if err := n.orphan.fetchOrphans(); err != nil {
			return false, err
		}
		if len(n.orphan.docs) > 0 {
			n.docsToYield = append(n.docsToYield, n.orphan.docs...)
			n.orphan.docs = nil
			return true, nil
		}
	}

	return false, nil
}

func (n *orphanWrapperNode) Explain(explainType request.ExplainType) (map[string]any, error) {
	switch explainType {
	case request.SimpleExplain:
		return map[string]any{}, nil

	case request.ExecuteExplain:
		return map[string]any{
			"iterations":   n.execInfo.iterations,
			"docFetches":   n.orphan.execInfo.fetches.DocsFetched,
			"fieldFetches": n.orphan.execInfo.fetches.FieldsFetched,
			"indexFetches": n.orphan.execInfo.fetches.IndexesFetched,
		}, nil

	default:
		return nil, ErrUnknownExplainRequestType
	}
}
