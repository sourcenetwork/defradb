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
	"strconv"
	"testing"

	blocks "github.com/ipfs/go-block-format"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/lock"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/immutable"
)

func newDocumentIDTestTxn(ctx context.Context) datastore.Txn {
	rootstore := memory.NewDatastore(ctx)
	return datastore.NewTxnFrom(rootstore, lock.NewLockSet(), 1, false, immutable.None[int]())
}

func TestFormatAndParseShortDocID(t *testing.T) {
	tests := []struct {
		seq       uint64
		formatted string
	}{
		{seq: 0, formatted: "0000000000000000"},
		{seq: 1, formatted: "0000000000000001"},
		{seq: 15, formatted: "000000000000000f"},
		{seq: 16, formatted: "0000000000000010"},
		{seq: 0xffffffffffffffff, formatted: "ffffffffffffffff"},
	}

	for _, test := range tests {
		t.Run(test.formatted, func(t *testing.T) {
			require.Equal(t, test.formatted, FormatShortDocID(test.seq))

			seq, err := ParseShortDocID(test.formatted)
			require.NoError(t, err)
			require.Equal(t, test.seq, seq)
		})
	}

	_, err := ParseShortDocID("")
	require.ErrorIs(t, err, strconv.ErrSyntax)

	_, err = ParseShortDocID("00000000000000000")
	require.ErrorIs(t, err, strconv.ErrSyntax)
}

func TestGenesisDocIDHelpers(t *testing.T) {
	const docID = "bae-public-doc"

	genesisDocID := NewGenesisDocID(docID)
	require.True(t, IsGenesisDocID(genesisDocID))
	require.False(t, IsGenesisDocID(docID))
	require.Equal(t, docID, UnwrapGenesisDocID(genesisDocID))
	require.Equal(t, docID, UnwrapGenesisDocID(docID))
}

func TestDocIDMappingMissingReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	txn := newDocumentIDTestTxn(ctx)
	defer txn.Discard()
	ctx = datastore.CtxSetTxn(ctx, txn)

	const (
		collectionShortID uint32 = 42
		shortDocID               = "0000000000000007"
		publicDocID              = "bae-public-doc"
	)

	_, found, err := GetPublicDocID(ctx, collectionShortID, shortDocID)
	require.NoError(t, err)
	require.False(t, found)

	_, found, err = GetShortDocID(ctx, collectionShortID, publicDocID)
	require.NoError(t, err)
	require.False(t, found)

	_, found, err = GetNodeShortDocID(ctx, publicDocID)
	require.NoError(t, err)
	require.False(t, found)

	_, found, err = ResolveShortDocID(ctx, collectionShortID, publicDocID)
	require.NoError(t, err)
	require.False(t, found)

	_, found, err = GetNodePublicDocID(ctx, publicDocID)
	require.NoError(t, err)
	require.False(t, found)

	aliases, err := GetNodeDocIDAliasesForShortDocID(ctx, txn.Systemstore(), "")
	require.NoError(t, err)
	require.Empty(t, aliases)

	undefinedCID := blocks.NewBlock(nil).Cid()
	docIDs, err := GetPublicDocIDsForGenesisFieldFromStore(ctx, txn.Systemstore(), collectionShortID, undefinedCID)
	require.NoError(t, err)
	require.Empty(t, docIDs)

	docID, found, err := GetPublicDocIDForGenesisFieldFromStore(ctx, txn.Systemstore(), collectionShortID, undefinedCID)
	require.NoError(t, err)
	require.False(t, found)
	require.Empty(t, docID)
}

func TestDocIDMappingRoundTrip(t *testing.T) {
	ctx := context.Background()
	txn := newDocumentIDTestTxn(ctx)
	defer txn.Discard()
	ctx = datastore.CtxSetTxn(ctx, txn)

	const (
		collectionShortID uint32 = 42
		shortDocID               = "0000000000000007"
		publicDocID              = "bae-public-doc"
		legacyDocID              = "bae-legacy-doc"
	)

	err := SetDocIDMapping(ctx, collectionShortID, shortDocID, publicDocID)
	require.NoError(t, err)

	gotPublicDocID, found, err := GetPublicDocID(ctx, collectionShortID, shortDocID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, publicDocID, gotPublicDocID)

	gotShortDocID, found, err := GetShortDocID(ctx, collectionShortID, publicDocID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, shortDocID, gotShortDocID)

	gotShortDocID, found, err = GetNodeShortDocID(ctx, publicDocID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, shortDocID, gotShortDocID)

	gotPublicDocID, found, err = GetNodePublicDocID(ctx, publicDocID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, publicDocID, gotPublicDocID)

	gotShortDocID, found, err = ResolveShortDocID(ctx, collectionShortID, publicDocID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, shortDocID, gotShortDocID)

	err = SetDocIDAlias(ctx, shortDocID, legacyDocID)
	require.NoError(t, err)

	gotPublicDocID, found, err = GetNodePublicDocID(ctx, legacyDocID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, publicDocID, gotPublicDocID)

	aliases, err := GetNodeDocIDAliasesForShortDocID(ctx, txn.Systemstore(), shortDocID)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{publicDocID, legacyDocID}, aliases)

	docIDIndexKey := keys.NewIndexDataStoreKey(
		collectionShortID,
		keys.DocIDIndexID,
		[]keys.IndexedField{
			{Value: client.NewNormalString(publicDocID)},
			{Value: client.NewNormalString(shortDocID)},
		},
	)
	_, err = txn.Datastore().Get(ctx, &docIDIndexKey)
	require.NoError(t, err)
}

func TestDocIDMappingFallsBackToCollectionMappings(t *testing.T) {
	ctx := context.Background()
	txn := newDocumentIDTestTxn(ctx)
	defer txn.Discard()
	ctx = datastore.CtxSetTxn(ctx, txn)

	const (
		collectionShortID uint32 = 42
		shortDocID               = "0000000000000007"
		publicDocID              = "bae-public-doc"
	)

	err := txn.Systemstore().Set(
		ctx,
		keys.NewNodeDocIDToShortIDKey(publicDocID).Bytes(),
		[]byte(shortDocID),
	)
	require.NoError(t, err)
	err = txn.Systemstore().Set(
		ctx,
		keys.NewShortIDToDocIDKey(collectionShortID, shortDocID).Bytes(),
		[]byte(publicDocID),
	)
	require.NoError(t, err)

	gotPublicDocID, found, err := GetNodePublicDocID(ctx, publicDocID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, publicDocID, gotPublicDocID)
}

func TestGenesisFieldDocIDMappings(t *testing.T) {
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

	err := SetGenesisFieldDocIDMapping(ctx, collectionShortID, fieldCID, docID1)
	require.NoError(t, err)
	err = SetGenesisFieldDocIDMapping(ctx, collectionShortID, fieldCID, docID2)
	require.NoError(t, err)

	docIDs, err := GetPublicDocIDsForGenesisFieldFromStore(ctx, txn.Systemstore(), collectionShortID, fieldCID)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{docID1, docID2}, docIDs)

	docID, found, err := GetPublicDocIDForGenesisFieldFromStore(ctx, txn.Systemstore(), collectionShortID, fieldCID)
	require.NoError(t, err)
	require.True(t, found)
	require.Contains(t, []string{docID1, docID2}, docID)

	err = DeleteGenesisFieldDocIDMappings(ctx, txn.Systemstore(), collectionShortID, docID1)
	require.NoError(t, err)

	docIDs, err = GetPublicDocIDsForGenesisFieldFromStore(ctx, txn.Systemstore(), collectionShortID, fieldCID)
	require.NoError(t, err)
	require.Equal(t, []string{docID2}, docIDs)

	err = SetGenesisFieldDocIDMapping(ctx, collectionShortID, fieldCID, "")
	require.NoError(t, err)
	docIDs, err = GetPublicDocIDsForGenesisFieldFromStore(ctx, txn.Systemstore(), collectionShortID, fieldCID)
	require.NoError(t, err)
	require.Equal(t, []string{docID2}, docIDs)
}

func TestDeleteNodeDocIDAliasesForShortDocID(t *testing.T) {
	ctx := context.Background()
	txn := newDocumentIDTestTxn(ctx)
	defer txn.Discard()
	ctx = datastore.CtxSetTxn(ctx, txn)

	const (
		shortDocID      = "0000000000000007"
		otherShortDocID = "0000000000000008"
		publicDocID     = "bae-public-doc"
		legacyDocID     = "bae-legacy-doc"
		otherDocID      = "bae-other-doc"
	)

	require.NoError(t, SetDocIDAlias(ctx, shortDocID, publicDocID))
	require.NoError(t, SetDocIDAlias(ctx, shortDocID, legacyDocID))
	require.NoError(t, SetDocIDAlias(ctx, otherShortDocID, otherDocID))

	err := DeleteNodeDocIDAliasesForShortDocID(ctx, txn.Systemstore(), shortDocID)
	require.NoError(t, err)

	_, found, err := GetNodeShortDocID(ctx, publicDocID)
	require.NoError(t, err)
	require.False(t, found)
	_, found, err = GetNodeShortDocID(ctx, legacyDocID)
	require.NoError(t, err)
	require.False(t, found)

	gotShortDocID, found, err := GetNodeShortDocID(ctx, otherDocID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, otherShortDocID, gotShortDocID)
}
