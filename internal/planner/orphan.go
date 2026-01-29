// Copyright 2025 Democratized Data Foundation
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
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// orphanNode handles fetching orphan parent documents (parents without children)
// when a join is inverted for ordering. Orphan parents have NULL values for the
// ordered field and must be included in results at the correct position:
// - ASC ordering: orphans come first (NULL sorts first)
// - DESC ordering: orphans come last (NULL sorts last)
//
// This node wraps the join and manages the ordering of orphan documents relative
// to the regular joined documents.
type orphanNode struct {
	docMapper

	// source is the wrapped join node (typeJoinOne or typeJoinMany)
	source planNode

	// join provides access to the join internals for orphan fetching
	join *invertibleTypeJoin

	// orderDirection determines where orphans are placed in results
	orderDirection mapper.SortDirection

	// state for iteration
	bufferedDocs    []core.Doc
	orphanDocs      []core.Doc
	docsToYield     []core.Doc
	sourceExhausted bool
	orphansFetched  bool
}

func newOrphanNode(source planNode, join *invertibleTypeJoin, orderDirection mapper.SortDirection) *orphanNode {
	return &orphanNode{
		docMapper:      join.docMapper,
		source:         source,
		join:           join,
		orderDirection: orderDirection,
	}
}

func (n *orphanNode) Kind() string {
	return "orphanNode"
}

func (n *orphanNode) Init() error {
	n.bufferedDocs = nil
	n.orphanDocs = nil
	n.docsToYield = nil
	n.sourceExhausted = false
	n.orphansFetched = false

	return n.source.Init()
}

func (n *orphanNode) Start() error {
	return n.source.Start()
}

func (n *orphanNode) Prefixes(prefixes []keys.Walkable) {
	n.source.Prefixes(prefixes)
}

func (n *orphanNode) Source() planNode {
	return n.source
}

func (n *orphanNode) Close() error {
	return n.source.Close()
}

func (n *orphanNode) Value() core.Doc {
	if len(n.docsToYield) == 0 {
		return core.Doc{}
	}
	return n.docsToYield[0]
}

func (n *orphanNode) Next() (bool, error) {
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

// nextASC handles ASC ordering where orphans come first.
// We must buffer all source docs, then fetch orphans, then yield orphans followed by buffered docs.
func (n *orphanNode) nextASC() (bool, error) {
	// If source is exhausted, we're in the yielding phase
	if n.sourceExhausted {
		if !n.orphansFetched {
			if err := n.fetchOrphans(); err != nil {
				return false, err
			}
			n.docsToYield = append(n.orphanDocs, n.bufferedDocs...)
			n.orphanDocs = nil
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

// nextDESC handles DESC ordering where orphans come last.
// We yield source docs first, then fetch and yield orphans at the end.
func (n *orphanNode) nextDESC() (bool, error) {
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
		if err := n.fetchOrphans(); err != nil {
			return false, err
		}
		if len(n.orphanDocs) > 0 {
			n.docsToYield = append(n.docsToYield, n.orphanDocs...)
			n.orphanDocs = nil
			return true, nil
		}
	}

	return false, nil
}

// fetchOrphans fetches parent documents that have no related children.
func (n *orphanNode) fetchOrphans() error {
	n.orphansFetched = true

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
		&parentScan.execInfo,
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
		orphans, err = fetcher.fetchAllExcluding(n.join.subFilter, n.join.state.encounteredDocIDs)
	}

	if err != nil {
		return err
	}

	n.orphanDocs = orphans
	return nil
}
