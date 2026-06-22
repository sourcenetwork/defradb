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
	"testing"

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

// stageRebuild stages the build+drop records for a rebuild of every index in col the way
// reindexNewActiveVersion does, and returns the deferred run function.
func stageRebuild(t *testing.T, ctx context.Context, db *DB, col client.CollectionVersion) func(context.Context) error {
	t.Helper()
	var run func(context.Context) error
	require.NoError(t, db.withTxnRetries(ctx, func(c context.Context) error {
		var err error
		run, err = db.reindexNewActiveVersion(c, col)
		return err
	}))
	return run
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

	desc, err := newNameIndex(t, ctx, col)
	require.NoError(t, err)

	oldEpoch := readEpoch(t, ctx, db, collectionID, desc.ID)
	require.Equal(t, uint32(1), oldEpoch)
	require.Equal(t, 3, countIndexEpochEntries(t, ctx, db, shortID, desc.ID, oldEpoch))

	def, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	run := stageRebuild(t, ctx, db, def.Version())
	require.NoError(t, run(ctx))

	// New epoch (2) holds every doc; old epoch is collected; the index resolves to the new epoch
	// and is ready.
	assert.Equal(t, 3, countIndexEpochEntries(t, ctx, db, shortID, desc.ID, 2))
	assert.Equal(t, 0, countIndexEpochEntries(t, ctx, db, shortID, desc.ID, oldEpoch))
	assert.Equal(t, uint32(2), readEpoch(t, ctx, db, collectionID, desc.ID))
	requireNoIndexState(t, ctx, db, collectionID, desc.ID)

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

	desc, err := newNameIndex(t, ctx, col)
	require.NoError(t, err)
	oldEpoch := readEpoch(t, ctx, db, collectionID, desc.ID) // epoch 1

	// Reproduce the interrupted state: advance to epoch 2 and build its entries, leaving epoch 1
	// stale and no action record at all.
	require.NoError(t, db.withTxnRetries(ctx, func(c context.Context) error {
		newEpoch, err := allocateIndexEpoch(c, collectionID, desc.ID)
		if err != nil {
			return err
		}
		require.Equal(t, uint32(2), newEpoch)
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

	require.NoError(t, db.recoverIndexStates(context.Background()))

	assert.Equal(t, 0, countIndexEpochEntries(t, ctx, db, shortID, desc.ID, oldEpoch))
	assert.Equal(t, 3, countIndexEpochEntries(t, ctx, db, shortID, desc.ID, 2))
	for _, name := range []string{"a", "b", "c"} {
		require.Len(t, queryUserByName(t, db, ctx, name), 1, "doc %q must be queryable after recovery", name)
	}
}
