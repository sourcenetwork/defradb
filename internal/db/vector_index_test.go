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
	"github.com/sourcenetwork/defradb/internal/db/vectorstore"
	"github.com/sourcenetwork/defradb/internal/index/hnsw"
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
			embedding: [Float32!] @vectorIndex(type: HNSW, metric: COSINE, dimensions: `+strconv.Itoa(dimensions)+`)
		}
	`)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "Users")
	require.NoError(t, err)

	return ctx, db, col
}

// vectorIndexGraph builds a read-only view of the graph maintained by col's vector index,
// against a fresh read transaction on ctx.
func vectorIndexGraph(t *testing.T, ctx context.Context, db *DB, col client.Collection) *hnsw.Graph {
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

	epoch, err := getIndexEpoch(readCtx, col.Version().CollectionID, desc.ID)
	require.NoError(t, err)

	params := hnsw.DefaultParams(int(desc.Vector.HNSW.M))
	params.EfConstruction = int(desc.Vector.HNSW.EfConstruction)
	params.EfSearch = int(desc.Vector.HNSW.EfSearch)
	return vectorstore.NewGraph(readCtx, collectionShortID, desc.ID, epoch, hnsw.Cosine, params)
}

func vectorIndexDocShortID(t *testing.T, ctx context.Context, db *DB, col client.Collection, docID string) hnsw.NodeID {
	t.Helper()

	txn, err := db.NewTxn(true)
	require.NoError(t, err)
	t.Cleanup(txn.Discard)
	readCtx := InitContext(ctx, txn)

	collectionShortID, err := id.GetCollectionShortID(readCtx, col.Version().CollectionID)
	require.NoError(t, err)

	docShortID, found, err := id.GetDocShortID(readCtx, collectionShortID, docID)
	require.NoError(t, err)
	require.True(t, found)
	return hnsw.NodeID(docShortID)
}

func TestCollectionVectorIndex_Save_InsertsIntoGraphAndIsSearchable(t *testing.T) {
	ctx, db, col := newVectorIndexTestDB(t, 3)

	doc1, err := client.NewDocFromJSON(ctx, []byte(`{"name": "near", "embedding": [1, 0, 0]}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.SaveDocument(ctx, doc1))

	doc2, err := client.NewDocFromJSON(ctx, []byte(`{"name": "far", "embedding": [0, 1, 0]}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.SaveDocument(ctx, doc2))

	graph := vectorIndexGraph(t, ctx, db, col)
	results, err := graph.Search([]float32{1, 0, 0}, 1, 10)
	require.NoError(t, err)
	require.Len(t, results, 1)

	wantID := vectorIndexDocShortID(t, ctx, db, col, doc1.ID().String())
	assert.Equal(t, wantID, results[0])
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

	graph := vectorIndexGraph(t, ctx, db, col)
	results, err := graph.Search([]float32{1, 0, 0}, 2, 10)
	require.NoError(t, err)

	wantExcludedID := vectorIndexDocShortID(t, ctx, db, col, doc1.ID().String())
	for _, id := range results {
		assert.NotEqual(t, wantExcludedID, id)
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

	graph := vectorIndexGraph(t, ctx, db, col)

	// The updated doc is now the nearest match to its new vector.
	results, err := graph.Search([]float32{0, 1, 0}, 1, 10)
	require.NoError(t, err)
	require.Len(t, results, 1)
	wantID := vectorIndexDocShortID(t, ctx, db, col, doc.ID().String())
	assert.Equal(t, wantID, results[0])

	// Searching near the old vector now finds "stationary" as the nearest match: "moving"'s old
	// entry was replaced in place (its node id now holds the new vector), not left behind as a
	// stale duplicate at the old location.
	oldResults, err := graph.Search([]float32{1, 0, 0}, 1, 10)
	require.NoError(t, err)
	require.Len(t, oldResults, 1)
	stationaryID := vectorIndexDocShortID(t, ctx, db, col, stationary.ID().String())
	assert.Equal(t, stationaryID, oldResults[0])
}

func TestCollectionVectorIndex_Save_DimensionMismatch_ReturnsTypedError(t *testing.T) {
	ctx, _, col := newVectorIndexTestDB(t, 3)

	doc, err := client.NewDocFromJSON(ctx, []byte(`{"name": "bad", "embedding": [1, 0]}`), col.Version())
	require.NoError(t, err)

	err = col.SaveDocument(ctx, doc)
	require.Error(t, err)
	assert.ErrorContains(t, err, "vector dimension mismatch")
}
