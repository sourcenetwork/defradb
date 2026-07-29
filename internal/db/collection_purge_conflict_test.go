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

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

// setupPrunedFieldBlock creates a document, then removes one of its field blocks from the
// blockstore while leaving the composite's link to it in place. It returns the now-absent
// block's CID and content so a test can store it back from a concurrent transaction, standing
// in for a DAG sync that recreates a block the purge walk is reading.
func setupPrunedFieldBlock(
	t *testing.T,
	ctx context.Context,
) (*DB, client.Collection, client.DocID, cid.Cid, blocks.Block) {
	t.Helper()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)

	_, err = db.AddCollection(ctx, userDocIDTestSchema)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON(ctx, []byte(`{"name":"Alice","age":40}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.AddDocument(ctx, doc))

	composite := loadTestBlock(t, ctx, db, doc.Head())
	require.NotEmpty(t, composite.Links)
	blockCID := composite.Links[0].Cid

	blockstore := datastore.BlockstoreFrom(db.rootstore, db.blockStoreChunkSize)
	block, err := blockstore.Get(ctx, blockCID)
	require.NoError(t, err)
	require.NoError(t, blockstore.DeleteBlock(ctx, blockCID))

	return db, col, doc.ID(), blockCID, block
}

// storeBlockInNewTxn stores block in its own committed transaction, standing in for a
// concurrent DAG sync that recreates a block.
func storeBlockInNewTxn(t *testing.T, ctx context.Context, db *DB, block blocks.Block) {
	t.Helper()

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	blockstore := datastore.BlockstoreFrom(db.rootstore, db.blockStoreChunkSize)
	require.NoError(t, blockstore.Put(InitContext(ctx, txn), block))
	require.NoError(t, txn.Commit())
}

// TestPurgeByDocIDsPruneHistoryReadsBlocksOutsideConflictSet checks that pruning a document's
// block history reads block content outside the purge transaction's conflict set, so a
// concurrent store of a block the walk visits does not abort the purge. The control subtest
// confirms the same store does abort a transaction that read the block itself, so the fix
// subtest is not passing vacuously.
func TestPurgeByDocIDsPruneHistoryReadsBlocksOutsideConflictSet(t *testing.T) {
	t.Run("reading the block through the write transaction conflicts", func(t *testing.T) {
		ctx := context.Background()
		db, _, _, blockCID, block := setupPrunedFieldBlock(t, ctx)
		defer db.Close()

		blockstore := datastore.BlockstoreFrom(db.rootstore, db.blockStoreChunkSize)

		txn, err := db.NewTxn(false)
		require.NoError(t, err)
		tctx := InitContext(ctx, txn)
		_, _, err = getBlock(tctx, blockstore, blockCID)
		require.NoError(t, err)
		// badger only checks for conflicts when the transaction has a pending write; the real
		// purge deletes as it walks, so delete the block here too.
		require.NoError(t, blockstore.DeleteBlock(tctx, blockCID))

		storeBlockInNewTxn(t, ctx, db, block)

		require.ErrorIs(t, txn.Commit(), corekv.ErrTxnConflict)
	})

	t.Run("pruning the document does not conflict", func(t *testing.T) {
		ctx := context.Background()
		db, col, docID, _, block := setupPrunedFieldBlock(t, ctx)
		defer db.Close()

		txn, err := db.NewTxn(false)
		require.NoError(t, err)
		require.NoError(t, col.PurgeByDocIDs(InitContext(ctx, txn), []client.DocID{docID}, true))

		storeBlockInNewTxn(t, ctx, db, block)

		require.NoError(t, txn.Commit())
	})
}
