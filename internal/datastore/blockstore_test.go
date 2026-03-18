// Copyright 2026 Democratized Data Foundation
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
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/immutable"
)

func TestBstore_IsMerged_CacheHit(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	noChunk := immutable.None[int]()

	// Use p2pBlockStore which creates the toMerge index alongside the block
	bs := P2PBlockstoreFrom(rootstore, noChunk)

	block := blocks.NewBlock([]byte("merged-cache-hit-test"))

	err := bs.Put(ctx, block)
	require.NoError(t, err)

	// Before marking as merged, IsMerged should return false
	// (block exists but toMerge index also exists)
	merged, err := bs.IsMerged(ctx, block.Cid())
	require.NoError(t, err)
	require.False(t, merged)

	// Mark as merged (removes the toMerge index and adds to cache)
	err = bs.MarkAsMerged(ctx, block.Cid())
	require.NoError(t, err)

	// Now IsMerged should return true (served from cache)
	merged, err = bs.IsMerged(ctx, block.Cid())
	require.NoError(t, err)
	require.True(t, merged)
}

func TestBstore_BatchMarkAsMerged(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	noChunk := immutable.None[int]()

	bs := P2PBlockstoreFrom(rootstore, noChunk)

	block1 := blocks.NewBlock([]byte("batch-merge-1"))
	block2 := blocks.NewBlock([]byte("batch-merge-2"))
	block3 := blocks.NewBlock([]byte("batch-merge-3"))

	// Put all blocks (creates toMerge index for each)
	for _, b := range []blocks.Block{block1, block2, block3} {
		err := bs.Put(ctx, b)
		require.NoError(t, err)
	}

	// Batch mark as merged
	cids := []cid.Cid{block1.Cid(), block2.Cid(), block3.Cid()}
	err := bs.BatchMarkAsMerged(ctx, cids)
	require.NoError(t, err)

	// All should now report as merged
	for _, c := range cids {
		merged, err := bs.IsMerged(ctx, c)
		require.NoError(t, err)
		require.True(t, merged, "CID %s should be merged", c.String())
	}
}

func TestBstore_BatchHas(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	noChunk := immutable.None[int]()

	bs := P2PBlockstoreFrom(rootstore, noChunk)

	block1 := blocks.NewBlock([]byte("batch-has-1"))
	block2 := blocks.NewBlock([]byte("batch-has-2"))
	missingBlock := blocks.NewBlock([]byte("batch-has-missing"))

	// Put only block1 and block2
	err := bs.Put(ctx, block1)
	require.NoError(t, err)
	err = bs.Put(ctx, block2)
	require.NoError(t, err)

	// BatchHas should show correct results
	result, err := bs.BatchHas(ctx, []cid.Cid{
		block1.Cid(),
		block2.Cid(),
		missingBlock.Cid(),
	})
	require.NoError(t, err)

	require.True(t, result[block1.Cid().String()])
	require.True(t, result[block2.Cid().String()])
	require.False(t, result[missingBlock.Cid().String()])
}

func TestBlindWriteBlockstore_PutSkipsHasCheck(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	noChunk := immutable.None[int]()

	bs := BlindWriteBlockstoreFrom(rootstore, noChunk)

	block := blocks.NewBlock([]byte("blind-write-test"))

	// First put
	err := bs.Put(ctx, block)
	require.NoError(t, err)

	// Verify the block exists
	has, err := bs.Has(ctx, block.Cid())
	require.NoError(t, err)
	require.True(t, has)

	// Second put should succeed without error (blind write doesn't check existence)
	err = bs.Put(ctx, block)
	require.NoError(t, err)

	// Block should still be retrievable
	got, err := bs.Get(ctx, block.Cid())
	require.NoError(t, err)
	require.Equal(t, block.RawData(), got.RawData())

	// Verify no toMerge index was created (blindWriteBlockstore doesn't create one)
	merged, err := bs.IsMerged(ctx, block.Cid())
	require.NoError(t, err)
	require.True(t, merged, "blindWriteBlockstore does not create toMerge index, so IsMerged should be true")
}
