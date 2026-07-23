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

import (
	"container/heap"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// heapUnderTest abstracts the two orderings so a single table can exercise both. root returns the
// distance currently at the heap root (h[0]) without popping, which is how the engine peeks the
// nearest (minHeap) or farthest (maxHeap) element.
type heapUnderTest struct {
	name string
	// new returns a fresh, empty heap.Interface backed by the concrete heap type.
	new func() heap.Interface
	// root reads h[0].dist from the concrete heap value.
	root func(h heap.Interface) float32
	// popOrder returns the distances in the order Pop yields them: ascending for minHeap,
	// descending for maxHeap.
	popOrder func(sorted []float32) []float32
	// extreme returns the element this heap would pop next from the given contents: the minimum
	// for minHeap, the maximum for maxHeap. Panics on empty input (never called that way).
	extreme func(contents []float32) float32
}

func minHeapUT() heapUnderTest {
	return heapUnderTest{
		name: "minHeap",
		new:  func() heap.Interface { return &minHeap{} },
		root: func(h heap.Interface) float32 { return (*h.(*minHeap))[0].dist },
		popOrder: func(sorted []float32) []float32 {
			out := append([]float32(nil), sorted...) // sorted is ascending → min pops ascending
			return out
		},
		extreme: func(contents []float32) float32 {
			m := contents[0]
			for _, v := range contents {
				if v < m {
					m = v
				}
			}
			return m
		},
	}
}

func maxHeapUT() heapUnderTest {
	return heapUnderTest{
		name: "maxHeap",
		new:  func() heap.Interface { return &maxHeap{} },
		root: func(h heap.Interface) float32 { return (*h.(*maxHeap))[0].dist },
		popOrder: func(sorted []float32) []float32 {
			out := make([]float32, len(sorted)) // sorted is ascending → max pops descending
			for i, v := range sorted {
				out[len(sorted)-1-i] = v
			}
			return out
		},
		extreme: func(contents []float32) float32 {
			m := contents[0]
			for _, v := range contents {
				if v > m {
					m = v
				}
			}
			return m
		},
	}
}

func popAllDists(h heap.Interface) []float32 {
	out := make([]float32, 0, h.Len())
	for h.Len() > 0 {
		out = append(out, heap.Pop(h).(candidate).dist)
	}
	return out
}

func TestHeap_PushThenPopAll_YieldsOrderedDistances(t *testing.T) {
	// A representative unordered input; sorted ascending is the reference the per-heap popOrder
	// derives its expectation from.
	inputs := map[string][]float32{
		"unordered distinct":  {5, 1, 4, 2, 3},
		"already ascending":   {1, 2, 3, 4},
		"already descending":  {4, 3, 2, 1},
		"single element":      {7},
		"duplicate distances": {2, 1, 2, 1, 3},
		"negatives and zero":  {0, -1.5, 2.5, -3},
	}

	for _, ut := range []heapUnderTest{minHeapUT(), maxHeapUT()} {
		for name, in := range inputs {
			t.Run(ut.name+"/"+name, func(t *testing.T) {
				h := ut.new()
				for i, d := range in {
					heap.Push(h, candidate{id: NodeID(i + 1), dist: d})
				}
				require.Equal(t, len(in), h.Len())

				got := popAllDists(h)
				want := ut.popOrder(ascendingSorted(in))
				assert.Equal(t, want, got)
				assert.Equal(t, 0, h.Len(), "heap must be empty after popping everything")
			})
		}
	}
}

func TestHeap_Root_IsExtremeElement(t *testing.T) {
	in := []float32{5, 1, 4, 2, 3}

	for _, ut := range []heapUnderTest{minHeapUT(), maxHeapUT()} {
		t.Run(ut.name, func(t *testing.T) {
			h := ut.new()
			for i, d := range in {
				heap.Push(h, candidate{id: NodeID(i + 1), dist: d})
			}
			// minHeap root is the smallest (1), maxHeap root is the largest (5).
			want := ut.popOrder(ascendingSorted(in))[0]
			assert.Equal(t, want, ut.root(h))
		})
	}
}

func TestHeap_InterleavedPushPop_ReturnsCurrentExtreme(t *testing.T) {
	// Interleave pushes and pops (the real SEARCH-LAYER access pattern). Each pop must return the
	// extreme of whatever is CURRENTLY in the heap, not the global extreme of everything ever
	// pushed. A plain slice mirrors the contents as the reference model.
	// Ops: +3 +1 POP +2 +5 POP POP POP  (mixed pushes before/after pops).
	for _, ut := range []heapUnderTest{minHeapUT(), maxHeapUT()} {
		t.Run(ut.name, func(t *testing.T) {
			h := ut.new()
			model := []float32{}

			push := func(d float32) {
				heap.Push(h, candidate{dist: d})
				model = append(model, d)
			}
			popAndCheck := func() {
				want := ut.extreme(model)
				got := heap.Pop(h).(candidate).dist
				assert.Equal(t, want, got)
				model = removeFirst(model, got)
			}

			push(3)
			push(1)
			popAndCheck()
			push(2)
			push(5)
			popAndCheck()
			popAndCheck()
			popAndCheck()
			assert.Equal(t, 0, h.Len())
		})
	}
}

func TestHeap_PreservesCandidatePayload(t *testing.T) {
	// Popping must return the whole candidate (id + vector), not just the distance, since the
	// neighbour-selection heuristic relies on the carried vector.
	h := &minHeap{}
	heap.Push(h, candidate{id: 9, dist: 0.5, vector: []float32{1, 2, 3}})
	got := heap.Pop(h).(candidate)
	assert.Equal(t, NodeID(9), got.id)
	assert.Equal(t, []float32{1, 2, 3}, got.vector)
}

// removeFirst returns s with the first occurrence of v removed (reference-model bookkeeping).
func removeFirst(s []float32, v float32) []float32 {
	for i, x := range s {
		if x == v {
			return append(s[:i:i], s[i+1:]...)
		}
	}
	return s
}

// ascendingSorted returns a new ascending-sorted copy of in, used as the reference ordering.
func ascendingSorted(in []float32) []float32 {
	out := append([]float32(nil), in...)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}
