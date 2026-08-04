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
	"time"

	blocks "github.com/ipfs/go-block-format"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/leveldb"

	"github.com/sourcenetwork/defradb/client"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

func newLevelDBForPurgeTest(t *testing.T, ctx context.Context) *DB {
	t.Helper()

	rootstore, err := leveldb.NewDatastore(t.TempDir(), nil)
	require.NoError(t, err)
	adminInfo, err := acpDB.NewNACInfo(ctx, "", false)
	require.NoError(t, err)
	db, err := newDB(ctx, rootstore, adminInfo)
	require.NoError(t, err)
	return db
}

func requireLevelDBOperationCompletes(t *testing.T, db *DB, operation func() error) {
	t.Helper()

	done := make(chan error, 1)
	go func() {
		done <- operation()
	}()

	select {
	case err := <-done:
		require.NoError(t, err)
		db.Close()
	case <-time.After(5 * time.Second):
		t.Fatal("operation blocked while opening a nested LevelDB transaction")
	}
}

func seedLevelDBPurgeDocument(
	t *testing.T,
	ctx context.Context,
	db *DB,
) (client.Collection, client.DocID) {
	t.Helper()

	_, err := db.AddCollection(ctx, userDocIDTestSchema)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	missingBlock := blocks.NewBlock([]byte("missing history block")).Cid()
	docID := client.NewDocIDV0(missingBlock)

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	dbTxn, ok := txn.(*Txn)
	require.True(t, ok)
	txnCtx := InitContext(ctx, txn)
	shortID, err := id.GetCollectionShortID(txnCtx, col.CollectionID())
	require.NoError(t, err)

	const docShortID = uint64(1)
	require.NoError(t, id.SetDocIDMapping(txnCtx, shortID, docShortID, docID.String()))
	require.NoError(t, id.SetBlockDocIDMapping(txnCtx, missingBlock, docID.String()))
	require.NoError(t, dbTxn.Datastore().Set(txnCtx, keys.PrimaryDataStoreKey{
		CollectionShortID: shortID,
		DocShortID:        docShortID,
	}, nil))
	require.NoError(t, dbTxn.Datastore().Set(txnCtx, keys.DataStoreKey{
		CollectionShortID: shortID,
		InstanceType:      keys.ValueKey,
		DocShortID:        docShortID,
		FieldID:           "1",
	}, []byte("alice")))
	require.NoError(t, dbTxn.Headstore().Set(txnCtx, keys.HeadstoreDocKey{
		DocShortID: docShortID,
		FieldID:    "C",
		Cid:        missingBlock,
	}.Bytes(), nil))
	require.NoError(t, txn.Commit())

	return col, docID
}

func TestPurgeAndTruncateDoNotOpenNestedLevelDBTransactions(t *testing.T) {
	t.Run("targeted purge", func(t *testing.T) {
		ctx := context.Background()
		db := newLevelDBForPurgeTest(t, ctx)
		col, docID := seedLevelDBPurgeDocument(t, ctx, db)

		requireLevelDBOperationCompletes(t, db, func() error {
			return col.PurgeByDocIDs(ctx, []client.DocID{docID}, true)
		})
	})

	t.Run("truncate", func(t *testing.T) {
		ctx := context.Background()
		db := newLevelDBForPurgeTest(t, ctx)
		col, _ := seedLevelDBPurgeDocument(t, ctx, db)

		requireLevelDBOperationCompletes(t, db, func() error {
			return col.Truncate(ctx)
		})
	})
}
