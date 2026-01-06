// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package datastore

import (
	"context"
	"testing"

	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/immutable"

	"github.com/stretchr/testify/require"
)

func TestBasicTxn_CommitWhileOperationInProgress_ReturnsError(t *testing.T) {
	ctx := context.Background()
	// Create a rootstore
	rootstore := memory.NewDatastore(ctx)

	// Create a transaction
	txn := NewTxnFrom(rootstore, 1, false, immutable.None[int]())

	// Simulate an operation in progress
	txn.StartOp()

	// Attempt to commit while operation is in progress
	err := txn.Commit()

	// Should receive an error
	require.ErrorIs(t, err, ErrTxnHasActiveOps)

	// End the operation
	txn.EndOp()

	// Now commit should work
	err = txn.Commit()
	require.NoError(t, err)
}

func TestBasicTxn_DiscardWhileOperationInProgress_IsIgnored(t *testing.T) {
	ctx := context.Background()
	// Create a rootstore
	rootstore := memory.NewDatastore(ctx)

	// Create a transaction
	txn := NewTxnFrom(rootstore, 1, false, immutable.None[int]())

	// Simulate an operation in progress
	txn.StartOp()

	// Discard is ignored when operation is in progress (logs error but returns)
	txn.Discard()

	// The txn should still be usable after the ignored discard
	// End the operation
	txn.EndOp()

	// Now discard should work
	txn.Discard()
}

func TestBasicTxn_MultipleOperations_TrackedCorrectly(t *testing.T) {
	ctx := context.Background()
	// Create a rootstore
	rootstore := memory.NewDatastore(ctx)

	// Create a transaction
	txn := NewTxnFrom(rootstore, 1, false, immutable.None[int]())

	// Start multiple operations
	txn.StartOp()
	txn.StartOp()
	txn.StartOp()

	// Should not be able to commit
	err := txn.Commit()
	require.ErrorIs(t, err, ErrTxnHasActiveOps)

	// End one operation
	txn.EndOp()

	// Still should not be able to commit
	err = txn.Commit()
	require.ErrorIs(t, err, ErrTxnHasActiveOps)

	// End remaining operations
	txn.EndOp()
	txn.EndOp()

	// Now commit should work
	err = txn.Commit()
	require.NoError(t, err)
}

func TestBasicTxn_NoOperations_CommitWorks(t *testing.T) {
	ctx := context.Background()
	// Create a rootstore
	rootstore := memory.NewDatastore(ctx)

	// Create a transaction
	txn := NewTxnFrom(rootstore, 1, false, immutable.None[int]())

	// Commit without any operations should work
	err := txn.Commit()
	require.NoError(t, err)
}
