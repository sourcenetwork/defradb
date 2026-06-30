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
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/id"
)

// TestRecoverIndexStates_BuildingResumesAndCompletes checks recovery rebuilds an interrupted
// build (wiped entries + building record) and clears the record on completion.
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

	// Empty watermark: the build committed its record but indexed no docs before interruption.
	err = db.withTxnRetries(ctx, func(txnCtx context.Context) error {
		if err := clearIndexEntries(t, txnCtx, shortID, desc.ID); err != nil {
			return err
		}
		return db.startIndexBuild(txnCtx, collectionID, desc.ID)
	})
	require.NoError(t, err)
	require.Equal(t, 0, countIndexEntries(t, ctx, db, shortID, desc.ID))

	require.NoError(t, db.recoverIndexStates(context.Background()))

	assert.Equal(t, len(names), countIndexEntries(t, ctx, db, shortID, desc.ID))
	requireNoIndexState(t, ctx, db, collectionID, desc.ID)
	for _, name := range names {
		require.Len(t, queryUserByName(t, db, ctx, name), 1, "name %q must be queryable", name)
	}
}

// TestRecoverIndexStates_DroppingResumesGC checks recovery finishes an interrupted GC
// (dropping record + leftover entries), removing both the entries and the record.
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

	before := countIndexEntries(t, ctx, db, shortID, desc.ID)
	assert.Equal(t, 5, before, "expected 5 index entries before recovery")

	// A dropping record with entries still present is the interrupted-GC state.
	err = db.withTxnRetries(ctx, func(txnCtx context.Context) error {
		return db.startIndexDrop(txnCtx, collectionID, desc.ID)
	})
	require.NoError(t, err)

	require.NoError(t, db.recoverIndexStates(context.Background()))

	after := countIndexEntries(t, ctx, db, shortID, desc.ID)
	assert.Equal(t, 0, after, "expected 0 index entries after recovery")
	requireNoIndexState(t, ctx, db, collectionID, desc.ID)
}

// TestRecoverIndexStates_BuildingResumesFromWatermark checks the non-empty watermark branch:
// recovery re-indexes only docs whose docID sorts strictly after the watermark, treating those
// at or before it as already done. Guards against an off-by-one at the resume boundary.
func TestRecoverIndexStates_BuildingResumesFromWatermark(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	docs := make([]*client.Document, 6)
	for i := range 6 {
		docs[i] = addUserDoc(t, ctx, col, "name"+string(rune('0'+i)))
	}

	desc, err := newNameIndex(t, ctx, col)
	require.NoError(t, err)

	collectionID := col.Version().CollectionID
	shortID := getCollectionShortID(t, ctx, db, collectionID)

	// Doc short IDs sort in storage order.
	docShortIDs := make([]uint64, len(docs))
	func() {
		rawTxn, err := db.NewTxn(true)
		require.NoError(t, err)
		defer rawTxn.Discard()
		txnCtx := InitContext(ctx, rawTxn)
		for i, d := range docs {
			docShortID, found, err := id.GetDocShortID(txnCtx, shortID, d.ID().String())
			require.NoError(t, err)
			require.True(t, found)
			docShortIDs[i] = docShortID
		}
	}()
	sort.Slice(docShortIDs, func(i, j int) bool {
		return docShortIDs[i] < docShortIDs[j]
	})

	// Watermark at the midpoint: docs 0-2 count as done, 3-5 must be re-indexed.
	watermark := docShortIDs[2]
	docsAfterWatermark := 3

	err = db.withTxnRetries(ctx, func(txnCtx context.Context) error {
		if err := clearIndexEntries(t, txnCtx, shortID, desc.ID); err != nil {
			return err
		}
		return db.advanceIndexWatermark(txnCtx, collectionID, desc.ID, watermark)
	})
	require.NoError(t, err)
	require.Equal(t, 0, countIndexEntries(t, ctx, db, shortID, desc.ID))

	require.NoError(t, db.recoverIndexStates(context.Background()))

	assert.Equal(t, docsAfterWatermark, countIndexEntries(t, ctx, db, shortID, desc.ID),
		"only docs after the watermark should be re-indexed")
	requireNoIndexState(t, ctx, db, collectionID, desc.ID)
}

// TestRecoverIndexStates_OrphanedBuildingRecord_Skipped checks that a building record with no
// matching index definition is skipped (not propagated as an error) and leaves the DB usable.
func TestRecoverIndexStates_OrphanedBuildingRecord_Skipped(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	collectionID := col.Version().CollectionID

	// No index has this ID, so findIndexDefinition fails and recovery must skip it.
	const orphanIndexID = uint32(999)
	err := db.withTxnRetries(ctx, func(txnCtx context.Context) error {
		return db.startIndexBuild(txnCtx, collectionID, orphanIndexID)
	})
	require.NoError(t, err)

	require.NoError(t, db.recoverIndexStates(context.Background()))

	_, err = newNameIndex(t, ctx, col)
	require.NoError(t, err, "system must remain usable after skipping orphan record")
}

// TestRecoverIndexStates_MixedRecords_OrphanSkippedValidHandled checks the recovery loop
// continues past a per-record error: an orphan building record errors while a valid dropping
// record alongside it is still processed.
func TestRecoverIndexStates_MixedRecords_OrphanSkippedValidHandled(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	addUserDoc(t, ctx, col, "Alice")
	addUserDoc(t, ctx, col, "Bob")

	desc, err := newNameIndex(t, ctx, col)
	require.NoError(t, err)

	collectionID := col.Version().CollectionID
	shortID := getCollectionShortID(t, ctx, db, collectionID)

	before := countIndexEntries(t, ctx, db, shortID, desc.ID)
	assert.Equal(t, 2, before, "expected 2 entries before recovery")

	const orphanIndexID = uint32(998)
	err = db.withTxnRetries(ctx, func(txnCtx context.Context) error {
		// Orphan building record (errors) alongside a valid dropping record.
		if err := db.startIndexBuild(txnCtx, collectionID, orphanIndexID); err != nil {
			return err
		}
		return db.startIndexDrop(txnCtx, collectionID, desc.ID)
	})
	require.NoError(t, err)

	require.NoError(t, db.recoverIndexStates(context.Background()))

	// The dropping record was still processed despite the orphan erroring.
	after := countIndexEntries(t, ctx, db, shortID, desc.ID)
	assert.Equal(t, 0, after, "expected dropping index entries to be GC'd")
	requireNoIndexState(t, ctx, db, collectionID, desc.ID)
}

// TestRecoverIndexStates_FailedAndNoRecords_NoOp checks recovery leaves a failed record
// untouched and is a no-op when there are no records.
func TestRecoverIndexStates_FailedAndNoRecords_NoOp(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	desc, err := newNameIndex(t, ctx, col)
	require.NoError(t, err)

	collectionID := col.Version().CollectionID

	err = db.withTxnRetries(ctx, func(txnCtx context.Context) error {
		return db.markIndexBuildFailed(txnCtx, collectionID, desc.ID, "some previous error")
	})
	require.NoError(t, err)

	require.NoError(t, db.recoverIndexStates(context.Background()))

	state := readIndexState(t, ctx, db, collectionID, desc.ID)
	assert.True(t, state.isFailed())
	assert.Equal(t, "some previous error", state.Reason)

	// A db with no records: recovery is a no-op.
	db2, _ := setupUserCollection(t, ctx)
	require.NoError(t, db2.recoverIndexStates(context.Background()))
}

// TestRecoverStaleEpochs_SweepsBelowBuildingEpoch checks the stale-epoch sweep collects a
// superseded epoch even while a rebuild is in flight, without touching the epoch being built. A
// migration invalidates the old epoch's values and a building index is excluded from query planning
// (queries full-scan), so the old epoch is neither valid nor read and is safe to collect early. The
// delete range is bounded strictly below the live epoch, so the in-progress build's epoch is left
// intact. recoverStaleEpochs is called directly here (not via recoverIndexStates, which would also
// resume the build) to isolate the sweep.
func TestRecoverStaleEpochs_SweepsBelowBuildingEpoch(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)
	collectionID := col.Version().CollectionID
	shortID := getCollectionShortID(t, ctx, db, collectionID)

	for _, name := range []string{"a", "b", "c"} {
		addUserDoc(t, ctx, col, name)
	}

	desc, err := newNameIndex(t, ctx, col)
	require.NoError(t, err)
	require.Equal(t, 3, countIndexEpochEntries(t, ctx, db, shortID, desc.ID, 1))

	// Advance to epoch 2 with a building record in place and a partial build in the new epoch. Epoch
	// 1 is the stale, pre-migration epoch; epoch 2 is the one being built.
	require.NoError(t, db.withTxnRetries(ctx, func(c context.Context) error {
		newEpoch, err := allocateIndexEpoch(c, collectionID, desc.ID)
		if err != nil {
			return err
		}
		require.Equal(t, uint32(2), newEpoch)
		if err := db.startIndexBuild(c, collectionID, desc.ID); err != nil {
			return err
		}
		coll, err := db.newCollection(c, col.Version(), datastore.CtxTryGetTxnOption(c))
		if err != nil {
			return err
		}
		colIndex, err := NewCollectionIndex(c, coll, desc, true)
		if err != nil {
			return err
		}
		return coll.indexExistingDocs(c, colIndex)
	}))
	require.Equal(t, 3, countIndexEpochEntries(t, ctx, db, shortID, desc.ID, 2))

	require.NoError(t, db.recoverStaleEpochs(ctx))

	assert.Equal(t, 0, countIndexEpochEntries(t, ctx, db, shortID, desc.ID, 1),
		"stale epoch below the building epoch must be collected")
	assert.Equal(t, 3, countIndexEpochEntries(t, ctx, db, shortID, desc.ID, 2),
		"the epoch being built must be left intact")
}
