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
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/db/vectorindex"
)

// newVectorIndexTestDB opens an in-memory badger-backed DB with a collection carrying a
// [Float32!] @vectorIndex field, ready for document writes.
func newVectorIndexTestDB(t *testing.T, dimensions int) (context.Context, *DB, client.Collection) {
	t.Helper()
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(db.Close)

	_, err = db.AddCollection(ctx, `
		type Users {
			name: String
			embedding: [Float32!] @vectorIndex(dimensions: `+strconv.Itoa(dimensions)+`, HNSW: {metric: COSINE})
		}
	`)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "Users")
	require.NoError(t, err)

	return ctx, db, col
}

// vectorIndexSearch runs a nearest-neighbour search against col's vector index over a fresh read
// transaction, returning the matched document ids nearest-first. It goes through the same
// vectorindex.Search the query planner uses, so the write-path tests verify the graph the way the
// database actually reads it.
func vectorIndexSearch(
	t *testing.T, ctx context.Context, db *DB, col client.Collection, query []float32, k int,
) []string {
	t.Helper()

	txn, err := db.NewTxn(true)
	require.NoError(t, err)
	t.Cleanup(txn.Discard)
	readCtx := InitContext(ctx, txn)

	collectionShortID, err := id.GetCollectionShortID(readCtx, col.Version().CollectionID)
	require.NoError(t, err)

	indexes := col.Version().Indexes
	require.Len(t, indexes, 1)
	desc := indexes[0]
	vectorDesc, ok := desc.GetVector()
	require.True(t, ok)

	epoch, err := getIndexEpoch(readCtx, col.Version().CollectionID, desc.ID)
	require.NoError(t, err)

	results, err := vectorindex.Search(readCtx, collectionShortID, desc.ID, epoch, *vectorDesc, query, k)
	require.NoError(t, err)

	docIDs := make([]string, len(results))
	for i, r := range results {
		docIDs[i] = r.DocID
	}
	return docIDs
}

func TestCollectionVectorIndex_Save_InsertsIntoGraphAndIsSearchable(t *testing.T) {
	ctx, db, col := newVectorIndexTestDB(t, 3)

	doc1, err := client.NewDocFromJSON(ctx, []byte(`{"name": "near", "embedding": [1, 0, 0]}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.SaveDocument(ctx, doc1))

	doc2, err := client.NewDocFromJSON(ctx, []byte(`{"name": "far", "embedding": [0, 1, 0]}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.SaveDocument(ctx, doc2))

	results := vectorIndexSearch(t, ctx, db, col, []float32{1, 0, 0}, 1)
	require.Len(t, results, 1)
	assert.Equal(t, doc1.ID().String(), results[0])
}

func TestCollectionVectorIndex_Delete_ExcludesDocFromSearch(t *testing.T) {
	ctx, db, col := newVectorIndexTestDB(t, 3)

	doc1, err := client.NewDocFromJSON(ctx, []byte(`{"name": "near", "embedding": [1, 0, 0]}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.SaveDocument(ctx, doc1))

	doc2, err := client.NewDocFromJSON(ctx, []byte(`{"name": "also near", "embedding": [0.9, 0.1, 0]}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.SaveDocument(ctx, doc2))

	deleted, err := col.DeleteDocument(ctx, doc1.ID())
	require.NoError(t, err)
	require.True(t, deleted)

	results := vectorIndexSearch(t, ctx, db, col, []float32{1, 0, 0}, 2)
	for _, docID := range results {
		assert.NotEqual(t, doc1.ID().String(), docID)
	}
}

func TestCollectionVectorIndex_Update_ReplacesVectorInGraph(t *testing.T) {
	ctx, db, col := newVectorIndexTestDB(t, 3)

	doc, err := client.NewDocFromJSON(ctx, []byte(`{"name": "moving", "embedding": [1, 0, 0]}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.SaveDocument(ctx, doc))

	// A second, stationary doc at the old vector's location, so that after the update below only
	// this one (and not "moving") should be the nearest match there.
	stationary, err := client.NewDocFromJSON(ctx, []byte(`{"name": "stationary", "embedding": [1, 0, 0]}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.SaveDocument(ctx, stationary))

	require.NoError(t, doc.Set(ctx, "embedding", []float32{0, 1, 0}))
	require.NoError(t, col.UpdateDocument(ctx, doc))

	// The updated doc is now the nearest match to its new vector.
	results := vectorIndexSearch(t, ctx, db, col, []float32{0, 1, 0}, 1)
	require.Len(t, results, 1)
	assert.Equal(t, doc.ID().String(), results[0])

	// Searching near the old vector now finds "stationary" as the nearest match: "moving"'s old
	// entry was replaced in place (its node id now holds the new vector), not left behind as a
	// stale duplicate at the old location.
	oldResults := vectorIndexSearch(t, ctx, db, col, []float32{1, 0, 0}, 1)
	require.Len(t, oldResults, 1)
	assert.Equal(t, stationary.ID().String(), oldResults[0])
}

func TestCollectionVectorIndex_Save_DimensionMismatch_ReturnsTypedError(t *testing.T) {
	ctx, _, col := newVectorIndexTestDB(t, 3)

	doc, err := client.NewDocFromJSON(ctx, []byte(`{"name": "bad", "embedding": [1, 0]}`), col.Version())
	require.NoError(t, err)

	err = col.SaveDocument(ctx, doc)
	require.Error(t, err)
	assert.ErrorContains(t, err, "vector dimension mismatch")
}

// Deleting a document the graph never held, from a built index, means the index has drifted from the
// data and must error. While the index is still building an absent document is expected instead, so
// that case is not tested here.
func TestCollectionVectorIndex_Delete_MissingNodeOnBuiltIndex_ReturnsCorruptedError(t *testing.T) {
	ctx, db, col := newVectorIndexTestDB(t, 3)

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	t.Cleanup(txn.Discard)
	txnCtx := InitContext(ctx, txn)

	desc := col.Version().Indexes[0]
	// building is false, so a missing node is treated as drift rather than a pending backfill.
	index, err := NewCollectionIndex(txnCtx, col, desc, false)
	require.NoError(t, err)

	// A document that was never saved has no short id, so the index cannot find its node.
	doc, err := client.NewDocFromJSON(ctx, []byte(`{"name": "ghost", "embedding": [1, 0, 0]}`), col.Version())
	require.NoError(t, err)

	err = index.Delete(txnCtx, doc)
	require.Error(t, err)
	assert.ErrorContains(t, err, "corrupted index")
	assert.ErrorContains(t, err, doc.ID().String())
}
