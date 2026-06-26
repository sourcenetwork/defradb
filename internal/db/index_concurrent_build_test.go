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
)

// waitForIndexReady blocks until the index has no build/drop state record (ready), or fails the test.
func waitForIndexReady(t *testing.T, ctx context.Context, db *DB, collectionID string, indexID uint32) {
	t.Helper()
	require.Eventually(t, func() bool {
		rawTxn, err := db.NewTxn(true)
		require.NoError(t, err)
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
		return countIndexEntries(t, ctx, db, shortID, idx.ID) == 0 &&
			noIndexActionRecords(t, ctx, db, collectionID, idx.ID)
	}, 20*time.Second, 10*time.Millisecond, "index entries/records not fully cleaned up after delete-while-building")
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

// noIndexActionRecords reports whether the index has no action record of any kind.
func noIndexActionRecords(t *testing.T, ctx context.Context, db *DB, collectionID string, indexID uint32) bool {
	t.Helper()
	rawTxn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer rawTxn.Discard()
	records, err := scanIndexStates(InitContext(ctx, rawTxn), indexActionCollectionPrefix(collectionID), false)
	require.NoError(t, err)
	for _, rec := range records {
		if rec.Key.IndexID == indexID {
			return false
		}
	}
	return true
}
