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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

func TestCollectionTruncateRemovesDocIDMappings(t *testing.T) {
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
	err = col.AddDocument(ctx, doc)
	require.NoError(t, err)

	publicDocID := doc.ID().String()
	compositeBlock := loadTestBlock(t, ctx, db, doc.Head())
	require.NotEmpty(t, compositeBlock.Links)
	genesisFieldCID := compositeBlock.Links[0].Cid

	txn, err := db.NewTxn(true)
	require.NoError(t, err)
	dbTxn, ok := txn.(*Txn)
	require.True(t, ok)
	txnCtx := InitContext(ctx, dbTxn)

	collectionShortID, err := id.GetShortCollectionID(txnCtx, col.CollectionID())
	require.NoError(t, err)
	docShortID, found, err := id.GetShortDocID(txnCtx, collectionShortID, publicDocID)
	require.NoError(t, err)
	require.True(t, found)

	blockDocID, found, err := id.GetDocIDForBlockFromStore(
		txnCtx,
		dbTxn.Systemstore(),
		genesisFieldCID,
	)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, publicDocID, blockDocID)
	txn.Discard()

	err = col.Truncate(ctx)
	require.NoError(t, err)

	txn, err = db.NewTxn(true)
	require.NoError(t, err)
	defer txn.Discard()
	dbTxn, ok = txn.(*Txn)
	require.True(t, ok)
	txnCtx = InitContext(ctx, dbTxn)

	_, found, err = id.GetShortDocID(txnCtx, collectionShortID, publicDocID)
	require.NoError(t, err)
	require.False(t, found)

	_, found, err = id.GetDocID(txnCtx, docShortID)
	require.NoError(t, err)
	require.False(t, found)

	_, found, err = id.GetDocRef(txnCtx, publicDocID)
	require.NoError(t, err)
	require.False(t, found)

	_, found, err = id.GetDocIDForBlockFromStore(
		txnCtx,
		dbTxn.Systemstore(),
		genesisFieldCID,
	)
	require.NoError(t, err)
	require.False(t, found)
}

func TestCollectionDeleteDocIDMappingsRemovesDocAliases(t *testing.T) {
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

	publicDocID := doc.ID().String()
	compositeBlock := loadTestBlock(t, ctx, db, doc.Head())
	require.NotEmpty(t, compositeBlock.Links)
	genesisFieldCID := compositeBlock.Links[0].Cid

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	dbTxn, ok := txn.(*Txn)
	require.True(t, ok)
	txnCtx := InitContext(ctx, dbTxn)

	collectionShortID, err := id.GetShortCollectionID(txnCtx, col.CollectionID())
	require.NoError(t, err)
	docShortID, found, err := id.GetShortDocID(txnCtx, collectionShortID, publicDocID)
	require.NoError(t, err)
	require.True(t, found)

	const legacyDocID = "bae-legacy-doc"
	require.NoError(t, id.SetDocIDToDocRefMapping(txnCtx, collectionShortID, docShortID, legacyDocID))

	require.NoError(t, id.DeleteDocIDMappings(txnCtx, dbTxn.Systemstore(), docShortID))

	_, found, err = id.GetShortDocID(txnCtx, collectionShortID, publicDocID)
	require.NoError(t, err)
	require.False(t, found)
	_, found, err = id.GetDocID(txnCtx, docShortID)
	require.NoError(t, err)
	require.False(t, found)
	_, found, err = id.GetDocRef(txnCtx, publicDocID)
	require.NoError(t, err)
	require.False(t, found)
	_, found, err = id.GetDocRef(txnCtx, legacyDocID)
	require.NoError(t, err)
	require.False(t, found)

	blockDocID, found, err := id.GetDocIDForBlockFromStore(
		txnCtx,
		dbTxn.Systemstore(),
		genesisFieldCID,
	)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, publicDocID, blockDocID)
}

func TestCollectionTruncateDeletesUnmappedStorageDoc(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddCollection(ctx, userDocIDTestSchema)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	c, ok := col.(*collection)
	require.True(t, ok)

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	defer txn.Discard()
	dbTxn, ok := txn.(*Txn)
	require.True(t, ok)
	txnCtx := InitContext(ctx, dbTxn)

	collectionShortID, err := id.GetShortCollectionID(txnCtx, col.CollectionID())
	require.NoError(t, err)

	key := keys.DataStoreKey{
		CollectionShortID: collectionShortID,
		InstanceType:      keys.ValueKey,
		DocShortID:        99,
		FieldID:           "1",
	}
	require.NoError(t, dbTxn.Datastore().Set(txnCtx, key, []byte("value")))
	require.NoError(t, txn.Commit())

	txn, err = db.NewTxn(false)
	require.NoError(t, err)
	dbTxn, ok = txn.(*Txn)
	require.True(t, ok)
	txnCtx = InitContext(ctx, dbTxn)
	require.NoError(t, c.hardDeleteDocKeysAndHeadstore(txnCtx, collectionShortID))
	require.NoError(t, txn.Commit())

	txn, err = db.NewTxn(true)
	require.NoError(t, err)
	defer txn.Discard()
	dbTxn, ok = txn.(*Txn)
	require.True(t, ok)
	txnCtx = InitContext(ctx, dbTxn)
	hasValue, err := dbTxn.Datastore().Has(txnCtx, key)
	require.NoError(t, err)
	require.False(t, hasValue)
}

func TestCollectionTruncateDeletesStorageKeyWithoutDocID(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddCollection(ctx, userDocIDTestSchema)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	dbTxn, ok := txn.(*Txn)
	require.True(t, ok)
	txnCtx := InitContext(ctx, dbTxn)

	collectionShortID, err := id.GetShortCollectionID(txnCtx, col.CollectionID())
	require.NoError(t, err)

	malformedKey := keys.DataStoreKey{
		CollectionShortID: collectionShortID,
		InstanceType:      keys.ValueKey,
	}
	require.NoError(t, dbTxn.Datastore().Set(txnCtx, malformedKey, []byte("value")))
	require.NoError(t, txn.Commit())

	err = col.Truncate(ctx)
	require.NoError(t, err)

	txn, err = db.NewTxn(true)
	require.NoError(t, err)
	defer txn.Discard()
	dbTxn, ok = txn.(*Txn)
	require.True(t, ok)
	txnCtx = InitContext(ctx, dbTxn)
	hasValue, err := dbTxn.Datastore().Has(txnCtx, malformedKey)
	require.NoError(t, err)
	require.False(t, hasValue)
}
