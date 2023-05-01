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
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/datastore/iterable"
	"github.com/sourcenetwork/defradb/datastore/memory"
)

func TestAsDSReaderWriterNonIterable(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	dsRW := AsDSReaderWriter(rootstore)

	key := ds.NewKey("source")

	err := dsRW.Put(ctx, key, []byte("hub"))
	require.NoError(t, err)

	val, err := dsRW.Get(ctx, key)
	require.NoError(t, err)
	require.Equal(t, []byte("hub"), val)
}

func TestAsDSReaderWriterIterable(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	store := shim{
		rootstore,
		iterable.NewIterable(rootstore),
	}
	dsRW := AsDSReaderWriter(store)

	key := ds.NewKey("source")

	err := dsRW.Put(ctx, key, []byte("hub"))
	require.NoError(t, err)

	val, err := dsRW.Get(ctx, key)
	require.NoError(t, err)
	require.Equal(t, []byte("hub"), val)
}
