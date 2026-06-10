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
// returned context and returns a cleanup function that discards the transaction.
func newIndexStateTestCtx(t *testing.T) (context.Context, func()) {
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

	return ctx, cleanup
}

func TestIndexState_SetThenGet_RoundTripsAllFields(t *testing.T) {
	ctx, cleanup := newIndexStateTestCtx(t)
	defer cleanup()

	want := indexState{
		Status:    client.IndexStatusBuilding,
		Watermark: "bafkreidoc1",
		Reason:    "",
	}

	err := setIndexState(ctx, "col1", 1, want)
	require.NoError(t, err)

	got, err := getIndexState(ctx, "col1", 1)
	require.NoError(t, err)

	assert.Equal(t, want.Status, got.Status)
	assert.Equal(t, want.Watermark, got.Watermark)
	assert.Equal(t, want.Reason, got.Reason)
}

func TestIndexState_SetThenGet_FailedStateIncludesReason(t *testing.T) {
	ctx, cleanup := newIndexStateTestCtx(t)
	defer cleanup()

	want := indexState{
		Status: client.IndexStatusFailed,
		Reason: "disk full",
	}

	err := setIndexState(ctx, "col2", 7, want)
	require.NoError(t, err)

	got, err := getIndexState(ctx, "col2", 7)
	require.NoError(t, err)

	assert.Equal(t, client.IndexStatusFailed, got.Status)
	assert.Equal(t, "disk full", got.Reason)
}

func TestIndexState_GetMissingRecord_ReturnsErrNotFound(t *testing.T) {
	ctx, cleanup := newIndexStateTestCtx(t)
	defer cleanup()

	_, err := getIndexState(ctx, "doesnotexist", 999)
	require.Error(t, err)
	assert.True(t, errors.Is(err, corekv.ErrNotFound), "expected corekv.ErrNotFound, got: %v", err)
}

func TestIndexState_GetIndexStates_ReturnsOnlyGivenCollection(t *testing.T) {
	ctx, cleanup := newIndexStateTestCtx(t)
	defer cleanup()

	// Seed two index states for collection A.
	err := setIndexState(ctx, "colA", 1, indexState{Status: client.IndexStatusReady})
	require.NoError(t, err)
	err = setIndexState(ctx, "colA", 2, indexState{Status: client.IndexStatusBuilding, Watermark: "w1"})
	require.NoError(t, err)

	// Seed two index states for collection B.
	err = setIndexState(ctx, "colB", 1, indexState{Status: client.IndexStatusReady})
	require.NoError(t, err)
	err = setIndexState(ctx, "colB", 3, indexState{Status: client.IndexStatusFailed, Reason: "oops"})
	require.NoError(t, err)

	// Query for colA only.
	states, err := getIndexStates(ctx, "colA")
	require.NoError(t, err)

	require.Len(t, states, 2)
	assert.Equal(t, client.IndexStatusReady, states[1].Status)
	assert.Equal(t, client.IndexStatusBuilding, states[2].Status)
	assert.Equal(t, "w1", states[2].Watermark)
}

func TestIndexState_ListIndexStates_ReturnsAll(t *testing.T) {
	ctx, cleanup := newIndexStateTestCtx(t)
	defer cleanup()

	err := setIndexState(ctx, "colA", 1, indexState{Status: client.IndexStatusReady})
	require.NoError(t, err)
	err = setIndexState(ctx, "colA", 2, indexState{Status: client.IndexStatusBuilding})
	require.NoError(t, err)
	err = setIndexState(ctx, "colB", 1, indexState{Status: client.IndexStatusReady})
	require.NoError(t, err)
	err = setIndexState(ctx, "colB", 3, indexState{Status: client.IndexStatusFailed})
	require.NoError(t, err)

	all, err := listIndexStates(ctx)
	require.NoError(t, err)

	assert.Len(t, all, 4)
}

func TestIndexState_DeleteThenGet_ReturnsErrNotFound(t *testing.T) {
	ctx, cleanup := newIndexStateTestCtx(t)
	defer cleanup()

	err := setIndexState(ctx, "col1", 5, indexState{Status: client.IndexStatusReady})
	require.NoError(t, err)

	err = deleteIndexState(ctx, "col1", 5)
	require.NoError(t, err)

	_, err = getIndexState(ctx, "col1", 5)
	require.Error(t, err)
	assert.True(t, errors.Is(err, corekv.ErrNotFound), "expected corekv.ErrNotFound, got: %v", err)
}

func TestIndexState_DeleteReducesGetIndexStatesCount(t *testing.T) {
	ctx, cleanup := newIndexStateTestCtx(t)
	defer cleanup()

	err := setIndexState(ctx, "colX", 1, indexState{Status: client.IndexStatusReady})
	require.NoError(t, err)
	err = setIndexState(ctx, "colX", 2, indexState{Status: client.IndexStatusBuilding})
	require.NoError(t, err)

	states, err := getIndexStates(ctx, "colX")
	require.NoError(t, err)
	require.Len(t, states, 2)

	err = deleteIndexState(ctx, "colX", 1)
	require.NoError(t, err)

	states, err = getIndexStates(ctx, "colX")
	require.NoError(t, err)
	assert.Len(t, states, 1)
	assert.Equal(t, client.IndexStatusBuilding, states[2].Status)
}
