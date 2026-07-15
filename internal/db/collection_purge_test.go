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
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/badger"

	"github.com/sourcenetwork/defradb/client"
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

	desc, err := col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
		Unique: true,
	})
	require.NoError(t, err)

	doc := addUserDoc(t, ctx, col, "alice")

	shortID := getCollectionShortID(t, ctx, db, col.Version().CollectionID)
	require.Equal(t, 1, countIndexEntries(t, ctx, db, shortID, desc.ID))

	require.NoError(t, col.PurgeByDocIDs(ctx, []client.DocID{doc.ID()}, false))

	require.Equal(t, 0, countIndexEntries(t, ctx, db, shortID, desc.ID),
		"purge must delete the document's index entries")
}
