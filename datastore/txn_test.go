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

func TestMemoryStoreTxn_TwoTransactionsWithPutConflict_ShouldErrorWithConflict(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	txn1, err := rootstore.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = txn1.Put(ctx, ds.NewKey("key"), []byte("value"))
	require.NoError(t, err)

	txn2, err := rootstore.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = txn2.Put(ctx, ds.NewKey("key"), []byte("value"))
	require.NoError(t, err)

	// Commit txn2 first to create a conflict
	err = txn2.Commit(ctx)
	require.NoError(t, err)

	err = txn1.Commit(ctx)
	require.ErrorIs(t, err, badger.ErrConflict)
}

func TestMemoryStoreTxn_TwoTransactionsWithGetPutConflict_ShouldErrorWithConflict(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	rootstore.Put(ctx, ds.NewKey("key"), []byte("value"))

	txn1, err := rootstore.NewTransaction(ctx, false)
	require.NoError(t, err)

	_, err = txn1.Get(ctx, ds.NewKey("key"))
	require.NoError(t, err)

	err = txn1.Put(ctx, ds.NewKey("other-key"), []byte("value"))
	require.NoError(t, err)

	txn2, err := rootstore.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = txn2.Put(ctx, ds.NewKey("key"), []byte("value"))
	require.NoError(t, err)

	// Commit txn2 first to create a conflict
	err = txn2.Commit(ctx)
	require.NoError(t, err)

	err = txn1.Commit(ctx)
	require.ErrorIs(t, err, badger.ErrConflict)
}

func TestMemoryStoreTxn_TwoTransactionsWithHasPutConflict_ShouldErrorWithConflict(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	rootstore.Put(ctx, ds.NewKey("key"), []byte("value"))

	txn1, err := rootstore.NewTransaction(ctx, false)
	require.NoError(t, err)

	_, err = txn1.Has(ctx, ds.NewKey("key"))
	require.NoError(t, err)

	err = txn1.Put(ctx, ds.NewKey("other-key"), []byte("value"))
	require.NoError(t, err)

	txn2, err := rootstore.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = txn2.Put(ctx, ds.NewKey("key"), []byte("value"))
	require.NoError(t, err)

	// Commit txn2 first to create a conflict
	err = txn2.Commit(ctx)
	require.NoError(t, err)

	err = txn1.Commit(ctx)
	require.ErrorIs(t, err, badger.ErrConflict)
}

func TestBadgerMemoryStoreTxn_TwoTransactionsWithPutConflict_ShouldErrorWithConflict(t *testing.T) {
	ctx := context.Background()
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	require.NoError(t, err)

	txn1, err := rootstore.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = txn1.Put(ctx, ds.NewKey("key"), []byte("value"))
	require.NoError(t, err)

	txn2, err := rootstore.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = txn2.Put(ctx, ds.NewKey("key"), []byte("value"))
	require.NoError(t, err)

	// Commit txn2 first to create a conflict
	err = txn2.Commit(ctx)
	require.NoError(t, err)

	err = txn1.Commit(ctx)
	// We are expecting this to fail because of the conflict but badger does not return an error.
	// Conflicts in badger only occurs when the value of a key was changed between the time you read and you rewrite it.
	// require.ErrorIs(t, err, badger.ErrConflict)
	require.NoError(t, err)
}

func TestBadgerMemoryStoreTxn_TwoTransactionsWithGetPutConflict_ShouldErrorWithConflict(t *testing.T) {
	ctx := context.Background()
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	require.NoError(t, err)

	rootstore.Put(ctx, ds.NewKey("key"), []byte("value"))

	txn1, err := rootstore.NewTransaction(ctx, false)
	require.NoError(t, err)

	_, err = txn1.Get(ctx, ds.NewKey("key"))
	require.NoError(t, err)

	err = txn1.Put(ctx, ds.NewKey("other-key"), []byte("value"))
	require.NoError(t, err)

	txn2, err := rootstore.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = txn2.Put(ctx, ds.NewKey("key"), []byte("value"))
	require.NoError(t, err)

	// Commit txn2 first to create a conflict
	err = txn2.Commit(ctx)
	require.NoError(t, err)

	err = txn1.Commit(ctx)
	require.ErrorIs(t, err, badger.ErrConflict)
}

func TestBadgerMemoryStoreTxn_TwoTransactionsWithHasPutConflict_ShouldErrorWithConflict(t *testing.T) {
	ctx := context.Background()
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	require.NoError(t, err)

	rootstore.Put(ctx, ds.NewKey("key"), []byte("value"))

	txn1, err := rootstore.NewTransaction(ctx, false)
	require.NoError(t, err)

	_, err = txn1.Has(ctx, ds.NewKey("key"))
	require.NoError(t, err)

	err = txn1.Put(ctx, ds.NewKey("other-key"), []byte("value"))
	require.NoError(t, err)

	txn2, err := rootstore.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = txn2.Put(ctx, ds.NewKey("key"), []byte("value"))
	require.NoError(t, err)

	// Commit txn2 first to create a conflict
	err = txn2.Commit(ctx)
	require.NoError(t, err)

	err = txn1.Commit(ctx)
	require.ErrorIs(t, err, badger.ErrConflict)
}

func TestBadgerFileStoreTxn_TwoTransactionsWithPutConflict_ShouldErrorWithConflict(t *testing.T) {
	ctx := context.Background()
	opts := badgerds.Options{Options: badger.DefaultOptions("")}
	rootstore, err := badgerds.NewDatastore(t.TempDir(), &opts)
	require.NoError(t, err)

	txn1, err := rootstore.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = txn1.Put(ctx, ds.NewKey("key"), []byte("value"))
	require.NoError(t, err)

	txn2, err := rootstore.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = txn2.Put(ctx, ds.NewKey("key"), []byte("value"))
	require.NoError(t, err)

	// Commit txn2 first to create a conflict
	err = txn2.Commit(ctx)
	require.NoError(t, err)

	err = txn1.Commit(ctx)
	// We are expecting this to fail because of the conflict but badger does not return an error.
	// Conflicts in badger only occurs when the value of a key was changed between the time you read and you rewrite it.
	// require.ErrorIs(t, err, badger.ErrConflict)
	require.NoError(t, err)
}

func TestBadgerFileStoreTxn_TwoTransactionsWithGetPutConflict_ShouldErrorWithConflict(t *testing.T) {
	ctx := context.Background()
	opts := badgerds.Options{Options: badger.DefaultOptions("")}
	rootstore, err := badgerds.NewDatastore(t.TempDir(), &opts)
	require.NoError(t, err)

	rootstore.Put(ctx, ds.NewKey("key"), []byte("value"))

	txn1, err := rootstore.NewTransaction(ctx, false)
	require.NoError(t, err)

	_, err = txn1.Get(ctx, ds.NewKey("key"))
	require.NoError(t, err)

	err = txn1.Put(ctx, ds.NewKey("other-key"), []byte("value"))
	require.NoError(t, err)

	txn2, err := rootstore.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = txn2.Put(ctx, ds.NewKey("key"), []byte("value"))
	require.NoError(t, err)

	// Commit txn2 first to create a conflict
	err = txn2.Commit(ctx)
	require.NoError(t, err)

	err = txn1.Commit(ctx)
	require.ErrorIs(t, err, badger.ErrConflict)
}

func TestBadgerFileStoreTxn_TwoTransactionsWithHasPutConflict_ShouldErrorWithConflict(t *testing.T) {
	ctx := context.Background()
	opts := badgerds.Options{Options: badger.DefaultOptions("")}
	rootstore, err := badgerds.NewDatastore(t.TempDir(), &opts)
	require.NoError(t, err)

	rootstore.Put(ctx, ds.NewKey("key"), []byte("value"))

	txn1, err := rootstore.NewTransaction(ctx, false)
	require.NoError(t, err)

	_, err = txn1.Has(ctx, ds.NewKey("key"))
	require.NoError(t, err)

	err = txn1.Put(ctx, ds.NewKey("other-key"), []byte("value"))
	require.NoError(t, err)

	txn2, err := rootstore.NewTransaction(ctx, false)
	require.NoError(t, err)

	err = txn2.Put(ctx, ds.NewKey("key"), []byte("value"))
	require.NoError(t, err)

	// Commit txn2 first to create a conflict
	err = txn2.Commit(ctx)
	require.NoError(t, err)

	err = txn1.Commit(ctx)
	require.ErrorIs(t, err, badger.ErrConflict)
}
