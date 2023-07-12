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

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	ccid "github.com/sourcenetwork/defradb/core/cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/datastore/memory"
)

var (
	data  = []byte("Source Inc")
	data2 = []byte("SourceHub")
)

func TestBStoreGet(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	dsRW := AsDSReaderWriter(rootstore)

	bs := bstore{
		store: dsRW,
	}

	cID, err := ccid.NewSHA256CidV1(data)
	require.NoError(t, err)
	b, err := blocks.NewBlockWithCid(data, cID)
	require.NoError(t, err)
	err = bs.Put(ctx, b)
	require.NoError(t, err)

	b2, err := bs.Get(ctx, cID)
	require.NoError(t, err)

	require.Equal(t, data, b2.RawData())
}

func TestBStoreGetWithUndefinedCID(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	dsRW := AsDSReaderWriter(rootstore)

	bs := bstore{
		store: dsRW,
	}

	cID, err := ccid.NewSHA256CidV1(data)
	require.NoError(t, err)
	b, err := blocks.NewBlockWithCid(data, cID)
	require.NoError(t, err)
	err = bs.Put(ctx, b)
	require.NoError(t, err)

	_, err = bs.Get(ctx, cid.Cid{})
	require.ErrorIs(t, err, ipld.ErrNotFound{})
}

func TestBStoreGetWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	dsRW := AsDSReaderWriter(rootstore)

	bs := bstore{
		store: dsRW,
	}

	cID, err := ccid.NewSHA256CidV1(data)
	require.NoError(t, err)
	b, err := blocks.NewBlockWithCid(data, cID)
	require.NoError(t, err)
	err = bs.Put(ctx, b)
	require.NoError(t, err)

	err = rootstore.Close()
	require.NoError(t, err)

	_, err = bs.Get(ctx, cID)
	require.ErrorIs(t, err, memory.ErrClosed)
}

func TestBStoreGetWithReHash(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	dsRW := AsDSReaderWriter(rootstore)

	bs := bstore{
		store: dsRW,
	}

	bs.HashOnRead(true)

	cID, err := ccid.NewSHA256CidV1(data)
	require.NoError(t, err)
	b, err := blocks.NewBlockWithCid(data, cID)
	require.NoError(t, err)
	err = bs.Put(ctx, b)
	require.NoError(t, err)

	b2, err := bs.Get(ctx, cID)
	require.NoError(t, err)

	require.Equal(t, data, b2.RawData())
}

func TestPutMany(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	dsRW := AsDSReaderWriter(rootstore)

	bs := bstore{
		store: dsRW,
	}

	cID, err := ccid.NewSHA256CidV1(data)
	require.NoError(t, err)
	b, err := blocks.NewBlockWithCid(data, cID)
	require.NoError(t, err)

	cID2, err := ccid.NewSHA256CidV1(data)
	require.NoError(t, err)
	b2, err := blocks.NewBlockWithCid(data2, cID2)
	require.NoError(t, err)

	err = bs.PutMany(ctx, []blocks.Block{b, b2})
	require.NoError(t, err)
}

func TestPutManyWithExists(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	dsRW := AsDSReaderWriter(rootstore)

	bs := bstore{
		store: dsRW,
	}

	cID, err := ccid.NewSHA256CidV1(data)
	require.NoError(t, err)
	b, err := blocks.NewBlockWithCid(data, cID)
	require.NoError(t, err)

	err = bs.Put(ctx, b)
	require.NoError(t, err)

	cID2, err := ccid.NewSHA256CidV1(data)
	require.NoError(t, err)
	b2, err := blocks.NewBlockWithCid(data2, cID2)
	require.NoError(t, err)

	err = bs.PutMany(ctx, []blocks.Block{b, b2})
	require.NoError(t, err)
}

func TestPutManyWithStoreClosed(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	dsRW := AsDSReaderWriter(rootstore)

	bs := bstore{
		store: dsRW,
	}

	cID, err := ccid.NewSHA256CidV1(data)
	require.NoError(t, err)
	b, err := blocks.NewBlockWithCid(data, cID)
	require.NoError(t, err)

	cID2, err := ccid.NewSHA256CidV1(data2)
	require.NoError(t, err)
	b2, err := blocks.NewBlockWithCid(data2, cID2)
	require.NoError(t, err)

	err = rootstore.Close()
	require.NoError(t, err)

	err = bs.PutMany(ctx, []blocks.Block{b, b2})
	require.ErrorIs(t, err, memory.ErrClosed)
}
