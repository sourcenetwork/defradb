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

	"github.com/stretchr/testify/require"

	badgerds "github.com/dgraph-io/badger/v4"
	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/memory"

	badgerkv "github.com/sourcenetwork/corekv/badger"
)

func getBadgerTxnDB(t *testing.T, ctx context.Context) corekv.TxnStore {
	opts := badgerds.DefaultOptions("").WithInMemory(true)
	rootstore, err := badgerkv.NewDatastore("", opts)
	require.NoError(t, err)

	return rootstore.(corekv.TxnStore)
}

func getMemoryTxnDB(t *testing.T, ctx context.Context) corekv.TxnStore {
	rootstore := memory.NewDatastore(ctx)

	return rootstore
}

func TestNewConcurrentTxnFrom(t *testing.T) {
	ctx := context.Background()
	rootstore := getBadgerTxnDB(t, ctx)

	txn, err := NewConcurrentTxnFrom(ctx, rootstore, 0, false)
	require.NoError(t, err)

	err = txn.Commit(ctx)
	require.NoError(t, err)
}

// TODO decide if corekv should expose errors for
// creating transactions when DB is closed.
// Note: Badger doesn't natively provide this either
//
// func TestNewConcurrentTxnFromWithStoreClosed(t *testing.T) {
// 	ctx := context.Background()
// 	rootstore := getTxnDB(t, ctx)

// 	rootstore.Close()

// 	_, err := NewConcurrentTxnFrom(ctx, rootstore, 0, false)
// 	require.ErrorIs(t, err, badgerds.ErrClosed)
// }

func TestConcurrentTxnSync(t *testing.T) {
	ctx := context.Background()
	rootstore := getBadgerTxnDB(t, ctx)

	txn := rootstore.NewTxn(false)

	cTxn := &concurrentTxn{Txn: txn}
	err := cTxn.Sync(ctx)
	require.NoError(t, err)
}

func TestConcurrentTxnClose(t *testing.T) {
	ctx := context.Background()
	opts := badgerds.Options{Options: badgerds.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	require.NoError(t, err)

	txn, err := rootstore.NewTransaction(ctx, false)
	require.NoError(t, err)

	cTxn := &concurrentTxn{Txn: txn}
	err = cTxn.Close(ctx)
	require.NoError(t, err)
}
