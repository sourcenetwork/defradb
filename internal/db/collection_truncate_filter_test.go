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
	"crypto/ed25519"
	"fmt"
	"testing"

	badgerds "github.com/dgraph-io/badger/v4"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/badger"
	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	defraCrypto "github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/internal/datastore"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
	"github.com/sourcenetwork/defradb/internal/db/id"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// Badger derives its transaction limit from the memtable size.
func newBadgerDBWithMemTableSize(ctx context.Context, memTableSize int64) (*DB, error) {
	// In-memory Badger has no value log for values above this threshold.
	rootstore, err := badger.NewDatastore(
		"",
		badgerds.DefaultOptions("").
			WithInMemory(true).
			WithMemTableSize(memTableSize).
			WithValueThreshold(1<<17),
	)
	if err != nil {
		return nil, err
	}

	adminInfo, err := acpDB.NewNACInfo(ctx, "", false)
	if err != nil {
		return nil, err
	}

	return newDB(ctx, rootstore, adminInfo)
}

func truncateDocuments(
	db *DB,
	ctx context.Context,
	collectionName string,
	docIDs []client.DocID,
	pruneHistory bool,
) error {
	col, err := db.GetCollectionByName(ctx, collectionName)
	if err != nil {
		return err
	}

	ids := make([]any, len(docIDs))
	for i, docID := range docIDs {
		ids[i] = docID.String()
	}
	truncateOpts := options.TruncateCollection().SetFilter(map[string]any{
		"_docID": map[string]any{"_in": ids},
	}).
		SetPruneHistory(pruneHistory)
	return col.Truncate(ctx, truncateOpts)
}

func TestTruncateWithFilterProcessesBatchOverTransactionLimit(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDBWithMemTableSize(ctx, 1<<21)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddCollection(ctx, userDocIDTestSchema)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	const docCount = 500
	docIDs := make([]client.DocID, 0, docCount)
	for i := range docCount {
		doc, err := client.NewDocFromJSON(
			ctx,
			fmt.Appendf(nil, `{"name":"user-%d","age":%d}`, i, i),
			col.Version(),
		)
		require.NoError(t, err)
		require.NoError(t, col.AddDocument(ctx, doc))
		docIDs = append(docIDs, doc.ID())
	}

	// Filtered truncate owns the transactions used to process each batch.
	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	dbTxn, ok := txn.(*Txn)
	require.True(t, ok)
	guardErr := truncateDocuments(db, InitContext(ctx, dbTxn), "User", docIDs, false)
	require.ErrorIs(t, guardErr, ErrFilteredTruncateInTransaction)
	txn.Discard()

	require.NoError(t, truncateDocuments(db, ctx, "User", docIDs, false))

	readTxn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer readTxn.Discard()
	readDBTxn, ok := readTxn.(*Txn)
	require.True(t, ok)
	readCtx := InitContext(ctx, readDBTxn)

	shortID, err := id.GetCollectionShortID(readCtx, col.CollectionID())
	require.NoError(t, err)
	for _, docID := range docIDs {
		_, found, err := id.GetDocShortID(readCtx, shortID, docID.String())
		require.NoError(t, err)
		require.False(t, found)
	}
}

func TestTruncateWithFilterChecksDedicatedNACPermission(t *testing.T) {
	ctx := context.Background()
	owner, err := acpIdentity.Generate(defraCrypto.KeyTypeEd25519)
	require.NoError(t, err)
	requestor, err := acpIdentity.Generate(defraCrypto.KeyTypeEd25519)
	require.NoError(t, err)

	ctx = iIdentity.WithContext(ctx, immutable.Some[acpIdentity.Identity](owner))
	rootstore, err := badger.NewDatastore("", badgerds.DefaultOptions("").WithInMemory(true))
	require.NoError(t, err)
	nacInfo, err := acpDB.NewNACInfo(ctx, "", true)
	require.NoError(t, err)
	db, err := newDB(ctx, rootstore, nacInfo)
	require.NoError(t, err)
	t.Cleanup(db.Close)

	_, err = db.AddCollection(ctx, userDocIDTestSchema, options.AddCollection().SetIdentity(owner))
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User", options.GetCollectionByName().SetIdentity(owner))
	require.NoError(t, err)
	doc, err := client.NewDocFromJSON(ctx, []byte(`{"name":"alice"}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.AddDocument(ctx, doc, options.AddDocument().SetIdentity(owner)))

	filter := map[string]any{"_docID": map[string]any{"_eq": doc.ID().String()}}
	err = col.Truncate(
		ctx,
		options.TruncateCollection().SetFilter(filter).SetIdentity(requestor),
	)
	require.ErrorIs(t, err, client.ErrNotAuthorizedToPerformOperation)
	require.ErrorContains(t, err, "Permission: truncate-collection")
	exists, err := col.ExistsDocument(ctx, doc.ID(), options.ExistsDocument().SetIdentity(owner))
	require.NoError(t, err)
	require.True(t, exists)

	_, err = db.AddNACActorRelationship(
		ctx,
		"admin",
		requestor.DID(),
		options.AddNACActorRelationship().SetIdentity(owner),
	)
	require.NoError(t, err)
	require.NoError(t, col.Truncate(
		ctx,
		options.TruncateCollection().SetFilter(filter).SetIdentity(requestor),
	))
}

func TestTruncateWithFilterPrunesSingleDocumentOverTransactionLimit(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDBWithMemTableSize(ctx, 1<<21)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddCollection(ctx, userDocIDTestSchema)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc := addUserDoc(t, ctx, col, "alice")
	for i := range 3000 {
		require.NoError(t, doc.Set(ctx, "age", i))
		require.NoError(t, col.UpdateDocument(ctx, doc))
	}
	latestHead := doc.Head()

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	err = truncateDocuments(db, InitContext(ctx, txn), "User", []client.DocID{doc.ID()}, true)
	require.ErrorIs(t, err, ErrFilteredTruncateInTransaction)
	txn.Discard()

	require.NoError(t, truncateDocuments(db, ctx, "User", []client.DocID{doc.ID()}, true))
	requireBlockPresent(
		t,
		ctx,
		datastore.BlockstoreFrom(db.rootstore, db.blockStoreChunkSize),
		latestHead,
		false,
	)
	readTxn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer readTxn.Discard()
	readCtx := InitContext(ctx, readTxn)
	shortID, err := id.GetCollectionShortID(readCtx, col.CollectionID())
	require.NoError(t, err)
	_, found, err := id.GetDocShortID(readCtx, shortID, doc.ID().String())
	require.NoError(t, err)
	require.False(t, found)
}

func TestTruncateWithFilterRejectsBranchableHistoryPruning(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddCollection(ctx, `type User @branchable { name: String }`)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	doc := addUserDoc(t, ctx, col, "alice")

	err = truncateDocuments(db, ctx, "User", []client.DocID{doc.ID()}, true)
	require.ErrorIs(t, err, ErrCannotPruneBranchableCollection)

	stored, err := col.GetDocument(ctx, doc.ID())
	require.NoError(t, err)
	name, err := stored.Get("name")
	require.NoError(t, err)
	require.Equal(t, "alice", name)
	requireBlockPresent(
		t,
		ctx,
		datastore.BlockstoreFrom(db.rootstore, db.blockStoreChunkSize),
		doc.Head(),
		true,
	)
}

func TestTruncateWithFilterRemovesPrimaryMarker(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)
	doc := addUserDoc(t, ctx, col, "alice")

	require.NoError(t, truncateDocuments(db, ctx, "User", []client.DocID{doc.ID()}, false))
	require.NoError(t, db.DeleteCollection(ctx, []string{"User"}))
}

func TestTruncateWithFilterWithoutPruningReleasesBlockOwnership(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)
	doc := addUserDoc(t, ctx, col, "alice")
	blockstore := datastore.BlockstoreFrom(db.rootstore, db.blockStoreChunkSize)

	require.NoError(t, truncateDocuments(db, ctx, "User", []client.DocID{doc.ID()}, false))
	requireBlockPresent(t, ctx, blockstore, doc.Head(), true)

	txn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer txn.Discard()
	dbTxn, ok := txn.(*Txn)
	require.True(t, ok)
	owners, err := id.GetDocIDsForBlockFromStore(ctx, dbTxn.Systemstore(), doc.Head())
	require.NoError(t, err)
	require.Empty(t, owners)
}

func TestTruncateWithFilterPrunesHistoryOwnedByAlias(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)
	doc := addUserDoc(t, ctx, col, "alice")

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	txnCtx := InitContext(ctx, txn)
	shortID, err := id.GetCollectionShortID(txnCtx, col.CollectionID())
	require.NoError(t, err)
	docShortID, found, err := id.GetDocShortID(txnCtx, shortID, doc.ID().String())
	require.NoError(t, err)
	require.True(t, found)
	const alias = "bae-alice-alias"
	require.NoError(t, id.SetDocIDToDocRefMapping(txnCtx, shortID, docShortID, alias))
	require.NoError(t, id.SetBlockDocIDMapping(txnCtx, doc.Head(), alias))
	require.NoError(t, txn.Commit())

	require.NoError(t, truncateDocuments(db, ctx, "User", []client.DocID{doc.ID()}, true))
	requireBlockPresent(
		t,
		ctx,
		datastore.BlockstoreFrom(db.rootstore, db.blockStoreChunkSize),
		doc.Head(),
		false,
	)
}

func TestTruncateWithFilterRemovesSearchableEncryptionArtifactsForAliases(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)
	docA := addUserDoc(t, ctx, col, "alice")
	docB := addUserDoc(t, ctx, col, "bob")

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	dbTxn, ok := txn.(*Txn)
	require.True(t, ok)
	txnCtx := InitContext(ctx, txn)
	shortID, err := id.GetCollectionShortID(txnCtx, col.CollectionID())
	require.NoError(t, err)
	docShortID, found, err := id.GetDocShortID(txnCtx, shortID, docA.ID().String())
	require.NoError(t, err)
	require.True(t, found)

	const alias = "bae-alice-alias"
	require.NoError(t, id.SetDocIDToDocRefMapping(txnCtx, shortID, docShortID, alias))
	keysToStore := []keys.DatastoreSE{
		{CollectionShortID: shortID, IndexID: "name", SearchTag: []byte{1}, DocID: docA.ID().String()},
		{CollectionShortID: shortID, IndexID: "name", SearchTag: []byte{2}, DocID: alias},
		{CollectionShortID: shortID, IndexID: "name", SearchTag: []byte{3}, DocID: docB.ID().String()},
	}
	for i := range keysToStore {
		require.NoError(t, dbTxn.Datastore().Set(txnCtx, &keysToStore[i], nil))
	}
	require.NoError(t, txn.Commit())

	require.NoError(t, truncateDocuments(db, ctx, "User", []client.DocID{docA.ID()}, false))

	readTxn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer readTxn.Discard()
	readDBTxn, ok := readTxn.(*Txn)
	require.True(t, ok)
	readCtx := InitContext(ctx, readTxn)
	for i := range keysToStore {
		has, err := readDBTxn.Datastore().Has(readCtx, &keysToStore[i])
		require.NoError(t, err)
		require.Equal(t, i == 2, has)
	}
}

func TestHardDeleteSearchableEncryptionResumesAfterChunk(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)
	dbCol, ok := col.(*collection)
	require.True(t, ok)

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	dbTxn, ok := txn.(*Txn)
	require.True(t, ok)
	txnCtx := InitContext(ctx, txn)
	shortID, err := id.GetCollectionShortID(txnCtx, col.CollectionID())
	require.NoError(t, err)

	const targetDocID = "target"
	keysToStore := []keys.DatastoreSE{
		{CollectionShortID: shortID, IndexID: "name", SearchTag: []byte{1}, DocID: targetDocID},
		{CollectionShortID: shortID, IndexID: "name", SearchTag: []byte{2}, DocID: "keep"},
		{CollectionShortID: shortID, IndexID: "name", SearchTag: []byte{3}, DocID: targetDocID},
		{CollectionShortID: shortID, IndexID: "name", SearchTag: []byte{4}, DocID: targetDocID},
	}
	for i := range keysToStore {
		require.NoError(t, dbTxn.Datastore().Set(txnCtx, &keysToStore[i], nil))
	}
	require.NoError(t, txn.Commit())

	deleteCtx, lockTxn, err := ensureContextTxnShim(ctx, db)
	require.NoError(t, err)
	defer lockTxn.Discard()
	db.lockSet.CollectionLock(lockTxn, shortID)
	require.NoError(t, dbCol.hardDeleteSearchableEncryptionInChunks(
		deleteCtx,
		shortID,
		map[string]struct{}{targetDocID: {}},
		2,
	))
	require.NoError(t, lockTxn.Commit())

	readTxn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer readTxn.Discard()
	readDBTxn, ok := readTxn.(*Txn)
	require.True(t, ok)
	readCtx := InitContext(ctx, readTxn)
	for i := range keysToStore {
		has, err := readDBTxn.Datastore().Has(readCtx, &keysToStore[i])
		require.NoError(t, err)
		require.Equal(t, i == 1, has)
	}
}

func TestTruncateWithFilterRemovesEncryptionBlocks(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)
	doc, err := client.NewDocFromJSON(ctx, []byte(`{"name":"alice"}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.AddDocument(ctx, doc, options.AddDocument().SetEncryptedFields([]string{"name"})))

	_, encryptionCID := encryptedFieldCIDs(t, ctx, db, doc.Head())
	encstore := datastore.EncstoreFrom(db.rootstore)
	has, err := encstore.Has(ctx, encryptionCID)
	require.NoError(t, err)
	require.True(t, has)

	require.NoError(t, truncateDocuments(db, ctx, "User", []client.DocID{doc.ID()}, true))
	has, err = encstore.Has(ctx, encryptionCID)
	require.NoError(t, err)
	require.False(t, has)
}

func TestTruncateWithFilterReleasesSharedEncryptionOwnership(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)
	doc, err := client.NewDocFromJSON(ctx, []byte(`{"name":"alice"}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.AddDocument(ctx, doc, options.AddDocument().SetEncryptedFields([]string{"name"})))
	fieldCID, encryptionCID := encryptedFieldCIDs(t, ctx, db, doc.Head())

	const otherDocID = "bae-shared-owner"
	setupTxn, err := db.NewTxn(false)
	require.NoError(t, err)
	setupCtx := InitContext(ctx, setupTxn)
	require.NoError(t, id.SetBlockDocIDMapping(setupCtx, fieldCID, otherDocID))
	require.NoError(t, id.SetBlockDocIDMapping(setupCtx, encryptionCID, otherDocID))
	require.NoError(t, setupTxn.Commit())

	require.NoError(t, truncateDocuments(db, ctx, "User", []client.DocID{doc.ID()}, true))
	readTxn, err := db.NewTxn(true)
	require.NoError(t, err)
	readDBTxn, ok := readTxn.(*Txn)
	require.True(t, ok)
	readCtx := InitContext(ctx, readTxn)
	owners, err := id.GetDocIDsForBlockFromStore(
		readCtx,
		readDBTxn.Systemstore(),
		encryptionCID,
	)
	require.NoError(t, err)
	require.Equal(t, []string{otherDocID}, owners)
	readTxn.Discard()
	requireBlockPresent(t, ctx, datastore.BlockstoreFrom(db.rootstore, db.blockStoreChunkSize), fieldCID, true)
	encstore := datastore.EncstoreFrom(db.rootstore)
	has, err := encstore.Has(ctx, encryptionCID)
	require.NoError(t, err)
	require.True(t, has)

	dbCol := col.(*collection) //nolint:forcetypeassert
	systemstore := datastore.NewMultistore(db.rootstore, db.lockSet, db.blockStoreChunkSize).Systemstore()
	require.NoError(t, dbCol.deleteBlocks(
		ctx,
		systemstore,
		[]string{otherDocID},
		fieldCID,
		true,
	))
	requireBlockPresent(t, ctx, datastore.BlockstoreFrom(db.rootstore, db.blockStoreChunkSize), fieldCID, false)
	has, err = encstore.Has(ctx, encryptionCID)
	require.NoError(t, err)
	require.False(t, has)
}

func TestTruncateWithFilterKeepsSharedSignatureWithParentBlock(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddCollection(ctx, userDocIDTestSchema)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	_, privateKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	ident, err := acpIdentity.FromPrivateKey(defraCrypto.NewPrivateKey(privateKey))
	require.NoError(t, err)
	addOpts := options.AddDocument().SetIdentity(ident).SetEnableSigning(true)

	docA, err := client.NewDocFromJSON(ctx, []byte(`{"name":"shared","age":1}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.AddDocument(ctx, docA, addOpts))
	docB, err := client.NewDocFromJSON(ctx, []byte(`{"name":"shared","age":2}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.AddDocument(ctx, docB, addOpts))

	shared := sharedFieldBlock(t, ctx, db, docA.Head(), docB.Head())
	sharedBlock := loadTestBlock(t, ctx, db, shared)
	require.NotNil(t, sharedBlock.Signature)
	signatureCID := sharedBlock.Signature.Cid
	blockstore := datastore.BlockstoreFrom(db.rootstore, db.blockStoreChunkSize)

	require.NoError(t, truncateDocuments(db, ctx, "User", []client.DocID{docA.ID()}, true))
	requireBlockPresent(t, ctx, blockstore, shared, true)
	has, err := blockstore.Has(ctx, signatureCID)
	require.NoError(t, err)
	require.True(t, has)
	require.NoError(t, db.VerifySignature(ctx, shared.String(), ident.PublicKey()))

	require.NoError(t, truncateDocuments(db, ctx, "User", []client.DocID{docB.ID()}, true))
	requireBlockPresent(t, ctx, blockstore, shared, false)
	has, err = blockstore.Has(ctx, signatureCID)
	require.NoError(t, err)
	require.False(t, has)
}

func TestTruncateWithFilterRemovesIndexEntries(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	desc, err := col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
		Unique: true,
	})
	require.NoError(t, err)

	doc := addUserDoc(t, ctx, col, "alice")

	shortID := getCollectionShortID(t, ctx, db, col.Version().CollectionID)
	require.Equal(t, 1, countIndexEntries(t, ctx, db, shortID, desc.ID))

	require.NoError(t, truncateDocuments(db, ctx, "User", []client.DocID{doc.ID()}, false))

	require.Equal(t, 0, countIndexEntries(t, ctx, db, shortID, desc.ID),
		"truncate must delete the document's index entries")
	addUserDoc(t, ctx, col, "alice")
}

func TestTruncateWithFilterRemovesSoftDeletedDocument(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	desc, err := newNameIndex(t, ctx, db, col)
	require.NoError(t, err)

	doc := addUserDoc(t, ctx, col, "alice")
	originalHead := doc.Head()
	deleted, err := col.DeleteDocument(ctx, doc.ID())
	require.NoError(t, err)
	require.True(t, deleted)

	shortID := getCollectionShortID(t, ctx, db, col.Version().CollectionID)
	require.Equal(t, 0, countIndexEntries(t, ctx, db, shortID, desc.ID))

	require.NoError(t, truncateDocuments(db, ctx, "User", []client.DocID{doc.ID()}, true))
	require.Equal(t, 0, countIndexEntries(t, ctx, db, shortID, desc.ID))

	txn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer txn.Discard()
	txnCtx := InitContext(ctx, txn)

	_, found, err := id.GetDocShortID(txnCtx, shortID, doc.ID().String())
	require.NoError(t, err)
	require.False(t, found)

	blockstore := datastore.BlockstoreFrom(db.rootstore, db.blockStoreChunkSize)
	requireBlockPresent(t, ctx, blockstore, originalHead, false)
}

// addSharedFieldDocs creates one shared field block and distinct composite blocks.
func addSharedFieldDocs(t *testing.T, ctx context.Context, col client.Collection) (*client.Document, *client.Document) {
	t.Helper()

	docA, err := client.NewDocFromJSON(ctx, []byte(`{"name":"shared","age":1}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.AddDocument(ctx, docA))

	docB, err := client.NewDocFromJSON(ctx, []byte(`{"name":"shared","age":2}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.AddDocument(ctx, docB))

	return docA, docB
}

func sharedFieldBlock(t *testing.T, ctx context.Context, db *DB, headA, headB cid.Cid) cid.Cid {
	t.Helper()

	inB := make(map[cid.Cid]struct{})
	for _, link := range loadTestBlock(t, ctx, db, headB).Links {
		inB[link.Cid] = struct{}{}
	}

	var shared []cid.Cid
	for _, link := range loadTestBlock(t, ctx, db, headA).Links {
		if _, ok := inB[link.Cid]; ok {
			shared = append(shared, link.Cid)
		}
	}
	require.Len(t, shared, 1, "documents must share exactly one field block")
	return shared[0]
}

func encryptedFieldCIDs(t *testing.T, ctx context.Context, db *DB, head cid.Cid) (cid.Cid, cid.Cid) {
	t.Helper()
	for _, link := range loadTestBlock(t, ctx, db, head).Links {
		fieldBlock := loadTestBlock(t, ctx, db, link.Cid)
		if fieldBlock.Encryption != nil {
			return link.Cid, fieldBlock.Encryption.Cid
		}
	}
	require.FailNow(t, "encrypted field block not found")
	return cid.Undef, cid.Undef
}

func requireBlockPresent(t *testing.T, ctx context.Context, bs datastore.Blockstore, blockCID cid.Cid, want bool) {
	t.Helper()
	_, found, err := getBlock(ctx, bs, blockCID)
	require.NoError(t, err)
	require.Equal(t, want, found)
}

func TestTruncateWithFilterPruneHistoryKeepsBlockOwnedByAnotherDoc(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddCollection(ctx, userDocIDTestSchema)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	docA, docB := addSharedFieldDocs(t, ctx, col)
	shared := sharedFieldBlock(t, ctx, db, docA.Head(), docB.Head())

	bs := datastore.BlockstoreFrom(db.rootstore, db.blockStoreChunkSize)
	requireBlockPresent(t, ctx, bs, shared, true)

	require.NoError(t, truncateDocuments(db, ctx, "User", []client.DocID{docA.ID()}, true))
	requireBlockPresent(t, ctx, bs, shared, true)

	require.NoError(t, truncateDocuments(db, ctx, "User", []client.DocID{docB.ID()}, true))
	requireBlockPresent(t, ctx, bs, shared, false)
}

func TestTruncateWithFilterPruneHistoryRemovesBlockWhenAllOwnersSelected(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddCollection(ctx, userDocIDTestSchema)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	docA, docB := addSharedFieldDocs(t, ctx, col)
	shared := sharedFieldBlock(t, ctx, db, docA.Head(), docB.Head())

	bs := datastore.BlockstoreFrom(db.rootstore, db.blockStoreChunkSize)
	requireBlockPresent(t, ctx, bs, shared, true)

	require.NoError(t, truncateDocuments(db,
		ctx,
		"User",
		[]client.DocID{docA.ID(), docB.ID()},
		true,
	))
	requireBlockPresent(t, ctx, bs, shared, false)
}
