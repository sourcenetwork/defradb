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
	"encoding/binary"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/blockowner"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

const userDocIDTestSchema = `
	type User {
		name: String
		age: Int
	}`

const projectDocIDTestSchema = `
	type Project {
		name: String
	}`

func setDocIDSequence(t *testing.T, ctx context.Context, db *DB, value uint64) {
	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	defer txn.Discard()
	dbTxn, ok := txn.(*Txn)
	require.True(t, ok)

	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], value)
	require.NoError(t, dbTxn.Systemstore().Set(ctx, keys.NewDocShortIDSequenceKey().Bytes(), buf[:]))
	require.NoError(t, txn.Commit())
}

func TestDocumentAdd_DerivesDocIDFromCompositeCID(t *testing.T) {
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

	docID, err := c.getDocIDFromPrimaryKey(txnCtx, primaryKey)
	require.NoError(t, err)
	require.Equal(t, doc.ID().String(), docID)
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

	setDocIDSequence(t, ctx, dbB, 100)

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
	collectionShortIDA, err := id.GetCollectionShortID(ctxA, colA.CollectionID())
	require.NoError(t, err)
	collectionShortIDB, err := id.GetCollectionShortID(ctxB, colB.CollectionID())
	require.NoError(t, err)

	shortIDA, found, err := id.GetDocShortID(ctxA, collectionShortIDA, docA.ID().String())
	require.NoError(t, err)
	require.True(t, found)
	shortIDB, found, err := id.GetDocShortID(ctxB, collectionShortIDB, docB.ID().String())
	require.NoError(t, err)
	require.True(t, found)
	require.NotEqual(t, shortIDA, shortIDB)
}

func TestNextShortDocIDUsesNodeSequence(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddCollection(ctx, userDocIDTestSchema)
	require.NoError(t, err)
	_, err = db.AddCollection(ctx, projectDocIDTestSchema)
	require.NoError(t, err)

	setDocIDSequence(t, ctx, db, 10)

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	defer txn.Discard()

	txnCtx := InitContext(ctx, txn)
	nextShortID, err := id.NextDocShortID(txnCtx)
	require.NoError(t, err)
	require.Equal(t, uint64(11), nextShortID)
	nextShortID, err = id.NextDocShortID(txnCtx)
	require.NoError(t, err)
	require.Equal(t, uint64(12), nextShortID)
	nextShortID, err = id.NextDocShortID(txnCtx)
	require.NoError(t, err)
	require.Equal(t, uint64(13), nextShortID)

	require.NoError(t, txn.Commit())
}

func TestBlockDocIDMappingSharedFieldBlockHasAllOwners(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddCollection(ctx, userDocIDTestSchema)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	docA, err := client.NewDocFromJSON(ctx, []byte(`{"name":"Shared","age":1}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.AddDocument(ctx, docA))

	docB, err := client.NewDocFromJSON(ctx, []byte(`{"name":"Shared","age":2}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.AddDocument(ctx, docB))
	require.NotEqual(t, docA.ID().String(), docB.ID().String())

	nameCIDA := nameFieldBlockCID(t, ctx, db, docA.Head())
	nameCIDB := nameFieldBlockCID(t, ctx, db, docB.Head())
	require.Equal(t, nameCIDA, nameCIDB)

	txn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer txn.Discard()
	dbTxn, ok := txn.(*Txn)
	require.True(t, ok)
	txnCtx := InitContext(ctx, dbTxn)

	docIDs, err := blockowner.DocIDs(txnCtx, dbTxn.Systemstore(), nameCIDA)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{docA.ID().String(), docB.ID().String()}, docIDs)
}

func nameFieldBlockCID(t *testing.T, ctx context.Context, db *DB, head cid.Cid) cid.Cid {
	t.Helper()
	composite := loadTestBlock(t, ctx, db, head)
	nameLink, found := composite.GetLinkByName("name")
	require.True(t, found)
	return nameLink.Cid
}
