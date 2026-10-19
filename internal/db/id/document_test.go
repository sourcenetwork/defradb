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
	"fmt"
	"testing"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/lock"
	"github.com/sourcenetwork/defradb/internal/keys"
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
		docShortID        uint64 = 7
		docID                    = "bae-public-doc"
	)

	_, found, err := GetDocID(ctx, docShortID)
	require.NoError(t, err)
	require.False(t, found)

	_, found, err = GetDocShortID(ctx, collectionShortID, docID)
	require.NoError(t, err)
	require.False(t, found)

	_, found, err = GetDocRef(ctx, docID)
	require.NoError(t, err)
	require.False(t, found)

	undefinedCID := blocks.NewBlock(nil).Cid()
	mappedDocIDs, err := GetDocIDsForBlockFromStore(ctx, txn.Systemstore(), undefinedCID)
	require.NoError(t, err)
	require.Empty(t, mappedDocIDs)
}

func TestDocIDMappingRoundTrip(t *testing.T) {
	ctx := context.Background()
	txn := newDocumentIDTestTxn(ctx)
	defer txn.Discard()
	ctx = datastore.CtxSetTxn(ctx, txn)

	const (
		collectionShortID uint32 = 42
		docShortID        uint64 = 7
		docID                    = "bae-public-doc"
		legacyDocID              = "bae-legacy-doc"
	)

	err := SetDocIDMapping(ctx, collectionShortID, docShortID, docID)
	require.NoError(t, err)

	gotDocID, found, err := GetDocID(ctx, docShortID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, docID, gotDocID)

	gotDocShortID, found, err := GetDocShortID(ctx, collectionShortID, docID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, docShortID, gotDocShortID)

	gotDocRef, found, err := GetDocRef(ctx, docID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, collectionShortID, gotDocRef.CollectionShortID)
	require.Equal(t, docShortID, gotDocRef.DocShortID)

	err = SetDocIDToDocRefMapping(ctx, collectionShortID, docShortID, legacyDocID)
	require.NoError(t, err)

	gotDocRef, found, err = GetDocRef(ctx, legacyDocID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, collectionShortID, gotDocRef.CollectionShortID)
	require.Equal(t, docShortID, gotDocRef.DocShortID)

	gotDocID, found, err = GetDocID(ctx, gotDocRef.DocShortID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, docID, gotDocID)
}

func TestGetDocShortIDDoesNotCrossCollections(t *testing.T) {
	ctx := context.Background()
	txn := newDocumentIDTestTxn(ctx)
	defer txn.Discard()
	ctx = datastore.CtxSetTxn(ctx, txn)

	const (
		collectionShortID uint32 = 42
		otherCollectionID uint32 = 43
		docShortID        uint64 = 7
		docID                    = "bae-public-doc"
		legacyDocID              = "bae-legacy-doc"
	)

	require.NoError(t, SetDocIDMapping(ctx, collectionShortID, docShortID, docID))
	require.NoError(t, SetDocIDToDocRefMapping(ctx, collectionShortID, docShortID, legacyDocID))

	gotDocRef, found, err := GetDocRef(ctx, docID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, collectionShortID, gotDocRef.CollectionShortID)
	require.Equal(t, docShortID, gotDocRef.DocShortID)

	_, found, err = GetDocShortID(ctx, otherCollectionID, docID)
	require.NoError(t, err)
	require.False(t, found)

	_, found, err = GetDocShortID(ctx, otherCollectionID, legacyDocID)
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
		docShortID1       uint64 = 7
		docShortID2       uint64 = 8
		docID1                   = "bae-doc-one"
		docID2                   = "bae-doc-two"
	)
	fieldCID := blocks.NewBlock([]byte("field value")).Cid()

	require.NoError(t, SetDocIDMapping(ctx, collectionShortID, docShortID1, docID1))
	require.NoError(t, SetDocIDMapping(ctx, collectionShortID, docShortID2, docID2))

	err := SetBlockDocIDMapping(ctx, fieldCID, docID1)
	require.NoError(t, err)
	err = SetBlockDocIDMapping(ctx, fieldCID, docID2)
	require.NoError(t, err)

	// A block CID can be owned by more than one document.
	docIDs, err := GetDocIDsForBlockFromStore(ctx, txn.Systemstore(), fieldCID)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{docID1, docID2}, docIDs)

	err = DeleteBlockDocIDMapping(ctx, txn.Systemstore(), fieldCID, docID1)
	require.NoError(t, err)

	docIDs, err = GetDocIDsForBlockFromStore(ctx, txn.Systemstore(), fieldCID)
	require.NoError(t, err)
	require.Equal(t, []string{docID2}, docIDs)

	err = DeleteBlockDocIDMapping(ctx, txn.Systemstore(), fieldCID, docID2)
	require.NoError(t, err)

	docIDs, err = GetDocIDsForBlockFromStore(ctx, txn.Systemstore(), fieldCID)
	require.NoError(t, err)
	require.Empty(t, docIDs)

	err = SetBlockDocIDMapping(ctx, fieldCID, "")
	require.NoError(t, err)
	docIDs, err = GetDocIDsForBlockFromStore(ctx, txn.Systemstore(), fieldCID)
	require.NoError(t, err)
	require.Empty(t, docIDs)
}

func TestBlockHasOwners(t *testing.T) {
	ctx := context.Background()
	txn := newDocumentIDTestTxn(ctx)
	defer txn.Discard()
	ctx = datastore.CtxSetTxn(ctx, txn)

	fieldCID := blocks.NewBlock([]byte("field value")).Cid()

	has, err := BlockHasOwners(ctx, txn.Systemstore(), fieldCID)
	require.NoError(t, err)
	require.False(t, has, "a block with no recorded owners has none")

	has, err = BlockHasOwners(ctx, txn.Systemstore(), cid.Undef)
	require.NoError(t, err)
	require.False(t, has, "an undefined CID has no owners")

	require.NoError(t, SetBlockDocIDMapping(ctx, fieldCID, "bae-doc-one"))
	require.NoError(t, SetBlockDocIDMapping(ctx, fieldCID, "bae-doc-two"))

	has, err = BlockHasOwners(ctx, txn.Systemstore(), fieldCID)
	require.NoError(t, err)
	require.True(t, has)

	// One of two owners removed: the block is still owned.
	require.NoError(t, DeleteBlockDocIDMapping(ctx, txn.Systemstore(), fieldCID, "bae-doc-one"))
	has, err = BlockHasOwners(ctx, txn.Systemstore(), fieldCID)
	require.NoError(t, err)
	require.True(t, has)

	require.NoError(t, DeleteBlockDocIDMapping(ctx, txn.Systemstore(), fieldCID, "bae-doc-two"))
	has, err = BlockHasOwners(ctx, txn.Systemstore(), fieldCID)
	require.NoError(t, err)
	require.False(t, has)
}

func TestBlockHasOwnersStopsAtFirstOwner(t *testing.T) {
	ctx := context.Background()
	txn := newDocumentIDTestTxn(ctx)
	defer txn.Discard()
	ctx = datastore.CtxSetTxn(ctx, txn)

	fieldCID := blocks.NewBlock([]byte("shared field value")).Cid()
	const owners = 500
	for i := range owners {
		require.NoError(t, SetBlockDocIDMapping(ctx, fieldCID, fmt.Sprintf("bae-doc-%d", i)))
	}

	var hasReads int
	has, err := BlockHasOwners(ctx, countingReader{txn.Systemstore(), &hasReads}, fieldCID)
	require.NoError(t, err)
	require.True(t, has)
	require.Equal(t, 1, hasReads, "BlockHasOwners must stop at the first owner")

	var listReads int
	all, err := GetDocIDsForBlockFromStore(ctx, countingReader{txn.Systemstore(), &listReads}, fieldCID)
	require.NoError(t, err)
	require.Len(t, all, owners)
	require.Greater(t, listReads, owners, "GetDocIDsForBlockFromStore reads every owner")
}

// countingReader counts iterator advances.
type countingReader struct {
	corekv.Reader
	nextCalls *int
}

func (r countingReader) Iterator(ctx context.Context, opts corekv.IterOptions) (corekv.Iterator, error) {
	iter, err := r.Reader.Iterator(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &countingIterator{Iterator: iter, nextCalls: r.nextCalls}, nil
}

type countingIterator struct {
	corekv.Iterator
	nextCalls *int
}

func (it *countingIterator) Next() (bool, error) {
	*it.nextCalls++
	return it.Iterator.Next()
}

func TestDeleteDocRefMappings(t *testing.T) {
	ctx := context.Background()
	txn := newDocumentIDTestTxn(ctx)
	defer txn.Discard()
	ctx = datastore.CtxSetTxn(ctx, txn)

	const (
		collectionShortID uint32 = 42
		docShortID        uint64 = 7
		otherDocShortID   uint64 = 8
		docID                    = "bae-public-doc"
		legacyDocID              = "bae-legacy-doc"
		otherDocID               = "bae-other-doc"
	)

	require.NoError(t, SetDocIDToDocRefMapping(ctx, collectionShortID, docShortID, docID))
	require.NoError(t, SetDocIDToDocRefMapping(ctx, collectionShortID, docShortID, legacyDocID))
	require.NoError(t, SetDocIDToDocRefMapping(ctx, collectionShortID, otherDocShortID, otherDocID))
	aliases, err := GetDocIDAliasesFromStore(ctx, txn.Systemstore(), docShortID)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{docID, legacyDocID}, aliases)

	err = DeleteDocRefMappings(ctx, txn.Systemstore(), docShortID)
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

func TestGetDocIDAliasesFromStoreIncludesPrimaryWithoutAliasMapping(t *testing.T) {
	ctx := context.Background()
	txn := newDocumentIDTestTxn(ctx)
	defer txn.Discard()
	ctx = datastore.CtxSetTxn(ctx, txn)

	const docShortID = uint64(1)
	const docID = "bae-primary"
	require.NoError(t, txn.Systemstore().Set(
		ctx,
		keys.NewDocShortIDToDocIDKey(docShortID).Bytes(),
		[]byte(docID),
	))

	docIDs, err := GetDocIDAliasesFromStore(ctx, txn.Systemstore(), docShortID)
	require.NoError(t, err)
	require.Equal(t, []string{docID}, docIDs)
}
