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

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/memory"
	"github.com/stretchr/testify/require"
)

func TestNewTxnFrom(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	txn := NewTxnFrom(ctx, rootstore, 0, false)

	err := txn.Commit(ctx)
	require.NoError(t, err)
}

func TestOnSuccess(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	txn := NewTxnFrom(ctx, rootstore, 0, false)

	text := "Source"
	txn.OnSuccess(func() {
		text += " Inc"
	})
	err := txn.Commit(ctx)
	require.NoError(t, err)

	require.Equal(t, text, "Source Inc")
}

func TestOnSuccessAsync(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	txn := NewTxnFrom(ctx, rootstore, 0, false)

	var wg sync.WaitGroup
	txn.OnSuccessAsync(func() {
		wg.Done()
	})

	wg.Add(1)
	err := txn.Commit(ctx)
	require.NoError(t, err)
	wg.Wait()
}

func TestOnError(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	txn := NewTxnFrom(ctx, rootstore, 0, false)

	text := "Source"
	txn.OnError(func() {
		text += " Inc"
	})

	err := rootstore.Close()
	require.NoError(t, err)

	err = txn.Commit(ctx)
	require.ErrorIs(t, err, corekv.ErrDBClosed)

	require.Equal(t, text, "Source Inc")
}

func TestOnErrorAsync(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	txn := NewTxnFrom(ctx, rootstore, 0, false)

	var wg sync.WaitGroup
	txn.OnErrorAsync(func() {
		wg.Done()
	})

	err := rootstore.Close()
	require.NoError(t, err)

	wg.Add(1)
	err = txn.Commit(ctx)
	require.ErrorIs(t, err, corekv.ErrDBClosed)
	wg.Wait()
}

func TestOnDiscard(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	txn := NewTxnFrom(ctx, rootstore, 0, false)

	text := "Source"
	txn.OnDiscard(func() {
		text += " Inc"
	})
	txn.Discard(ctx)

	require.Equal(t, text, "Source Inc")
}

func TestOnDiscardAsync(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	txn := NewTxnFrom(ctx, rootstore, 0, false)

	var wg sync.WaitGroup
	txn.OnDiscardAsync(func() {
		wg.Done()
	})

	wg.Add(1)
	txn.Discard(ctx)
	wg.Wait()
}
