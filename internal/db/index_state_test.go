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

	"github.com/sourcenetwork/corekv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
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

func TestIndexState_SetThenGet_RoundTripsBuildingWatermark(t *testing.T) {
	db, ctx, cleanup := newIndexStateTestCtx(t)
	defer cleanup()

	want := indexState{
		Status:    client.IndexStatusBuilding,
		Watermark: "bafkreidoc1",
	}

	err := db.setIndexState(ctx, "col1", 1, want)
	require.NoError(t, err)

	got, err := getIndexState(ctx, "col1", 1)
	require.NoError(t, err)

	assert.Equal(t, client.IndexStatusBuilding, got.Status)
	assert.Equal(t, "bafkreidoc1", got.Watermark)
	assert.Empty(t, got.Reason)
}

func TestIndexState_SetThenGet_FailedStateIncludesReason(t *testing.T) {
	db, ctx, cleanup := newIndexStateTestCtx(t)
	defer cleanup()

	want := indexState{
		Status: client.IndexStatusFailed,
		Reason: "disk full",
	}

	err := db.setIndexState(ctx, "col2", 7, want)
	require.NoError(t, err)

	got, err := getIndexState(ctx, "col2", 7)
	require.NoError(t, err)

	assert.Equal(t, client.IndexStatusFailed, got.Status)
	assert.Equal(t, "disk full", got.Reason)
}

func TestIndexState_SetThenGet_DroppingState(t *testing.T) {
	db, ctx, cleanup := newIndexStateTestCtx(t)
	defer cleanup()

	err := db.setIndexState(ctx, "col3", 2, indexState{Status: client.IndexStatusDropping})
	require.NoError(t, err)

	got, err := getIndexState(ctx, "col3", 2)
	require.NoError(t, err)
	assert.Equal(t, client.IndexStatusDropping, got.Status)
}

func TestIndexState_SetReady_IsRejected(t *testing.T) {
	db, ctx, cleanup := newIndexStateTestCtx(t)
	defer cleanup()

	// Ready is represented by the absence of a record, so it cannot be stored.
	err := db.setIndexState(ctx, "col1", 1, indexState{Status: client.IndexStatusReady})
	require.Error(t, err)
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
	err := db.setIndexState(ctx, "colA", 1, indexState{Status: client.IndexStatusFailed, Reason: "boom"})
	require.NoError(t, err)
	err = db.setIndexState(ctx, "colA", 2, indexState{Status: client.IndexStatusBuilding, Watermark: "w1"})
	require.NoError(t, err)

	// Seed for collection B, which must not leak into the colA query.
	err = db.setIndexState(ctx, "colB", 1, indexState{Status: client.IndexStatusDropping})
	require.NoError(t, err)
	err = db.setIndexState(ctx, "colB", 3, indexState{Status: client.IndexStatusFailed, Reason: "oops"})
	require.NoError(t, err)

	states, err := getIndexStates(ctx, "colA")
	require.NoError(t, err)

	require.Len(t, states, 2)
	assert.Equal(t, client.IndexStatusFailed, states[1].Status)
	assert.Equal(t, "boom", states[1].Reason)
	assert.Equal(t, client.IndexStatusBuilding, states[2].Status)
	assert.Equal(t, "w1", states[2].Watermark)
}

func TestIndexState_ListIndexStates_ReturnsAll(t *testing.T) {
	db, ctx, cleanup := newIndexStateTestCtx(t)
	defer cleanup()

	err := db.setIndexState(ctx, "colA", 1, indexState{Status: client.IndexStatusFailed})
	require.NoError(t, err)
	err = db.setIndexState(ctx, "colA", 2, indexState{Status: client.IndexStatusBuilding})
	require.NoError(t, err)
	err = db.setIndexState(ctx, "colB", 1, indexState{Status: client.IndexStatusDropping})
	require.NoError(t, err)
	err = db.setIndexState(ctx, "colB", 3, indexState{Status: client.IndexStatusFailed})
	require.NoError(t, err)

	all, err := listIndexStates(ctx)
	require.NoError(t, err)

	assert.Len(t, all, 4)
}

func TestIndexState_DeleteThenGet_ReturnsErrNotFound(t *testing.T) {
	db, ctx, cleanup := newIndexStateTestCtx(t)
	defer cleanup()

	err := db.setIndexState(ctx, "col1", 5, indexState{Status: client.IndexStatusBuilding})
	require.NoError(t, err)

	err = db.deleteIndexState(ctx, "col1", 5)
	require.NoError(t, err)

	_, err = getIndexState(ctx, "col1", 5)
	require.Error(t, err)
	assert.True(t, errors.Is(err, corekv.ErrNotFound), "expected corekv.ErrNotFound, got: %v", err)
}

func TestIndexState_DeleteReducesGetIndexStatesCount(t *testing.T) {
	db, ctx, cleanup := newIndexStateTestCtx(t)
	defer cleanup()

	err := db.setIndexState(ctx, "colX", 1, indexState{Status: client.IndexStatusFailed})
	require.NoError(t, err)
	err = db.setIndexState(ctx, "colX", 2, indexState{Status: client.IndexStatusBuilding})
	require.NoError(t, err)

	states, err := getIndexStates(ctx, "colX")
	require.NoError(t, err)
	require.Len(t, states, 2)

	err = db.deleteIndexState(ctx, "colX", 1)
	require.NoError(t, err)

	states, err = getIndexStates(ctx, "colX")
	require.NoError(t, err)
	assert.Len(t, states, 1)
	assert.Equal(t, client.IndexStatusBuilding, states[2].Status)
}
