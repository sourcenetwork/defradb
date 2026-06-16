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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

// TestRecoverIndexStates_BuildingResumesAndCompletes simulates a shutdown mid-backfill
// by clearing the index entries and seeding a building record, then asserts that recovery
// rebuilds every entry and clears the record (a completed build keeps no record).
func TestRecoverIndexStates_BuildingResumesAndCompletes(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	names := []string{"Alice", "Bob", "Carol"}
	for _, name := range names {
		addUserDoc(t, ctx, col, name)
	}

	desc, err := newNameIndex(t, ctx, col)
	require.NoError(t, err)

	collectionID := col.Version().CollectionID
	shortID := getCollectionShortID(t, ctx, db, collectionID)

	// Wipe the entries and seed a building record with no watermark, mimicking a build
	// that committed its definition and state but was interrupted before indexing docs.
	err = db.withTxnRetries(ctx, func(txnCtx context.Context) error {
		if err := clearIndexEntries(t, txnCtx, shortID, desc.ID); err != nil {
			return err
		}
		return setIndexState(txnCtx, collectionID, desc.ID, indexState{
			Status: client.IndexStatusBuilding,
		})
	})
	require.NoError(t, err)
	require.Equal(t, 0, countIndexEntries(t, ctx, db, shortID, desc.ID))

	require.NoError(t, db.recoverIndexStates(context.Background()))

	// The resumed build indexes every document and clears the record (missing == ready).
	assert.Equal(t, len(names), countIndexEntries(t, ctx, db, shortID, desc.ID))
	requireNoIndexState(t, ctx, db, collectionID, desc.ID)
	for _, name := range names {
		require.Len(t, queryUserByName(t, db, ctx, name), 1, "name %q must be queryable", name)
	}
}

// TestRecoverIndexStates_DroppingResumesGC simulates a shutdown mid-GC by seeding a
// dropping record while leaving index entries in place, then asserts that recovery
// deletes all entries and removes the state record.
func TestRecoverIndexStates_DroppingResumesGC(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	for _, name := range []string{"Alice", "Bob", "Carol", "Dave", "Eve"} {
		addUserDoc(t, ctx, col, name)
	}

	desc, err := newNameIndex(t, ctx, col)
	require.NoError(t, err)

	collectionID := col.Version().CollectionID
	shortID := getCollectionShortID(t, ctx, db, collectionID)

	// Confirm entries exist before simulating interrupted GC.
	before := countIndexEntries(t, ctx, db, shortID, desc.ID)
	assert.Equal(t, 5, before, "expected 5 index entries before recovery")

	// Seed a dropping record to simulate an interrupted GC run.
	err = db.withTxnRetries(ctx, func(txnCtx context.Context) error {
		return setIndexState(txnCtx, collectionID, desc.ID, indexState{
			Status: client.IndexStatusDropping,
		})
	})
	require.NoError(t, err)

	require.NoError(t, db.recoverIndexStates(context.Background()))

	// All entries must be gone.
	after := countIndexEntries(t, ctx, db, shortID, desc.ID)
	assert.Equal(t, 0, after, "expected 0 index entries after recovery")

	// The state record must be removed.
	requireNoIndexState(t, ctx, db, collectionID, desc.ID)
}

// TestRecoverIndexStates_FailedAndNoRecords_NoOp verifies that a failed record is left
// untouched and that recovery on a fresh collection with no records returns nil.
func TestRecoverIndexStates_FailedAndNoRecords_NoOp(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	desc, err := newNameIndex(t, ctx, col)
	require.NoError(t, err)

	collectionID := col.Version().CollectionID

	// Seed a failed record.
	err = db.withTxnRetries(ctx, func(txnCtx context.Context) error {
		return setIndexState(txnCtx, collectionID, desc.ID, indexState{
			Status: client.IndexStatusFailed,
			Reason: "some previous error",
		})
	})
	require.NoError(t, err)

	require.NoError(t, db.recoverIndexStates(context.Background()))

	// Failed record must remain untouched.
	state := readIndexState(t, ctx, db, collectionID, desc.ID)
	assert.Equal(t, client.IndexStatusFailed, state.Status)
	assert.Equal(t, "some previous error", state.Reason)

	// Recovery on a db with no records must also be a no-op.
	db2, _ := setupUserCollection(t, ctx)
	require.NoError(t, db2.recoverIndexStates(context.Background()))
}
