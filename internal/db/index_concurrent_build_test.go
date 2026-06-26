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
	origBatch := indexBackfillBatchSize
	indexBackfillBatchSize = 5
	defer func() { indexBackfillBatchSize = origBatch }()

	origConc := indexBuildConcurrency
	indexBuildConcurrency = 4
	defer func() { indexBuildConcurrency = origConc }()

	origDelay := indexBuildRetryDelay
	indexBuildRetryDelay = 10 * time.Millisecond
	defer func() { indexBuildRetryDelay = origDelay }()

	// Gate: park each build at its first batch boundary until released, so both are in flight at
	// once. Each build blocks on its first gate call, so the two signal exactly once before parking.
	origGate := IndexBuildGate
	defer func() { IndexBuildGate = origGate }()
	release := make(chan struct{})
	var releaseOnce sync.Once
	var entered sync.WaitGroup
	entered.Add(2)
	IndexBuildGate = func(gateCtx context.Context, _ string, _ uint32) {
		select {
		case <-release:
			return // already released; let later batches run freely
		default:
		}
		entered.Done()
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
	origBatch := indexBackfillBatchSize
	indexBackfillBatchSize = 5
	defer func() { indexBackfillBatchSize = origBatch }()

	origDelay := indexBuildRetryDelay
	indexBuildRetryDelay = 10 * time.Millisecond
	defer func() { indexBuildRetryDelay = origDelay }()

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
	assertNoActionRecordsForIndex(t, ctx, db, collectionID, failedID)
}

// assertNoActionRecordsForIndex fails if any action record (any action type) exists for the index.
func assertNoActionRecordsForIndex(t *testing.T, ctx context.Context, db *DB, collectionID string, indexID uint32) {
	t.Helper()
	rawTxn, err := db.NewTxn(true)
	require.NoError(t, err)
	t.Cleanup(func() { rawTxn.Discard() })
	txnCtx := InitContext(ctx, rawTxn)

	records, err := scanIndexStates(txnCtx, indexActionCollectionPrefix(collectionID), false)
	require.NoError(t, err)
	for _, rec := range records {
		require.NotEqual(t, indexID, rec.Key.IndexID,
			"expected no action record for dropped index, found %+v", rec.State)
	}
}
