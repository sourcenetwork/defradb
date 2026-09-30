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
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// waitForIndexReady blocks until the index has no build/drop state record (ready), or fails the test.
func waitForIndexReady(t *testing.T, ctx context.Context, db *DB, collectionID string, indexID uint32) {
	t.Helper()
	require.Eventually(t, func() bool {
		rawTxn, err := db.NewTxn(true)
		if err != nil {
			return false
		}
		defer rawTxn.Discard()
		_, err = getIndexState(InitContext(ctx, rawTxn), collectionID, indexID)
		return err != nil // a missing record means ready
	}, 20*time.Second, 5*time.Millisecond, "index %d did not become ready", indexID)
}

// TestConcurrentBuilds_SameCollection_BothComplete is the regression for the conflict-wedge bug:
// two backfills on the same collection used to conflict (each batch read the other's watermark
// record via newCollection's index-state scan and wrote its own), leaving one stuck building
// forever. The gate forces the two builds to overlap; both must still converge to ready with full
// entries.
func TestConcurrentBuilds_SameCollection_BothComplete(t *testing.T) {
	setForTest(t, &indexBackfillBatchSize, 5)
	setForTest(t, &indexBuildConcurrency, 4)
	setForTest(t, &indexBuildRetryDelay, 10*time.Millisecond)

	// Gate: park each build at its first batch boundary until released, so both are in flight at
	// once. entered fires once per distinct index (tracked in seen), so a build re-entering the gate
	// across batches cannot over-signal the WaitGroup regardless of batch count.
	origGate := IndexBuildGate
	defer func() { IndexBuildGate = origGate }()
	release := make(chan struct{})
	var releaseOnce sync.Once
	var entered sync.WaitGroup
	entered.Add(2)
	var seen sync.Map
	IndexBuildGate = func(gateCtx context.Context, _ string, indexID uint32) {
		select {
		case <-release:
			return // already released; let later batches run freely
		default:
		}
		if _, dup := seen.LoadOrStore(indexID, struct{}{}); !dup {
			entered.Done()
		}
		// Honour ctx so Close (e.g. on a failed assertion) can unblock a parked build.
		select {
		case <-release:
		case <-gateCtx.Done():
		}
	}

	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.AddCollection(ctx, `type User { name: String, age: Int }`)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	const docCount = 40
	for i := range docCount {
		doc, err := client.NewDocFromJSON(ctx,
			fmt.Appendf(nil, `{"name":"n%02d","age":%d}`, i, i), col.Version())
		require.NoError(t, err)
		require.NoError(t, col.AddDocument(ctx, doc))
	}

	nameIdx, err := col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
	})
	require.NoError(t, err)
	ageIdx, err := col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "age"}},
	})
	require.NoError(t, err)

	// Both builds are parked at the gate (mid-build at once); release them.
	entered.Wait()
	releaseOnce.Do(func() { close(release) })

	collectionID := col.Version().CollectionID
	waitForIndexReady(t, ctx, db, collectionID, nameIdx.ID)
	waitForIndexReady(t, ctx, db, collectionID, ageIdx.ID)

	shortID := getCollectionShortID(t, ctx, db, collectionID)
	assert.Equal(t, docCount, countIndexEntries(t, ctx, db, shortID, nameIdx.ID),
		"name index must have one entry per doc")
	assert.Equal(t, docCount, countIndexEntries(t, ctx, db, shortID, ageIdx.ID),
		"age index must have one entry per doc")
}

// TestConcurrentBuilds_ManyIndexes_AllComplete stresses the in-flight guard and the bounded pool
// with more indexes than the concurrency limit, over a multi-batch collection, all on one
// collection (so the cross-build conflict surface is exercised). Every index must converge.
func TestConcurrentBuilds_ManyIndexes_AllComplete(t *testing.T) {
	setForTest(t, &indexBackfillBatchSize, 5)
	setForTest(t, &indexBuildRetryDelay, 10*time.Millisecond)

	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.AddCollection(ctx, `type User { a: Int, b: Int, c: Int, d: Int, e: Int, f: Int }`)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	const docCount = 30
	for i := range docCount {
		doc, err := client.NewDocFromJSON(ctx,
			fmt.Appendf(nil, `{"a":%d,"b":%d,"c":%d,"d":%d,"e":%d,"f":%d}`, i, i, i, i, i, i),
			col.Version())
		require.NoError(t, err)
		require.NoError(t, col.AddDocument(ctx, doc))
	}

	fields := []string{"a", "b", "c", "d", "e", "f"}
	ids := make([]uint32, 0, len(fields))
	for _, f := range fields {
		desc, err := col.NewIndex(ctx, client.NewIndexRequest{
			Fields: []client.IndexedFieldDescription{{Name: f}},
		})
		require.NoError(t, err)
		ids = append(ids, desc.ID)
	}

	collectionID := col.Version().CollectionID
	shortID := getCollectionShortID(t, ctx, db, collectionID)
	for _, id := range ids {
		waitForIndexReady(t, ctx, db, collectionID, id)
		assert.Equal(t, docCount, countIndexEntries(t, ctx, db, shortID, id),
			"index %d must have one entry per doc", id)
	}
}

// TestDeleteFailedIndex_ClearsBackfillRecord is the regression for the orphaned-record bug: deleting
// a failed index used to leave its BackfillIndexAction Errored status + reason behind, since the
// drop only cleared the DropIndexAction record. After delete, NO action record of any kind must
// remain for the index.
func TestDeleteFailedIndex_ClearsBackfillRecord(t *testing.T) {
	ctx := context.Background()
	db := newBadgerDBNoIndexWorker(t, ctx)

	_, err := db.AddCollection(ctx, "type User { name: String\n age: Int }")
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	// Two docs share an age so a unique build on age fails.
	for _, j := range []string{`{"name":"Alice","age":21}`, `{"name":"Bob","age":21}`} {
		doc, err := client.NewDocFromJSON(ctx, []byte(j), col.Version())
		require.NoError(t, err)
		require.NoError(t, col.AddDocument(ctx, doc))
	}

	_, err = col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "age"}},
		Unique: true,
	})
	require.NoError(t, err)
	db.indexBuildWorker.drainSync(ctx)

	collectionID := col.Version().CollectionID
	indexes, err := col.ListIndexes(ctx)
	require.NoError(t, err)
	require.Len(t, indexes, 1)
	failedID := indexes[0].Description.ID
	require.True(t, readIndexState(t, ctx, db, collectionID, failedID).isFailed(),
		"index must be failed before delete")

	// Delete the failed index and drain its GC.
	col, err = db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	require.NoError(t, col.DeleteIndex(ctx, indexes[0].Description.Name))
	db.indexBuildWorker.drainSync(ctx)

	// No action record of any kind must remain (the leaked Errored backfill record is the bug).
	requireNoIndexState(t, ctx, db, collectionID, failedID)
}

// TestDeleteIndex_WhileBuilding_NoOrphanEntries is the regression for the drop-vs-build race: a
// whole-index drop range-deletes every epoch assuming no writer touches them, so it must not run
// while a backfill of the same index is still writing entries. The in-flight guard keys build and
// drop by index (not action) so they are mutually exclusive; the drop only GCs after the build
// finishes, leaving no orphaned entries behind the dropped definition.
func TestDeleteIndex_WhileBuilding_NoOrphanEntries(t *testing.T) {
	setForTest(t, &indexBackfillBatchSize, 5)
	setForTest(t, &indexBuildRetryDelay, 10*time.Millisecond)

	// Gate: hold the build at its first batch boundary until released.
	origGate := IndexBuildGate
	defer func() { IndexBuildGate = origGate }()
	release := make(chan struct{})
	var releaseOnce sync.Once
	var entered sync.WaitGroup
	entered.Add(1)
	var enteredOnce sync.Once
	IndexBuildGate = func(gateCtx context.Context, _ string, _ uint32) {
		select {
		case <-release:
			return
		default:
		}
		enteredOnce.Do(entered.Done)
		select {
		case <-release:
		case <-gateCtx.Done():
		}
	}

	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.AddCollection(ctx, `type User { name: String }`)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	const docCount = 40
	for i := range docCount {
		doc, err := client.NewDocFromJSON(ctx, fmt.Appendf(nil, `{"name":"n%02d"}`, i), col.Version())
		require.NoError(t, err)
		require.NoError(t, col.AddDocument(ctx, doc))
	}

	idx, err := col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
	})
	require.NoError(t, err)

	// Wait until the build is parked mid-flight, then delete the index while it is building.
	entered.Wait()
	col, err = db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	require.NoError(t, col.DeleteIndex(ctx, idx.Name))

	// Release the build and let everything settle.
	releaseOnce.Do(func() { close(release) })

	collectionID := col.Version().CollectionID
	shortID := getCollectionShortID(t, ctx, db, collectionID)
	require.Eventually(t, func() bool {
		return indexFullyCleaned(t, ctx, db, shortID, collectionID, idx.ID)
	}, 20*time.Second, 10*time.Millisecond, "index entries/records not fully cleaned up after delete-while-building")
}

// indexFullyCleaned reports whether the index has zero entries and no action records. It is safe to
// call from a require.Eventually condition (a separate goroutine): a read error returns false to let
// the poll continue rather than calling require, which would Goexit off the test goroutine.
func indexFullyCleaned(t *testing.T, ctx context.Context, db *DB, shortID uint32, collectionID string, indexID uint32) bool {
	t.Helper()
	rawTxn, err := db.NewTxn(true)
	if err != nil {
		return false
	}
	defer rawTxn.Discard()
	txnCtx := InitContext(ctx, rawTxn)

	prefix := &keys.IndexDataStoreKey{CollectionShortID: shortID, IndexID: indexID}
	iter, err := datastore.CtxMustGetTxn(txnCtx).Datastore().Iterator(txnCtx, datastore.IterOptions{Prefix: prefix, KeysOnly: true})
	if err != nil {
		return false
	}
	hasEntry, err := iter.Next()
	if closeErr := iter.Close(); err != nil || closeErr != nil || hasEntry {
		return false
	}

	records, err := scanIndexStates(txnCtx, indexActionCollectionPrefix(collectionID), false)
	if err != nil {
		return false
	}
	for _, rec := range records {
		if rec.Key.IndexID == indexID {
			return false
		}
	}
	return true
}

// TestDeleteIndex_WhileFailingBuild_NoOrphanRecord is the regression for a delete racing a build that
// then fails. deleteIndex clears the build record up front, but the still-in-flight build can
// re-create one after that clear: here a unique-index build over duplicate values fails and records
// an Errored backfill record once its guard releases the drop. The drop's GC used to clear only the
// drop record, leaving that Errored record orphaned for a deleted index. The GC now also clears any
// leftover backfill record, so no action record of any kind survives.
func TestDeleteIndex_WhileFailingBuild_NoOrphanRecord(t *testing.T) {
	setForTest(t, &indexBackfillBatchSize, 5)
	setForTest(t, &indexBuildRetryDelay, 10*time.Millisecond)

	// Gate: hold the build at its first batch boundary until released.
	origGate := IndexBuildGate
	defer func() { IndexBuildGate = origGate }()
	release := make(chan struct{})
	var releaseOnce sync.Once
	var entered sync.WaitGroup
	entered.Add(1)
	var enteredOnce sync.Once
	IndexBuildGate = func(gateCtx context.Context, _ string, _ uint32) {
		select {
		case <-release:
			return
		default:
		}
		enteredOnce.Do(entered.Done)
		select {
		case <-release:
		case <-gateCtx.Done():
		}
	}

	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.AddCollection(ctx, `type User { name: String, tag: String }`)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	// Distinct tag keeps each doc's ID unique; a duplicate name for docs 6+ makes the unique
	// name-index build fail on a later batch, after it has parked at the gate on the first.
	const docCount = 40
	for i := range docCount {
		name := "dup"
		if i < 6 {
			name = fmt.Sprintf("u%02d", i)
		}
		doc, err := client.NewDocFromJSON(ctx, fmt.Appendf(nil, `{"name":"%s","tag":"t%02d"}`, name, i), col.Version())
		require.NoError(t, err)
		require.NoError(t, col.AddDocument(ctx, doc))
	}

	idx, err := col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
		Unique: true,
	})
	require.NoError(t, err)

	// Delete while the build is parked, then release it so it runs on to its unique-violation failure.
	entered.Wait()
	col, err = db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	require.NoError(t, col.DeleteIndex(ctx, idx.Name))
	releaseOnce.Do(func() { close(release) })

	collectionID := col.Version().CollectionID
	shortID := getCollectionShortID(t, ctx, db, collectionID)
	require.Eventually(t, func() bool {
		return indexFullyCleaned(t, ctx, db, shortID, collectionID, idx.ID)
	}, 20*time.Second, 10*time.Millisecond, "a failed build racing a delete left an orphaned action record")
}

// TestBuild_ConflictWithLiveWrite_ConvergesToReady checks that a backfill batch that conflicts with
// a concurrent live document write still converges to ready with a full entry set, rather than
// wedging in building. The gate parks the build at its first batch; a racing goroutine updates a
// document the batch reads and commits, so the batch's commit conflicts. The build must recover
// (inner retry, or a re-drained retry via scheduleRetry if the inner retries are exhausted) and
// finish.
func TestBuild_ConflictWithLiveWrite_ConvergesToReady(t *testing.T) {
	setForTest(t, &indexBackfillBatchSize, 5)
	setForTest(t, &indexBuildRetryDelay, 10*time.Millisecond)

	origGate := IndexBuildGate
	defer func() { IndexBuildGate = origGate }()
	release := make(chan struct{})
	var releaseOnce sync.Once
	var entered sync.WaitGroup
	entered.Add(1)
	var enteredOnce sync.Once
	IndexBuildGate = func(gateCtx context.Context, _ string, _ uint32) {
		select {
		case <-release:
			return
		default:
		}
		enteredOnce.Do(entered.Done)
		select {
		case <-release:
		case <-gateCtx.Done():
		}
	}

	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.AddCollection(ctx, `type User { name: String }`)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	const docCount = 30
	docs := make([]*client.Document, docCount)
	for i := range docCount {
		doc, err := client.NewDocFromJSON(ctx, fmt.Appendf(nil, `{"name":"n%02d"}`, i), col.Version())
		require.NoError(t, err)
		require.NoError(t, col.AddDocument(ctx, doc))
		docs[i] = doc
	}

	idx, err := col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
	})
	require.NoError(t, err)

	// While the build is parked at its first batch, update a document it will read, then release.
	// The build's batch commit then conflicts with this committed write.
	entered.Wait()
	require.NoError(t, docs[0].Set(ctx, "name", "updated"))
	require.NoError(t, col.UpdateDocument(ctx, docs[0]))
	releaseOnce.Do(func() { close(release) })

	// Despite the conflict, the index must converge to ready with one entry per live doc.
	collectionID := col.Version().CollectionID
	waitForIndexReady(t, ctx, db, collectionID, idx.ID)
	shortID := getCollectionShortID(t, ctx, db, collectionID)
	require.Equal(t, docCount, countIndexEntries(t, ctx, db, shortID, idx.ID))
	require.Len(t, queryUserByName(t, db, ctx, "updated"), 1)
}

// gateAtSecondBatch drives IndexBuildGate so the build runs its first batch, then parks at the
// second batch boundary. It returns a channel that closes once parked (first batch committed) and
// a release func that lets the rest of the build proceed. This makes the mid-build window
// deterministic: no timing luck decides whether a batch has run.
func gateAtSecondBatch(t *testing.T) (parked <-chan struct{}, release func()) {
	t.Helper()
	orig := IndexBuildGate
	t.Cleanup(func() { IndexBuildGate = orig })

	parkedCh := make(chan struct{})
	releaseCh := make(chan struct{})
	var boundary int
	var mu sync.Mutex
	var parkedOnce, releaseOnce sync.Once
	IndexBuildGate = func(gateCtx context.Context, _ string, _ uint32) {
		mu.Lock()
		boundary++
		b := boundary
		mu.Unlock()
		if b != 2 {
			return // let the first batch run, then run freely after the release boundary
		}
		parkedOnce.Do(func() { close(parkedCh) })
		select {
		case <-releaseCh:
		case <-gateCtx.Done():
		}
	}
	return parkedCh, func() { releaseOnce.Do(func() { close(releaseCh) }) }
}

// TestConcurrentWrite_StaleHandleDuringBuild_MaintainsIndex guards against a document written
// through a collection handle that predates the index creation being left out of the index.
//
// A handle caches its index set when it is built and is reused across writes. A handle obtained
// before an index is created does not list that index. If such a handle updates a document while the
// backfill runs, the update must still maintain the new index: the document's old value must lose
// its (backfill-written) entry and its new value must gain one. Otherwise the live value is
// unqueryable via the index once the build finishes.
func TestConcurrentWrite_StaleHandleDuringBuild_MaintainsIndex(t *testing.T) {
	setForTest(t, &indexBackfillBatchSize, 5)
	setForTest(t, &indexBuildRetryDelay, 10*time.Millisecond)

	parked, release := gateAtSecondBatch(t)

	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.AddCollection(ctx, `type User { name: String }`)
	require.NoError(t, err)

	const docCount = 20
	docs := make([]*client.Document, docCount)
	freshCol, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	for i := range docCount {
		doc, err := client.NewDocFromJSON(ctx, fmt.Appendf(nil, `{"name":"n%02d"}`, i), freshCol.Version())
		require.NoError(t, err)
		require.NoError(t, freshCol.AddDocument(ctx, doc))
		docs[i] = doc
	}

	// Handle captured before the index exists; its cached index set has no name index.
	staleCol, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	idxCol, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	idx, err := idxCol.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
	})
	require.NoError(t, err)

	// Wait until the first batch (docs 0..4) is indexed and the build is parked, so doc 0 already has
	// an entry at its old value.
	<-parked

	// Change doc 0's indexed field through the stale handle, then let the build finish.
	require.NoError(t, docs[0].Set(ctx, "name", "CHANGED"))
	require.NoError(t, staleCol.UpdateDocument(ctx, docs[0]))
	release()

	collectionID := idxCol.Version().CollectionID
	waitForIndexReady(t, ctx, db, collectionID, idx.ID)

	// The live value must be found via the index, and no stale entry for the old value may remain.
	shortID := getCollectionShortID(t, ctx, db, collectionID)
	require.Equal(t, docCount, countIndexEntries(t, ctx, db, shortID, idx.ID))
	require.Len(t, queryUserByName(t, db, ctx, "CHANGED"), 1)
	require.Empty(t, queryUserByName(t, db, ctx, "n00"))
}

// TestConcurrentWrite_HandleHeldAcrossRebuild_LandsInLiveEpoch guards the epoch-rebuild case: a
// handle that already maintained an index (so it cached that index's epoch) and then writes while a
// rebuild is in flight must land in the NEW epoch, not the one it cached.
//
// A rebuild advances the epoch while the index ID is unchanged, so a handle keyed only on the index
// ID would reuse its cached instance and write the old, superseded epoch, leaving the live epoch
// missing the doc. The index-mutation counter is bumped when the rebuild advances the epoch, so the
// handle re-resolves and picks up the live epoch. The write happens while the rebuild's build is
// parked (epoch already advanced, build not yet complete), so only the epoch-advance bump can save it,
// not the build-completion bump. The warm-up write is load-bearing: it makes the handle cache the old
// epoch. Removing the epoch-advance bump makes the mid-rebuild write land in the stale epoch, and this
// test fails.
func TestConcurrentWrite_HandleHeldAcrossRebuild_LandsInLiveEpoch(t *testing.T) {
	setForTest(t, &indexBackfillBatchSize, 100)
	setForTest(t, &indexBuildRetryDelay, 10*time.Millisecond)

	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.AddCollection(ctx, `type User { name: String }`)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	for i := range 5 {
		doc, err := client.NewDocFromJSON(ctx, fmt.Appendf(nil, `{"name":"n%d"}`, i), col.Version())
		require.NoError(t, err)
		require.NoError(t, col.AddDocument(ctx, doc))
	}
	idx, err := col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
	})
	require.NoError(t, err)
	db.indexBuildWorker.drainSync(ctx)

	collectionID := col.Version().CollectionID

	// Handle held across the rebuild. Warm it with a write so it resolves and caches the current
	// (pre-rebuild) epoch; without this the next write would resolve fresh regardless of the counter.
	handle, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	warm, err := client.NewDocFromJSON(ctx, []byte(`{"name":"WARM"}`), handle.Version())
	require.NoError(t, err)
	require.NoError(t, handle.AddDocument(ctx, warm))

	// Park the rebuild's build at its first batch so the epoch is advanced but the build has not yet
	// completed (its completion bump has not fired).
	parked, release := gateAtFirstBatch(t)

	stageRebuild(t, ctx, db, col.Version())

	<-parked

	// The epoch is now advanced (bumped on the epoch-advance) but the build is still in flight. Write
	// through the warm handle: it must re-resolve on the bumped counter and land in the live epoch.
	after, err := client.NewDocFromJSON(ctx, []byte(`{"name":"AFTER"}`), handle.Version())
	require.NoError(t, err)
	require.NoError(t, handle.AddDocument(ctx, after))

	liveEpoch := readEpoch(t, ctx, db, collectionID, idx.ID)
	shortID := getCollectionShortID(t, ctx, db, collectionID)
	liveCount := countIndexEpochEntries(t, ctx, db, shortID, idx.ID, liveEpoch)
	t.Logf("liveEpoch=%d liveEpochEntries=%d", liveEpoch, liveCount)
	require.NotZero(t, liveCount, "AFTER must have been written to the live epoch")

	release()
	waitForIndexReady(t, ctx, db, collectionID, idx.ID)
	require.Len(t, queryUserByName(t, db, ctx, "AFTER"), 1,
		"doc written through a handle held across a rebuild must be queryable via the live epoch")
}

// gateAtFirstBatch parks the build at its first batch boundary until released. It returns a channel
// that closes once parked and a release func. Unlike gateAtSecondBatch it stops before any batch
// commits, which for a rebuild means the epoch is advanced but no build progress is committed yet.
func gateAtFirstBatch(t *testing.T) (parked <-chan struct{}, release func()) {
	t.Helper()
	orig := IndexBuildGate
	t.Cleanup(func() { IndexBuildGate = orig })

	parkedCh := make(chan struct{})
	releaseCh := make(chan struct{})
	var parkedOnce, releaseOnce sync.Once
	IndexBuildGate = func(gateCtx context.Context, _ string, _ uint32) {
		select {
		case <-releaseCh:
			return
		default:
		}
		parkedOnce.Do(func() { close(parkedCh) })
		select {
		case <-releaseCh:
		case <-gateCtx.Done():
		}
	}
	return parkedCh, func() { releaseOnce.Do(func() { close(releaseCh) }) }
}

// TestConcurrentWrite_FreshHandleDuringBuild_MaintainsIndex is the counterpart control: the
// mid-build update uses a handle obtained after the index was created, which already lists the
// building index. It must keep the index correct too. Passing here alongside the stale-handle test
// shows the fix covers the reused stale handle without regressing the ordinary case.
func TestConcurrentWrite_FreshHandleDuringBuild_MaintainsIndex(t *testing.T) {
	setForTest(t, &indexBackfillBatchSize, 5)
	setForTest(t, &indexBuildRetryDelay, 10*time.Millisecond)

	parked, release := gateAtSecondBatch(t)

	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.AddCollection(ctx, `type User { name: String }`)
	require.NoError(t, err)

	const docCount = 20
	docs := make([]*client.Document, docCount)
	freshCol, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	for i := range docCount {
		doc, err := client.NewDocFromJSON(ctx, fmt.Appendf(nil, `{"name":"n%02d"}`, i), freshCol.Version())
		require.NoError(t, err)
		require.NoError(t, freshCol.AddDocument(ctx, doc))
		docs[i] = doc
	}

	idxCol, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	idx, err := idxCol.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
	})
	require.NoError(t, err)

	<-parked

	// Handle obtained after the index was created; its cached index set includes the building index.
	updCol, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	require.NoError(t, docs[0].Set(ctx, "name", "CHANGED"))
	require.NoError(t, updCol.UpdateDocument(ctx, docs[0]))
	release()

	collectionID := idxCol.Version().CollectionID
	waitForIndexReady(t, ctx, db, collectionID, idx.ID)

	shortID := getCollectionShortID(t, ctx, db, collectionID)
	require.Equal(t, docCount, countIndexEntries(t, ctx, db, shortID, idx.ID))
	require.Len(t, queryUserByName(t, db, ctx, "CHANGED"), 1)
	require.Empty(t, queryUserByName(t, db, ctx, "n00"))
}
