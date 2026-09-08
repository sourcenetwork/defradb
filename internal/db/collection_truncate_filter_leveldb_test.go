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
	"time"

	"github.com/fxamacker/cbor/v2"
	blocks "github.com/ipfs/go-block-format"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/leveldb"

	"github.com/sourcenetwork/defradb/client"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

func newLevelDBForHistoryCleanupTest(t *testing.T, ctx context.Context) *DB {
	t.Helper()

	rootstore, err := leveldb.NewDatastore(t.TempDir(), nil)
	require.NoError(t, err)
	adminInfo, err := acpDB.NewNACInfo(ctx, "", false)
	require.NoError(t, err)
	db, err := newDB(ctx, rootstore, adminInfo)
	require.NoError(t, err)
	t.Cleanup(db.Close)
	return db
}

func requireLevelDBOperationCompletes(t *testing.T, operation func() error) {
	t.Helper()

	done := make(chan error, 1)
	go func() {
		done <- operation()
	}()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("operation blocked while opening a nested LevelDB transaction")
	}
}

func seedLevelDBHistoryDocument(
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
	encodedName, err := cbor.Marshal("alice")
	require.NoError(t, err)
	require.NoError(t, dbTxn.Datastore().Set(txnCtx, keys.DataStoreKey{
		CollectionShortID: shortID,
		InstanceType:      keys.ValueKey,
		DocShortID:        docShortID,
		FieldID:           "1",
	}, encodedName))
	require.NoError(t, dbTxn.Headstore().Set(txnCtx, keys.HeadstoreDocKey{
		DocShortID: docShortID,
		FieldID:    "C",
		Cid:        missingBlock,
	}.Bytes(), nil))
	require.NoError(t, txn.Commit())

	return col, docID
}

func TestTruncateWithFilterDoesNotOpenNestedLevelDBTransaction(t *testing.T) {
	ctx := context.Background()
	db := newLevelDBForHistoryCleanupTest(t, ctx)
	_, docID := seedLevelDBHistoryDocument(t, ctx, db)

	requireLevelDBOperationCompletes(t, func() error {
		return truncateDocuments(db, ctx, "User", []client.DocID{docID})
	})
}

func TestTruncateDoesNotOpenNestedLevelDBTransaction(t *testing.T) {
	ctx := context.Background()
	db := newLevelDBForHistoryCleanupTest(t, ctx)
	col, _ := seedLevelDBHistoryDocument(t, ctx, db)

	requireLevelDBOperationCompletes(t, func() error {
		return col.Truncate(ctx)
	})
}

func TestGraphQLTruncateDoesNotOpenNestedLevelDBTransaction(t *testing.T) {
	ctx := context.Background()
	db := newLevelDBForHistoryCleanupTest(t, ctx)
	_, docID := seedLevelDBHistoryDocument(t, ctx, db)

	done := make(chan *client.RequestResult, 1)
	go func() {
		done <- db.ExecRequest(ctx, fmt.Sprintf(`mutation {
			truncate_User(filter: {_docID: {_eq: %q}})
		}`, docID.String()))
	}()

	select {
	case result := <-done:
		require.Empty(t, result.GQL.Errors)
		require.Equal(t, map[string]any{"truncate_User": true}, result.GQL.Data)
	case <-time.After(5 * time.Second):
		t.Fatal("GraphQL truncate blocked while opening a nested LevelDB transaction")
	}
}
