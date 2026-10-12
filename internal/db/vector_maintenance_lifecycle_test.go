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
	"math"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/db/vectorindex"
)

// openLifecycleTestDB opens a deferred-index DB with neither the vector writer nor the index build
// worker running, so a test drives both and the initial (empty) index build is settled first.
func openLifecycleTestDB(t *testing.T) (context.Context, *DB, client.Collection) {
	t.Helper()
	setForTest(t, &suppressIndexWorkerRun, true)
	ctx, db, col := openDeferredTestDB(t, "", true)
	db.indexBuildWorker.drainSync(ctx)
	return ctx, db, col
}

// pendingKeys returns the pending entries of the collection's index in the given epoch.
func pendingAt(t *testing.T, ctx context.Context, db *DB, col client.Collection, epoch uint32) []vectorindex.PendingEntry {
	t.Helper()
	txn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer txn.Discard()
	readCtx := InitContext(ctx, txn)
	shortID, err := id.GetCollectionShortID(readCtx, col.Version().CollectionID)
	require.NoError(t, err)
	entries, err := vectorindex.ListPending(readCtx, shortID, col.Version().Indexes[0].ID, epoch, 0)
	require.NoError(t, err)
	return entries
}

func dirtyCount(w *vectorMaintenanceWriter) int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.dirty)
}

// Regression guard: dropping a deferred index with undrained pending entries leaves no keys behind
// (the whole-index GC covers the pending prefix) and the writer forgets the index instead of looping.
func TestVectorMaintenance_DropIndexWithPending_GCsPendingAndWriterForgets(t *testing.T) {
	ctx, db, col := openLifecycleTestDB(t)
	for i := range 3 {
		addDeferredDoc(t, ctx, col, "d", []float32{1, float32(i), 0})
	}
	require.Equal(t, 3, pendingCount(t, ctx, db, col))
	require.Equal(t, 1, dirtyCount(db.vectorMaintenance))

	collectionID := col.Version().CollectionID
	shortID := getCollectionShortID(t, ctx, db, collectionID)
	idx := col.Version().Indexes[0]
	oldEpoch := readEpoch(t, ctx, db, collectionID, idx.ID)
	require.Positive(t, countIndexEpochEntries(t, ctx, db, shortID, idx.ID, oldEpoch))

	db.vectorMaintenance.setBatchLimit(vectorIndexRef{collectionID, idx.ID}, 4)

	deleteIndexSync(t, ctx, db, col, idx.Name)

	assert.Equal(t, 0, countIndexEntries(t, ctx, db, shortID, idx.ID), "no key may remain under the index prefix")

	// The writer still has the index marked dirty; draining must succeed and forget it.
	require.NoError(t, db.vectorMaintenance.drainSync(ctx))
	assert.Equal(t, 0, dirtyCount(db.vectorMaintenance))
	assert.Equal(t, 0, countIndexEntries(t, ctx, db, shortID, idx.ID), "the writer must not write to a dropped index")
	_, hasCounter := db.vectorMaintenance.pending.Load(vectorIndexRef{col.Version().CollectionID, idx.ID})
	assert.False(t, hasCounter, "the writer must drop the counter of a dropped index")
	db.vectorMaintenance.mu.Lock()
	_, hasBatchSize := db.vectorMaintenance.batchSizes[vectorIndexRef{col.Version().CollectionID, idx.ID}]
	db.vectorMaintenance.mu.Unlock()
	assert.False(t, hasBatchSize, "the writer must drop the batch size of a dropped index")
}

// Regression guard: truncating a collection with pending entries removes them with the rest of the
// index keyspace; a later drain is a no-op and new documents are searchable (pending then drained).
func TestVectorMaintenance_TruncateWithPending_NoErrorLoopAndNewDocsSearchable(t *testing.T) {
	ctx, db, col := openLifecycleTestDB(t)
	for i := range 4 {
		addDeferredDoc(t, ctx, col, "old", []float32{1, float32(i), 0})
	}
	require.Equal(t, 4, pendingCount(t, ctx, db, col))

	require.NoError(t, col.Truncate(ctx))

	assert.Equal(t, 0, pendingCount(t, ctx, db, col))
	require.NoError(t, db.vectorMaintenance.drainSync(ctx))
	assert.Equal(t, 0, dirtyCount(db.vectorMaintenance))
	assert.Empty(t, vectorIndexSearch(t, ctx, db, col, []float32{1, 0, 0}, 10))

	near := addDeferredDoc(t, ctx, col, "near", []float32{1, 0, 0})
	far := addDeferredDoc(t, ctx, col, "far", []float32{0, 1, 0})
	want := []string{near.ID().String(), far.ID().String()}
	assert.Equal(t, want, vectorIndexSearch(t, ctx, db, col, []float32{1, 0, 0}, 2))
	require.NoError(t, db.vectorMaintenance.drainSync(ctx))
	assert.Equal(t, 0, pendingCount(t, ctx, db, col))
	assert.Equal(t, want, vectorIndexSearch(t, ctx, db, col, []float32{1, 0, 0}, 2))
}

// Regression guard: a filtered truncate deletes some documents' index entries, which are recorded as
// pending deletes; the drain applies them without needing the removed doc short-id mappings.
func TestVectorMaintenance_FilteredTruncateWithPending_DrainsCleanly(t *testing.T) {
	ctx, db, col := openLifecycleTestDB(t)
	a := addDeferredDoc(t, ctx, col, "a", []float32{1, 0, 0})
	b := addDeferredDoc(t, ctx, col, "b", []float32{0.9, 0.4, 0})
	c := addDeferredDoc(t, ctx, col, "c", []float32{0, 1, 0})

	opts := options.TruncateCollection().SetFilter(map[string]any{
		"_docID": map[string]any{"_in": []any{a.ID().String()}},
	})
	require.NoError(t, col.Truncate(ctx, opts))

	want := []string{b.ID().String(), c.ID().String()}
	assert.Equal(t, want, vectorIndexSearch(t, ctx, db, col, []float32{1, 0, 0}, 10))
	require.NoError(t, db.vectorMaintenance.drainSync(ctx))
	assert.Equal(t, 0, dirtyCount(db.vectorMaintenance))
	assert.Equal(t, 0, pendingCount(t, ctx, db, col))
	assert.Equal(t, want, vectorIndexSearch(t, ctx, db, col, []float32{1, 0, 0}, 10))
}

func cosine(a, b []float32) float64 {
	var dot, na, nb float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		na += float64(a[i]) * float64(a[i])
		nb += float64(b[i]) * float64(b[i])
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

// Regression guard: with entries pending in epoch 1, a rebuild moves the index to epoch 2. While it is
// building, writes go inline (no pending entry is written), the writer leaves the stale epoch's entries
// alone, and once the build and stale-epoch GC finish the stale pending keys are gone and the index
// matches brute force over the current documents.
func TestVectorMaintenance_RebuildWithPending_StaleEpochNotApplied(t *testing.T) {
	ctx, db, col := openLifecycleTestDB(t)
	collectionID := col.Version().CollectionID
	shortID := getCollectionShortID(t, ctx, db, collectionID)
	idx := col.Version().Indexes[0]
	require.Equal(t, uint32(1), readEpoch(t, ctx, db, collectionID, idx.ID))

	vecs := map[string][]float32{
		"a": {1, 0, 0}, "b": {0.9, 0.4, 0}, "c": {0, 1, 0}, "d": {0.5, 0.5, 0.5}, "e": {0, 0, 1},
	}
	docs := map[string]*client.Document{}
	for _, n := range []string{"a", "b", "c", "d", "e"} {
		docs[n] = addDeferredDoc(t, ctx, col, n, vecs[n])
	}
	// c moves next to the query and e is deleted, so the stale epoch holds an overwrite and a delete.
	vecs["c"] = []float32{0.95, 0.05, 0}
	require.NoError(t, docs["c"].Set(ctx, "embedding", vecs["c"]))
	require.NoError(t, col.UpdateDocument(ctx, docs["c"]))
	_, err := col.DeleteDocument(ctx, docs["e"].ID())
	require.NoError(t, err)
	delete(vecs, "e")
	require.Len(t, pendingAt(t, ctx, db, col, 1), 5)

	def, err := db.GetCollectionByName(ctx, "Users")
	require.NoError(t, err)
	stageRebuild(t, ctx, db, def.Version())
	require.Equal(t, uint32(2), readEpoch(t, ctx, db, collectionID, idx.ID))

	// Building: the write is applied inline, and the writer neither applies nor drops the stale entries.
	vecs["f"] = []float32{0.99, 0.01, 0}
	docs["f"] = addDeferredDoc(t, ctx, col, "f", vecs["f"])
	assert.Empty(t, pendingAt(t, ctx, db, col, 2), "writes during a build must not be deferred")
	require.NoError(t, db.vectorMaintenance.drainSync(ctx))
	assert.Len(t, pendingAt(t, ctx, db, col, 1), 5, "stale-epoch entries are not applied")
	assert.Empty(t, pendingAt(t, ctx, db, col, 2))

	db.indexBuildWorker.drainSync(ctx)

	assert.Equal(t, 0, countIndexEpochEntries(t, ctx, db, shortID, idx.ID, 1), "stale epoch must be collected")
	assert.Empty(t, pendingAt(t, ctx, db, col, 2))
	require.NoError(t, db.vectorMaintenance.drainSync(ctx))

	query := []float32{1, 0, 0}
	names := make([]string, 0, len(vecs))
	for n := range vecs {
		names = append(names, n)
	}
	sort.Slice(names, func(i, j int) bool { return cosine(vecs[names[i]], query) > cosine(vecs[names[j]], query) })
	want := make([]string, len(names))
	for i, n := range names {
		want[i] = docs[n].ID().String()
	}
	assert.Equal(t, want, vectorIndexSearch(t, ctx, db, col, query, 10))

	// Once the build is done, writes defer again in the new epoch and drain. A handle resolved during
	// the build keeps its building flag (a build finishing does not invalidate handles), so use a fresh one.
	col, err = db.GetCollectionByName(ctx, "Users")
	require.NoError(t, err)
	g := addDeferredDoc(t, ctx, col, "g", []float32{1, 0, 0.001})
	assert.Len(t, pendingAt(t, ctx, db, col, 2), 1)
	require.NoError(t, db.vectorMaintenance.drainSync(ctx))
	assert.Empty(t, pendingAt(t, ctx, db, col, 2))
	got := vectorIndexSearch(t, ctx, db, col, query, 2)
	assert.Equal(t, []string{docs["a"].ID().String(), g.ID().String()}, got)
}
