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

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

// setupUserCollection opens a new DB, registers a cleanup to close it, adds the
// `type User { name: String }` schema, and returns the open DB and User collection.
func setupUserCollection(t *testing.T, ctx context.Context) (*DB, client.Collection) {
	t.Helper()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.AddCollection(ctx, `type User { name: String }`)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	return db, col
}

// addUserDoc creates a User document with the given name and saves it to col.
func addUserDoc(t *testing.T, ctx context.Context, col client.Collection, name string) *client.Document {
	t.Helper()
	doc, err := client.NewDocFromJSON(ctx, fmt.Appendf(nil, `{"name":%q}`, name), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.AddDocument(ctx, doc))
	return doc
}

// readIndexState opens a read-only transaction, reads the index state for the given
// collection and index IDs, and returns it. The transaction is discarded on cleanup.
func readIndexState(t *testing.T, ctx context.Context, db *DB, collectionID string, indexID uint32) indexState {
	t.Helper()
	rawTxn, err := db.NewTxn(true)
	require.NoError(t, err)
	t.Cleanup(func() { rawTxn.Discard() })
	txnCtx := InitContext(ctx, rawTxn)

	state, err := getIndexState(txnCtx, collectionID, indexID)
	require.NoError(t, err)
	return state
}

// requireNoIndexState asserts that no state record exists for the given index.
// A missing record means the index is ready.
func requireNoIndexState(t *testing.T, ctx context.Context, db *DB, collectionID string, indexID uint32) {
	t.Helper()
	rawTxn, err := db.NewTxn(true)
	require.NoError(t, err)
	t.Cleanup(func() { rawTxn.Discard() })
	txnCtx := InitContext(ctx, rawTxn)

	_, err = getIndexState(txnCtx, collectionID, indexID)
	require.True(t, errors.Is(err, corekv.ErrNotFound),
		"expected no state record, got err: %v", err)
}

// queryUserByName executes a filtered query for User documents with the given name
// and returns the result rows. The test fails immediately on any GQL error.
func queryUserByName(t *testing.T, db *DB, ctx context.Context, name string) []map[string]any {
	t.Helper()
	res := db.ExecRequest(ctx, fmt.Sprintf(`query { User(filter: {name: {_eq: %q}}) { name } }`, name))
	require.Empty(t, res.GQL.Errors, "query error for name %q", name)
	data, ok := res.GQL.Data.(map[string]any)
	require.True(t, ok, "unexpected data type for name %q", name)

	v := data["User"]
	slice, ok := v.([]map[string]any)
	if !ok {
		rawSlice, ok2 := v.([]any)
		require.True(t, ok2, "expected slice result, got %T", v)
		slice = make([]map[string]any, len(rawSlice))
		for i, el := range rawSlice {
			m, ok3 := el.(map[string]any)
			require.True(t, ok3, "expected map element at index %d, got %T", i, el)
			slice[i] = m
		}
	}
	return slice
}

// newNameIndex creates an index on the "name" field of col. The error is returned
// to the caller so tests that expect failure can assert on it.
func newNameIndex(t *testing.T, ctx context.Context, col client.Collection) (client.IndexDescription, error) {
	t.Helper()
	return col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
	})
}

// TestBackfillIndex_MultiBatch_IndexesAllDocsAndClearsState creates 10 documents and then
// runs NewIndex with indexBackfillBatchSize overridden to 3.  This exercises 4 batches
// (3+3+3+1) and verifies:
//   - The state record is deleted on completion (missing record means ready).
//   - Every document is reachable through the index (filtered queries succeed).
func TestBackfillIndex_MultiBatch_IndexesAllDocsAndClearsState(t *testing.T) {
	// Override batch size so 10 docs trigger multiple batches.
	origBatchSize := indexBackfillBatchSize
	indexBackfillBatchSize = 3
	defer func() { indexBackfillBatchSize = origBatchSize }()

	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	names := make([]string, 10)
	for i := range 10 {
		names[i] = fmt.Sprintf("name%02d", i)
		addUserDoc(t, ctx, col, names[i])
	}

	desc, err := newNameIndex(t, ctx, col)
	require.NoError(t, err)

	requireNoIndexState(t, ctx, db, col.Version().CollectionID, desc.ID)

	for _, name := range names {
		users := queryUserByName(t, db, ctx, name)
		require.Len(t, users, 1, "expected exactly 1 result for name %q", name)
		assert.Equal(t, name, users[0]["name"], "name mismatch")
	}
}

// TestBackfillIndex_EmptyCollection_ClearsState ensures that creating an index on a
// collection with no documents completes without error and leaves no state record.
func TestBackfillIndex_EmptyCollection_ClearsState(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	desc, err := newNameIndex(t, ctx, col)
	require.NoError(t, err)

	requireNoIndexState(t, ctx, db, col.Version().CollectionID, desc.ID)
}

// TestWithTxnRetries_ConflictThenSuccess verifies that withTxnRetries retries after a
// conflict and succeeds on the second attempt, persisting the work from that attempt.
func TestWithTxnRetries_ConflictThenSuccess(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)
	collectionID := col.Version().CollectionID

	attempts := 0
	err := db.withTxnRetries(ctx, func(txnCtx context.Context) error {
		attempts++
		if attempts == 1 {
			return corekv.ErrTxnConflict
		}
		return setIndexState(txnCtx, collectionID, 1, indexState{Status: client.IndexStatusBuilding})
	})
	require.NoError(t, err)
	assert.Equal(t, 2, attempts)

	state := readIndexState(t, ctx, db, collectionID, 1)
	assert.Equal(t, client.IndexStatusBuilding, state.Status)
}

// TestWithTxnRetries_ConflictEveryAttempt_ReturnsConflict verifies that when every attempt
// conflicts, withTxnRetries exhausts MaxTxnRetries and returns the conflict error.
func TestWithTxnRetries_ConflictEveryAttempt_ReturnsConflict(t *testing.T) {
	ctx := context.Background()
	db, _ := setupUserCollection(t, ctx)

	attempts := 0
	err := db.withTxnRetries(ctx, func(txnCtx context.Context) error {
		attempts++
		return corekv.ErrTxnConflict
	})
	assert.True(t, errors.Is(err, corekv.ErrTxnConflict))
	assert.Equal(t, db.MaxTxnRetries(), attempts)
}

// TestWithTxnRetries_NonRetryableError_NoRetry verifies that a non-conflict error causes
// withTxnRetries to abort immediately with no further retries.
func TestWithTxnRetries_NonRetryableError_NoRetry(t *testing.T) {
	ctx := context.Background()
	db, _ := setupUserCollection(t, ctx)

	sentinel := errors.New("boom")
	attempts := 0
	err := db.withTxnRetries(ctx, func(txnCtx context.Context) error {
		attempts++
		return sentinel
	})
	assert.True(t, errors.Is(err, sentinel))
	assert.Equal(t, 1, attempts)
}

// TestBackfillIndex_ConflictOnEveryAttempt_FailsAndMarksFailed verifies that when
// backfill fails (here via a unique-constraint violation on every attempt), the index
// definition remains listed with a failed state record instead of being rolled back.
func TestBackfillIndex_ConflictOnEveryAttempt_FailsAndMarksFailed(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.AddCollection(ctx, "type User { name: String\n age: Int }")
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	// Add two documents with the same age to trigger a unique violation during backfill.
	doc1, err := client.NewDocFromJSON(ctx, []byte(`{"name":"Alice","age":21}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.AddDocument(ctx, doc1))

	doc2, err := client.NewDocFromJSON(ctx, []byte(`{"name":"Bob","age":21}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.AddDocument(ctx, doc2))

	_, err = col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "age"}},
		Unique: true,
	})

	// Backfill must fail due to the uniqueness violation.
	require.Error(t, err)
	require.Contains(t, err.Error(), "can not index a doc's field(s) that violates unique index")

	// The index definition must still be listed — it was not rolled back.
	indexes, listErr := col.ListIndexes(ctx)
	require.NoError(t, listErr)
	require.Len(t, indexes, 1, "index definition must persist after failed backfill")

	// The state record must reflect the failure with a non-empty reason.
	state := readIndexState(t, ctx, db, col.Version().CollectionID, indexes[0].ID)
	assert.Equal(t, client.IndexStatusFailed, state.Status)
	assert.NotEmpty(t, state.Reason)
}

// TestBackfillBatchTxn_ConflictsWhenReadDocIsModified verifies that a backfill batch
// transaction conflicts when a document it read is updated before commit.
//
// Backfill batches always write (index entries + state), so conflict detection at the
// storage level applies. This test replicates the exact batch body: iterateDocsBatch
// feeds docs into colIndex.Save, which writes an index entry into txn1. A concurrent
// update to the same document must cause txn1.Commit to return ErrTxnConflict.
func TestBackfillBatchTxn_ConflictsWhenReadDocIsModified(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	doc := addUserDoc(t, ctx, col, "old")

	// Build a CollectionVersion that carries a staged index definition so that
	// NewCollectionIndex can resolve the "name" field. The definition is not
	// persisted — this test exercises storage-level conflict behavior, not the API.
	nameDesc := client.IndexDescription{
		Name:   "name_idx",
		ID:     1,
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
	}
	colVersion := col.Version()
	colVersion.Indexes = append(colVersion.Indexes, nameDesc)

	// Open a read-write transaction that will act as the backfill batch transaction.
	rawTxn1, err := db.NewTxn(false)
	require.NoError(t, err)
	txn1 := rawTxn1.(*Txn)
	ctx1 := InitContext(ctx, txn1)

	col1, err := db.newCollection(ctx1, colVersion, immutable.Some[datastore.Txn](txn1))
	require.NoError(t, err)

	colIndex, err := NewCollectionIndex(col1, nameDesc)
	require.NoError(t, err)

	// Replicate the backfill batch body: scan docs and write an index entry for each.
	// This puts the doc's key range in txn1's read set AND produces a write, which
	// is what makes conflict detection fire at commit time.
	fields := col1.Version().CollectIndexedFields()
	_, _, err = col1.iterateDocsBatch(ctx1, fields, immutable.None[string](), 10, func(d *client.Document) error {
		return colIndex.Save(ctx1, d)
	})
	require.NoError(t, err)

	// Update the same document through the public API using the original col handle,
	// which has no index definition, so the update touches only document storage keys
	// and does not attempt to maintain the staged index.
	require.NoError(t, doc.Set(ctx, "name", "new"))
	require.NoError(t, col.UpdateDocument(ctx, doc))

	// txn1 read the doc's storage keys and wrote an index entry; the intervening
	// document update overlaps that read set, so commit must conflict.
	commitErr := txn1.Commit()
	require.True(t, errors.Is(commitErr, corekv.ErrTxnConflict),
		"expected ErrTxnConflict but got: %v", commitErr)
}
