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
)

// Reading one document must not put another document's keys in the transaction's read set.
// Documents sit consecutively in the store, so a scan of one that runs past its end lands on
// the next document, and a concurrent write there then rejects a transaction that never
// touched it.
func TestDocumentFetcher_ReadingOneDocDoesNotConflictWithItsNeighbour(t *testing.T) {
	ctx := context.Background()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.AddCollection(ctx, userSchema)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	first, err := client.NewDocFromJSON(ctx, []byte(`{"name":"alice","age":1}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.AddDocument(ctx, first))

	second, err := client.NewDocFromJSON(ctx, []byte(`{"name":"bob","age":2}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.AddDocument(ctx, second))

	// Read the first document and write it back, all in one transaction.
	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	defer txn.Discard()
	txnCtx := InitContext(ctx, txn)

	read, err := col.GetDocument(txnCtx, first.ID())
	require.NoError(t, err)
	require.NoError(t, read.Set(ctx, "age", 10))
	require.NoError(t, col.UpdateDocument(txnCtx, read))

	// Someone else rewrites the neighbour and commits, in between. Every field is written so
	// the test does not depend on which one holds the lowest field ID.
	other, err := col.GetDocument(ctx, second.ID())
	require.NoError(t, err)
	require.NoError(t, other.Set(ctx, "name", "robert"))
	require.NoError(t, other.Set(ctx, "age", 99))
	require.NoError(t, col.UpdateDocument(ctx, other))

	require.NoError(t, txn.Commit(),
		"reading one document must not conflict with a write to the next one")
}
