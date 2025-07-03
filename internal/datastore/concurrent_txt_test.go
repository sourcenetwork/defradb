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

	badgerds "github.com/dgraph-io/badger/v4"
	"github.com/sourcenetwork/corekv/badger"
	"github.com/sourcenetwork/corekv/memory"

	"github.com/stretchr/testify/require"
)

func getBadgerTxnDB(t *testing.T) *badger.Datastore {
	opts := badgerds.DefaultOptions("").WithInMemory(true)
	rootstore, err := badger.NewDatastore("", opts)
	require.NoError(t, err)

	return rootstore
}

func TestNewConcurrentTxnFrom(t *testing.T) {
	ctx := context.Background()
	rootstore := getBadgerTxnDB(t)

	txn := NewConcurrentTxnFrom(ctx, rootstore, 0, false)

	err := txn.Commit(ctx)
	require.NoError(t, err)
}

func TestNewConcurrentTxnFromNonIterable(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	txn := NewConcurrentTxnFrom(ctx, rootstore, 0, false)

	err := txn.Commit(ctx)
	require.NoError(t, err)
}

func TestConcurrentTxnSync(t *testing.T) {
	ctx := context.Background()
	rootstore := getBadgerTxnDB(t)

	txn := rootstore.NewTxn(false)

	cTxn := &concurrentTxn{Txn: txn}
	err := cTxn.Sync(ctx)
	require.NoError(t, err)
}
