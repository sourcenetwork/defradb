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
)

func TestPublicDocIDForSignatureBlockResolvesGenesisCompositeAndField(t *testing.T) {
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
	docID, err := db.publicDocIDForSignatureBlock(ctx, doc.Head(), compositeBlock, col)
	require.NoError(t, err)
	require.Equal(t, doc.ID().String(), docID)
	require.NotEmpty(t, compositeBlock.Links)

	fieldCID := compositeBlock.Links[0].Cid
	fieldBlock := loadTestBlock(t, ctx, db, fieldCID)
	require.True(t, fieldBlock.Delta.IsField())

	docID, err = db.publicDocIDForSignatureBlock(ctx, fieldCID, fieldBlock, col)
	require.NoError(t, err)
	require.Equal(t, doc.ID().String(), docID)
}

func TestPublicDocIDForSignatureBlockMapsStoredBlockCID(t *testing.T) {
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

	block := &coreblock.Block{
		Delta: crdt.NewCRDT(&crdt.DocCompositeDelta{
			CollectionVersionID: col.Version().VersionID,
			Status:              client.Active,
		}),
	}

	docID, err := db.publicDocIDForSignatureBlock(ctx, doc.Head(), block, col)
	require.NoError(t, err)
	require.Equal(t, doc.ID().String(), docID)
}

func loadTestBlock(t *testing.T, ctx context.Context, db *DB, blockCID cid.Cid) *coreblock.Block {
	t.Helper()

	rawBlock, err := datastore.BlockstoreFrom(db.rootstore, db.blockStoreChunkSize).Get(ctx, blockCID)
	require.NoError(t, err)

	block, err := coreblock.GetFromBytes(rawBlock.RawData())
	require.NoError(t, err)
	return block
}
