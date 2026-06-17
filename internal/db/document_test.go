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

const userDocIDTestSchema = `
	type User {
		name: String
		age: Int
	}`

func TestDocumentAdd_DerivesPublicDocIDFromCompositeCID(t *testing.T) {
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

	legacyDocID := doc.ID().String()
	err = col.AddDocument(ctx, doc)
	require.NoError(t, err)

	expectedDocID := client.NewDocIDV0(doc.Head()).String()
	require.Equal(t, expectedDocID, doc.ID().String())
	require.NotEqual(t, legacyDocID, doc.ID().String())

	gotDoc, err := col.GetDocument(ctx, doc.ID())
	require.NoError(t, err)
	require.Equal(t, doc.ID().String(), gotDoc.ID().String())

	c, ok := col.(*collection)
	require.True(t, ok)
	txn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer txn.Discard()
	txnCtx := InitContext(ctx, txn)

	primaryKey, err := c.getPrimaryKeyFromDocID(txnCtx, doc.ID())
	require.NoError(t, err)
	require.NotEqual(t, doc.ID().String(), primaryKey.DocShortID)

	publicDocID, err := c.getPublicDocIDFromPrimaryKey(txnCtx, primaryKey)
	require.NoError(t, err)
	require.Equal(t, doc.ID().String(), publicDocID)
}

func TestUnsignedGenesisProducesEqualCIDAcrossNodes(t *testing.T) {
	ctx := context.Background()
	dbA, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer dbA.Close()
	dbB, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer dbB.Close()

	_, err = dbA.AddCollection(ctx, userDocIDTestSchema)
	require.NoError(t, err)
	_, err = dbB.AddCollection(ctx, userDocIDTestSchema)
	require.NoError(t, err)

	colA, err := dbA.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	colB, err := dbB.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	dbB.docIDSequence.Store(100)

	docA, err := client.NewDocFromJSON(ctx, []byte(`{"name":"Alice","age":40}`), colA.Version())
	require.NoError(t, err)
	docB, err := client.NewDocFromJSON(ctx, []byte(`{"name":"Alice","age":40}`), colB.Version())
	require.NoError(t, err)

	err = colA.AddDocument(ctx, docA)
	require.NoError(t, err)
	err = colB.AddDocument(ctx, docB)
	require.NoError(t, err)

	require.Equal(t, docA.Head(), docB.Head())
	require.Equal(t, docA.ID(), docB.ID())

	txnA, err := dbA.NewTxn(true)
	require.NoError(t, err)
	defer txnA.Discard()
	txnB, err := dbB.NewTxn(true)
	require.NoError(t, err)
	defer txnB.Discard()

	ctxA := InitContext(ctx, txnA)
	ctxB := InitContext(ctx, txnB)
	collectionShortIDA, err := id.GetShortCollectionID(ctxA, colA.CollectionID())
	require.NoError(t, err)
	collectionShortIDB, err := id.GetShortCollectionID(ctxB, colB.CollectionID())
	require.NoError(t, err)

	shortIDA, found, err := id.GetShortDocID(ctxA, collectionShortIDA, docA.ID().String())
	require.NoError(t, err)
	require.True(t, found)
	shortIDB, found, err := id.GetShortDocID(ctxB, collectionShortIDB, docB.ID().String())
	require.NoError(t, err)
	require.True(t, found)
	require.NotEqual(t, shortIDA, shortIDB)
}

func TestSeedDocIDSequenceRestoresPrimaryKeyWithSlashEncodedShortID(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	const slashEncodedShortID uint32 = 303
	require.Contains(t, keys.EncodeDocShortID(slashEncodedShortID), byte('/'))

	_, err = db.AddCollection(ctx, userDocIDTestSchema)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	db.docIDSequence.Store(uint64(slashEncodedShortID - 1))

	doc, err := client.NewDocFromJSON(ctx, []byte(`{"name":"Alice","age":40}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.AddDocument(ctx, doc))

	db.docIDSequence.Store(0)
	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	txnCtx := InitContext(ctx, txn)
	require.NoError(t, db.seedDocIDSequence(txnCtx))
	require.NoError(t, txn.Commit())

	nextShortID, err := db.nextShortDocID()
	require.NoError(t, err)
	require.Equal(t, slashEncodedShortID+1, nextShortID)
}
