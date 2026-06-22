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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/db/action"
)

// newIndexStateTestCtx creates a DB, opens a read-write transaction, stores it in the
// returned context and returns the DB plus a cleanup function that discards the transaction.
//
// Index state is persisted as action records; a ready index has no record at all
// (missing ⇒ ready), so the tests below seed only the non-ready states.
func newIndexStateTestCtx(t *testing.T) (*DB, context.Context, func()) {
	t.Helper()

	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)

	txn, err := db.NewTxn(false)
	require.NoError(t, err)

	ctx = InitContext(ctx, txn)

	cleanup := func() {
		txn.Discard()
		db.Close()
	}

	return db, ctx, cleanup
}

func TestIndexState_BuildThenGet_RoundTripsWatermark(t *testing.T) {
	db, ctx, cleanup := newIndexStateTestCtx(t)
	defer cleanup()

	require.NoError(t, db.advanceIndexWatermark(ctx, "col1", 1, "bafkreidoc1"))

	got, err := getIndexState(ctx, "col1", 1)
	require.NoError(t, err)

	assert.True(t, got.isBuilding())
	assert.Equal(t, "bafkreidoc1", got.Watermark)
	assert.Empty(t, got.Reason)
}

func TestIndexState_FailedIncludesReason(t *testing.T) {
	db, ctx, cleanup := newIndexStateTestCtx(t)
	defer cleanup()

	require.NoError(t, db.markIndexBuildFailed(ctx, "col2", 7, "disk full"))

	got, err := getIndexState(ctx, "col2", 7)
	require.NoError(t, err)

	assert.True(t, got.isFailed())
	assert.Equal(t, "disk full", got.Reason)
}

func TestIndexState_Dropping(t *testing.T) {
	db, ctx, cleanup := newIndexStateTestCtx(t)
	defer cleanup()

	require.NoError(t, db.startIndexDrop(ctx, "col3", 2))

	got, err := getIndexState(ctx, "col3", 2)
	require.NoError(t, err)
	assert.True(t, got.isDropping())
}

func TestIndexState_GetMissingRecord_ReturnsErrNotFound(t *testing.T) {
	_, ctx, cleanup := newIndexStateTestCtx(t)
	defer cleanup()

	_, err := getIndexState(ctx, "doesnotexist", 999)
	require.Error(t, err)
	assert.True(t, errors.Is(err, corekv.ErrNotFound), "expected corekv.ErrNotFound, got: %v", err)
}

func TestIndexState_GetIndexStates_ReturnsOnlyGivenCollection(t *testing.T) {
	db, ctx, cleanup := newIndexStateTestCtx(t)
	defer cleanup()

	// Seed non-ready index states for collection A. A ready index (index 3) has no record.
	require.NoError(t, db.markIndexBuildFailed(ctx, "colA", 1, "boom"))
	require.NoError(t, db.advanceIndexWatermark(ctx, "colA", 2, "w1"))

	// Seed for collection B, which must not leak into the colA query.
	require.NoError(t, db.startIndexDrop(ctx, "colB", 1))
	require.NoError(t, db.markIndexBuildFailed(ctx, "colB", 3, "oops"))

	states, err := getIndexStates(ctx, "colA")
	require.NoError(t, err)

	require.Len(t, states, 2)
	assert.True(t, states[1].isFailed())
	assert.Equal(t, "boom", states[1].Reason)
	assert.True(t, states[2].isBuilding())
	assert.Equal(t, "w1", states[2].Watermark)
}

func TestIndexState_ListIndexStates_ReturnsAll(t *testing.T) {
	db, ctx, cleanup := newIndexStateTestCtx(t)
	defer cleanup()

	require.NoError(t, db.markIndexBuildFailed(ctx, "colA", 1, ""))
	require.NoError(t, db.startIndexBuild(ctx, "colA", 2))
	require.NoError(t, db.startIndexDrop(ctx, "colB", 1))
	require.NoError(t, db.markIndexBuildFailed(ctx, "colB", 3, ""))

	all, err := listIndexStates(ctx)
	require.NoError(t, err)

	assert.Len(t, all, 4)
}

func TestIndexState_CompleteThenGet_ReturnsErrNotFound(t *testing.T) {
	db, ctx, cleanup := newIndexStateTestCtx(t)
	defer cleanup()

	require.NoError(t, db.startIndexBuild(ctx, "col1", 5))

	require.NoError(t, db.completeIndexBuild(ctx, "col1", 5))

	_, err := getIndexState(ctx, "col1", 5)
	require.Error(t, err)
	assert.True(t, errors.Is(err, corekv.ErrNotFound), "expected corekv.ErrNotFound, got: %v", err)
}

func TestIndexState_CompleteReducesGetIndexStatesCount(t *testing.T) {
	db, ctx, cleanup := newIndexStateTestCtx(t)
	defer cleanup()

	require.NoError(t, db.markIndexBuildFailed(ctx, "colX", 1, ""))
	require.NoError(t, db.startIndexBuild(ctx, "colX", 2))

	states, err := getIndexStates(ctx, "colX")
	require.NoError(t, err)
	require.Len(t, states, 2)

	require.NoError(t, db.completeIndexBuild(ctx, "colX", 1))

	states, err = getIndexStates(ctx, "colX")
	require.NoError(t, err)
	assert.Len(t, states, 1)
	assert.True(t, states[2].isBuilding())
}

// TestScanIndexStates_CorruptPayload_StrictVsLenient pins the two error policies of scanIndexStates
// against an undecodable build payload. The lenient scan (skipCorrupt=true, used by collection open
// and listings) skips the bad record so one corrupt entry does not deny access to the collection;
// the strict scan (skipCorrupt=false, used by recovery via listIndexStates) surfaces the error so a
// corrupt record is not silently left unrecovered.
func TestScanIndexStates_CorruptPayload_StrictVsLenient(t *testing.T) {
	db, ctx, cleanup := newIndexStateTestCtx(t)
	defer cleanup()

	const collectionID = "colCorrupt"
	const indexID = uint32(1)

	// A build record whose payload is not valid JSON, so loadIndexState fails to decode it.
	require.NoError(t, action.SetTxn(
		ctx, db.events, collectionID, client.BackfillIndexAction, indexSubject(indexID),
		client.InProgressActionStatus, "", json.RawMessage("}not json{"), false,
	))

	prefix := indexActionCollectionPrefix(collectionID)

	// Lenient: the corrupt record is skipped, leaving a successful (record-free) result.
	lenient, err := scanIndexStates(ctx, prefix, true)
	require.NoError(t, err, "lenient scan must not fail on a corrupt payload")
	for _, rec := range lenient {
		assert.NotEqual(t, indexID, rec.Key.IndexID, "corrupt record must be skipped by the lenient scan")
	}

	// Strict: the corrupt record aborts the scan with the corrupt-payload error.
	_, err = scanIndexStates(ctx, prefix, false)
	require.ErrorIs(t, err, NewErrCorruptIndexPayload(nil), "strict scan must surface a corrupt payload")
}
