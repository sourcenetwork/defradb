// Copyright 2025 Democratized Data Foundation
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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/internal/datastore"
)

// TestTxn_CommitWhileOperationInProgress verifies that a transaction cannot be committed
// while operations are still running on it. This prevents database corruption that could
// occur if a user commits a transaction before an operation using it has completed.
func TestTxn_CommitWhileOperationInProgress(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	// Create a manual transaction as a user would
	txn, err := db.NewTxn(false)
	require.NoError(t, err)

	// Simulate that an operation is in progress by calling StartOp
	// (this happens automatically in ensureContextTxn when operations use the txn)
	basicTxn := datastore.MustGetFromClientTxn(txn)
	basicTxn.StartOp()

	// Now try to commit the transaction while "operation" is in progress
	err = txn.Commit()

	// Should return an error since operations are in progress
	require.ErrorIs(t, err, datastore.ErrTxnHasActiveOps,
		"Commit should fail when operations are in progress")

	// End the operation
	basicTxn.EndOp()

	// Now commit should succeed
	err = txn.Commit()
	require.NoError(t, err)
}

// TestTxn_ConcurrentCommitDuringOperation verifies that concurrent commit attempts
// are rejected while an operation is using the transaction.
func TestTxn_ConcurrentCommitDuringOperation(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	txn, err := db.NewTxn(false)
	require.NoError(t, err)

	basicTxn := datastore.MustGetFromClientTxn(txn)

	var wg sync.WaitGroup
	var commitErr error

	// Goroutine 1: Simulates a long-running operation
	wg.Add(1)
	go func() {
		defer wg.Done()
		basicTxn.StartOp()
		defer basicTxn.EndOp()
		time.Sleep(100 * time.Millisecond)
	}()

	// Goroutine 2: Attempts to commit while operation is running
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(20 * time.Millisecond)
		commitErr = txn.Commit()
	}()

	wg.Wait()

	// Concurrent commit should have failed
	require.ErrorIs(t, commitErr, datastore.ErrTxnHasActiveOps)

	// After operation completes, commit should succeed
	err = txn.Commit()
	require.NoError(t, err)
}

// TestTxn_ExplicitTxnOperationTracking verifies that ensureContextTxn correctly
// tracks operations and prevents premature commits.
func TestTxn_ExplicitTxnOperationTracking(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	userTxn, err := db.NewTxn(false)
	require.NoError(t, err)

	// Set transaction on context and create explicit wrapper
	ctx = datastore.CtxSetFromClientTxn(ctx, userTxn)
	_, opTxn, err := ensureContextTxn(ctx, db, false)
	require.NoError(t, err)

	// User cannot commit while operation wrapper exists
	err = userTxn.Commit()
	require.ErrorIs(t, err, datastore.ErrTxnHasActiveOps)

	// Operation completes
	opTxn.Discard()

	// Now user can commit
	err = userTxn.Commit()
	require.NoError(t, err)
}

// TestTxn_MultipleDiscardCalls verifies that calling Discard multiple times
// on an explicit transaction does not cause incorrect activeOps count.
func TestTxn_MultipleDiscardCalls(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	userTxn, err := db.NewTxn(false)
	require.NoError(t, err)

	ctx = datastore.CtxSetFromClientTxn(ctx, userTxn)
	_, opTxn, err := ensureContextTxn(ctx, db, false)
	require.NoError(t, err)

	// Multiple Discard calls should not cause negative activeOps
	opTxn.Discard()
	opTxn.Discard()
	opTxn.Discard()

	// Commit should still work (activeOps should be 0, not negative)
	err = userTxn.Commit()
	require.NoError(t, err)
}

// TestTxn_MultipleOperationsThenCommit verifies that multiple sequential operations
// can be performed on a transaction before a single commit.
func TestTxn_MultipleOperationsThenCommit(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	userTxn, err := db.NewTxn(false)
	require.NoError(t, err)

	ctx = datastore.CtxSetFromClientTxn(ctx, userTxn)

	// First operation
	_, opTxn1, err := ensureContextTxn(ctx, db, false)
	require.NoError(t, err)
	opTxn1.Discard()

	// Second operation
	_, opTxn2, err := ensureContextTxn(ctx, db, false)
	require.NoError(t, err)
	opTxn2.Discard()

	// Third operation
	_, opTxn3, err := ensureContextTxn(ctx, db, false)
	require.NoError(t, err)
	opTxn3.Discard()

	// Single commit at the end
	err = userTxn.Commit()
	require.NoError(t, err)
}

// TestTxn_DiscardThenOperationThenCommit verifies that a transaction can be
// used after being discarded, if supported by the underlying store.
func TestTxn_DiscardThenOperationThenCommit(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	userTxn, err := db.NewTxn(false)
	require.NoError(t, err)

	// Discard multiple times before any operation
	userTxn.Discard()
	userTxn.Discard()

	// Perform an operation
	ctx = datastore.CtxSetFromClientTxn(context.Background(), userTxn)
	_, opTxn, err := ensureContextTxn(ctx, db, false)
	require.NoError(t, err)

	opTxn.Discard()

	// Commit should work
	err = userTxn.Commit()
	require.NoError(t, err)
}

// TestTxn_ImmediateDiscard verifies that discarding a transaction immediately
// after creation without performing any operations works correctly.
func TestTxn_ImmediateDiscard(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	userTxn, err := db.NewTxn(false)
	require.NoError(t, err)

	// Discard immediately without any operations
	userTxn.Discard()
	userTxn.Discard()

	// Verify database is still functional
	userTxn2, err := db.NewTxn(false)
	require.NoError(t, err)
	err = userTxn2.Commit()
	require.NoError(t, err)
}
