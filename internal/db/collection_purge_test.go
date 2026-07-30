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
	"fmt"
	"testing"

	badgerds "github.com/dgraph-io/badger/v4"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/badger"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
	"github.com/sourcenetwork/defradb/internal/db/id"
)

// newBadgerDBWithMemTableSize builds an in-memory test DB with a reduced memtable. Badger's
// per-transaction size limit is a fixed fraction of the memtable, so shrinking it lets a
// purge exceed one transaction with a modest number of documents.
func newBadgerDBWithMemTableSize(ctx context.Context, memTableSize int64) (*DB, error) {
	// The value threshold must sit below the memtable-derived max batch size (badger
	// validates this on open) yet above the largest value the store holds inline, since an
	// in-memory badger has no value log to offload larger values to.
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

// TestPurgeByDocIDsChunksPurgeOverTransactionLimit verifies that a purge too large for a
// single transaction succeeds by committing in chunks. The guard first confirms the set
// really does exceed one transaction, so the success below is meaningful rather than a set
// that would have fit anyway.
func TestPurgeByDocIDsChunksPurgeOverTransactionLimit(t *testing.T) {
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

	// Guard: run the purge inside a caller transaction, which cannot chunk, to confirm the
	// document set really does exceed badger's per-transaction limit at this memtable size.
	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	dbTxn, ok := txn.(*Txn)
	require.True(t, ok)
	guardErr := col.PurgeByDocIDs(InitContext(ctx, dbTxn), docIDs, false)
	require.ErrorContains(t, guardErr, "Txn is too big")
	txn.Discard()

	// Without a caller transaction the purge commits per chunk and completes.
	require.NoError(t, col.PurgeByDocIDs(ctx, docIDs, false))

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

// TestPurgeByDocIDsRemovesIndexEntries verifies a purge removes the pruned document's
// secondary-index entries, so re-indexing the same document later does not collide with a
// stale unique entry.
func TestPurgeByDocIDsRemovesIndexEntries(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)
	purgeCol, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	desc, err := col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
		Unique: true,
	})
	require.NoError(t, err)

	doc := addUserDoc(t, ctx, col, "alice")

	shortID := getCollectionShortID(t, ctx, db, col.Version().CollectionID)
	require.Equal(t, 1, countIndexEntries(t, ctx, db, shortID, desc.ID))

	require.NoError(t, purgeCol.PurgeByDocIDs(ctx, []client.DocID{doc.ID()}, false))

	require.Equal(t, 0, countIndexEntries(t, ctx, db, shortID, desc.ID),
		"purge must delete the document's index entries")
	addUserDoc(t, ctx, col, "alice")
}

func TestPurgeByDocIDsRemovesSoftDeletedDocument(t *testing.T) {
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

	require.NoError(t, col.PurgeByDocIDs(ctx, []client.DocID{doc.ID()}, true))
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

func TestPurgeByDocIDsUsesCallerTransaction(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	discarded := addUserDoc(t, ctx, col, "discarded")
	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	require.NoError(t, col.PurgeByDocIDs(InitContext(ctx, txn), []client.DocID{discarded.ID()}, false))
	txn.Discard()

	shortID := getCollectionShortID(t, ctx, db, col.Version().CollectionID)
	readTxn, err := db.NewTxn(true)
	require.NoError(t, err)
	_, found, err := id.GetDocShortID(InitContext(ctx, readTxn), shortID, discarded.ID().String())
	require.NoError(t, err)
	require.True(t, found)
	readTxn.Discard()

	committed := addUserDoc(t, ctx, col, "committed")
	txn, err = db.NewTxn(false)
	require.NoError(t, err)
	require.NoError(t, col.PurgeByDocIDs(InitContext(ctx, txn), []client.DocID{committed.ID()}, false))
	require.NoError(t, txn.Commit())

	readTxn, err = db.NewTxn(true)
	require.NoError(t, err)
	defer readTxn.Discard()
	_, found, err = id.GetDocShortID(InitContext(ctx, readTxn), shortID, committed.ID().String())
	require.NoError(t, err)
	require.False(t, found)
}

func TestPurgeByDocIDsPrunesBlocksCreatedInCallerTransaction(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	txnCtx := InitContext(ctx, txn)

	doc, err := client.NewDocFromJSON(ctx, []byte(`{"name":"alice"}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.AddDocument(txnCtx, doc))
	require.NoError(t, col.PurgeByDocIDs(txnCtx, []client.DocID{doc.ID()}, true))
	require.NoError(t, txn.Commit())

	blockstore := datastore.BlockstoreFrom(db.rootstore, db.blockStoreChunkSize)
	requireBlockPresent(t, ctx, blockstore, doc.Head(), false)
}

// addSharedFieldDocs adds two User documents with the same name, so they share the name field
// block, and different ages, so their age field and composite blocks differ.
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

// sharedFieldBlock returns the one field block that the documents with composite heads headA and
// headB have in common, failing if they do not share exactly one. Field blocks are
// content-addressed and carry no document identity, so an identical field value yields a shared
// block owned by both documents.
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

func requireBlockPresent(t *testing.T, ctx context.Context, bs datastore.Blockstore, blockCID cid.Cid, want bool) {
	t.Helper()
	_, found, err := getBlock(ctx, bs, blockCID)
	require.NoError(t, err)
	require.Equal(t, want, found)
}

// TestPurgeByDocIDsPruneHistoryKeepsBlockOwnedByAnotherDoc checks that pruning one document's
// history keeps a field block a second document still owns, and removes it once the second
// document is pruned too. The two purges run in separate transactions, so the second reads the
// first's edge deletion from committed state.
func TestPurgeByDocIDsPruneHistoryKeepsBlockOwnedByAnotherDoc(t *testing.T) {
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

	require.NoError(t, col.PurgeByDocIDs(ctx, []client.DocID{docA.ID()}, true))
	requireBlockPresent(t, ctx, bs, shared, true)

	require.NoError(t, col.PurgeByDocIDs(ctx, []client.DocID{docB.ID()}, true))
	requireBlockPresent(t, ctx, bs, shared, false)
}

// TestPurgeByDocIDsPruneHistoryRemovesBlockWhenAllOwnersPurgedTogether checks the same block is
// removed when both owners are purged in one chunk. Here the first document's edge deletion is
// still uncommitted when the second is checked, so the per-chunk owner tracking is what lets the
// shared block be recognised as unowned rather than leaked.
func TestPurgeByDocIDsPruneHistoryRemovesBlockWhenAllOwnersPurgedTogether(t *testing.T) {
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

	require.NoError(t, col.PurgeByDocIDs(ctx, []client.DocID{docA.ID(), docB.ID()}, true))
	requireBlockPresent(t, ctx, bs, shared, false)
}
