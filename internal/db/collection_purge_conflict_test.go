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

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

func TestPurgeByDocIDsPruneHistoryTracksCallerTransactionConflicts(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddCollection(ctx, userDocIDTestSchema)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	doc, err := client.NewDocFromJSON(ctx, []byte(`{"name":"Alice","age":40}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.AddDocument(ctx, doc))

	composite := loadTestBlock(t, ctx, db, doc.Head())
	require.NotEmpty(t, composite.Links)
	fieldCID := composite.Links[0].Cid
	blockstore := datastore.BlockstoreFrom(db.rootstore, db.blockStoreChunkSize)
	fieldBlock, err := blockstore.Get(ctx, fieldCID)
	require.NoError(t, err)
	require.NoError(t, blockstore.DeleteBlock(ctx, fieldCID))

	purgeTxn, err := db.NewTxn(false)
	require.NoError(t, err)
	purgeCtx := InitContext(ctx, purgeTxn)
	require.NoError(t, col.PurgeByDocIDs(purgeCtx, []client.DocID{doc.ID()}, true))

	writeTxn, err := db.NewTxn(false)
	require.NoError(t, err)
	require.NoError(t, blockstore.Put(InitContext(ctx, writeTxn), fieldBlock))
	require.NoError(t, writeTxn.Commit())

	require.ErrorIs(t, purgeTxn.Commit(), corekv.ErrTxnConflict)
}
