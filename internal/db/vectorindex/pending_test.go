// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package vectorindex

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

func TestPendingEntry_EncodeDecode_RoundTrips(t *testing.T) {
	for _, e := range []PendingEntry{
		{NodeID: 3, Op: PendingUpsert, Vector: []float32{1, -2.5, 0}},
		{NodeID: 4, Op: PendingDelete},
	} {
		got, err := DecodePendingEntry(e.NodeID, e.Encode())
		require.NoError(t, err)
		assert.Equal(t, e, got)
	}
}

func TestDecodePendingEntry_Malformed_ReturnsError(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   []byte
		want string
	}{
		{"empty", nil, errMalformedPending},
		{"partial float", []byte{byte(PendingUpsert), 1, 2}, errMalformedPending},
		{"unknown op", []byte{9}, errUnknownPendingOp},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := DecodePendingEntry(1, tc.in)
			require.ErrorContains(t, err, tc.want)
		})
	}
}

// One corrupt stored entry fails the whole listing, and so every search on the index. Storage
// corruption is not tolerated on purpose.
func TestListPending_CorruptStoredValue_ReturnsError(t *testing.T) {
	ctx := newStoreTestCtx(t)
	require.NoError(t, PutPending(ctx, 1, 1, 1, PendingEntry{NodeID: 1, Op: PendingDelete}))
	key := keys.NewVectorPendingKey(1, 1, 1, 2)
	require.NoError(t, datastore.CtxMustGetTxn(ctx).Datastore().Set(ctx, &key, []byte{9}))

	_, err := ListPending(ctx, 1, 1, 1, 0)
	require.ErrorContains(t, err, errVectorIndexStore)
	require.ErrorContains(t, err, errUnknownPendingOp)
}

func TestPending_PutListDelete(t *testing.T) {
	ctx := newStoreTestCtx(t)

	require.NoError(t, PutPending(ctx, 1, 1, 1, PendingEntry{NodeID: 5, Op: PendingUpsert, Vector: []float32{1, 2}}))
	require.NoError(t, PutPending(ctx, 1, 1, 1, PendingEntry{NodeID: 2, Op: PendingDelete}))
	require.NoError(t, PutPending(ctx, 1, 1, 1, PendingEntry{NodeID: 9, Op: PendingUpsert, Vector: []float32{3, 4}}))
	// Another epoch and another index must not leak in.
	require.NoError(t, PutPending(ctx, 1, 1, 2, PendingEntry{NodeID: 7, Op: PendingDelete}))
	require.NoError(t, PutPending(ctx, 1, 2, 1, PendingEntry{NodeID: 8, Op: PendingDelete}))
	// Overwrite coalesces.
	require.NoError(t, PutPending(ctx, 1, 1, 1, PendingEntry{NodeID: 5, Op: PendingDelete}))

	all, err := ListPending(ctx, 1, 1, 1, 0)
	require.NoError(t, err)
	assert.Equal(t, []PendingEntry{
		{NodeID: 2, Op: PendingDelete},
		{NodeID: 5, Op: PendingDelete},
		{NodeID: 9, Op: PendingUpsert, Vector: []float32{3, 4}},
	}, all)

	limited, err := ListPending(ctx, 1, 1, 1, 2)
	require.NoError(t, err)
	assert.Len(t, limited, 2)

	require.NoError(t, DeletePending(ctx, 1, 1, 1, 5))
	all, err = ListPending(ctx, 1, 1, 1, 0)
	require.NoError(t, err)
	assert.Len(t, all, 2)
}

func deferredTestDesc() client.VectorIndexDescription {
	d := testDesc()
	d.Maintenance = client.VectorIndexMaintenanceDeferred
	return d
}

// setupSearch maps node ids 1..n to doc ids "doc<n>" so Search can resolve its hits.
func setupSearch(t *testing.T, n int) context.Context {
	ctx := newStoreTestCtx(t)
	for i := 1; i <= n; i++ {
		require.NoError(t, id.SetDocIDMapping(ctx, 1, uint64(i), fmt.Sprintf("doc%d", i)))
	}
	return ctx
}

func searchDocs(t *testing.T, ctx context.Context, desc client.VectorIndexDescription, k int) []string {
	t.Helper()
	res, err := Search(ctx, 1, 1, 1, desc, []float32{1, 0}, k)
	require.NoError(t, err)
	docs := make([]string, len(res))
	for i, r := range res {
		docs[i] = r.DocID
	}
	return docs
}

func TestSearch_DeferredPendingUpsert_OutranksGraphNodes(t *testing.T) {
	ctx := setupSearch(t, 4)
	desc := deferredTestDesc()
	index, err := Open(ctx, 1, 1, 1, desc)
	require.NoError(t, err)
	require.NoError(t, index.Insert(1, []float32{0.7, 0.7}))
	require.NoError(t, index.Insert(2, []float32{0, 1}))

	require.NoError(t, PutPending(ctx, 1, 1, 1, PendingEntry{NodeID: 3, Op: PendingUpsert, Vector: []float32{1, 0}}))

	assert.Equal(t, []string{"doc3", "doc1"}, searchDocs(t, ctx, desc, 2))
}

func TestSearch_DeferredPendingDelete_RemovesGraphHit(t *testing.T) {
	ctx := setupSearch(t, 3)
	desc := deferredTestDesc()
	index, err := Open(ctx, 1, 1, 1, desc)
	require.NoError(t, err)
	require.NoError(t, index.Insert(1, []float32{1, 0}))
	require.NoError(t, index.Insert(2, []float32{0.7, 0.7}))

	require.NoError(t, PutPending(ctx, 1, 1, 1, PendingEntry{NodeID: 1, Op: PendingDelete}))

	assert.Equal(t, []string{"doc2"}, searchDocs(t, ctx, desc, 2))
}

func TestSearch_DeferredPendingUpsert_ReplacesGraphVectorOfSameNode(t *testing.T) {
	ctx := setupSearch(t, 3)
	desc := deferredTestDesc()
	index, err := Open(ctx, 1, 1, 1, desc)
	require.NoError(t, err)
	require.NoError(t, index.Insert(1, []float32{1, 0}))
	require.NoError(t, index.Insert(2, []float32{0.7, 0.7}))

	// Node 1 moves from the nearest to the farthest.
	require.NoError(t, PutPending(ctx, 1, 1, 1, PendingEntry{NodeID: 1, Op: PendingUpsert, Vector: []float32{0, 1}}))

	assert.Equal(t, []string{"doc2", "doc1"}, searchDocs(t, ctx, desc, 2))
}

func TestSearch_DeferredDroppedHits_WidenToFillK(t *testing.T) {
	ctx := setupSearch(t, 6)
	desc := deferredTestDesc()
	index, err := Open(ctx, 1, 1, 1, desc)
	require.NoError(t, err)
	// Nodes 1..6 get progressively farther from the query.
	vecs := [][]float32{{1, 0}, {0.9, 0.1}, {0.8, 0.2}, {0.5, 0.5}, {0.2, 0.8}, {0, 1}}
	for i, v := range vecs {
		require.NoError(t, index.Insert(uint64(i+1), v))
	}
	// The three nearest are deleted, so k=2 needs a wider graph search to find 4 and 5.
	for n := uint64(1); n <= 3; n++ {
		require.NoError(t, PutPending(ctx, 1, 1, 1, PendingEntry{NodeID: n, Op: PendingDelete}))
	}

	assert.Equal(t, []string{"doc4", "doc5"}, searchDocs(t, ctx, desc, 2))
}

func TestSearch_SyncIgnoresPending(t *testing.T) {
	ctx := setupSearch(t, 2)
	desc := testDesc()
	index, err := Open(ctx, 1, 1, 1, desc)
	require.NoError(t, err)
	require.NoError(t, index.Insert(1, []float32{0, 1}))
	require.NoError(t, PutPending(ctx, 1, 1, 1, PendingEntry{NodeID: 2, Op: PendingUpsert, Vector: []float32{1, 0}}))

	assert.Equal(t, []string{"doc1"}, searchDocs(t, ctx, desc, 2))
}
