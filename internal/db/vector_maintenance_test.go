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

const bigVectorSDL = `
	type Users {
		name: String
		embedding: [Float32!] @index(vector: {dimensions: 128, maintenance: DEFERRED, hnsw: {metric: COSINE}})
	}
`

// deferredIndexFixture is a DB with a Users collection that has one deferred vector index.
type deferredIndexFixture struct {
	t   *testing.T
	ctx context.Context
	db  *DB
	col client.Collection

	// pendingReadCalls counts how often the hook set by onPendingRead has fired.
	pendingReadCalls int
}

// newDeferredFixture opens a DB (in memory when path is empty) and defines the collection if it
// does not exist yet. The background writer is not started when suppressRun is set.
func newDeferredFixture(t *testing.T, path string, suppressRun bool) *deferredIndexFixture {
	t.Helper()
	opts := badgerds.DefaultOptions(path)
	if path == "" {
		opts = opts.WithInMemory(true)
	}
	return newDeferredFixtureWith(t, opts, deferredVectorSDL, suppressRun)
}

// newDeferredFixtureWith is newDeferredFixture with custom Badger options and schema.
func newDeferredFixtureWith(
	t *testing.T, opts badgerds.Options, sdl string, suppressRun bool,
) *deferredIndexFixture {
	t.Helper()
	ctx := context.Background()
	rootstore, err := badger.NewDatastore(opts.Dir, opts)
	require.NoError(t, err)
	adminInfo, err := acpDB.NewNACInfo(ctx, "", false)
	require.NoError(t, err)

	setForTest(t, &suppressVectorMaintenanceRun, suppressRun)
	db, err := newDB(ctx, rootstore, adminInfo)
	require.NoError(t, err)
	t.Cleanup(db.Close)

	f := &deferredIndexFixture{t: t, ctx: ctx, db: db}
	col, err := db.GetCollectionByName(ctx, "Users")
	if err != nil {
		_, err = db.AddCollection(ctx, sdl)
		require.NoError(t, err)
		col = f.fetch()
	}
	f.col = col
	return f
}

// fetch resolves a fresh handle of the collection without replacing the fixture's one.
func (f *deferredIndexFixture) fetch() client.Collection {
	f.t.Helper()
	col, err := f.db.GetCollectionByName(f.ctx, "Users")
	require.NoError(f.t, err)
	return col
}

// reload replaces the collection handle with a fresh one. A handle keeps the index state it saw
// when it was resolved, so tests reload after the index state changes.
func (f *deferredIndexFixture) reload() {
	f.t.Helper()
	f.col = f.fetch()
}

func (f *deferredIndexFixture) add(name string, vec []float32) *client.Document {
	f.t.Helper()
	return f.addIn(f.ctx, name, vec)
}

// addIn saves a document using the transaction carried by ctx.
func (f *deferredIndexFixture) addIn(ctx context.Context, name string, vec []float32) *client.Document {
	f.t.Helper()
	data, err := json.Marshal(map[string]any{"name": name, "embedding": vec})
	require.NoError(f.t, err)
	doc, err := client.NewDocFromJSON(ctx, data, f.col.Version())
	require.NoError(f.t, err)
	require.NoError(f.t, f.col.SaveDocument(ctx, doc))
	return doc
}

func (f *deferredIndexFixture) update(doc *client.Document, vec []float32) {
	f.t.Helper()
	require.NoError(f.t, doc.Set(f.ctx, "embedding", vec))
	require.NoError(f.t, f.col.UpdateDocument(f.ctx, doc))
}

func (f *deferredIndexFixture) delete(doc *client.Document) {
	f.t.Helper()
	_, err := f.col.DeleteDocument(f.ctx, doc.ID())
	require.NoError(f.t, err)
}

// drain applies every pending entry, like the background writer would.
func (f *deferredIndexFixture) drain() {
	f.t.Helper()
	require.NoError(f.t, f.db.vectorMaintenance.drainSync(f.ctx))
}

// read runs fn in a read transaction that is discarded right after. Open read transactions would
// block later truncates and index drops, so they must not outlive the call.
func (f *deferredIndexFixture) read(fn func(ctx context.Context)) {
	f.t.Helper()
	txn, err := f.db.NewTxn(true)
	require.NoError(f.t, err)
	defer txn.Discard()
	fn(InitContext(f.ctx, txn))
}

// resolvedIndex is what a read of the collection's only index needs to know about it.
type resolvedIndex struct {
	shortID uint32
	indexID uint32
	epoch   uint32
	vector  *client.VectorIndexDescription
}

// resolve looks up the ids, current epoch and vector settings of the collection's only index.
func (f *deferredIndexFixture) resolve(ctx context.Context) resolvedIndex {
	f.t.Helper()
	version := f.col.Version()
	shortID, err := id.GetCollectionShortID(ctx, version.CollectionID)
	require.NoError(f.t, err)
	desc := version.Indexes[0]
	vd, ok := desc.GetVector()
	require.True(f.t, ok)
	epoch, err := getIndexEpoch(ctx, version.CollectionID, desc.ID)
	require.NoError(f.t, err)
	return resolvedIndex{shortID: shortID, indexID: desc.ID, epoch: epoch, vector: vd}
}

// pending returns the number of pending entries in the index's current epoch.
func (f *deferredIndexFixture) pending() (n int) {
	f.t.Helper()
	f.read(func(ctx context.Context) {
		ix := f.resolve(ctx)
		n = len(f.listPending(ctx, ix, ix.epoch))
	})
	return n
}

// pendingAt returns the pending entries of the index in the given epoch.
func (f *deferredIndexFixture) pendingAt(epoch uint32) (entries []vectorindex.PendingEntry) {
	f.t.Helper()
	f.read(func(ctx context.Context) {
		entries = f.listPending(ctx, f.resolve(ctx), epoch)
	})
	return entries
}

func (f *deferredIndexFixture) listPending(
	ctx context.Context, ix resolvedIndex, epoch uint32,
) []vectorindex.PendingEntry {
	f.t.Helper()
	entries, err := vectorindex.ListPending(ctx, ix.shortID, ix.indexID, epoch, 0)
	require.NoError(f.t, err)
	return entries
}

// search returns the ids of the k nearest documents, pending entries included.
func (f *deferredIndexFixture) search(query []float32, k int) []string {
	f.t.Helper()
	return vectorIndexSearch(f.t, f.ctx, f.db, f.col, query, k)
}

// searchIn is search inside the transaction carried by ctx, with at most 10 hits.
func (f *deferredIndexFixture) searchIn(ctx context.Context, query []float32) []string {
	f.t.Helper()
	ix := f.resolve(ctx)
	results, err := vectorindex.Search(ctx, ix.shortID, ix.indexID, ix.epoch, *ix.vector, query, 10)
	require.NoError(f.t, err)
	docIDs := make([]string, len(results))
	for i, r := range results {
		docIDs[i] = r.DocID
	}
	return docIDs
}

// graphHits counts the hits (at most 10) from the index graph alone, ignoring pending entries.
func (f *deferredIndexFixture) graphHits(query []float32) (n int) {
	f.t.Helper()
	f.read(func(ctx context.Context) {
		ix := f.resolve(ctx)
		index, err := vectorindex.Open(ctx, ix.shortID, ix.indexID, ix.epoch, *ix.vector)
		require.NoError(f.t, err)
		hits, err := index.Search(query, 10)
		require.NoError(f.t, err)
		n = len(hits)
	})
	return n
}

func (f *deferredIndexFixture) ref() vectorIndexRef {
	version := f.col.Version()
	return vectorIndexRef{version.CollectionID, version.Indexes[0].ID}
}

// dirty returns how many indexes the writer has marked as needing a drain.
func (f *deferredIndexFixture) dirty() int {
	w := f.db.vectorMaintenance
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.dirty)
}

// hasCounter and hasBatchSize take the ref explicitly because the collection handle no longer lists
// an index once it is dropped.
func (f *deferredIndexFixture) hasCounter(ref vectorIndexRef) bool {
	_, ok := f.db.vectorMaintenance.pending.Load(ref)
	return ok
}

func (f *deferredIndexFixture) hasBatchSize(ref vectorIndexRef) bool {
	w := f.db.vectorMaintenance
	w.mu.Lock()
	defer w.mu.Unlock()
	_, ok := w.batchSizes[ref]
	return ok
}

// onPendingRead runs fn each time a drain batch has read the pending list, before it commits. fn
// gets the 1-based call number. The hook is removed when the test ends.
func (f *deferredIndexFixture) onPendingRead(fn func(call int)) {
	f.pendingReadCalls = 0
	afterPendingRead = func() {
		f.pendingReadCalls++
		fn(f.pendingReadCalls)
	}
	f.t.Cleanup(func() { afterPendingRead = nil })
}

// ids returns the string ids of the documents, in order.
func ids(docs ...*client.Document) []string {
	out := make([]string, len(docs))
	for i, doc := range docs {
		out[i] = doc.ID().String()
	}
	return out
}

func TestVectorMaintenance_Drain_AppliesUpsertsAndDeletesAndKeepsResults(t *testing.T) {
	f := newDeferredFixture(t, "", true)
	setForTest(t, &deferredBatchSize, 2) // several batches

	docs := make([]*client.Document, 6)
	for i := range docs {
		docs[i] = f.add(fmt.Sprint("d", i), []float32{1, float32(i) / 4, 0})
	}
	// Move one document and delete another, so the list holds an overwrite and a delete.
	f.update(docs[0], []float32{0, 0, 1})
	f.delete(docs[1])

	require.Equal(t, 6, f.pending(), "one coalesced entry per document")
	query := []float32{1, 0, 0}
	before := f.search(query, 10)
	require.Len(t, before, 5)
	assert.NotContains(t, before, docs[1].ID().String())

	f.drain()

	assert.Equal(t, 0, f.pending())
	assert.Equal(t, before, f.search(query, 10))
}

func TestVectorMaintenance_Drain_OverwrittenEntryIsNotLost(t *testing.T) {
	f := newDeferredFixture(t, "", true)
	doc := f.add("moving", []float32{1, 0, 0})
	f.add("other", []float32{0.5, 0.5, 0})

	// Overwrite the document's pending entry after the batch has read it and before it commits.
	f.onPendingRead(func(call int) {
		if call == 1 {
			f.update(doc, []float32{0, 0, 1})
		}
	})

	f.drain()

	assert.GreaterOrEqual(t, f.pendingReadCalls, 2, "the conflicting batch must have been retried")
	assert.Equal(t, 0, f.pending())
	// Only the newer vector is in the graph: the doc is nearest to (0,0,1), not (1,0,0).
	require.Equal(t, ids(doc), f.search([]float32{0, 0, 1}, 1))
	// Had the old vector been applied, the doc would be nearest to (1,0,0).
	got := f.search([]float32{1, 0, 0}, 2)
	require.Len(t, got, 2)
	assert.NotEqual(t, doc.ID().String(), got[0], "old vector must not be applied")
}

func TestVectorMaintenance_CrashResume_DrainsPendingOnRestart(t *testing.T) {
	path := t.TempDir()
	f := newDeferredFixture(t, path, true)
	d1 := f.add("a", []float32{1, 0, 0})
	f.add("b", []float32{0, 1, 0})
	require.Equal(t, 2, f.pending())
	f.db.Close()

	// Reopen with the writer running: it must find the durable pending entries by itself.
	f2 := newDeferredFixture(t, path, false)
	require.Eventually(t, func() bool { return f2.pending() == 0 }, 10e9, 10e6)
	require.Equal(t, ids(d1), f2.search([]float32{1, 0, 0}, 1))
}

func TestVectorMaintenance_BackPressure_AppliesInlineWhenPendingIsFull(t *testing.T) {
	f := newDeferredFixture(t, "", true)
	// The pending list holds at most two entries.
	setForTest(t, &vectorMaintenancePendingLimit, 2)

	docs := make([]*client.Document, 5)
	for i := range docs {
		docs[i] = f.add(fmt.Sprint("d", i), []float32{1, float32(i), 0})
	}
	// Two were deferred; the other three went to the graph directly.
	assert.Equal(t, 2, f.pending())
	assert.Len(t, f.search([]float32{1, 0, 0}, 10), 5)

	// Updating a document that still has a pending upsert applies inline and must remove that
	// entry, or the writer would later overwrite the newer vector with the older one.
	f.update(docs[0], []float32{0, 0, 1})
	assert.Equal(t, 1, f.pending())

	f.delete(docs[1])
	assert.Equal(t, 0, f.pending())

	before := f.search([]float32{0, 0, 1}, 10)
	require.Len(t, before, 4)
	assert.Equal(t, docs[0].ID().String(), before[0])
	assert.NotContains(t, before, docs[1].ID().String())

	f.drain()
	assert.Equal(t, before, f.search([]float32{0, 0, 1}, 10))

	// An inline write to a full index still wakes the writer, since the graph it touched may have
	// been changed by a drain at the same time.
	require.Equal(t, 0, f.dirty())
	f.db.vectorMaintenance.pendingCounter(f.ref()).Store(2)
	f.add("inline", []float32{0, 1, 0})
	assert.Equal(t, 0, f.pending(), "the write should have gone to the graph")
	assert.Equal(t, 1, f.dirty())
}

// Badger rejects a transaction that grows past a share of its memtable. With a small memtable and
// big vectors the first batch of 32 does not fit, so the writer must halve it and keep going.
func TestVectorMaintenance_Drain_ShrinksBatchWhenBadgerTxnIsTooBig(t *testing.T) {
	opts := badgerds.DefaultOptions("").WithInMemory(true).
		WithMemTableSize(256 << 10).
		WithValueThreshold(8 << 10)
	f := newDeferredFixtureWith(t, opts, bigVectorSDL, true)

	rng := rand.New(rand.NewSource(1))
	const docCount = 20
	for i := range docCount {
		vec := make([]float32, 128)
		for j := range vec {
			vec[j] = rng.Float32()
		}
		f.add(fmt.Sprintf("d%d", i), vec)
	}
	require.Equal(t, docCount, f.pending())

	f.drain()

	assert.Equal(t, 0, f.pending())
	// The search returns at most 10 hits.
	assert.Equal(t, 10, f.graphHits(make([]float32, 128)))
	assert.Less(t, f.db.vectorMaintenance.batchLimit(f.ref()), deferredBatchSize)
}

func TestVectorMaintenance_Drain_IncrementDuringDrainSurvives(t *testing.T) {
	f := newDeferredFixture(t, "", true)
	f.add("a", []float32{1, 0, 0})
	f.add("b", []float32{0, 1, 0})
	counter := f.db.vectorMaintenance.pendingCounter(f.ref())
	require.Equal(t, int64(2), counter.Load())

	// A write that commits after the drain's scan bumps the counter while the drain is running.
	f.onPendingRead(func(call int) {
		if call == 1 {
			counter.Add(1)
		}
	})

	f.drain()
	assert.Equal(t, int64(1), counter.Load(), "the increment after the scan must not be erased")
}

// With the writer suppressed, a write stays pending. A search in the same transaction sees it, a
// separate transaction does not until the commit.
func TestVectorMaintenance_SearchInTransaction_SeesOwnPendingWrites(t *testing.T) {
	f := newDeferredFixture(t, "", true)
	far := f.add("far", []float32{0, 1, 0})
	query := []float32{1, 0, 0}

	txn, err := f.db.NewTxn(false)
	require.NoError(t, err)
	defer txn.Discard()
	txnCtx := InitContext(f.ctx, txn)

	near := f.addIn(txnCtx, "near", []float32{1, 0, 0})

	assert.Equal(t, ids(near, far), f.searchIn(txnCtx, query))

	// Another transaction only sees the committed document.
	assert.Equal(t, ids(far), f.search(query, 10))

	require.NoError(t, txn.Commit())
	assert.Equal(t, ids(near, far), f.search(query, 10))
	// Both entries are still pending because the writer is suppressed.
	assert.Equal(t, 2, f.pending())
}

// A merge into a deferred index writes pending entries, not graph nodes. The writer applies them
// later.
func TestVectorMaintenance_Merge_WritesPendingEntriesInsteadOfGraphNodes(t *testing.T) {
	source := newDeferredFixture(t, "", true)
	target := newDeferredFixture(t, "", true)

	doc := source.add("merged", []float32{1, 0, 0})
	copyDAGBlocks(t, target.ctx, source.db, target.db, doc.Head())

	err := target.db.executeMerge(target.ctx, target.col.(*collection), event.Merge{
		DocID:        doc.ID().String(),
		Cid:          doc.Head(),
		CollectionID: target.col.CollectionID(),
	})
	require.NoError(t, err)

	query := []float32{1, 0, 0}
	assert.Equal(t, 1, target.pending())
	assert.Equal(t, 0, target.graphHits(query))
	assert.Equal(t, ids(doc), target.search(query, 10))

	target.drain()
	assert.Equal(t, 0, target.pending())
	assert.Equal(t, 1, target.graphHits(query))
}

func TestVectorMaintenance_Startup_SeedsPendingCounterFromStoredEntries(t *testing.T) {
	path := t.TempDir()
	f := newDeferredFixture(t, path, true)
	for i := range 3 {
		f.add(fmt.Sprint("d", i), []float32{1, float32(i), 0})
	}
	f.db.Close()

	f2 := newDeferredFixture(t, path, true)
	w := f2.db.vectorMaintenance
	require.Equal(t, int64(0), w.pendingCounter(f2.ref()).Load())

	require.NoError(t, w.markAllDeferredDirty(f2.ctx))

	assert.Equal(t, int64(3), w.pendingCounter(f2.ref()).Load())
	assert.Equal(t, 1, f2.dirty())
}

func TestVectorMaintenance_Update_WritesOnePendingEntryPerUpdate(t *testing.T) {
	f := newDeferredFixture(t, "", true)
	counter := f.db.vectorMaintenance.pendingCounter(f.ref())
	doc := f.add("a", []float32{1, 0, 0})
	require.Equal(t, int64(1), counter.Load())

	f.update(doc, []float32{0, 1, 0})

	assert.Equal(t, int64(2), counter.Load())
	entries := f.pendingAt(1)
	require.Len(t, entries, 1)
	assert.Equal(t, vectorindex.PendingUpsert, entries[0].Op)
	assert.Equal(t, []float32{0, 1, 0}, entries[0].Vector)
}

func TestVectorMaintenance_Update_RemovingVectorWritesPendingDelete(t *testing.T) {
	f := newDeferredFixture(t, "", true)
	doc := f.add("a", []float32{1, 0, 0})
	f.drain()
	require.Equal(t, ids(doc), f.search([]float32{1, 0, 0}, 10))

	require.NoError(t, doc.Set(f.ctx, "embedding", nil))
	require.NoError(t, f.col.UpdateDocument(f.ctx, doc))

	entries := f.pendingAt(1)
	require.Len(t, entries, 1)
	assert.Equal(t, vectorindex.PendingDelete, entries[0].Op)
	assert.Empty(t, f.search([]float32{1, 0, 0}, 10))
}
