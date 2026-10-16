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
	"math"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
)

// newLifecycleFixture opens a deferred-index DB with neither the vector writer nor the index build
// worker running, so a test drives both. The initial (empty) index build is settled first.
func newLifecycleFixture(t *testing.T) *deferredIndexFixture {
	t.Helper()
	setForTest(t, &suppressIndexWorkerRun, true)
	f := newDeferredFixture(t, "", true)
	f.db.indexBuildWorker.drainSync(f.ctx)
	return f
}

// Regression guard: dropping a deferred index with undrained pending entries leaves no keys behind
// (the whole-index GC covers the pending prefix) and the writer forgets the index instead of looping.
func TestVectorMaintenance_DropIndexWithPending_GCsPendingAndWriterForgets(t *testing.T) {
	f := newLifecycleFixture(t)
	for i := range 3 {
		f.add("d", []float32{1, float32(i), 0})
	}
	require.Equal(t, 3, f.pending())
	require.Equal(t, 1, f.dirty())

	collectionID := f.col.Version().CollectionID
	shortID := getCollectionShortID(t, f.ctx, f.db, collectionID)
	idx := f.col.Version().Indexes[0]
	oldEpoch := readEpoch(t, f.ctx, f.db, collectionID, idx.ID)
	require.Positive(t, countIndexEpochEntries(t, f.ctx, f.db, shortID, idx.ID, oldEpoch))

	ref := f.ref()
	f.db.vectorMaintenance.setBatchLimit(ref, 4)

	deleteIndexSync(t, f.ctx, f.db, f.col, idx.Name)

	assert.Equal(t, 0, countIndexEntries(t, f.ctx, f.db, shortID, idx.ID), "no key may remain under the index prefix")

	// The writer still has the index marked dirty; draining must succeed and forget it.
	f.drain()
	assert.Equal(t, 0, f.dirty())
	assert.Equal(t, 0, countIndexEntries(t, f.ctx, f.db, shortID, idx.ID), "the writer must not write to a dropped index")
	assert.False(t, f.hasCounter(ref), "the writer must drop the counter of a dropped index")
	assert.False(t, f.hasBatchSize(ref), "the writer must drop the batch size of a dropped index")
}

// Regression guard: truncating a collection with pending entries removes them with the rest of the
// index keyspace; a later drain is a no-op and new documents are searchable (pending then drained).
func TestVectorMaintenance_TruncateWithPending_NoErrorLoopAndNewDocsSearchable(t *testing.T) {
	f := newLifecycleFixture(t)
	for i := range 4 {
		f.add("old", []float32{1, float32(i), 0})
	}
	require.Equal(t, 4, f.pending())

	require.NoError(t, f.col.Truncate(f.ctx))

	assert.Equal(t, 0, f.pending())
	f.drain()
	assert.Equal(t, 0, f.dirty())
	assert.Empty(t, f.search([]float32{1, 0, 0}, 10))

	near := f.add("near", []float32{1, 0, 0})
	far := f.add("far", []float32{0, 1, 0})
	want := ids(near, far)
	assert.Equal(t, want, f.search([]float32{1, 0, 0}, 2))
	f.drain()
	assert.Equal(t, 0, f.pending())
	assert.Equal(t, want, f.search([]float32{1, 0, 0}, 2))
}

// Regression guard: a filtered truncate deletes some documents' index entries, which are recorded as
// pending deletes; the drain applies them without needing the removed doc short-id mappings.
func TestVectorMaintenance_FilteredTruncateWithPending_DrainsCleanly(t *testing.T) {
	f := newLifecycleFixture(t)
	a := f.add("a", []float32{1, 0, 0})
	b := f.add("b", []float32{0.9, 0.4, 0})
	c := f.add("c", []float32{0, 1, 0})

	opts := options.TruncateCollection().SetFilter(map[string]any{
		"_docID": map[string]any{"_in": []any{a.ID().String()}},
	})
	require.NoError(t, f.col.Truncate(f.ctx, opts))

	want := ids(b, c)
	assert.Equal(t, want, f.search([]float32{1, 0, 0}, 10))
	f.drain()
	assert.Equal(t, 0, f.dirty())
	assert.Equal(t, 0, f.pending())
	assert.Equal(t, want, f.search([]float32{1, 0, 0}, 10))
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

// nearestFirst returns the ids of the named documents ordered by cosine similarity to the query,
// as a brute-force reference for the index results.
func nearestFirst(vecs map[string][]float32, docs map[string]*client.Document, query []float32) []string {
	names := make([]string, 0, len(vecs))
	for n := range vecs {
		names = append(names, n)
	}
	sort.Slice(names, func(i, j int) bool { return cosine(vecs[names[i]], query) > cosine(vecs[names[j]], query) })
	want := make([]string, len(names))
	for i, n := range names {
		want[i] = docs[n].ID().String()
	}
	return want
}

// Regression guard: with entries pending in epoch 1, a rebuild moves the index to epoch 2. While it
// is building, writes go inline (no pending entry is written), the writer leaves the stale epoch's
// entries alone, and once the build and stale-epoch GC finish the stale pending keys are gone and
// the index matches brute force over the current documents.
func TestVectorMaintenance_RebuildWithPending_StaleEpochNotApplied(t *testing.T) {
	f := newLifecycleFixture(t)
	collectionID := f.col.Version().CollectionID
	shortID := getCollectionShortID(t, f.ctx, f.db, collectionID)
	idx := f.col.Version().Indexes[0]
	require.Equal(t, uint32(1), readEpoch(t, f.ctx, f.db, collectionID, idx.ID))

	vecs := map[string][]float32{
		"a": {1, 0, 0}, "b": {0.9, 0.4, 0}, "c": {0, 1, 0}, "d": {0.5, 0.5, 0.5}, "e": {0, 0, 1},
	}
	docs := map[string]*client.Document{}
	for _, n := range []string{"a", "b", "c", "d", "e"} {
		docs[n] = f.add(n, vecs[n])
	}
	// c moves next to the query and e is deleted, so the stale epoch holds an overwrite and a delete.
	vecs["c"] = []float32{0.95, 0.05, 0}
	f.update(docs["c"], vecs["c"])
	f.delete(docs["e"])
	delete(vecs, "e")
	require.Len(t, f.pendingAt(1), 5)

	stageRebuild(t, f.ctx, f.db, f.fetch().Version())
	require.Equal(t, uint32(2), readEpoch(t, f.ctx, f.db, collectionID, idx.ID))

	// Building: the write is applied inline, and the writer neither applies nor drops the stale entries.
	vecs["f"] = []float32{0.99, 0.01, 0}
	docs["f"] = f.add("f", vecs["f"])
	assert.Empty(t, f.pendingAt(2), "writes during a build must not be deferred")
	f.drain()
	assert.Len(t, f.pendingAt(1), 5, "stale-epoch entries are not applied")
	assert.Empty(t, f.pendingAt(2))

	f.db.indexBuildWorker.drainSync(f.ctx)

	assert.Equal(t, 0, countIndexEpochEntries(t, f.ctx, f.db, shortID, idx.ID, 1), "stale epoch must be collected")
	assert.Empty(t, f.pendingAt(2))
	f.drain()

	query := []float32{1, 0, 0}
	assert.Equal(t, nearestFirst(vecs, docs, query), f.search(query, 10))

	// Once the build is done, writes defer again in the new epoch and drain. A handle resolved during
	// the build keeps its building flag (a build finishing does not invalidate handles), so use a fresh one.
	f.reload()
	g := f.add("g", []float32{1, 0, 0.001})
	assert.Len(t, f.pendingAt(2), 1)
	f.drain()
	assert.Empty(t, f.pendingAt(2))
	assert.Equal(t, ids(docs["a"], g), f.search(query, 2))
}

// A handle resolved during a build keeps building=true for good, so it writes inline. Its inline
// write must still remove the doc's older pending entry, or the writer would later apply that
// entry over the newer vector.
func TestVectorMaintenance_StaleBuildingHandle_InlineWriteRemovesPendingEntry(t *testing.T) {
	f := newLifecycleFixture(t)
	x := f.add("x", []float32{1, 0, 0})
	// z sits between x's two later vectors, so the final search shows which one the graph kept.
	f.add("z", []float32{0.5, 0.5, 0})
	f.drain()

	stageRebuild(t, f.ctx, f.db, f.fetch().Version())
	// Resolved while the build is running, so this handle thinks the index is still building.
	stale := f.fetch()
	f.db.indexBuildWorker.drainSync(f.ctx)
	f.drain()

	// A fresh handle defers the write, leaving an older pending entry for x.
	f.reload()
	f.update(x, []float32{0, 0, 1})
	require.Equal(t, 1, f.pending())

	// The stale handle writes inline with a newer vector and must clear that entry.
	require.NoError(t, x.Set(f.ctx, "embedding", []float32{1, 0, 0.01}))
	require.NoError(t, stale.UpdateDocument(f.ctx, x))
	assert.Equal(t, 0, f.pending(), "inline write must remove the older pending entry")

	// Had the older pending vector (0,0,1) been applied, z would be nearest to (1,0,0).
	f.drain()
	assert.Equal(t, ids(x), f.search([]float32{1, 0, 0}, 1))
}
