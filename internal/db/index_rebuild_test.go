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

	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/sequence"
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

// TestRunIndexRebuild_FillsNewEpochFlipsAndGCsOld drives a rebuild directly and checks the epoch
// mechanics: a fresh epoch is filled, the index flips to it (ready, no state record), the old
// epoch's entries are collected, and every doc stays queryable through the index.
func TestRunIndexRebuild_FillsNewEpochFlipsAndGCsOld(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)
	collectionID := col.Version().CollectionID
	shortID := getCollectionShortID(t, ctx, db, collectionID)

	for _, name := range []string{"a", "b", "c"} {
		addUserDoc(t, ctx, col, name)
	}

	desc, err := newNameIndex(t, ctx, col)
	require.NoError(t, err)

	// The initial build occupies epoch 1; the whole-index prefix and epoch 1 agree.
	oldEpoch := readEpoch(t, ctx, db, collectionID, desc.ID)
	require.Equal(t, uint32(1), oldEpoch)
	require.Equal(t, 3, countIndexEpochEntries(t, ctx, db, shortID, desc.ID, oldEpoch))

	def, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	// Stage the rebuild epochs the way reindexNewActiveVersion does, then run it.
	var buildingEpoch uint32
	require.NoError(t, db.withTxnRetries(ctx, func(c context.Context) error {
		seq, err := sequence.Get(c, keys.NewIndexEpochSequenceKey(collectionID, desc.ID))
		if err != nil {
			return err
		}
		next, err := seq.Next(c)
		if err != nil {
			return err
		}
		buildingEpoch = uint32(next)
		return db.startIndexRebuild(c, collectionID, desc.ID, buildingEpoch, oldEpoch)
	}))
	require.Equal(t, uint32(2), buildingEpoch)

	require.NoError(t, db.runIndexRebuild(
		ctx, def.Version(), desc, immutable.None[string](), buildingEpoch, oldEpoch,
	))

	// New epoch holds every doc; old epoch is collected; the index resolves to the new epoch
	// and is ready (no state record).
	assert.Equal(t, 3, countIndexEpochEntries(t, ctx, db, shortID, desc.ID, buildingEpoch))
	assert.Equal(t, 0, countIndexEpochEntries(t, ctx, db, shortID, desc.ID, oldEpoch))
	assert.Equal(t, buildingEpoch, readEpoch(t, ctx, db, collectionID, desc.ID))
	requireNoIndexState(t, ctx, db, collectionID, desc.ID)

	for _, name := range []string{"a", "b", "c"} {
		require.Len(t, queryUserByName(t, db, ctx, name), 1, "doc %q must be queryable after rebuild", name)
	}
}
