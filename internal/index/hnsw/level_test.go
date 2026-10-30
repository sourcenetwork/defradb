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
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The index is reopened for each write, so levels must depend on the seed and node id alone.
func TestGraph_RandomLevel_SameSeedAndIDGiveSameLevelAcrossInstances(t *testing.T) {
	const m = 4
	a := New(NewMemStore(), Cosine, DefaultParams(m), 7)
	b := New(NewMemStore(), Cosine, DefaultParams(m), 7)

	distinct := map[int]struct{}{}
	for id := NodeID(1); id <= 1000; id++ {
		require.Equal(t, a.randomLevel(id), b.randomLevel(id))
		require.Equal(t, a.randomLevel(id), a.randomLevel(id))
		distinct[a.randomLevel(id)] = struct{}{}
	}
	assert.Greater(t, len(distinct), 1, "every node got the same level")
}

func TestGraph_RandomLevel_FollowsDecay(t *testing.T) {
	const (
		m = 4
		n = 10000
	)
	g := New(NewMemStore(), Cosine, DefaultParams(m), 1)

	var upper int
	for id := NodeID(1); id <= n; id++ {
		if g.randomLevel(id) >= 1 {
			upper++
		}
	}
	// P(level >= 1) is 1/M, so about 2500 of 10000.
	assert.InDelta(t, float64(n)/m, float64(upper), 0.2*float64(n)/m)
}

// When every node a layer reaches is deleted, a new node must still link into the lower layers.
// Otherwise it becomes the entry point alone and the rest of the graph is cut off.
func TestGraph_InsertWhenUpperLayersAreDeleted_KeepsGraphReachable(t *testing.T) {
	g := New(NewMemStore(), Euclidean, DefaultParams(4), 3)
	rng := rand.New(rand.NewSource(5))

	vecs := map[NodeID][]float32{}
	var upper []NodeID
	for id := NodeID(1); id <= 300; id++ {
		vecs[id] = randomVector(rng, 8)
		require.NoError(t, g.Insert(id, vecs[id]))
		if g.randomLevel(id) > 0 {
			upper = append(upper, id)
		}
	}
	require.NotEmpty(t, upper)
	for _, id := range upper {
		require.NoError(t, g.Delete(id))
		delete(vecs, id)
	}

	for id := NodeID(1000); id < 1020; id++ {
		vecs[id] = randomVector(rng, 8)
		require.NoError(t, g.Insert(id, vecs[id]))
	}

	missed := 0
	for id, v := range vecs {
		got, err := g.Search(v, 1, 200)
		require.NoError(t, err)
		if len(got) != 1 || got[0] != id {
			missed++
		}
	}
	// Search is approximate, so allow a few misses. Without the fix almost every node is lost.
	assert.LessOrEqual(t, missed, len(vecs)/20, "too many nodes unreachable")
}

// Graphs built before levels came from the node id have every node on one level. Re-inserting the
// entry point of such a graph, as an update does, must not shrink it below the graph's top layer.
func TestGraph_ReinsertEntryPointOfSingleLevelGraph_KeepsGraphSearchable(t *testing.T) {
	const (
		n     = 60
		level = 3
	)
	store := NewMemStore()
	g := New(store, Euclidean, DefaultParams(4), 3)
	rng := rand.New(rand.NewSource(11))

	// Build the graph the old way: every node on the same level, linked to every other node.
	vecs := map[NodeID][]float32{}
	for id := NodeID(1); id <= n; id++ {
		vecs[id] = randomVector(rng, 8)
	}
	for id, v := range vecs {
		layers := make([][]NodeID, level+1)
		for l := range layers {
			for other := range vecs {
				if other != id {
					layers[l] = append(layers[l], other)
				}
			}
		}
		require.NoError(t, store.PutNode(Node{ID: id, Vector: vectorForMetric(Euclidean, v), Neighbours: layers}))
	}

	// Pick an entry point whose own level is lower, so a plain re-insert would shrink it.
	var entry NodeID
	for id := NodeID(1); id <= n; id++ {
		if g.randomLevel(id) < level {
			entry = id
			break
		}
	}
	require.NotZero(t, entry)
	require.NoError(t, store.PutMeta(Meta{EntryPoint: entry, TopLayer: level}))

	vecs[entry] = randomVector(rng, 8)
	require.NoError(t, g.Delete(entry))
	require.NoError(t, g.Insert(entry, vecs[entry]))

	node, found, err := store.GetNode(entry)
	require.NoError(t, err)
	require.True(t, found)
	assert.Len(t, node.Neighbours, level+1, "entry point lost its upper layers")

	for id, v := range vecs {
		got, err := g.Search(v, 1, 100)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, id, got[0])
	}
}
