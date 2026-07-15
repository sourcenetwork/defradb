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

// countIndexEpochEntries counts the stored entries for an index in a single epoch.
func countIndexEpochEntries(t *testing.T, ctx context.Context, db *DB, shortID, indexID, epoch uint32) int {
	t.Helper()
	rawTxn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer rawTxn.Discard()
	txnCtx := InitContext(ctx, rawTxn)
	txn := datastore.CtxMustGetTxn(txnCtx)

	prefix := &keys.IndexDataStoreKey{CollectionShortID: shortID, IndexID: indexID, Epoch: epoch}
	iter, err := txn.Datastore().Iterator(txnCtx, datastore.IterOptions{Prefix: prefix, KeysOnly: true})
	require.NoError(t, err)

	var count int
	for {
		hasNext, err := iter.Next()
		require.NoError(t, err)
		if !hasNext {
			break
		}
		count++
	}
	require.NoError(t, iter.Close())
	return count
}

// hasStaleEpochMarker reports whether the index's stale-epoch marker is present.
func hasStaleEpochMarker(t *testing.T, ctx context.Context, db *DB, shortID, indexID uint32) bool {
	t.Helper()
	rawTxn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer rawTxn.Discard()
	txn := datastore.CtxMustGetTxn(InitContext(ctx, rawTxn))
	has, err := txn.Systemstore().Has(ctx, keys.NewIndexStaleEpochKey(shortID, indexID).Bytes())
	require.NoError(t, err)
	return has
}

// readEpoch returns the index's current epoch (the epoch sequence value).
func readEpoch(t *testing.T, ctx context.Context, db *DB, collectionID string, indexID uint32) uint32 {
	t.Helper()
	rawTxn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer rawTxn.Discard()
	epoch, err := getIndexEpoch(InitContext(ctx, rawTxn), collectionID, indexID)
	require.NoError(t, err)
	return epoch
}

// stageRebuild stages the fresh-epoch build records for every index in col the way
// reindexNewActiveVersion does (advance epoch, mark stale, record building). The worker does the
// actual rebuild; callers drive it with drainSync.
func stageRebuild(t *testing.T, ctx context.Context, db *DB, col client.CollectionVersion) {
	t.Helper()
	require.NoError(t, db.withTxnRetries(ctx, func(c context.Context) error {
		return db.reindexNewActiveVersion(c, col)
	}))
}

// TestReindexNewActiveVersion_BuildsNewEpochAndCollectsOld checks the rebuild mechanics: a fresh
// epoch is filled, the old epoch's entries are collected, the index resolves to the new epoch and
// is ready, and every doc is queryable.
func TestReindexNewActiveVersion_BuildsNewEpochAndCollectsOld(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)
	collectionID := col.Version().CollectionID
	shortID := getCollectionShortID(t, ctx, db, collectionID)

	for _, name := range []string{"a", "b", "c"} {
		addUserDoc(t, ctx, col, name)
	}

	desc, err := newNameIndex(t, ctx, db, col)
	require.NoError(t, err)

	oldEpoch := readEpoch(t, ctx, db, collectionID, desc.ID)
	require.Equal(t, uint32(1), oldEpoch)
	require.Equal(t, 3, countIndexEpochEntries(t, ctx, db, shortID, desc.ID, oldEpoch))

	def, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	stageRebuild(t, ctx, db, def.Version())
	db.indexBuildWorker.drainSync(ctx)

	// New epoch (2) holds every doc; old epoch is collected; the index resolves to the new epoch
	// and is ready.
	assert.Equal(t, 3, countIndexEpochEntries(t, ctx, db, shortID, desc.ID, 2))
	assert.Equal(t, 0, countIndexEpochEntries(t, ctx, db, shortID, desc.ID, oldEpoch))
	assert.Equal(t, uint32(2), readEpoch(t, ctx, db, collectionID, desc.ID))
	requireNoIndexState(t, ctx, db, collectionID, desc.ID)
	assert.False(t, hasStaleEpochMarker(t, ctx, db, shortID, desc.ID),
		"the stale-epoch marker must be cleared after the rebuild collects the old epoch")

	for _, name := range []string{"a", "b", "c"} {
		require.Len(t, queryUserByName(t, db, ctx, name), 1, "doc %q must be queryable after rebuild", name)
	}
}

// TestReindexNewActiveVersion_RecoversStaleEpochs reproduces a crash after a rebuild's build
// finished but before its superseded epoch was collected: epoch 2 is live and complete, epoch 1 is
// stale, and there is NO record of either (the build record was deleted at completion, and a
// rebuild writes no drop record). Recovery must collect the stale epoch by scanning, leave the
// live one intact, and the index stays queryable throughout.
func TestReindexNewActiveVersion_RecoversStaleEpochs(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)
	collectionID := col.Version().CollectionID
	shortID := getCollectionShortID(t, ctx, db, collectionID)

	for _, name := range []string{"a", "b", "c"} {
		addUserDoc(t, ctx, col, name)
	}

	desc, err := newNameIndex(t, ctx, db, col)
	require.NoError(t, err)
	oldEpoch := readEpoch(t, ctx, db, collectionID, desc.ID) // epoch 1

	// Reproduce the interrupted state: advance to epoch 2, mark the stale epochs (as a real rebuild
	// does atomically with the advance), and build the new epoch's entries, leaving epoch 1 stale
	// and no build/drop action record.
	require.NoError(t, db.withTxnRetries(ctx, func(c context.Context) error {
		newEpoch, err := allocateIndexEpoch(c, collectionID, desc.ID)
		if err != nil {
			return err
		}
		require.Equal(t, uint32(2), newEpoch)
		if err := db.markStaleEpochs(c, shortID, desc.ID); err != nil {
			return err
		}
		coll, err := db.newCollection(c, col.Version(), datastore.CtxTryGetTxnOption(c))
		if err != nil {
			return err
		}
		colIndex, err := NewCollectionIndex(c, coll, desc, false)
		if err != nil {
			return err
		}
		return coll.indexExistingDocs(c, colIndex)
	}))

	require.Equal(t, 3, countIndexEpochEntries(t, ctx, db, shortID, desc.ID, 2))
	require.Equal(t, 3, countIndexEpochEntries(t, ctx, db, shortID, desc.ID, oldEpoch))
	requireNoIndexState(t, ctx, db, collectionID, desc.ID)

	// Queryable while the stale epoch is still present.
	for _, name := range []string{"a", "b", "c"} {
		require.Len(t, queryUserByName(t, db, ctx, name), 1, "doc %q must be queryable with stale epoch", name)
	}

	db.indexBuildWorker.drainSync(context.Background())

	assert.Equal(t, 0, countIndexEpochEntries(t, ctx, db, shortID, desc.ID, oldEpoch))
	assert.Equal(t, 3, countIndexEpochEntries(t, ctx, db, shortID, desc.ID, 2))
	assert.False(t, hasStaleEpochMarker(t, ctx, db, shortID, desc.ID),
		"the stale-epoch marker must be cleared after recovery collects the stale epoch")
	for _, name := range []string{"a", "b", "c"} {
		require.Len(t, queryUserByName(t, db, ctx, name), 1, "doc %q must be queryable after recovery", name)
	}
}

// TestBackfill_EpochAdvancedMidBuild_LiveEpochComplete is the regression for the mid-build epoch
// split: an initial build resolves its epoch once and pins it, so a version switch that advances the
// sequence while the build is in flight cannot make the build write its later batches into the new
// epoch while its earlier batches landed in the old one. If the epoch were re-read per batch, the new
// live epoch would be missing every document indexed before the advance. Here the initial build is
// parked mid-flight, the epoch is advanced underneath it (as a version switch does), and after
// everything settles the live epoch must hold every document.
func TestBackfill_EpochAdvancedMidBuild_LiveEpochComplete(t *testing.T) {
	setForTest(t, &indexBackfillBatchSize, 5)
	setForTest(t, &indexBuildRetryDelay, 10*time.Millisecond)

	// Gate: park the initial build at its first batch boundary until released.
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
	collectionID := col.Version().CollectionID
	shortID := getCollectionShortID(t, ctx, db, collectionID)

	// Wait until the initial build is parked mid-flight, then advance the epoch underneath it exactly
	// as a version switch does: allocate the next epoch, mark the old one stale, and stage a fresh
	// build record.
	entered.Wait()
	def, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	require.NoError(t, db.withTxnRetries(ctx, func(c context.Context) error {
		return db.reindexNewActiveVersion(c, def.Version())
	}))

	// Release the parked build and let the worker settle: the pinned build finishes the old epoch
	// (now stale, collected by the sweep), and the staged rebuild fills the new live epoch in full.
	releaseOnce.Do(func() { close(release) })

	waitForIndexReady(t, ctx, db, collectionID, idx.ID)
	liveEpoch := readEpoch(t, ctx, db, collectionID, idx.ID)
	assert.Equal(t, docCount, countIndexEpochEntries(t, ctx, db, shortID, idx.ID, liveEpoch),
		"the live epoch must hold every document, not just those indexed after the advance")
	for i := range docCount {
		name := fmt.Sprintf("n%02d", i)
		require.Len(t, queryUserByName(t, db, ctx, name), 1, "doc %q must be queryable", name)
	}
}
