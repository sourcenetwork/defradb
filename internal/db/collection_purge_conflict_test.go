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

// setupPrunedFieldBlock leaves a composite pointing to an absent field block.
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

func storeBlockInNewTxn(t *testing.T, ctx context.Context, db *DB, block blocks.Block) {
	t.Helper()

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	blockstore := datastore.BlockstoreFrom(db.rootstore, db.blockStoreChunkSize)
	require.NoError(t, blockstore.Put(InitContext(ctx, txn), block))
	require.NoError(t, txn.Commit())
}

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
		// Badger only checks read conflicts on transactions with pending writes.
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
