// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/internal/index/hnsw"
)

// newVectorIndexStoreTestCtx opens an in-memory badger DB and returns a context carrying a live
// write transaction, suitable for exercising a datastoreNodeStore directly.
func newVectorIndexStoreTestCtx(t *testing.T) context.Context {
	t.Helper()
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	t.Cleanup(txn.Discard)

	return InitContext(ctx, txn)
}

func TestDatastoreNodeStore_PutNodeThenGetNode_RoundTripsNode(t *testing.T) {
	ctx := newVectorIndexStoreTestCtx(t)
	store := newDatastoreNodeStore(ctx, 1, 1, 1)

	node := hnsw.Node{
		ID:     5,
		Vector: []float32{1, 2, 3},
		Layers: [][]hnsw.NodeID{{1, 2}, {3}},
	}

	err := store.PutNode(node)
	require.NoError(t, err)

	got, ok, err := store.GetNode(5)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, node, got)
}

func TestDatastoreNodeStore_GetNode_IfMissing_ReturnsNotFound(t *testing.T) {
	ctx := newVectorIndexStoreTestCtx(t)
	store := newDatastoreNodeStore(ctx, 1, 1, 1)

	got, ok, err := store.GetNode(999)
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Equal(t, hnsw.Node{}, got)
}

func TestDatastoreNodeStore_PutMetaThenGetMeta_RoundTripsMeta(t *testing.T) {
	ctx := newVectorIndexStoreTestCtx(t)
	store := newDatastoreNodeStore(ctx, 1, 1, 1)

	meta := hnsw.Meta{
		EntryPoint: 3,
		TopLayer:   2,
	}

	err := store.PutMeta(meta)
	require.NoError(t, err)

	got, err := store.GetMeta()
	require.NoError(t, err)
	assert.Equal(t, meta, got)
}

func TestDatastoreNodeStore_GetMeta_IfEmptyKeyspace_ReturnsEmptyMeta(t *testing.T) {
	ctx := newVectorIndexStoreTestCtx(t)
	store := newDatastoreNodeStore(ctx, 1, 1, 1)

	got, err := store.GetMeta()
	require.NoError(t, err)
	assert.Equal(t, hnsw.Meta{Empty: true}, got)
}

func TestDatastoreNodeStore_IterateNodes_VisitsAllNodesAndSkipsDeleted(t *testing.T) {
	ctx := newVectorIndexStoreTestCtx(t)
	store := newDatastoreNodeStore(ctx, 1, 1, 1)

	require.NoError(t, store.PutNode(hnsw.Node{ID: 1, Vector: []float32{1}}))
	require.NoError(t, store.PutNode(hnsw.Node{ID: 2, Vector: []float32{2}}))
	require.NoError(t, store.PutNode(hnsw.Node{ID: 3, Vector: []float32{3}, Deleted: true}))

	visited := make(map[hnsw.NodeID]bool)
	err := store.IterateNodes(func(n hnsw.Node) error {
		visited[n.ID] = true
		return nil
	})
	require.NoError(t, err)

	assert.Equal(t, map[hnsw.NodeID]bool{1: true, 2: true}, visited)
}

func TestDatastoreNodeStore_DifferentIndexIDs_AreIsolated(t *testing.T) {
	ctx := newVectorIndexStoreTestCtx(t)
	store1 := newDatastoreNodeStore(ctx, 1, 1, 1)
	store2 := newDatastoreNodeStore(ctx, 1, 2, 1)

	require.NoError(t, store1.PutNode(hnsw.Node{ID: 1, Vector: []float32{1}}))

	_, ok, err := store2.GetNode(1)
	require.NoError(t, err)
	assert.False(t, ok)

	var visited int
	err = store2.IterateNodes(func(n hnsw.Node) error {
		visited++
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 0, visited)
}

func TestDatastoreNodeStore_DifferentEpochs_AreIsolated(t *testing.T) {
	ctx := newVectorIndexStoreTestCtx(t)
	store1 := newDatastoreNodeStore(ctx, 1, 1, 1)
	store2 := newDatastoreNodeStore(ctx, 1, 1, 2)

	require.NoError(t, store1.PutNode(hnsw.Node{ID: 1, Vector: []float32{1}}))
	require.NoError(t, store1.PutMeta(hnsw.Meta{EntryPoint: 1, TopLayer: 0}))

	_, ok, err := store2.GetNode(1)
	require.NoError(t, err)
	assert.False(t, ok)

	meta, err := store2.GetMeta()
	require.NoError(t, err)
	assert.Equal(t, hnsw.Meta{Empty: true}, meta)
}
