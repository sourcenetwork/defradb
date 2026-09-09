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
	"sync"
	"testing"
	"time"

	blocks "github.com/ipfs/go-block-format"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/immutable"
)

// The refresh keeps a re-received block out of the sweep while it is still waiting to merge.
func TestRefresh_MovesMarkerWhileUnmerged(t *testing.T) {
	ctx := context.Background()
	_, _, rootstore := newTestBlockstore(t)
	bs := P2PBlockstoreFrom(rootstore, immutable.None[int]())
	block := blocks.NewBlock([]byte("unmerged"))

	require.NoError(t, bs.Put(ctx, block))
	first := backdateMarker(t, ctx, rootstore, block)

	require.NoError(t, bs.Put(ctx, block))
	second := markerStamp(t, ctx, rootstore, block)

	require.True(t, second.After(first), "re-put of an unmerged block should move the marker forward")
}

// A block that has merged has no marker, and re-receiving it must not make it look unmerged again.
func TestRefresh_LeavesMergedBlockMerged(t *testing.T) {
	ctx := context.Background()
	_, _, rootstore := newTestBlockstore(t)
	bs := P2PBlockstoreFrom(rootstore, immutable.None[int]())
	block := blocks.NewBlock([]byte("merged"))

	require.NoError(t, bs.Put(ctx, block))
	require.NoError(t, bs.MarkAsMerged(ctx, block.Cid()))

	require.NoError(t, bs.Put(ctx, block))

	merged, err := bs.IsMerged(ctx, block.Cid())
	require.NoError(t, err)
	require.True(t, merged, "re-put of a merged block must leave it merged")
}

// A block that merges while it is being re-received must stay merged, whichever of the two commits
// first. The refresh reads the marker inside its own transaction, so a merge clearing that marker
// concurrently fails the refresh commit rather than putting the marker back.
func TestRefresh_ConcurrentMergeLeavesBlockMerged(t *testing.T) {
	ctx := context.Background()
	_, _, rootstore := newTestBlockstore(t)
	bs := P2PBlockstoreFrom(rootstore, immutable.None[int]())

	const attempts = 300
	resurrected := 0

	for i := 0; i < attempts; i++ {
		block := blocks.NewBlock([]byte{byte(i), byte(i >> 8), 'r'})
		require.NoError(t, bs.Put(ctx, block))

		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); _ = bs.Put(ctx, block) }()
		go func() { defer wg.Done(); _ = bs.MarkAsMerged(ctx, block.Cid()) }()
		wg.Wait()

		merged, err := bs.IsMerged(ctx, block.Cid())
		require.NoError(t, err)
		if !merged {
			resurrected++
		}
	}

	require.Zero(t, resurrected, "%d of %d merged blocks were marked unmerged again", resurrected, attempts)
}

// PutMany takes the same early return as Put for a block that is already stored, and a batch can
// hold both states at once: one block still waiting to merge, another already merged.
func TestRefresh_PutManyGuardsEachBlock(t *testing.T) {
	ctx := context.Background()
	_, _, rootstore := newTestBlockstore(t)
	bs := P2PBlockstoreFrom(rootstore, immutable.None[int]())

	unmerged := blocks.NewBlock([]byte("batched, still waiting"))
	merged := blocks.NewBlock([]byte("batched, already merged"))

	require.NoError(t, bs.PutMany(ctx, []blocks.Block{unmerged, merged}))
	require.NoError(t, bs.MarkAsMerged(ctx, merged.Cid()))
	first := backdateMarker(t, ctx, rootstore, unmerged)

	require.NoError(t, bs.PutMany(ctx, []blocks.Block{unmerged, merged}))

	second := markerStamp(t, ctx, rootstore, unmerged)
	require.True(t, second.After(first), "re-put of an unmerged block should move the marker forward")

	stillMerged, err := bs.IsMerged(ctx, merged.Cid())
	require.NoError(t, err)
	require.True(t, stillMerged, "re-put of a merged block must leave it merged")
}

func markerStamp(t *testing.T, ctx context.Context, rootstore corekv.Reader, block blocks.Block) time.Time {
	t.Helper()
	raw, err := rootstore.Get(ctx, append([]byte{blockStoreKey}, newToMergeKey(block.Cid().Bytes())...))
	require.NoError(t, err)
	stamp, ok := toMergeTime(raw)
	require.True(t, ok, "marker should carry a timestamp")
	return stamp
}

// backdateMarker rewinds a block's marker and returns the time it now carries, so a refresh is
// observable without waiting out the marker's one-second resolution.
func backdateMarker(t *testing.T, ctx context.Context, rootstore corekv.Writer, block blocks.Block) time.Time {
	t.Helper()
	stamp := time.Now().Add(-time.Hour)
	require.NoError(t, rootstore.Set(ctx,
		append([]byte{blockStoreKey}, newToMergeKey(block.Cid().Bytes())...), newToMergeValue(stamp)))
	return stamp
}
