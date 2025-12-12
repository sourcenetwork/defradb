// Copyright 2022 Democratized Data Foundation
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
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/immutable"
)

func TestNewTxnFrom(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	txn := NewTxnFrom(rootstore, 0, false, immutable.None[int]())

	err := txn.Commit()
	require.NoError(t, err)
}

func TestOnSuccess(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	txn := NewTxnFrom(rootstore, 0, false, immutable.None[int]())

	runOnSuccessTest(t, txn)
}

func TestOnSuccessAsync(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	txn := NewTxnFrom(rootstore, 0, false, immutable.None[int]())

	runOnSuccessAsyncTest(t, txn)
}

func TestOnError(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	txn := NewTxnFrom(rootstore, 0, false, immutable.None[int]())

	err := rootstore.Close()
	require.NoError(t, err)

	runOnErrorTest(t, txn)
}

func TestOnErrorAsync(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	txn := NewTxnFrom(rootstore, 0, false, immutable.None[int]())

	err := rootstore.Close()
	require.NoError(t, err)

	runOnErrorAsyncTest(t, txn)
}

func TestOnDiscard(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	txn := NewTxnFrom(rootstore, 0, false, immutable.None[int]())

	runOnDiscardTest(t, txn)
}

func TestOnDiscardAsync(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	txn := NewTxnFrom(rootstore, 0, false, immutable.None[int]())

	runOnDiscardAsyncTest(t, txn)
}

// Helper functions for common Txn tests

func runOnSuccessTest(t *testing.T, txn Txn) {
	text := "Source"
	txn.OnSuccess(func() {
		text += " Inc"
	})
	err := txn.Commit()
	require.NoError(t, err)

	require.Equal(t, "Source Inc", text)
}

func runOnSuccessAsyncTest(t *testing.T, txn Txn) {
	var wg sync.WaitGroup
	txn.OnSuccessAsync(func() {
		wg.Done()
	})

	wg.Add(1)
	err := txn.Commit()
	require.NoError(t, err)
	wg.Wait()
}

func runOnErrorTest(t *testing.T, txn Txn) {
	text := "Source"
	txn.OnError(func() {
		text += " Inc"
	})

	err := txn.Commit()
	require.ErrorIs(t, err, corekv.ErrDBClosed)

	require.Equal(t, "Source Inc", text)
}

func runOnErrorAsyncTest(t *testing.T, txn Txn) {
	var wg sync.WaitGroup
	txn.OnErrorAsync(func() {
		wg.Done()
	})

	wg.Add(1)
	err := txn.Commit()
	require.ErrorIs(t, err, corekv.ErrDBClosed)
	wg.Wait()
}

func runOnDiscardTest(t *testing.T, txn Txn) {
	text := "Source"
	txn.OnDiscard(func() {
		text += " Inc"
	})
	txn.Discard()

	require.Equal(t, "Source Inc", text)
}

func runOnDiscardAsyncTest(t *testing.T, txn Txn) {
	var wg sync.WaitGroup
	txn.OnDiscardAsync(func() {
		wg.Done()
	})

	wg.Add(1)
	txn.Discard()
	wg.Wait()
}
