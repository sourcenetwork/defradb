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

	badger "github.com/dgraph-io/badger/v4"
	ds "github.com/ipfs/go-datastore"
	"github.com/stretchr/testify/require"

	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v4"
)

func TestNewTxnFrom(t *testing.T) {
	ctx := context.Background()
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	require.NoError(t, err)

	txn, err := NewTxnFrom(ctx, rootstore, 0, false)
	require.NoError(t, err)

	err = txn.Commit(ctx)
	require.NoError(t, err)
}

func TestNewTxnFromWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	require.NoError(t, err)

	err = rootstore.Close()
	require.NoError(t, err)

	_, err = NewTxnFrom(ctx, rootstore, 0, false)
	require.ErrorIs(t, err, badgerds.ErrClosed)
}

func TestOnSuccess(t *testing.T) {
	ctx := context.Background()
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	require.NoError(t, err)

	txn, err := NewTxnFrom(ctx, rootstore, 0, false)
	require.NoError(t, err)

	txn.OnSuccess(nil)

	text := "Source"
	txn.OnSuccess(func() {
		text += " Inc"
	})
	err = txn.Commit(ctx)
	require.NoError(t, err)

	require.Equal(t, text, "Source Inc")
}

func TestOnError(t *testing.T) {
	ctx := context.Background()
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	require.NoError(t, err)

	txn, err := NewTxnFrom(ctx, rootstore, 0, false)
	require.NoError(t, err)

	txn.OnError(nil)

	text := "Source"
	txn.OnError(func() {
		text += " Inc"
	})

	rootstore.Close()
	require.NoError(t, err)

	err = txn.Commit(ctx)
	require.ErrorIs(t, err, badgerds.ErrClosed)

	require.Equal(t, text, "Source Inc")
}

func TestShimTxnStoreSync(t *testing.T) {
	ctx := context.Background()
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	require.NoError(t, err)

	txn, err := rootstore.NewTransaction(ctx, false)
	require.NoError(t, err)

	shimTxn := ShimTxnStore{txn}
	err = shimTxn.Sync(ctx, ds.Key{})
	require.NoError(t, err)
}

func TestShimTxnStoreClose(t *testing.T) {
	ctx := context.Background()
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	require.NoError(t, err)

	txn, err := rootstore.NewTransaction(ctx, false)
	require.NoError(t, err)

	shimTxn := ShimTxnStore{txn}
	err = shimTxn.Close()
	require.NoError(t, err)
}
