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
	"math"
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// bruteForceKNN computes the true k nearest neighbours of query among
// vectors (already normalized), by exhaustive cosine-distance
// comparison. Returns node ids nearest-first.
func bruteForceKNN(metric Metric, query []float32, vectors map[NodeID][]float32, k int) []NodeID {
	type scored struct {
		id   NodeID
		dist float32
	}
	scoredList := make([]scored, 0, len(vectors))
	for id, v := range vectors {
		scoredList = append(scoredList, scored{id: id, dist: distance(metric, query, v)})
	}
	sort.Slice(scoredList, func(i, j int) bool { return scoredList[i].dist < scoredList[j].dist })
	if k > len(scoredList) {
		k = len(scoredList)
	}
	out := make([]NodeID, k)
	for i := 0; i < k; i++ {
		out[i] = scoredList[i].id
	}
	return out
}

func randomVector(rng *rand.Rand, dim int) []float32 {
	v := make([]float32, dim)
	for i := range v {
		v[i] = float32(rng.NormFloat64())
	}
	return v
}

func TestGraph_RandomVectors_RecallMeetsThreshold(t *testing.T) {
	const (
		n           = 1000
		dim         = 32
		k           = 10
		numQueries  = 50
		efSearch    = 64
		recallFloor = 0.90
	)

	rng := rand.New(rand.NewSource(42))
	store := NewMemStore()
	params := DefaultParams(16)
	params.EfSearch = efSearch
	g := New(store, Cosine, params, 42)

	truth := make(map[NodeID][]float32, n)
	for i := range n {
		v := randomVector(rng, dim)
		id := NodeID(i + 1)
		require.NoError(t, g.Insert(id, v))
		truth[id] = normalize(v)
	}

	var totalRecall float64
	for range numQueries {
		query := randomVector(rng, dim)
		expected := bruteForceKNN(Cosine, normalize(query), truth, k)
		actual, err := g.Search(query, k, efSearch)
		require.NoError(t, err)

		expectedSet := make(map[NodeID]struct{}, len(expected))
		for _, id := range expected {
			expectedSet[id] = struct{}{}
		}
		hits := 0
		for _, id := range actual {
			if _, ok := expectedSet[id]; ok {
				hits++
			}
		}
		totalRecall += float64(hits) / float64(len(expected))
	}

	avgRecall := totalRecall / float64(numQueries)
	t.Logf("average recall@%d over %d queries: %.4f", k, numQueries, avgRecall)
	assert.GreaterOrEqual(t, avgRecall, recallFloor, "recall below threshold")
}

func TestGraph_SmallHandPlaced_ReturnsNearestInOrder(t *testing.T) {
	store := NewMemStore()
	g := New(store, Cosine, DefaultParams(8), 1)

	// Vectors placed at increasing angles from the +X axis so their
	// cosine similarity to the query (+X axis) decreases monotonically.
	ids := []NodeID{1, 2, 3, 4}
	vectors := [][]float32{
		{1, 0},      // 0 degrees -- closest to query
		{0.9, 0.1},  // small angle
		{0.5, 0.86}, // ~60 degrees
		{-1, 0},     // 180 degrees -- farthest
	}
	for i, v := range vectors {
		require.NoError(t, g.Insert(ids[i], v))
	}

	result, err := g.Search([]float32{1, 0}, 4, 64)
	require.NoError(t, err)
	require.Equal(t, []NodeID{1, 2, 3, 4}, result)
}

func TestGraph_DeletedNodes_ExcludedFromResults(t *testing.T) {
	const (
		n        = 300
		dim      = 16
		k        = 10
		efSearch = 64
	)

	rng := rand.New(rand.NewSource(7))
	store := NewMemStore()
	g := New(store, Cosine, DefaultParams(16), 7)

	truth := make(map[NodeID][]float32, n)
	allIDs := make([]NodeID, 0, n)
	for i := range n {
		v := randomVector(rng, dim)
		id := NodeID(i + 1)
		require.NoError(t, g.Insert(id, v))
		truth[id] = normalize(v)
		allIDs = append(allIDs, id)
	}

	// Delete every 3rd node.
	deleted := make(map[NodeID]struct{})
	for i, id := range allIDs {
		if i%3 == 0 {
			require.NoError(t, g.Delete(id))
			deleted[id] = struct{}{}
			delete(truth, id)
		}
	}

	var totalRecall float64
	const numQueries = 30
	for range numQueries {
		query := randomVector(rng, dim)
		result, err := g.Search(query, k, efSearch)
		require.NoError(t, err)

		for _, id := range result {
			_, isDeleted := deleted[id]
			assert.False(t, isDeleted, "deleted node %d appeared in search results", id)
		}

		expected := bruteForceKNN(Cosine, normalize(query), truth, k)
		expectedSet := make(map[NodeID]struct{}, len(expected))
		for _, id := range expected {
			expectedSet[id] = struct{}{}
		}
		hits := 0
		for _, id := range result {
			if _, ok := expectedSet[id]; ok {
				hits++
			}
		}
		totalRecall += float64(hits) / float64(len(expected))
	}

	avgRecall := totalRecall / float64(numQueries)
	t.Logf("average recall@%d after deletions: %.4f", k, avgRecall)
	assert.GreaterOrEqual(t, avgRecall, 0.85, "recall degraded too much after deletions")
}

func TestGraph_KExceedsNodeCount_ReturnsAll(t *testing.T) {
	store := NewMemStore()
	g := New(store, Cosine, DefaultParams(8), 3)

	ids := []NodeID{1, 2, 3, 4, 5}
	rng := rand.New(rand.NewSource(3))
	for _, id := range ids {
		require.NoError(t, g.Insert(id, randomVector(rng, 8)))
	}

	result, err := g.Search(randomVector(rng, 8), 100, 64)
	require.NoError(t, err)
	assert.Len(t, result, len(ids))

	seen := make(map[NodeID]struct{})
	for _, id := range result {
		seen[id] = struct{}{}
	}
	for _, id := range ids {
		_, ok := seen[id]
		assert.True(t, ok, "expected id %d to be present in results", id)
	}
}

func TestGraph_SameSeed_Deterministic(t *testing.T) {
	buildGraph := func(seed int64) (*Graph, []NodeID) {
		store := NewMemStore()
		g := New(store, Cosine, DefaultParams(8), seed)
		rng := rand.New(rand.NewSource(99))
		ids := make([]NodeID, 200)
		for i := range 200 {
			id := NodeID(i + 1)
			ids[i] = id
			require.NoError(t, g.Insert(id, randomVector(rng, 16)))
		}
		return g, ids
	}

	g1, _ := buildGraph(123)
	g2, _ := buildGraph(123)

	meta1, err := g1.store.GetMeta()
	require.NoError(t, err)
	meta2, err := g2.store.GetMeta()
	require.NoError(t, err)
	assert.Equal(t, meta1, meta2)

	queryRng := rand.New(rand.NewSource(55))
	query := randomVector(queryRng, 16)

	result1, err := g1.Search(query, 10, 64)
	require.NoError(t, err)
	result2, err := g2.Search(query, 10, 64)
	require.NoError(t, err)
	assert.Equal(t, result1, result2)
}

func TestGraph_EmptyGraph_ReturnsNoResults(t *testing.T) {
	store := NewMemStore()
	g := New(store, Cosine, DefaultParams(8), 1)

	result, err := g.Search([]float32{1, 0, 0}, 5, 64)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestNormalize_ZeroVector_ReturnsUnchanged(t *testing.T) {
	v := []float32{0, 0, 0}
	out := normalize(v)
	assert.Equal(t, []float32{0, 0, 0}, out)
}

func TestNormalize_NegligibleNorm_ReturnsUnchangedWithoutInf(t *testing.T) {
	// A vector whose squared norm is below normThreshold must be returned unchanged rather than
	// divided by a near-zero norm (which would overflow float32 to ±Inf).
	v := []float32{1e-25, -1e-25}
	out := normalize(v)
	assert.Equal(t, v, out)
	for _, x := range out {
		assert.False(t, math.IsInf(float64(x), 0), "normalize must not produce Inf for a negligible-norm vector")
	}
}

func TestNormalize_NonZeroVector_ReturnsUnitLength(t *testing.T) {
	v := []float32{3, 4}
	out := normalize(v)

	var sumSq float64
	for _, x := range out {
		sumSq += float64(x) * float64(x)
	}
	assert.InDelta(t, 1.0, sumSq, 1e-6)
	assert.InDelta(t, 0.6, out[0], 1e-6)
	assert.InDelta(t, 0.8, out[1], 1e-6)
}
