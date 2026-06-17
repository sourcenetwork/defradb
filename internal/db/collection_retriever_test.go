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
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/immutable"
)

func TestCollectionRetrieverResolvesPublicDocIDAliases(t *testing.T) {
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
	const legacyDocID = "bae-legacy-doc"

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	dbTxn, ok := txn.(*Txn)
	require.True(t, ok)
	txnCtx := InitContext(ctx, dbTxn)

	collectionShortID, err := id.GetShortCollectionID(txnCtx, col.CollectionID())
	require.NoError(t, err)
	shortDocID, found, err := id.GetShortDocID(txnCtx, collectionShortID, publicDocID)
	require.NoError(t, err)
	require.True(t, found)

	blockCID := blocks.NewBlock([]byte("encryption-key")).Cid()
	require.NoError(t, id.SetDocIDAlias(txnCtx, collectionShortID, shortDocID, legacyDocID))
	require.NoError(t, id.SetBlockDocIDMapping(txnCtx, collectionShortID, blockCID, publicDocID))
	require.NoError(t, txn.Commit())

	retriever := NewCollectionRetriever(db)

	resolvedDocID, err := retriever.ResolvePublicDocID(ctx, legacyDocID)
	require.NoError(t, err)
	require.Equal(t, publicDocID, resolvedDocID)

	resolvedDocID, err = retriever.ResolvePublicDocID(
		ctx,
		string(keys.EncodeLocalDocID(collectionShortID, shortDocID)),
	)
	require.NoError(t, err)
	require.Equal(t, publicDocID, resolvedDocID)

	resolvedDocID, err = retriever.ResolvePublicDocID(ctx, blockCID.String())
	require.NoError(t, err)
	require.Equal(t, publicDocID, resolvedDocID)

	retrievedCol, err := retriever.RetrieveCollectionFromDocID(
		ctx,
		legacyDocID,
		immutable.None[identity.Identity](),
	)
	require.NoError(t, err)
	require.Equal(t, col.CollectionID(), retrievedCol.CollectionID())
}
