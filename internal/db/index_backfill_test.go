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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/sequence"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// setupUserCollection opens a DB with a `User { name: String }` collection, closed on cleanup.
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

// addUserDoc saves a User document with the given name.
func addUserDoc(t *testing.T, ctx context.Context, col client.Collection, name string) *client.Document {
	t.Helper()
	doc, err := client.NewDocFromJSON(ctx, fmt.Appendf(nil, `{"name":%q}`, name), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.AddDocument(ctx, doc))
	return doc
}

// readIndexState returns the stored state for an index. Fails if no record exists.
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

// requireNoIndexState asserts the index has no action record of any kind (build or drop),
// i.e. it is fully ready with nothing in flight.
func requireNoIndexState(t *testing.T, ctx context.Context, db *DB, collectionID string, indexID uint32) {
	t.Helper()
	rawTxn, err := db.NewTxn(true)
	require.NoError(t, err)
	t.Cleanup(func() { rawTxn.Discard() })
	txnCtx := InitContext(ctx, rawTxn)

	records, err := scanIndexStates(txnCtx, indexActionCollectionPrefix(collectionID), false)
	require.NoError(t, err)
	for _, rec := range records {
		require.NotEqual(t, indexID, rec.Key.IndexID, "expected no state record, found %+v", rec.State)
	}
}

// queryUserByName returns the rows from a name-filtered User query.
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

// newNameIndex creates an index on "name", returning the error for the caller to assert.
func newNameIndex(t *testing.T, ctx context.Context, col client.Collection) (client.IndexDescription, error) {
	t.Helper()
	return col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
	})
}

// TestBackfillIndex_MultiBatch_IndexesAllDocsAndClearsState builds an index over 10 docs in
// batches of 3, then checks the state record is cleared and every doc is queryable.
func TestBackfillIndex_MultiBatch_IndexesAllDocsAndClearsState(t *testing.T) {
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

// TestBackfillIndex_EmptyCollection_ClearsState builds an index on an empty collection
// and checks it leaves no state record.
func TestBackfillIndex_EmptyCollection_ClearsState(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	desc, err := newNameIndex(t, ctx, col)
	require.NoError(t, err)

	requireNoIndexState(t, ctx, db, col.Version().CollectionID, desc.ID)
}

// TestWithTxnRetries_ConflictThenSuccess checks a conflict is retried and the second
// attempt's work is persisted.
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
		return db.startIndexBuild(txnCtx, collectionID, 1)
	})
	require.NoError(t, err)
	assert.Equal(t, 2, attempts)

	state := readIndexState(t, ctx, db, collectionID, 1)
	assert.True(t, state.isBuilding())
}

// TestWithTxnRetries_ConflictEveryAttempt_ReturnsConflict checks that persistent conflicts
// exhaust MaxTxnRetries and return the conflict error.
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

// TestWithTxnRetries_NonRetryableError_NoRetry checks a non-conflict error aborts on the
// first attempt with no retry.
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

// TestBackfillIndex_NonRetryableError_MarksFailed checks that a unique-violation backfill
// (non-retryable) leaves the definition listed with a failed state, not rolled back.
func TestBackfillIndex_NonRetryableError_MarksFailed(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.AddCollection(ctx, "type User { name: String\n age: Int }")
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	// Two docs sharing an age violate the unique index during backfill.
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

	require.Error(t, err)
	require.Contains(t, err.Error(), "can not index a doc's field(s) that violates unique index")

	indexes, listErr := col.ListIndexes(ctx)
	require.NoError(t, listErr)
	require.Len(t, indexes, 1, "index definition must persist after failed backfill")

	state := readIndexState(t, ctx, db, col.Version().CollectionID, indexes[0].Description.ID)
	assert.True(t, state.isFailed())
	assert.NotEmpty(t, state.Reason)
}

// TestBackfillIndex_DocUpdatedAfterIndexing_NoStaleEntry checks the write path replaces an
// index entry on update: after "old"→"new" exactly one entry remains and "old" is gone.
// This is the deterministic stand-in for the concurrent live-write-vs-backfill race, which
// the conflict and withTxnRetries tests cover structurally.
func TestBackfillIndex_DocUpdatedAfterIndexing_NoStaleEntry(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	doc := addUserDoc(t, ctx, col, "old")

	desc, err := newNameIndex(t, ctx, col)
	require.NoError(t, err)

	collectionID := col.Version().CollectionID
	shortID := getCollectionShortID(t, ctx, db, collectionID)

	require.Equal(t, 1, countIndexEntries(t, ctx, db, shortID, desc.ID))

	require.NoError(t, doc.Set(ctx, "name", "new"))
	require.NoError(t, col.UpdateDocument(ctx, doc))

	require.Equal(t, 1, countIndexEntries(t, ctx, db, shortID, desc.ID))

	require.Empty(t, queryUserByName(t, db, ctx, "old"), "stale entry must not be queryable")
	require.Len(t, queryUserByName(t, db, ctx, "new"), 1, "updated name must be queryable")
}

// TestBackfillIndex_UniqueIndex_ToleratesSameDocEntry checks saveUniqueKey's tolerateSameDoc
// branch: re-running backfill over a doc whose unique entry already exists (same docID, as if
// a live write got there first) skips it without error and adds no duplicate.
func TestBackfillIndex_UniqueIndex_ToleratesSameDocEntry(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.AddCollection(ctx, "type User { name: String\n age: Int }")
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	addUserDoc(t, ctx, col, "alice")

	desc, err := col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
		Unique: true,
	})
	require.NoError(t, err)

	shortID := getCollectionShortID(t, ctx, db, col.Version().CollectionID)
	require.Equal(t, 1, countIndexEntries(t, ctx, db, shortID, desc.ID))

	// Re-fetch to get the version carrying the index definition.
	col, err = db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	// Re-run backfill: the entry already exists for the same docID, so it must be skipped.
	err = db.backfillIndex(ctx, col.Version(), desc, immutable.None[string]())
	require.NoError(t, err, "re-running backfill over an already-indexed doc must not error")

	require.Equal(t, 1, countIndexEntries(t, ctx, db, shortID, desc.ID))
}

// TestBackfillBatchTxn_ConflictsWhenReadDocIsModified checks that a batch txn, which reads a
// doc and writes its index entry, conflicts on commit when that doc is updated concurrently.
// This is the storage-level guarantee the backfill retry loop relies on.
func TestBackfillBatchTxn_ConflictsWhenReadDocIsModified(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	doc := addUserDoc(t, ctx, col, "old")

	// Stage an in-memory index definition so NewCollectionIndex can resolve "name";
	// it is not persisted — this exercises storage-level conflict behavior, not the API.
	nameDesc := client.IndexDescription{
		Name:   "name_idx",
		ID:     1,
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
	}
	colVersion := col.Version()
	colVersion.Indexes = append(colVersion.Indexes, nameDesc)

	// Seed the index's epoch sequence as real index creation does, so the index resolves to
	// epoch 1; without it a created index could not exist.
	require.NoError(t, db.withTxnRetries(ctx, func(c context.Context) error {
		seq, err := sequence.Get(c, keys.NewIndexEpochSequenceKey(colVersion.CollectionID, nameDesc.ID))
		if err != nil {
			return err
		}
		_, err = seq.Next(c)
		return err
	}))

	// txn1 stands in for the backfill batch transaction.
	rawTxn1, err := db.NewTxn(false)
	require.NoError(t, err)
	txn1, ok := rawTxn1.(*Txn)
	require.True(t, ok, "expected *Txn")
	ctx1 := InitContext(ctx, txn1)

	col1, err := db.newCollection(ctx1, colVersion, immutable.Some[datastore.Txn](txn1))
	require.NoError(t, err)

	colIndex, err := NewCollectionIndex(ctx1, col1, nameDesc, true)
	require.NoError(t, err)

	// Run the batch body: reading the docs and writing entries puts the doc key range
	// in txn1's read set and produces a write, both needed for a commit-time conflict.
	fields := col1.Version().CollectIndexedFields()
	_, _, err = col1.iterateDocsBatch(ctx1, fields, immutable.None[string](), 10, func(d *client.Document) error {
		return colIndex.Save(ctx1, d)
	})
	require.NoError(t, err)

	// Update via the original handle (no index definition), touching only document keys.
	require.NoError(t, doc.Set(ctx, "name", "new"))
	require.NoError(t, col.UpdateDocument(ctx, doc))

	// The update overlaps txn1's read set, so its commit must conflict.
	commitErr := txn1.Commit()
	require.True(t, errors.Is(commitErr, corekv.ErrTxnConflict),
		"expected ErrTxnConflict but got: %v", commitErr)
}
