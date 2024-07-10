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
	"github.com/ipfs/go-datastore/query"
	badger "github.com/sourcenetwork/badger/v4"
	"github.com/stretchr/testify/require"

	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v4"
	"github.com/sourcenetwork/defradb/datastore/memory"
)

var prefixKey = ds.NewKey("mystore")

func TestGetIterator(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	dsRW := AsDSReaderWriter(rootstore)
	dsRW = prefix(dsRW, prefixKey)

	_, err := dsRW.GetIterator(query.Query{})
	require.NoError(t, err)
}

type testOrder struct{}

func (t *testOrder) Compare(a, b query.Entry) int {
	return 0
}

func TestGetIteratorWithInvalidOrderType(t *testing.T) {
	ctx := context.Background()
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	require.NoError(t, err)

	iter, err := rootstore.NewIterableTransaction(ctx, false)
	require.NoError(t, err)

	dsRW := prefix(iter, prefixKey)

	_, err = dsRW.GetIterator(query.Query{
		Orders: []query.Order{
			&testOrder{},
		},
	})
	require.Error(t, err)
}

func TestQuery(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	dsRW := AsDSReaderWriter(rootstore)
	dsRW = prefix(dsRW, prefixKey)

	_, err := dsRW.Query(ctx, query.Query{})
	require.NoError(t, err)
}

func TestQueryWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	require.NoError(t, err)

	dsRW := AsDSReaderWriter(rootstore)
	dsRW = prefix(dsRW, prefixKey)

	err = rootstore.Close()
	require.NoError(t, err)

	_, err = dsRW.Query(ctx, query.Query{})
	require.ErrorIs(t, err, ErrClosed)
}

func TestIteratePrefix(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	dsRW := AsDSReaderWriter(rootstore)
	dsRW = prefix(dsRW, prefixKey)

	iter, err := dsRW.GetIterator(query.Query{})
	require.NoError(t, err)

	_, err = iter.IteratePrefix(ctx, ds.NewKey("key1"), ds.NewKey("key1"))
	require.NoError(t, err)
}

func TestIteratePrefixWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	require.NoError(t, err)

	dsRW := AsDSReaderWriter(rootstore)
	dsRW = prefix(dsRW, prefixKey)

	iter, err := dsRW.GetIterator(query.Query{})
	require.NoError(t, err)

	err = rootstore.Close()
	require.NoError(t, err)

	_, err = iter.IteratePrefix(ctx, ds.NewKey("key1"), ds.NewKey("key1"))
	require.ErrorIs(t, err, ErrClosed)
}
