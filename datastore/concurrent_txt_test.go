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
	"testing"

	ds "github.com/ipfs/go-datastore"
	badger "github.com/sourcenetwork/badger/v4"
	"github.com/stretchr/testify/require"

	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v4"
	"github.com/sourcenetwork/defradb/datastore/memory"
)

func TestNewConcurrentTxnFrom(t *testing.T) {
	ctx := context.Background()
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	require.NoError(t, err)

	txn, err := NewConcurrentTxnFrom(ctx, rootstore, 0, false)
	require.NoError(t, err)

	err = txn.Commit(ctx)
	require.NoError(t, err)
}

func TestNewConcurrentTxnFromWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	require.NoError(t, err)

	err = rootstore.Close()
	require.NoError(t, err)

	_, err = NewConcurrentTxnFrom(ctx, rootstore, 0, false)
	require.ErrorIs(t, err, badgerds.ErrClosed)
}

func TestNewConcurrentTxnFromNonIterable(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	txn, err := NewConcurrentTxnFrom(ctx, rootstore, 0, false)
	require.NoError(t, err)

	err = txn.Commit(ctx)
	require.NoError(t, err)
}

func TestNewConcurrentTxnFromNonIterableWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	err := rootstore.Close()
	require.NoError(t, err)

	_, err = NewConcurrentTxnFrom(ctx, rootstore, 0, false)
	require.ErrorIs(t, err, badgerds.ErrClosed)
}

func TestConcurrentTxnSync(t *testing.T) {
	ctx := context.Background()
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	require.NoError(t, err)

	txn, err := rootstore.NewTransaction(ctx, false)
	require.NoError(t, err)

	cTxn := &concurrentTxn{Txn: txn}
	err = cTxn.Sync(ctx, ds.Key{})
	require.NoError(t, err)
}

func TestConcurrentTxnClose(t *testing.T) {
	ctx := context.Background()
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	require.NoError(t, err)

	txn, err := rootstore.NewTransaction(ctx, false)
	require.NoError(t, err)

	cTxn := &concurrentTxn{Txn: txn}
	err = cTxn.Close()
	require.NoError(t, err)
}
