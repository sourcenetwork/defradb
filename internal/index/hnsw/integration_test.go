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
	"errors"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errFaultStore is a NodeStore that delegates to a real store but injects an error on the first
// call to a chosen operation. It exists to exercise the engine's error-propagation paths, which
// the failure-free in-memory store can never reach — proving a storage failure surfaces as an
// error rather than silently corrupting the graph.
type errFaultStore struct {
	NodeStore
	failGetNode bool
	failPutNode bool
	failGetMeta bool
	err         error
}

func (s *errFaultStore) GetNode(id NodeID) (Node, bool, error) {
	if s.failGetNode {
		s.failGetNode = false
		return Node{}, false, s.err
	}
	return s.NodeStore.GetNode(id)
}

func (s *errFaultStore) PutNode(n Node) error {
	if s.failPutNode {
		s.failPutNode = false
		return s.err
	}
	return s.NodeStore.PutNode(n)
}

func (s *errFaultStore) GetMeta() (Meta, bool, error) {
	if s.failGetMeta {
		s.failGetMeta = false
		return Meta{}, false, s.err
	}
	return s.NodeStore.GetMeta()
}

// reloadThroughCodec serializes every node and the meta of src through the on-the-wire codec and
// decodes them into a fresh store — simulating a process restart that rebuilds the graph purely
// from persisted bytes (no in-memory state carried over).
func reloadThroughCodec(t *testing.T, src NodeStore) NodeStore {
	t.Helper()
	dst := NewMemStore()

	require.NoError(t, src.IterateNodes(func(n Node) error {
		b, err := MarshalNode(n)
		require.NoError(t, err)
		decoded, err := UnmarshalNode(b)
		require.NoError(t, err)
		return dst.PutNode(decoded)
	}))

	meta, found, err := src.GetMeta()
	require.NoError(t, err)
	if found {
		mb, err := MarshalMeta(meta)
		require.NoError(t, err)
		decodedMeta, err := UnmarshalMeta(mb)
		require.NoError(t, err)
		require.NoError(t, dst.PutMeta(decodedMeta))
	}

	return dst
}

// TestGraph_PersistReloadSearch_MatchesOriginal is the full-stack round trip: build a graph, persist
// every node + meta through the codec into a fresh store, wrap a new HNSWIndex around that store, and
// confirm searches return identical results to the original in-memory graph. This proves the
// algorithm's on-disk representation is complete — search survives a cold reload.
func TestGraph_PersistReloadSearch_MatchesOriginal(t *testing.T) {
	const (
		n        = 500
		dim      = 24
		k        = 10
		efSearch = 64
	)

	rng := rand.New(rand.NewSource(101))
	store := NewMemStore()
	g := New(store, Cosine, DefaultParams(16), 101)

	for i := range n {
		require.NoError(t, g.Insert(NodeID(i+1), randomVector(rng, dim)))
	}

	// Rebuild a graph from persisted bytes only.
	reloaded := reloadThroughCodec(t, store)
	g2 := New(reloaded, Cosine, DefaultParams(16), 101)

	queryRng := rand.New(rand.NewSource(202))
	for range 30 {
		query := randomVector(queryRng, dim)
		before, err := g.Search(query, k, efSearch)
		require.NoError(t, err)
		after, err := g2.Search(query, k, efSearch)
		require.NoError(t, err)
		assert.Equal(t, before, after, "search results must be identical after a persist/reload cycle")
	}
}

func TestGraph_DeleteNonExistentID_IsNoOpWithoutError(t *testing.T) {
	g := New(NewMemStore(), Cosine, DefaultParams(8), 1)
	require.NoError(t, g.Insert(1, []float32{1, 0}))

	// Deleting an id that was never inserted must not error and must not affect existing nodes.
	require.NoError(t, g.Delete(999))

	res, err := g.Search([]float32{1, 0}, 5, 16)
	require.NoError(t, err)
	assert.Equal(t, []NodeID{1}, res)
}

func TestGraph_InsertWithMissingEntryPoint_ReturnsError(t *testing.T) {
	// A torn graph: meta names an entry point node the store does not have. Insert must fail rather
	// than descend from a zero-value node.
	store := NewMemStore()
	require.NoError(t, store.PutMeta(Meta{EntryPoint: 42, TopLayer: 0}))

	g := New(store, Cosine, DefaultParams(8), 1)
	err := g.Insert(1, []float32{1, 0})
	require.ErrorIs(t, err, ErrEntryPointNotFound)
}

func TestGraph_DeleteAlreadyDeletedID_IsIdempotent(t *testing.T) {
	g := New(NewMemStore(), Cosine, DefaultParams(8), 1)
	require.NoError(t, g.Insert(1, []float32{1, 0}))
	require.NoError(t, g.Insert(2, []float32{0, 1}))

	require.NoError(t, g.Delete(1))
	require.NoError(t, g.Delete(1)) // second delete is a no-op

	res, err := g.Search([]float32{1, 0}, 5, 16)
	require.NoError(t, err)
	assert.NotContains(t, res, NodeID(1))
	assert.Contains(t, res, NodeID(2))
}

func TestGraph_DeleteEntryPoint_StillSearchesRemaining(t *testing.T) {
	// Deleting the graph's entry point must not break traversal: search traverses through the
	// tombstone to reach the live nodes behind it, and never returns the deleted entry point.
	const (
		n        = 200
		dim      = 16
		efSearch = 64
	)
	rng := rand.New(rand.NewSource(5))
	store := NewMemStore()
	g := New(store, Cosine, DefaultParams(16), 5)

	truth := make(map[NodeID][]float32, n)
	for i := range n {
		v := randomVector(rng, dim)
		id := NodeID(i + 1)
		require.NoError(t, g.Insert(id, v))
		truth[id] = normalize(v)
	}

	meta, _, err := store.GetMeta()
	require.NoError(t, err)
	entryPoint := meta.EntryPoint
	require.NoError(t, g.Delete(entryPoint))
	delete(truth, entryPoint)

	// Search must still work, exclude the deleted entry point, and keep decent recall by traversing
	// through the tombstoned entry point to the rest of the graph.
	queryRng := rand.New(rand.NewSource(6))
	var totalRecall float64
	const numQueries = 20
	for range numQueries {
		query := randomVector(queryRng, dim)
		res, err := g.Search(query, 10, efSearch)
		require.NoError(t, err)
		assert.NotContains(t, res, entryPoint, "deleted entry point must never appear in results")

		expected := bruteForceKNN(Cosine, normalize(query), truth, 10)
		expectedSet := make(map[NodeID]struct{}, len(expected))
		for _, id := range expected {
			expectedSet[id] = struct{}{}
		}
		hits := 0
		for _, id := range res {
			if _, ok := expectedSet[id]; ok {
				hits++
			}
		}
		totalRecall += float64(hits) / float64(len(expected))
	}
	assert.GreaterOrEqual(t, totalRecall/numQueries, 0.80,
		"recall must stay reasonable after deleting the entry point (traversal through tombstone)")
}

func TestGraph_TraversesThroughTombstoneToReachLiveNode(t *testing.T) {
	// A small hand-built case: three points on a line where the middle one is the entry point and
	// gets deleted. The query is nearest to the far node, which is only reachable by traversing
	// through the deleted middle node — it must still be found.
	store := NewMemStore()
	g := New(store, Cosine, DefaultParams(8), 1)

	require.NoError(t, g.Insert(1, []float32{1, 0}))     // becomes entry point (first inserted)
	require.NoError(t, g.Insert(2, []float32{0.7, 0.7})) // middle
	require.NoError(t, g.Insert(3, []float32{0, 1}))     // far

	// Delete whatever the entry point is (id 1, the first insert).
	meta, _, err := store.GetMeta()
	require.NoError(t, err)
	require.NoError(t, g.Delete(meta.EntryPoint))

	// Query nearest to id 3; it must be reachable and returned despite the deleted entry point.
	res, err := g.Search([]float32{0, 1}, 2, 16)
	require.NoError(t, err)
	assert.NotContains(t, res, meta.EntryPoint)
	assert.Contains(t, res, NodeID(3))
}

// --- Update-as-reinsert (the collectionVectorIndex.Update strategy: tombstone old, insert new) ---

func TestGraph_UpdateAsReinsert_ReflectsNewVector(t *testing.T) {
	// The DB layer models an update as delete(old) + insert(newID). Simulate that and confirm the
	// old vector is gone from results and the new one is findable.
	store := NewMemStore()
	g := New(store, Cosine, DefaultParams(8), 2)

	// Seed some neighbours so the graph is non-trivial.
	rng := rand.New(rand.NewSource(2))
	for i := 3; i <= 40; i++ {
		require.NoError(t, g.Insert(NodeID(i), randomVector(rng, 8)))
	}

	// Insert the "original" version of a document's vector, pointing along +X.
	orig := []float32{1, 0, 0, 0, 0, 0, 0, 0}
	require.NoError(t, g.Insert(1, orig))

	// Query along +X finds it.
	res, err := g.Search(orig, 5, 32)
	require.NoError(t, err)
	assert.Contains(t, res, NodeID(1))

	// "Update": tombstone the old node, reinsert under a new id with a vector pointing along +Y.
	require.NoError(t, g.Delete(1))
	updated := []float32{0, 1, 0, 0, 0, 0, 0, 0}
	require.NoError(t, g.Insert(2, updated))

	// The old node must be gone; a query along the NEW direction finds the new node.
	resNew, err := g.Search(updated, 5, 32)
	require.NoError(t, err)
	assert.NotContains(t, resNew, NodeID(1), "tombstoned pre-update node must not appear")
	assert.Contains(t, resNew, NodeID(2), "reinserted node must be findable at its new vector")
}

func TestGraph_InsertStoreError_Propagates(t *testing.T) {
	sentinel := errors.New("boom")

	// First insert (empty graph) must fail if the initial PutNode fails.
	s1 := &errFaultStore{NodeStore: NewMemStore(), failPutNode: true, err: sentinel}
	g1 := New(s1, Cosine, DefaultParams(8), 1)
	assert.ErrorIs(t, g1.Insert(1, []float32{1, 0}), sentinel)

	// A GetMeta failure on a later insert must propagate.
	s2 := &errFaultStore{NodeStore: NewMemStore(), err: sentinel}
	g2 := New(s2, Cosine, DefaultParams(8), 1)
	require.NoError(t, g2.Insert(1, []float32{1, 0}))
	s2.failGetMeta = true
	assert.ErrorIs(t, g2.Insert(2, []float32{0, 1}), sentinel)
}

func TestGraph_SearchStoreError_Propagates(t *testing.T) {
	sentinel := errors.New("boom")
	s := &errFaultStore{NodeStore: NewMemStore(), err: sentinel}
	g := New(s, Cosine, DefaultParams(8), 1)
	require.NoError(t, g.Insert(1, []float32{1, 0}))

	s.failGetMeta = true
	_, err := g.Search([]float32{1, 0}, 5, 16)
	assert.ErrorIs(t, err, sentinel)
}

func TestGraph_DeleteStoreError_Propagates(t *testing.T) {
	sentinel := errors.New("boom")
	s := &errFaultStore{NodeStore: NewMemStore(), err: sentinel}
	g := New(s, Cosine, DefaultParams(8), 1)
	require.NoError(t, g.Insert(1, []float32{1, 0}))

	s.failGetNode = true
	assert.ErrorIs(t, g.Delete(1), sentinel)
}

func TestGraph_InsertsBeyondM_ShrinksNeighboursToCap(t *testing.T) {
	// With a small M, inserting many mutually-close vectors forces addLink's shrink path. After the
	// inserts, no node may exceed the per-layer cap (Mmax0 at layer 0, M above).
	const (
		m   = 4
		n   = 60
		dim = 8
	)
	params := DefaultParams(m)
	store := NewMemStore()
	g := New(store, Cosine, params, 9)

	rng := rand.New(rand.NewSource(9))
	for i := range n {
		require.NoError(t, g.Insert(NodeID(i+1), randomVector(rng, dim)))
	}

	sawShrunkLayer0 := false
	require.NoError(t, store.IterateNodes(func(node Node) error {
		for layer, neighbours := range node.Neighbours {
			maxDeg := params.M
			if layer == 0 {
				maxDeg = params.Mmax0
			}
			assert.LessOrEqualf(t, len(neighbours), maxDeg,
				"node %d layer %d has %d neighbours, exceeds cap %d", node.ID, layer, len(neighbours), maxDeg)
			if layer == 0 && len(neighbours) == params.Mmax0 {
				sawShrunkLayer0 = true
			}
		}
		return nil
	}))
	assert.True(t, sawShrunkLayer0, "expected at least one node at the layer-0 degree cap (shrink exercised)")
}

func TestDefaultParams_NonPositiveM_FallsBackToDefault(t *testing.T) {
	for _, m := range []int{0, -1, -100} {
		p := DefaultParams(m)
		assert.Equal(t, 16, p.M, "non-positive M should fall back to 16")
		assert.Equal(t, 32, p.Mmax0)
	}
}
