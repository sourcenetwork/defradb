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
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"

	badgerds "github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/badger"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/db/vectorindex"
)

// drainSync drains every dirty index until none is left, returning the first error. Test-only: the
// run loop does the same in the background, so tests that call it suppress the run loop.
func (w *vectorMaintenanceWriter) drainSync(ctx context.Context) error {
	w.drainMu.Lock()
	defer w.drainMu.Unlock()
	for {
		w.mu.Lock()
		empty := len(w.dirty) == 0
		w.mu.Unlock()
		if empty {
			return nil
		}
		if _, err := w.drainPassLocked(ctx); err != nil {
			return err
		}
	}
}

const deferredVectorSDL = `
	type Users {
		name: String
		embedding: [Float32!] @index(vector: {dimensions: 3, maintenance: DEFERRED, hnsw: {metric: COSINE}})
	}
`

// openDeferredTestDB opens a DB (in memory when path is empty) with the deferred-index collection
// defined if it does not exist yet. The background writer is not started when suppressRun is set.
func openDeferredTestDB(
	t *testing.T, path string, suppressRun bool,
) (context.Context, *DB, client.Collection) {
	t.Helper()
	opts := badgerds.DefaultOptions(path)
	if path == "" {
		opts = opts.WithInMemory(true)
	}
	return openDeferredTestDBWith(t, opts, deferredVectorSDL, suppressRun)
}

// openDeferredTestDBWith is openDeferredTestDB with custom Badger options and schema.
func openDeferredTestDBWith(
	t *testing.T, opts badgerds.Options, sdl string, suppressRun bool,
) (context.Context, *DB, client.Collection) {
	t.Helper()
	ctx := context.Background()
	path := opts.Dir
	rootstore, err := badger.NewDatastore(path, opts)
	require.NoError(t, err)
	adminInfo, err := acpDB.NewNACInfo(ctx, "", false)
	require.NoError(t, err)

	setForTest(t, &suppressVectorMaintenanceRun, suppressRun)
	db, err := newDB(ctx, rootstore, adminInfo)
	require.NoError(t, err)
	t.Cleanup(db.Close)

	col, err := db.GetCollectionByName(ctx, "Users")
	if err != nil {
		_, err = db.AddCollection(ctx, sdl)
		require.NoError(t, err)
		col, err = db.GetCollectionByName(ctx, "Users")
		require.NoError(t, err)
	}
	return ctx, db, col
}

func addDeferredDoc(t *testing.T, ctx context.Context, col client.Collection, name string, vec []float32) *client.Document {
	t.Helper()
	doc, err := client.NewDocFromJSON(ctx, fmt.Appendf(nil, `{"name": %q, "embedding": %s}`, name, vecJSON(vec)), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.SaveDocument(ctx, doc))
	return doc
}

func vecJSON(vec []float32) string {
	return fmt.Sprint("[", vec[0], ",", vec[1], ",", vec[2], "]")
}

// pendingCount returns the number of pending entries of the collection's only index.
func pendingCount(t *testing.T, ctx context.Context, db *DB, col client.Collection) int {
	t.Helper()
	txn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer txn.Discard()
	readCtx := InitContext(ctx, txn)

	shortID, err := id.GetCollectionShortID(readCtx, col.Version().CollectionID)
	require.NoError(t, err)
	indexID := col.Version().Indexes[0].ID
	epoch, err := getIndexEpoch(readCtx, col.Version().CollectionID, indexID)
	require.NoError(t, err)
	entries, err := vectorindex.ListPending(readCtx, shortID, indexID, epoch, 0)
	require.NoError(t, err)
	return len(entries)
}

func TestVectorMaintenance_Drain_AppliesUpsertsAndDeletesAndKeepsResults(t *testing.T) {
	ctx, db, col := openDeferredTestDB(t, "", true)
	setForTest(t, &deferredBatchSize, 2) // several batches

	docs := make([]*client.Document, 6)
	for i := range docs {
		docs[i] = addDeferredDoc(t, ctx, col, fmt.Sprint("d", i), []float32{1, float32(i) / 4, 0})
	}
	// Move one document and delete another, so the list holds an overwrite and a delete.
	require.NoError(t, docs[0].Set(ctx, "embedding", []float32{0, 0, 1}))
	require.NoError(t, col.UpdateDocument(ctx, docs[0]))
	_, err := col.DeleteDocument(ctx, docs[1].ID())
	require.NoError(t, err)

	require.Equal(t, 6, pendingCount(t, ctx, db, col), "one coalesced entry per document")
	query := []float32{1, 0, 0}
	before := vectorIndexSearch(t, ctx, db, col, query, 10)
	require.Len(t, before, 5)
	assert.NotContains(t, before, docs[1].ID().String())

	require.NoError(t, db.vectorMaintenance.drainSync(ctx))

	assert.Equal(t, 0, pendingCount(t, ctx, db, col))
	assert.Equal(t, before, vectorIndexSearch(t, ctx, db, col, query, 10))
}

func TestVectorMaintenance_Drain_OverwrittenEntryIsNotLost(t *testing.T) {
	ctx, db, col := openDeferredTestDB(t, "", true)
	doc := addDeferredDoc(t, ctx, col, "moving", []float32{1, 0, 0})
	addDeferredDoc(t, ctx, col, "other", []float32{0.5, 0.5, 0})

	// Overwrite the document's pending entry after the batch has read it and before it commits.
	calls := 0
	afterPendingRead = func() {
		calls++
		if calls != 1 {
			return
		}
		require.NoError(t, doc.Set(ctx, "embedding", []float32{0, 0, 1}))
		require.NoError(t, col.UpdateDocument(ctx, doc))
	}
	t.Cleanup(func() { afterPendingRead = nil })

	require.NoError(t, db.vectorMaintenance.drainSync(ctx))

	assert.GreaterOrEqual(t, calls, 2, "the conflicting batch must have been retried")
	assert.Equal(t, 0, pendingCount(t, ctx, db, col))
	// Only the newer vector is in the graph: the doc is nearest to (0,0,1), not (1,0,0).
	got := vectorIndexSearch(t, ctx, db, col, []float32{0, 0, 1}, 1)
	require.Equal(t, []string{doc.ID().String()}, got)
	// Had the old vector been applied, the doc would be nearest to (1,0,0).
	got = vectorIndexSearch(t, ctx, db, col, []float32{1, 0, 0}, 2)
	require.Len(t, got, 2)
	assert.NotEqual(t, doc.ID().String(), got[0], "old vector must not be applied")
}

func TestVectorMaintenance_CrashResume_DrainsPendingOnRestart(t *testing.T) {
	path := t.TempDir()
	ctx, db, col := openDeferredTestDB(t, path, true)
	d1 := addDeferredDoc(t, ctx, col, "a", []float32{1, 0, 0})
	addDeferredDoc(t, ctx, col, "b", []float32{0, 1, 0})
	require.Equal(t, 2, pendingCount(t, ctx, db, col))
	db.Close()

	// Reopen with the writer running: it must find the durable pending entries by itself.
	ctx, db2, col2 := openDeferredTestDB(t, path, false)
	require.Eventually(t, func() bool {
		return pendingCount(t, ctx, db2, col2) == 0
	}, 10e9, 10e6)
	got := vectorIndexSearch(t, ctx, db2, col2, []float32{1, 0, 0}, 1)
	require.Equal(t, []string{d1.ID().String()}, got)
}

func TestVectorMaintenance_BackPressure_AppliesInlineWhenPendingIsFull(t *testing.T) {
	ctx, db, col := openDeferredTestDB(t, "", true)
	// Dimensions 3 is 12 bytes a vector, so 24 bytes is two entries.
	setForTest(t, &vectorMaintenancePendingLimit, 2)

	docs := make([]*client.Document, 5)
	for i := range docs {
		docs[i] = addDeferredDoc(t, ctx, col, fmt.Sprint("d", i), []float32{1, float32(i), 0})
	}
	// Two were deferred; the other three went to the graph directly.
	assert.Equal(t, 2, pendingCount(t, ctx, db, col))
	assert.Len(t, vectorIndexSearch(t, ctx, db, col, []float32{1, 0, 0}, 10), 5)

	// Updating a document that still has a pending upsert applies inline and must remove that
	// entry, or the writer would later overwrite the newer vector with the older one.
	require.NoError(t, docs[0].Set(ctx, "embedding", []float32{0, 0, 1}))
	require.NoError(t, col.UpdateDocument(ctx, docs[0]))
	assert.Equal(t, 1, pendingCount(t, ctx, db, col))

	_, err := col.DeleteDocument(ctx, docs[1].ID())
	require.NoError(t, err)
	assert.Equal(t, 0, pendingCount(t, ctx, db, col))

	before := vectorIndexSearch(t, ctx, db, col, []float32{0, 0, 1}, 10)
	require.Len(t, before, 4)
	assert.Equal(t, docs[0].ID().String(), before[0])
	assert.NotContains(t, before, docs[1].ID().String())

	require.NoError(t, db.vectorMaintenance.drainSync(ctx))
	assert.Equal(t, before, vectorIndexSearch(t, ctx, db, col, []float32{0, 0, 1}, 10))
}

const bigVectorSDL = `
	type Users {
		name: String
		embedding: [Float32!] @index(vector: {dimensions: 128, maintenance: DEFERRED, hnsw: {metric: COSINE}})
	}
`

// Badger rejects a transaction that grows past a share of its memtable. With a small memtable and
// big vectors the first batch of 32 does not fit, so the writer must halve it and keep going.
func TestVectorMaintenance_Drain_ShrinksBatchWhenBadgerTxnIsTooBig(t *testing.T) {
	opts := badgerds.DefaultOptions("").WithInMemory(true).
		WithMemTableSize(256 << 10).
		WithValueThreshold(8 << 10)
	ctx, db, col := openDeferredTestDBWith(t, opts, bigVectorSDL, true)

	rng := rand.New(rand.NewSource(1))
	const docCount = 20
	for i := range docCount {
		vec := make([]float32, 128)
		for j := range vec {
			vec[j] = rng.Float32()
		}
		doc, err := client.NewDocFromJSON(ctx,
			fmt.Appendf(nil, `{"name": "d%d", "embedding": %s}`, i, mustJSON(t, vec)), col.Version())
		require.NoError(t, err)
		require.NoError(t, col.SaveDocument(ctx, doc))
	}
	require.Equal(t, docCount, pendingCount(t, ctx, db, col))

	require.NoError(t, db.vectorMaintenance.drainSync(ctx))

	assert.Equal(t, 0, pendingCount(t, ctx, db, col))
	// The search returns at most 10 hits.
	assert.Equal(t, 10, graphOnlySearch(t, ctx, db, col, make([]float32, 128)))
	key := vectorIndexRef{col.Version().CollectionID, col.Version().Indexes[0].ID}
	assert.Less(t, db.vectorMaintenance.batchLimit(key), deferredBatchSize)
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return string(b)
}

func TestVectorMaintenance_Drain_IncrementDuringDrainSurvives(t *testing.T) {
	ctx, db, col := openDeferredTestDB(t, "", true)
	addDeferredDoc(t, ctx, col, "a", []float32{1, 0, 0})
	addDeferredDoc(t, ctx, col, "b", []float32{0, 1, 0})
	key := vectorIndexRef{col.Version().CollectionID, col.Version().Indexes[0].ID}
	counter := db.vectorMaintenance.pendingCounter(key)
	require.Equal(t, int64(2), counter.Load())

	// A write that commits after the drain's scan bumps the counter while the drain is running.
	calls := 0
	afterPendingRead = func() {
		calls++
		if calls == 1 {
			counter.Add(1)
		}
	}
	t.Cleanup(func() { afterPendingRead = nil })

	require.NoError(t, db.vectorMaintenance.drainSync(ctx))
	assert.Equal(t, int64(1), counter.Load(), "the increment after the scan must not be erased")
}

// graphOnlySearch searches the index graph without the pending list overlay.
func graphOnlySearch(t *testing.T, ctx context.Context, db *DB, col client.Collection, query []float32) int {
	t.Helper()
	txn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer txn.Discard()
	readCtx := InitContext(ctx, txn)

	shortID, err := id.GetCollectionShortID(readCtx, col.Version().CollectionID)
	require.NoError(t, err)
	desc := col.Version().Indexes[0]
	vd, ok := desc.GetVector()
	require.True(t, ok)
	epoch, err := getIndexEpoch(readCtx, col.Version().CollectionID, desc.ID)
	require.NoError(t, err)
	index, err := vectorindex.Open(readCtx, shortID, desc.ID, epoch, *vd)
	require.NoError(t, err)
	hits, err := index.Search(query, 10)
	require.NoError(t, err)
	return len(hits)
}

// searchIn searches the deferred index inside the transaction carried by ctx.
func searchIn(t *testing.T, ctx context.Context, col client.Collection, query []float32) []string {
	t.Helper()
	shortID, err := id.GetCollectionShortID(ctx, col.Version().CollectionID)
	require.NoError(t, err)
	desc := col.Version().Indexes[0]
	vd, ok := desc.GetVector()
	require.True(t, ok)
	epoch, err := getIndexEpoch(ctx, col.Version().CollectionID, desc.ID)
	require.NoError(t, err)
	results, err := vectorindex.Search(ctx, shortID, desc.ID, epoch, *vd, query, 10)
	require.NoError(t, err)
	docIDs := make([]string, len(results))
	for i, r := range results {
		docIDs[i] = r.DocID
	}
	return docIDs
}

// With the writer suppressed, a write stays pending. A search in the same transaction sees it, a
// separate transaction does not until the commit.
func TestVectorMaintenance_SearchInTransaction_SeesOwnPendingWrites(t *testing.T) {
	ctx, db, col := openDeferredTestDB(t, "", true)
	far := addDeferredDoc(t, ctx, col, "far", []float32{0, 1, 0})
	query := []float32{1, 0, 0}

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	defer txn.Discard()
	txnCtx := InitContext(ctx, txn)

	near, err := client.NewDocFromJSON(txnCtx, []byte(`{"name": "near", "embedding": [1, 0, 0]}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.SaveDocument(txnCtx, near))

	assert.Equal(t, []string{near.ID().String(), far.ID().String()}, searchIn(t, txnCtx, col, query))

	// Another transaction only sees the committed document.
	assert.Equal(t, []string{far.ID().String()}, vectorIndexSearch(t, ctx, db, col, query, 10))

	require.NoError(t, txn.Commit())
	assert.Equal(t, []string{near.ID().String(), far.ID().String()}, vectorIndexSearch(t, ctx, db, col, query, 10))
	// Both entries are still pending because the writer is suppressed.
	assert.Equal(t, 2, pendingCount(t, ctx, db, col))
}

// A merge into a deferred index writes pending entries, not graph nodes. The writer applies them
// later.
func TestVectorMaintenance_Merge_WritesPendingEntriesInsteadOfGraphNodes(t *testing.T) {
	sourceCtx, sourceDB, sourceCol := openDeferredTestDB(t, "", true)
	ctx, targetDB, targetCol := openDeferredTestDB(t, "", true)

	doc := addDeferredDoc(t, sourceCtx, sourceCol, "merged", []float32{1, 0, 0})
	copyDAGBlocks(t, ctx, sourceDB, targetDB, doc.Head())

	err := targetDB.executeMerge(ctx, targetCol.(*collection), event.Merge{
		DocID:        doc.ID().String(),
		Cid:          doc.Head(),
		CollectionID: targetCol.CollectionID(),
	})
	require.NoError(t, err)

	query := []float32{1, 0, 0}
	assert.Equal(t, 1, pendingCount(t, ctx, targetDB, targetCol))
	assert.Equal(t, 0, graphOnlySearch(t, ctx, targetDB, targetCol, query))
	assert.Equal(t, []string{doc.ID().String()}, vectorIndexSearch(t, ctx, targetDB, targetCol, query, 10))

	require.NoError(t, targetDB.vectorMaintenance.drainSync(ctx))
	assert.Equal(t, 0, pendingCount(t, ctx, targetDB, targetCol))
	assert.Equal(t, 1, graphOnlySearch(t, ctx, targetDB, targetCol, query))
}
