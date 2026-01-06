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

	// Now commit should work
	err = txn.Commit()
	require.NoError(t, err)
}

// TestTxn_ConcurrentCommitDuringOperation verifies that concurrent commit attempts
// are correctly rejected while an operation is using the transaction.
func TestTxn_ConcurrentCommitDuringOperation(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	// Create a manual transaction
	txn, err := db.NewTxn(false)
	require.NoError(t, err)

	basicTxn := datastore.MustGetFromClientTxn(txn)

	var wg sync.WaitGroup
	var commitErr error
	operationDone := make(chan struct{})

	// Goroutine 1: Simulates a long-running operation
	wg.Add(1)
	go func() {
		defer wg.Done()
		basicTxn.StartOp()
		defer basicTxn.EndOp()

		// Simulate some work
		time.Sleep(100 * time.Millisecond)
		close(operationDone)
	}()

	// Goroutine 2: User tries to commit while operation is running
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Wait a bit to ensure operation has started
		time.Sleep(20 * time.Millisecond)

		// Try to commit - should fail because operation is in progress
		commitErr = txn.Commit()
	}()

	wg.Wait()

	// The concurrent commit should have failed
	require.ErrorIs(t, commitErr, datastore.ErrTxnHasActiveOps,
		"Concurrent commit during operation should fail")

	// After operation is done, commit should work
	err = txn.Commit()
	require.NoError(t, err)
}

// TestTxn_ExplicitTxnOperationTracking verifies that when a user-created transaction
// is used in an operation via ensureContextTxn, the operation tracking works correctly
// and prevents premature commits.
func TestTxn_ExplicitTxnOperationTracking(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	// Create a manual transaction
	userTxn, err := db.NewTxn(false)
	require.NoError(t, err)

	// Set the transaction on the context (as a user would do)
	ctx = datastore.CtxSetFromClientTxn(ctx, userTxn)

	// Call ensureContextTxn (which is what operations do internally)
	// This should call StartOp on the underlying BasicTxn
	ctx, opTxn, err := ensureContextTxn(ctx, db, false)
	require.NoError(t, err)

	// Now the user-created txn should not be committable because ensureContextTxn
	// called StartOp
	err = userTxn.Commit()
	require.ErrorIs(t, err, datastore.ErrTxnHasActiveOps,
		"User txn should not be committable while operation is using it")

	// When the operation "completes" (calls Commit/Discard on explicit wrapper),
	// EndOp is called automatically
	opTxn.Discard()

	// Now user should be able to commit
	err = userTxn.Commit()
	require.NoError(t, err)
}
