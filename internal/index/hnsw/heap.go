// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package hnsw

import "container/heap"

// candidate pairs a node id with its distance to the current query vector
// and (optionally) the node's own vector, so that pairwise distances
// between candidates can be computed without further store lookups (used
// by the neighbour-selection heuristic).
type candidate struct {
	id     NodeID
	dist   float32
	vector []float32
}

// candidateHeap is a heap of candidates, ordered by nearestFirst. The search needs both orderings,
// which are the same heap with a flipped comparison, so one type covers both:
//   - nearestFirst true: nearest pops first (the frontier of nodes still to explore).
//   - nearestFirst false: farthest pops first, so the farthest result is cheap to drop once there
//     are more than ef results.
type candidateHeap struct {
	items        []candidate
	nearestFirst bool
}

var _ heap.Interface = (*candidateHeap)(nil)

func newMinHeap() *candidateHeap { return &candidateHeap{nearestFirst: true} }
func newMaxHeap() *candidateHeap { return &candidateHeap{nearestFirst: false} }

func (h candidateHeap) Len() int { return len(h.items) }
func (h candidateHeap) Less(i, j int) bool {
	if h.nearestFirst {
		return h.items[i].dist < h.items[j].dist
	}
	return h.items[i].dist > h.items[j].dist
}
func (h candidateHeap) Swap(i, j int) { h.items[i], h.items[j] = h.items[j], h.items[i] }
func (h *candidateHeap) Push(x any) {
	if c, ok := x.(candidate); ok {
		h.items = append(h.items, c)
	}
}
func (h *candidateHeap) Pop() any {
	n := len(h.items)
	item := h.items[n-1]
	h.items = h.items[:n-1]
	return item
}
