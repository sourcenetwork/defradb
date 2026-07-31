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

// minHeap is a min-heap of candidates ordered by ascending distance
// (nearest first out). Used as the candidate frontier "C" in SEARCH-LAYER.
type minHeap []candidate

func (h minHeap) Len() int           { return len(h) }
func (h minHeap) Less(i, j int) bool { return h[i].dist < h[j].dist }
func (h minHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x any) {
	if c, ok := x.(candidate); ok {
		*h = append(*h, c)
	}
}
func (h *minHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

// maxHeap is a max-heap of candidates ordered by descending distance
// (farthest first out). Used as the dynamic result set "W" in
// SEARCH-LAYER, so that the farthest element can be evicted cheaply once
// the set exceeds ef.
type maxHeap []candidate

func (h maxHeap) Len() int           { return len(h) }
func (h maxHeap) Less(i, j int) bool { return h[i].dist > h[j].dist }
func (h maxHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *maxHeap) Push(x any) {
	if c, ok := x.(candidate); ok {
		*h = append(*h, c)
	}
}
func (h *maxHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

var (
	_ heap.Interface = (*minHeap)(nil)
	_ heap.Interface = (*maxHeap)(nil)
)
