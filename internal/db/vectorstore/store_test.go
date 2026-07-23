// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package vectorstore

import (
	"context"
	"testing"

	badgerds "github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/badger"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/lock"
	"github.com/sourcenetwork/defradb/internal/index/hnsw"
)

// newStoreTestCtx opens an in-memory badger store and returns a context carrying a live write
// transaction, the way the store resolves it (datastore.CtxMustGetTxn / corekv txn on context).
// This is self-contained so the store test does not depend on the parent db package.
func newStoreTestCtx(t *testing.T) context.Context {
	t.Helper()
	ctx := context.Background()

	rootstore, err := badger.NewDatastore("", badgerds.DefaultOptions("").WithInMemory(true))
	require.NoError(t, err)
	t.Cleanup(func() { _ = rootstore.Close() })

	txn := datastore.NewTxnFrom(rootstore, lock.NewLockSet(), 1, false, immutable.None[int]())
	t.Cleanup(txn.Discard)

	ctx = datastore.CtxSetTxn(ctx, txn)
	ctx = corekv.SetCtxTxn(ctx, txn.Txn())
	return ctx
}

func TestNodeStore_PutNodeThenGetNode_RoundTripsNode(t *testing.T) {
	ctx := newStoreTestCtx(t)
	store := NewNodeStore(ctx, 1, 1, 1)

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

func TestNodeStore_GetNode_IfMissing_ReturnsNotFound(t *testing.T) {
	ctx := newStoreTestCtx(t)
	store := NewNodeStore(ctx, 1, 1, 1)

	got, ok, err := store.GetNode(999)
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Equal(t, hnsw.Node{}, got)
}

func TestNodeStore_PutMetaThenGetMeta_RoundTripsMeta(t *testing.T) {
	ctx := newStoreTestCtx(t)
	store := NewNodeStore(ctx, 1, 1, 1)

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

func TestNodeStore_GetMeta_IfEmptyKeyspace_ReturnsEmptyMeta(t *testing.T) {
	ctx := newStoreTestCtx(t)
	store := NewNodeStore(ctx, 1, 1, 1)

	got, err := store.GetMeta()
	require.NoError(t, err)
	assert.Equal(t, hnsw.Meta{Empty: true}, got)
}

func TestNodeStore_IterateNodes_VisitsAllNodesAndSkipsDeleted(t *testing.T) {
	ctx := newStoreTestCtx(t)
	store := NewNodeStore(ctx, 1, 1, 1)

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

func TestNodeStore_DifferentIndexIDs_AreIsolated(t *testing.T) {
	ctx := newStoreTestCtx(t)
	store1 := NewNodeStore(ctx, 1, 1, 1)
	store2 := NewNodeStore(ctx, 1, 2, 1)

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

func TestNodeStore_DifferentEpochs_AreIsolated(t *testing.T) {
	ctx := newStoreTestCtx(t)
	store1 := NewNodeStore(ctx, 1, 1, 1)
	store2 := NewNodeStore(ctx, 1, 1, 2)

	require.NoError(t, store1.PutNode(hnsw.Node{ID: 1, Vector: []float32{1}}))
	require.NoError(t, store1.PutMeta(hnsw.Meta{EntryPoint: 1, TopLayer: 0}))

	_, ok, err := store2.GetNode(1)
	require.NoError(t, err)
	assert.False(t, ok)

	meta, err := store2.GetMeta()
	require.NoError(t, err)
	assert.Equal(t, hnsw.Meta{Empty: true}, meta)
}
