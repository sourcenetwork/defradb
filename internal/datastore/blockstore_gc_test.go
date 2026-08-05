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
	"time"

	badgerds "github.com/dgraph-io/badger/v4"
	blocks "github.com/ipfs/go-block-format"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv"
	ckbadger "github.com/sourcenetwork/corekv/badger"
	"github.com/sourcenetwork/corekv/namespace"

	"github.com/sourcenetwork/defradb/internal/keys"
)

func TestToMergeValueRoundTrip(t *testing.T) {
	want := time.Unix(1_700_000_000, 0)
	got, ok := toMergeTime(newToMergeValue(want))
	require.True(t, ok)
	require.Equal(t, want.Unix(), got.Unix())
}

func TestToMergeTimeRejectsLegacyMarker(t *testing.T) {
	// The older marker format is a single byte with no timestamp.
	_, ok := toMergeTime([]byte{0xff})
	require.False(t, ok)
}

func TestReclaimOrphanBlocks(t *testing.T) {
	ctx := context.Background()
	blockNS, _, rootstore := newTestBlockstore(t)
	defer func() { require.NoError(t, rootstore.Close()) }()

	cutoff := time.Unix(1_700_000_000, 0)

	// Marker older than the cutoff: an abandoned fetch, must be reclaimed.
	stale := putMarkedBlock(t, ctx, blockNS, []byte("stale"), newToMergeValue(cutoff.Add(-time.Hour)))
	// Marker newer than the cutoff: a fetch still in flight, must be kept.
	fresh := putMarkedBlock(t, ctx, blockNS, []byte("fresh"), newToMergeValue(cutoff.Add(time.Hour)))
	// A marker with no timestamp cannot be aged, and comes from a store with no owner edges
	// to fall back on, so it is left alone.
	legacy := putMarkedBlock(t, ctx, blockNS, []byte("legacy"), []byte{0xff})
	// A merged block has no marker and must never be touched.
	merged := putUnmarkedBlock(t, ctx, blockNS, []byte("merged"))

	result, err := ReclaimOrphanBlocks(ctx, rootstore, cutoff, nil, 100)
	require.NoError(t, err)
	require.Nil(t, result.NextKey, "a completed sweep returns no resume key")
	require.Equal(t, 1, result.Reclaimed)
	require.Equal(t, 3, result.Scanned, "only the three marked blocks are scanned, not the merged block")
	require.Zero(t, result.Repaired)
	require.Zero(t, result.Conflicts)

	requireReclaimed(t, ctx, blockNS, stale)
	requireKept(t, ctx, blockNS, legacy, true)
	requireKept(t, ctx, blockNS, fresh, true)
	requireKept(t, ctx, blockNS, merged, false)
}

// A store upgraded in place carries single-byte markers over blocks that merged before the
// index recorded fetch times, and that merge wrote no owner edge, so an age-only decision
// deletes blocks a live document still links to.
func TestReclaimOrphanBlocksNeverReclaimsUntimestampedMarker(t *testing.T) {
	ctx := context.Background()
	blockNS, _, rootstore := newTestBlockstore(t)
	defer func() { require.NoError(t, rootstore.Close()) }()

	legacy := putMarkedBlock(t, ctx, blockNS, []byte("legacy"), []byte{0xff})

	// Far enough ahead that any readable timestamp would be expired.
	result, err := ReclaimOrphanBlocks(ctx, rootstore, time.Now().Add(time.Hour), nil, 100)
	require.NoError(t, err)
	require.Equal(t, 1, result.Scanned)
	require.Zero(t, result.Reclaimed)
	require.Zero(t, result.Repaired, "the marker stays, so it is not counted as repaired either")

	requireKept(t, ctx, blockNS, legacy, true)
}

// A block a document owns must survive an expired marker over it. The marker is not proof
// the block is garbage; ownership is what decides.
func TestReclaimOrphanBlocksKeepsBlockOwnedByDocument(t *testing.T) {
	ctx := context.Background()
	blockNS, systemNS, rootstore := newTestBlockstore(t)
	defer func() { require.NoError(t, rootstore.Close()) }()

	cutoff := time.Unix(1_700_000_000, 0)
	owned := putMarkedBlock(t, ctx, blockNS, []byte("owned"), newToMergeValue(cutoff.Add(-time.Hour)))

	ownerKey := keys.NewBlockCIDToDocIDKey(owned.Cid().String(), "bae-some-document").Bytes()
	require.NoError(t, systemNS.Set(ctx, ownerKey, []byte{}))

	result, err := ReclaimOrphanBlocks(ctx, rootstore, cutoff, nil, 100)
	require.NoError(t, err)
	require.Zero(t, result.Reclaimed)
	require.Equal(t, 1, result.Repaired)

	requireKept(t, ctx, blockNS, owned, false)
}

// A batch holding both kinds of outcome: the ownership lookup iterates the store from
// inside the same transaction that has already deleted an earlier block, so the two must
// not interfere.
func TestReclaimOrphanBlocksMixesReclaimAndRepairInOneBatch(t *testing.T) {
	ctx := context.Background()
	blockNS, systemNS, rootstore := newTestBlockstore(t)
	defer func() { require.NoError(t, rootstore.Close()) }()

	cutoff := time.Unix(1_700_000_000, 0)
	expired := newToMergeValue(cutoff.Add(-time.Hour))

	orphans := make([]blocks.Block, 3)
	for i := range orphans {
		orphans[i] = putMarkedBlock(t, ctx, blockNS, []byte{byte('a' + i)}, expired)
	}
	owned := putMarkedBlock(t, ctx, blockNS, []byte("owned"), expired)
	ownerKey := keys.NewBlockCIDToDocIDKey(owned.Cid().String(), "bae-some-document").Bytes()
	require.NoError(t, systemNS.Set(ctx, ownerKey, []byte{}))

	result, err := ReclaimOrphanBlocks(ctx, rootstore, cutoff, nil, 100)
	require.NoError(t, err)
	require.Equal(t, 3, result.Reclaimed)
	require.Equal(t, 1, result.Repaired)
	require.Equal(t, 4, result.Scanned)

	for _, block := range orphans {
		requireReclaimed(t, ctx, blockNS, block)
	}
	requireKept(t, ctx, blockNS, owned, false)
}

// The scan nominates candidates, then the delete phase re-reads each marker. A merge
// that commits in that window clears the marker, and the block must be left alone.
func TestReclaimBatchSkipsMarkerClearedAfterScan(t *testing.T) {
	ctx := context.Background()
	blockNS, _, rootstore := newTestBlockstore(t)
	defer func() { require.NoError(t, rootstore.Close()) }()

	cutoff := time.Unix(1_700_000_000, 0)
	block := putMarkedBlock(t, ctx, blockNS, []byte("merged-after-scan"), newToMergeValue(cutoff.Add(-time.Hour)))
	marker := newToMergeKey(block.Cid().Bytes())

	// Stand in for the merge committing between the scan and the delete phase:
	// MarkAsMerged removes the marker inside the merge transaction.
	require.NoError(t, blockNS.Delete(ctx, marker))

	var result SweepResult
	require.NoError(t, reclaimBatch(ctx, rootstore, cutoff, [][]byte{marker}, &result))
	require.Zero(t, result.Reclaimed)

	hasBlock, err := blockNS.Has(ctx, block.Cid().Bytes())
	require.NoError(t, err)
	require.True(t, hasBlock, "a block whose merge committed after the scan must survive")
}

// The delete phase re-decides on its own, so it has to apply the same timestamp rule the scan
// does. Handed an untimestamped marker directly, it must still leave the block alone.
func TestReclaimBatchSkipsUntimestampedMarker(t *testing.T) {
	ctx := context.Background()
	blockNS, _, rootstore := newTestBlockstore(t)
	defer func() { require.NoError(t, rootstore.Close()) }()

	block := putMarkedBlock(t, ctx, blockNS, []byte("legacy"), []byte{0xff})
	marker := newToMergeKey(block.Cid().Bytes())

	var result SweepResult
	require.NoError(t, reclaimBatch(ctx, rootstore, time.Now().Add(time.Hour), [][]byte{marker}, &result))
	require.Zero(t, result.Reclaimed)

	hasBlock, err := blockNS.Has(ctx, block.Cid().Bytes())
	require.NoError(t, err)
	require.True(t, hasBlock, "an unageable marker is not grounds to delete a block")
}

// The sweep's safety rests on the store aborting a transaction that deletes a marker it
// read when another transaction wrote that same key first. That is a storage-layer
// property, so it is asserted directly: turning conflict detection off would otherwise
// reintroduce the race silently.
func TestStoreDetectsConflictOnMarkerReadThenDelete(t *testing.T) {
	ctx := context.Background()
	blockNS, _, rootstore := newTestBlockstore(t)
	defer func() { require.NoError(t, rootstore.Close()) }()

	block := putMarkedBlock(t, ctx, blockNS, []byte("contended"), newToMergeValue(time.Unix(1, 0)))
	marker := newToMergeKey(block.Cid().Bytes())

	sweep := rootstore.NewTxn(false)
	defer sweep.Discard()
	sweepCtx := corekv.SetCtxTxn(ctx, sweep)
	sweepNS := namespace.Wrap(rootstore, []byte{blockStoreKey})

	// The sweep reads the marker, putting it in the transaction's read set.
	_, err := sweepNS.Get(sweepCtx, marker)
	require.NoError(t, err)

	// A merge commits against the same marker.
	require.NoError(t, blockNS.Delete(ctx, marker))

	require.NoError(t, sweepNS.Delete(sweepCtx, block.Cid().Bytes()))
	require.ErrorIs(t, sweep.Commit(), corekv.ErrTxnConflict,
		"a read marker written by another transaction must abort the sweep's commit")
}

func TestReclaimOrphanBlocksStopsOnCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	blockNS, _, rootstore := newTestBlockstore(t)
	defer func() { require.NoError(t, rootstore.Close()) }()

	cutoff := time.Unix(1_700_000_000, 0)
	block := putMarkedBlock(t, ctx, blockNS, []byte("stale"), newToMergeValue(cutoff.Add(-time.Hour)))

	cancel()
	_, err := ReclaimOrphanBlocks(ctx, rootstore, cutoff, nil, 100)
	require.ErrorIs(t, err, context.Canceled)
	requireKept(t, ctx, blockNS, block, true, "a cancelled sweep must not have deleted anything")
}

func TestReclaimOrphanBlocksResumesAcrossCalls(t *testing.T) {
	ctx := context.Background()
	blockNS, _, rootstore := newTestBlockstore(t)
	defer func() { require.NoError(t, rootstore.Close()) }()

	cutoff := time.Unix(1_700_000_000, 0)
	stale := newToMergeValue(cutoff.Add(-time.Hour))

	const total = 5
	created := make([]blocks.Block, total)
	for i := range created {
		created[i] = putMarkedBlock(t, ctx, blockNS, []byte{byte('a' + i)}, stale)
	}

	// A scan limit below the marker count forces the sweep to resume via the cursor.
	var cursor []byte
	var reclaimedTotal, scannedTotal, calls int
	for {
		result, err := ReclaimOrphanBlocks(ctx, rootstore, cutoff, cursor, 2)
		require.NoError(t, err)
		reclaimedTotal += result.Reclaimed
		scannedTotal += result.Scanned
		calls++
		require.LessOrEqual(t, calls, total+1, "sweep did not make progress")
		if result.NextKey == nil {
			break
		}
		cursor = result.NextKey
	}

	require.Greater(t, calls, 1, "a scan limit below the index size must take several calls")
	require.Equal(t, total, reclaimedTotal)
	require.Equal(t, total, scannedTotal, "each marker is scanned exactly once across the resumed calls")
	for _, block := range created {
		requireReclaimed(t, ctx, blockNS, block)
	}
}

func newTestBlockstore(t *testing.T) (blockNS, systemNS corekv.ReaderWriter, rootstore corekv.TxnStore) {
	t.Helper()
	store, err := ckbadger.NewDatastore("", badgerds.DefaultOptions("").WithInMemory(true))
	require.NoError(t, err)
	return namespace.Wrap(store, []byte{blockStoreKey}),
		namespace.Wrap(store, []byte{systemStoreKey}),
		store
}

func requireReclaimed(t *testing.T, ctx context.Context, blockNS corekv.ReaderWriter, block blocks.Block) {
	t.Helper()
	hasBlock, err := blockNS.Has(ctx, block.Cid().Bytes())
	require.NoError(t, err)
	require.False(t, hasBlock, "block should be reclaimed")
	hasMarker, err := blockNS.Has(ctx, newToMergeKey(block.Cid().Bytes()))
	require.NoError(t, err)
	require.False(t, hasMarker, "marker should be reclaimed")
}

func requireKept(
	t *testing.T,
	ctx context.Context,
	blockNS corekv.ReaderWriter,
	block blocks.Block,
	withMarker bool,
	msgAndArgs ...any,
) {
	t.Helper()
	hasBlock, err := blockNS.Has(ctx, block.Cid().Bytes())
	require.NoError(t, err)
	require.True(t, hasBlock, append([]any{"block should be kept"}, msgAndArgs...)...)
	hasMarker, err := blockNS.Has(ctx, newToMergeKey(block.Cid().Bytes()))
	require.NoError(t, err)
	require.Equal(t, withMarker, hasMarker, "marker presence should be unchanged")
}
