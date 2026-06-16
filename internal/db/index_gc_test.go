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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// countIndexEntries returns the number of raw keys under the given index prefix
// using a fresh read-only transaction.
func countIndexEntries(t *testing.T, ctx context.Context, db *DB, shortID uint32, indexID uint32) int {
	t.Helper()
	rawTxn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer rawTxn.Discard()
	txnCtx := InitContext(ctx, rawTxn)
	txn := datastore.CtxMustGetTxn(txnCtx)

	prefix := &keys.IndexDataStoreKey{
		CollectionShortID: shortID,
		IndexID:           indexID,
	}
	iter, err := txn.Datastore().Iterator(txnCtx, datastore.IterOptions{
		Prefix:   prefix,
		KeysOnly: true,
	})
	require.NoError(t, err)

	var count int
	for {
		hasNext, err := iter.Next()
		if err != nil {
			require.NoError(t, errors.Join(err, iter.Close()))
		}
		if !hasNext {
			break
		}
		count++
	}
	require.NoError(t, iter.Close())
	return count
}

// clearIndexEntries deletes all raw keys under the given index prefix using the
// transaction already set on ctx. Used to simulate a build that has not yet run.
func clearIndexEntries(t *testing.T, ctx context.Context, shortID uint32, indexID uint32) error {
	t.Helper()
	txn := datastore.CtxMustGetTxn(ctx)

	prefix := &keys.IndexDataStoreKey{CollectionShortID: shortID, IndexID: indexID}
	iter, err := txn.Datastore().Iterator(ctx, datastore.IterOptions{Prefix: prefix, KeysOnly: true})
	if err != nil {
		return err
	}

	var rawKeys [][]byte
	for {
		hasNext, err := iter.Next()
		if err != nil {
			return errors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}
		rawKeys = append(rawKeys, append([]byte(nil), iter.Key()...))
	}
	if err := iter.Close(); err != nil {
		return err
	}

	for _, k := range rawKeys {
		if err := txn.Datastore().Delete(ctx, rawBytesKey{k}); err != nil {
			return err
		}
	}
	return nil
}

// getCollectionShortID resolves the short collection ID for tests.
func getCollectionShortID(t *testing.T, ctx context.Context, db *DB, collectionID string) uint32 {
	t.Helper()
	var shortID uint32
	err := db.withTxnRetries(ctx, func(txnCtx context.Context) error {
		var resolveErr error
		shortID, resolveErr = id.GetShortCollectionID(txnCtx, collectionID)
		return resolveErr
	})
	require.NoError(t, err)
	return shortID
}

// TestGCIndex_MultiBatch_DeletesAllEntries creates 10 documents, creates an index,
// then deletes it with a reduced batch size to exercise multi-batch GC. It asserts:
//   - No index state record remains after deletion.
//   - The index is absent from ListIndexes.
//   - A raw prefix scan over the index keys returns zero entries.
func TestGCIndex_MultiBatch_DeletesAllEntries(t *testing.T) {
	origBatchSize := indexBackfillBatchSize
	indexBackfillBatchSize = 3
	defer func() { indexBackfillBatchSize = origBatchSize }()

	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	for i := range 10 {
		addUserDoc(t, ctx, col, fmt.Sprintf("name%02d", i))
	}

	desc, err := newNameIndex(t, ctx, col)
	require.NoError(t, err)

	collectionID := col.Version().CollectionID
	shortID := getCollectionShortID(t, ctx, db, collectionID)

	// Sanity check: index entries exist before deletion.
	before := countIndexEntries(t, ctx, db, shortID, desc.ID)
	assert.Equal(t, 10, before, "expected 10 index entries before delete")

	// Re-fetch collection so we have the latest definition.
	col, err = db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	require.NoError(t, col.DeleteIndex(ctx, desc.Name))

	// The state record must be gone.
	requireNoIndexState(t, ctx, db, collectionID, desc.ID)

	// The index must be absent from the collection listing.
	col, err = db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	indexes, err := col.ListIndexes(ctx)
	require.NoError(t, err)
	assert.Empty(t, indexes, "expected no indexes after delete")

	// All index entries must be gone.
	after := countIndexEntries(t, ctx, db, shortID, desc.ID)
	assert.Equal(t, 0, after, "expected 0 index entries after GC")
}

// TestGCIndex_EmptyIndex_Succeeds creates an index on an empty collection and deletes it.
// No entries to GC; the operation must still succeed and leave a clean state.
func TestGCIndex_EmptyIndex_Succeeds(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	desc, err := newNameIndex(t, ctx, col)
	require.NoError(t, err)

	collectionID := col.Version().CollectionID
	shortID := getCollectionShortID(t, ctx, db, collectionID)

	require.NoError(t, col.DeleteIndex(ctx, desc.Name))

	requireNoIndexState(t, ctx, db, collectionID, desc.ID)

	col, err = db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	indexes, err := col.ListIndexes(ctx)
	require.NoError(t, err)
	assert.Empty(t, indexes, "expected no indexes after delete")

	after := countIndexEntries(t, ctx, db, shortID, desc.ID)
	assert.Equal(t, 0, after)
}
