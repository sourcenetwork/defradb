// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package id

import (
	"context"
	"testing"

	blocks "github.com/ipfs/go-block-format"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/lock"
	"github.com/sourcenetwork/immutable"
)

func newDocumentIDTestTxn(ctx context.Context) datastore.Txn {
	rootstore := memory.NewDatastore(ctx)
	return datastore.NewTxnFrom(rootstore, lock.NewLockSet(), 1, false, immutable.None[int]())
}

func TestDocIDMappingMissingReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	txn := newDocumentIDTestTxn(ctx)
	defer txn.Discard()
	ctx = datastore.CtxSetTxn(ctx, txn)

	const (
		collectionShortID uint32 = 42
		docShortID        uint32 = 7
		docID                    = "bae-public-doc"
	)

	_, found, err := GetDocID(ctx, collectionShortID, docShortID)
	require.NoError(t, err)
	require.False(t, found)

	_, found, err = GetShortDocID(ctx, collectionShortID, docID)
	require.NoError(t, err)
	require.False(t, found)

	_, found, err = GetDocRef(ctx, docID)
	require.NoError(t, err)
	require.False(t, found)

	undefinedCID := blocks.NewBlock(nil).Cid()
	mappedDocID, found, err := GetDocIDForBlockFromStore(ctx, txn.Systemstore(), collectionShortID, undefinedCID)
	require.NoError(t, err)
	require.False(t, found)
	require.Empty(t, mappedDocID)
}

func TestDocIDMappingRoundTrip(t *testing.T) {
	ctx := context.Background()
	txn := newDocumentIDTestTxn(ctx)
	defer txn.Discard()
	ctx = datastore.CtxSetTxn(ctx, txn)

	const (
		collectionShortID uint32 = 42
		docShortID        uint32 = 7
		docID                    = "bae-public-doc"
		legacyDocID              = "bae-legacy-doc"
	)

	err := SetDocIDMapping(ctx, collectionShortID, docShortID, docID)
	require.NoError(t, err)

	gotDocID, found, err := GetDocID(ctx, collectionShortID, docShortID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, docID, gotDocID)

	gotDocShortID, found, err := GetShortDocID(ctx, collectionShortID, docID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, docShortID, gotDocShortID)

	gotDocRef, found, err := GetDocRef(ctx, docID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, collectionShortID, gotDocRef.CollectionShortID)
	require.Equal(t, docShortID, gotDocRef.DocShortID)

	err = SetDocIDAlias(ctx, collectionShortID, docShortID, legacyDocID)
	require.NoError(t, err)

	gotDocRef, found, err = GetDocRef(ctx, legacyDocID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, collectionShortID, gotDocRef.CollectionShortID)
	require.Equal(t, docShortID, gotDocRef.DocShortID)

	gotDocID, found, err = GetDocID(ctx, gotDocRef.CollectionShortID, gotDocRef.DocShortID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, docID, gotDocID)
}

func TestGetShortDocIDDoesNotCrossCollections(t *testing.T) {
	ctx := context.Background()
	txn := newDocumentIDTestTxn(ctx)
	defer txn.Discard()
	ctx = datastore.CtxSetTxn(ctx, txn)

	const (
		collectionShortID uint32 = 42
		otherCollectionID uint32 = 43
		docShortID        uint32 = 7
		docID                    = "bae-public-doc"
		legacyDocID              = "bae-legacy-doc"
	)

	require.NoError(t, SetDocIDMapping(ctx, collectionShortID, docShortID, docID))
	require.NoError(t, SetDocIDAlias(ctx, collectionShortID, docShortID, legacyDocID))

	gotDocRef, found, err := GetDocRef(ctx, docID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, collectionShortID, gotDocRef.CollectionShortID)
	require.Equal(t, docShortID, gotDocRef.DocShortID)

	_, found, err = GetShortDocID(ctx, otherCollectionID, docID)
	require.NoError(t, err)
	require.False(t, found)

	_, found, err = GetShortDocID(ctx, otherCollectionID, legacyDocID)
	require.NoError(t, err)
	require.False(t, found)
}

func TestBlockDocIDMappings(t *testing.T) {
	ctx := context.Background()
	txn := newDocumentIDTestTxn(ctx)
	defer txn.Discard()
	ctx = datastore.CtxSetTxn(ctx, txn)

	const (
		collectionShortID uint32 = 42
		docID1                   = "bae-doc-one"
		docID2                   = "bae-doc-two"
	)
	fieldCID := blocks.NewBlock([]byte("field value")).Cid()

	err := SetBlockDocIDMapping(ctx, collectionShortID, fieldCID, docID1)
	require.NoError(t, err)
	err = SetBlockDocIDMapping(ctx, collectionShortID, fieldCID, docID2)
	require.NoError(t, err)

	docID, found, err := GetDocIDForBlockFromStore(ctx, txn.Systemstore(), collectionShortID, fieldCID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, docID2, docID)

	docID, found, err = GetDocIDForBlockFromStore(ctx, txn.Systemstore(), 0, fieldCID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, docID2, docID)

	err = DeleteBlockDocIDMapping(ctx, txn.Systemstore(), collectionShortID, fieldCID)
	require.NoError(t, err)

	docID, found, err = GetDocIDForBlockFromStore(ctx, txn.Systemstore(), collectionShortID, fieldCID)
	require.NoError(t, err)
	require.False(t, found)
	require.Empty(t, docID)

	err = SetBlockDocIDMapping(ctx, collectionShortID, fieldCID, "")
	require.NoError(t, err)
	docID, found, err = GetDocIDForBlockFromStore(ctx, txn.Systemstore(), collectionShortID, fieldCID)
	require.NoError(t, err)
	require.False(t, found)
	require.Empty(t, docID)
}

func TestDeleteNodeDocIDAliasesForShortDocID(t *testing.T) {
	ctx := context.Background()
	txn := newDocumentIDTestTxn(ctx)
	defer txn.Discard()
	ctx = datastore.CtxSetTxn(ctx, txn)

	const (
		collectionShortID uint32 = 42
		docShortID        uint32 = 7
		otherDocShortID   uint32 = 8
		docID                    = "bae-public-doc"
		legacyDocID              = "bae-legacy-doc"
		otherDocID               = "bae-other-doc"
	)

	require.NoError(t, SetDocIDAlias(ctx, collectionShortID, docShortID, docID))
	require.NoError(t, SetDocIDAlias(ctx, collectionShortID, docShortID, legacyDocID))
	require.NoError(t, SetDocIDAlias(ctx, collectionShortID, otherDocShortID, otherDocID))

	err := DeleteNodeDocIDAliasesForShortDocID(ctx, txn.Systemstore(), collectionShortID, docShortID)
	require.NoError(t, err)

	_, found, err := GetDocRef(ctx, docID)
	require.NoError(t, err)
	require.False(t, found)
	_, found, err = GetDocRef(ctx, legacyDocID)
	require.NoError(t, err)
	require.False(t, found)

	gotDocRef, found, err := GetDocRef(ctx, otherDocID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, collectionShortID, gotDocRef.CollectionShortID)
	require.Equal(t, otherDocShortID, gotDocRef.DocShortID)
}
