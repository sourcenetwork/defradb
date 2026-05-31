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

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/id"
)

func TestPublicDocIDsForSignatureBlockResolvesGenesisCompositeAndField(t *testing.T) {
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

	compositeBlock := loadTestBlock(t, ctx, db, doc.Head())
	docIDs, err := db.publicDocIDsForSignatureBlock(ctx, doc.Head(), compositeBlock, col)
	require.NoError(t, err)
	require.Equal(t, []string{doc.ID().String()}, docIDs)
	require.NotEmpty(t, compositeBlock.Links)

	fieldCID := compositeBlock.Links[0].Cid
	fieldBlock := loadTestBlock(t, ctx, db, fieldCID)
	require.True(t, fieldBlock.Delta.IsField())

	docIDs, err = db.publicDocIDsForSignatureBlock(ctx, fieldCID, fieldBlock, col)
	require.NoError(t, err)
	require.Equal(t, []string{doc.ID().String()}, docIDs)
}

func TestPublicDocIDsForSignatureBlockMapsStoredShortDocID(t *testing.T) {
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

	txn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer txn.Discard()
	txnCtx := InitContext(ctx, txn)

	collectionShortID, err := id.GetShortCollectionID(txnCtx, col.CollectionID())
	require.NoError(t, err)
	shortDocID, found, err := id.GetShortDocID(txnCtx, collectionShortID, doc.ID().String())
	require.NoError(t, err)
	require.True(t, found)

	block := &coreblock.Block{
		Delta: crdt.NewCRDT(&crdt.DocCompositeDelta{
			DocID:               []byte(shortDocID),
			CollectionVersionID: col.Version().VersionID,
			Status:              client.Active,
		}),
	}

	docIDs, err := db.publicDocIDsForSignatureBlock(ctx, doc.Head(), block, col)
	require.NoError(t, err)
	require.Equal(t, []string{doc.ID().String()}, docIDs)
}

func loadTestBlock(t *testing.T, ctx context.Context, db *DB, blockCID cid.Cid) *coreblock.Block {
	t.Helper()

	rawBlock, err := datastore.BlockstoreFrom(db.rootstore, db.blockStoreChunkSize).Get(ctx, blockCID)
	require.NoError(t, err)

	block, err := coreblock.GetFromBytes(rawBlock.RawData())
	require.NoError(t, err)
	return block
}
