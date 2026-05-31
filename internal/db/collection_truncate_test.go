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
	shortDocID, found, err := id.GetShortDocID(txnCtx, collectionShortID, publicDocID)
	require.NoError(t, err)
	require.True(t, found)

	docIDIndexKey := keys.NewIndexDataStoreKey(
		collectionShortID,
		keys.DocIDIndexID,
		[]keys.IndexedField{
			{Value: client.NewNormalString(publicDocID)},
			{Value: client.NewNormalString(shortDocID)},
		},
	)
	_, err = dbTxn.Datastore().Get(txnCtx, &docIDIndexKey)
	require.NoError(t, err)

	docIDs, err := id.GetPublicDocIDsForGenesisFieldFromStore(
		txnCtx,
		dbTxn.Systemstore(),
		collectionShortID,
		genesisFieldCID,
	)
	require.NoError(t, err)
	require.Equal(t, []string{publicDocID}, docIDs)
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

	_, found, err = id.GetPublicDocID(txnCtx, collectionShortID, shortDocID)
	require.NoError(t, err)
	require.False(t, found)

	_, found, err = id.GetNodeShortDocID(txnCtx, publicDocID)
	require.NoError(t, err)
	require.False(t, found)

	_, found, err = id.GetNodePublicDocID(txnCtx, publicDocID)
	require.NoError(t, err)
	require.False(t, found)

	docIDs, err = id.GetPublicDocIDsForGenesisFieldFromStore(
		txnCtx,
		dbTxn.Systemstore(),
		collectionShortID,
		genesisFieldCID,
	)
	require.NoError(t, err)
	require.Empty(t, docIDs)

	_, err = dbTxn.Datastore().Get(txnCtx, &docIDIndexKey)
	require.Error(t, err)
}
