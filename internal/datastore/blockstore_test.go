// Copyright 2025 Democratized Data Foundation
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
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/corekv/namespace"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/internal/db/lock"
)

func TestIsMerged_RolledBackTxnLeavesBlockUnmerged(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	// A fetched but unmerged block: stored, with its to-merge marker set.
	blockNS := namespace.Wrap(rootstore, []byte{blockStoreKey})
	block := putMarkedBlock(t, ctx, blockNS, []byte("rolled back"), []byte{objectMarker})

	txn := NewTxnFrom(rootstore, lock.NewLockSet(), 1, false, immutable.None[int]())
	require.NoError(t, txn.Blockstore().MarkAsMerged(ctx, block.Cid()))
	txn.Discard()

	merged, err := BlockstoreFrom(rootstore, immutable.None[int]()).IsMerged(ctx, block.Cid())
	require.NoError(t, err)
	require.False(t, merged)
}

func TestIsMerged_CommittedTxnMarksBlockMerged(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	blockNS := namespace.Wrap(rootstore, []byte{blockStoreKey})
	block := putMarkedBlock(t, ctx, blockNS, []byte("committed"), []byte{objectMarker})

	txn := NewTxnFrom(rootstore, lock.NewLockSet(), 1, false, immutable.None[int]())
	require.NoError(t, txn.Blockstore().MarkAsMerged(ctx, block.Cid()))
	require.NoError(t, txn.Commit())

	merged, err := BlockstoreFrom(rootstore, immutable.None[int]()).IsMerged(ctx, block.Cid())
	require.NoError(t, err)
	require.True(t, merged)
}

func TestIsMerged_DeletedBlockIsNotMerged(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	blockNS := namespace.Wrap(rootstore, []byte{blockStoreKey})
	block := putMarkedBlock(t, ctx, blockNS, []byte("deleted"), []byte{objectMarker})

	blockstore := BlockstoreFrom(rootstore, immutable.None[int]())
	require.NoError(t, blockstore.MarkAsMerged(ctx, block.Cid()))

	merged, err := blockstore.IsMerged(ctx, block.Cid())
	require.NoError(t, err)
	require.True(t, merged)

	// History prune and the orphan sweep both delete blocks under a merged peer. Callers
	// treat a merged answer as "no need to fetch", so it has to follow the store.
	require.NoError(t, blockstore.DeleteBlock(ctx, block.Cid()))

	merged, err = blockstore.IsMerged(ctx, block.Cid())
	require.NoError(t, err)
	require.False(t, merged)
}

// A CAR import re-puts blocks the node already holds. Storing one unconditionally stamps a fresh
// to-merge marker over a block that already merged, which reads as an abandoned fetch and leaves
// the orphan sweep free to delete a block a live document still links to.
func TestIsMerged_ReputtingAMergedBlockLeavesItMerged(t *testing.T) {
	for _, tc := range []struct {
		name string
		put  func(context.Context, Blockstore, blocks.Block) error
	}{
		{
			name: "Put",
			put:  func(ctx context.Context, bs Blockstore, b blocks.Block) error { return bs.Put(ctx, b) },
		},
		{
			name: "PutMany",
			put: func(ctx context.Context, bs Blockstore, b blocks.Block) error {
				return bs.PutMany(ctx, []blocks.Block{b})
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			rootstore := memory.NewDatastore(ctx)

			blockNS := namespace.Wrap(rootstore, []byte{blockStoreKey})
			block := putMarkedBlock(t, ctx, blockNS, []byte("merged then re-imported"),
				[]byte{objectMarker})

			blockstore := BlockstoreFrom(rootstore, immutable.None[int]())
			require.NoError(t, blockstore.MarkAsMerged(ctx, block.Cid()))
			merged, err := blockstore.IsMerged(ctx, block.Cid())
			require.NoError(t, err)
			require.True(t, merged)

			require.NoError(t, tc.put(ctx, P2PBlockstoreFrom(rootstore, immutable.None[int]()), block))

			merged, err = blockstore.IsMerged(ctx, block.Cid())
			require.NoError(t, err)
			require.True(t, merged, "storing a block the node already holds must not unmerge it")
		})
	}
}

func TestIsMerged_UnmergedAndAbsentBlocks(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	blockNS := namespace.Wrap(rootstore, []byte{blockStoreKey})
	blockstore := BlockstoreFrom(rootstore, immutable.None[int]())

	marked := putMarkedBlock(t, ctx, blockNS, []byte("still marked"), []byte{objectMarker})
	merged, err := blockstore.IsMerged(ctx, marked.Cid())
	require.NoError(t, err)
	require.False(t, merged, "a block still carrying its to-merge marker is not merged")

	// Never stored at all: absent must not read as merged just because no marker exists.
	absent := putUnmarkedBlock(t, ctx, namespace.Wrap(memory.NewDatastore(ctx), []byte{blockStoreKey}), []byte("absent"))
	merged, err = blockstore.IsMerged(ctx, absent.Cid())
	require.NoError(t, err)
	require.False(t, merged)
}

// putUnmarkedBlock stores a block with no to-merge marker, the shape of one that has
// already merged.
func putUnmarkedBlock(t *testing.T, ctx context.Context, blockNS corekv.ReaderWriter, data []byte) blocks.Block {
	t.Helper()
	block := blocks.NewBlock(data)
	require.NoError(t, blockNS.Set(ctx, block.Cid().Bytes(), block.RawData()))
	return block
}

// putMarkedBlock stores a block along with a to-merge marker carrying the given value,
// the shape of one that has been fetched but not yet merged.
func putMarkedBlock(
	t *testing.T,
	ctx context.Context,
	blockNS corekv.ReaderWriter,
	data, marker []byte,
) blocks.Block {
	t.Helper()
	block := putUnmarkedBlock(t, ctx, blockNS, data)
	require.NoError(t, blockNS.Set(ctx, newToMergeKey(block.Cid().Bytes()), marker))
	return block
}
