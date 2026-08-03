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
	blockNS, rootstore := newTestBlockstore(t)
	defer rootstore.Close()

	cutoff := time.Unix(1_700_000_000, 0)

	// Marker older than the cutoff: an abandoned fetch, must be reclaimed.
	stale := putMarkedBlock(t, ctx, blockNS, []byte("stale"), newToMergeValue(cutoff.Add(-time.Hour)))
	// Marker newer than the cutoff: a fetch still in flight, must be kept.
	fresh := putMarkedBlock(t, ctx, blockNS, []byte("fresh"), newToMergeValue(cutoff.Add(time.Hour)))
	// Older single-byte marker: reclaimed, since it decodes as long abandoned.
	legacy := putMarkedBlock(t, ctx, blockNS, []byte("legacy"), []byte{0xff})
	// A merged block has no marker and must never be touched.
	merged := putUnmarkedBlock(t, ctx, blockNS, []byte("merged"))

	next, reclaimed, scanned, err := ReclaimOrphanBlocks(ctx, rootstore, cutoff, nil, 100)
	require.NoError(t, err)
	require.Nil(t, next, "a completed sweep returns no resume key")
	require.Equal(t, 2, reclaimed)
	require.Equal(t, 3, scanned, "only the three marked blocks are scanned, not the merged block")

	requireReclaimed(t, ctx, blockNS, stale)
	requireReclaimed(t, ctx, blockNS, legacy)
	requireKept(t, ctx, blockNS, fresh, true)
	requireKept(t, ctx, blockNS, merged, false)
}

func TestReclaimOrphanBlocksResumesAcrossCalls(t *testing.T) {
	ctx := context.Background()
	blockNS, rootstore := newTestBlockstore(t)
	defer rootstore.Close()

	cutoff := time.Unix(1_700_000_000, 0)
	stale := newToMergeValue(cutoff.Add(-time.Hour))

	const total = 5
	cids := make([][]byte, total)
	for i := range cids {
		cids[i] = putMarkedBlock(t, ctx, blockNS, []byte{byte('a' + i)}, stale)
	}

	// A scan limit below the marker count forces the sweep to resume via the cursor.
	var cursor []byte
	var reclaimedTotal, scannedTotal, calls int
	for {
		next, reclaimed, scanned, err := ReclaimOrphanBlocks(ctx, rootstore, cutoff, cursor, 2)
		require.NoError(t, err)
		reclaimedTotal += reclaimed
		scannedTotal += scanned
		calls++
		require.LessOrEqual(t, calls, total+1, "sweep did not make progress")
		if next == nil {
			break
		}
		cursor = next
	}

	require.Greater(t, calls, 1, "a scan limit below the index size must take several calls")
	require.Equal(t, total, reclaimedTotal)
	require.Equal(t, total, scannedTotal, "each marker is scanned exactly once across the resumed calls")
	for _, cid := range cids {
		requireReclaimed(t, ctx, blockNS, cid)
	}
}

func newTestBlockstore(t *testing.T) (corekv.ReaderWriter, corekv.Store) {
	t.Helper()
	rootstore, err := ckbadger.NewDatastore("", badgerds.DefaultOptions("").WithInMemory(true))
	require.NoError(t, err)
	return namespace.Wrap(rootstore, []byte{blockStoreKey}), rootstore
}

func putMarkedBlock(t *testing.T, ctx context.Context, blockNS corekv.ReaderWriter, data, marker []byte) []byte {
	t.Helper()
	block := blocks.NewBlock(data)
	cid := block.Cid().Bytes()
	require.NoError(t, blockNS.Set(ctx, newToMergeKey(cid), marker))
	require.NoError(t, blockNS.Set(ctx, cid, block.RawData()))
	return cid
}

func putUnmarkedBlock(t *testing.T, ctx context.Context, blockNS corekv.ReaderWriter, data []byte) []byte {
	t.Helper()
	block := blocks.NewBlock(data)
	cid := block.Cid().Bytes()
	require.NoError(t, blockNS.Set(ctx, cid, block.RawData()))
	return cid
}

func requireReclaimed(t *testing.T, ctx context.Context, blockNS corekv.ReaderWriter, cid []byte) {
	t.Helper()
	hasBlock, err := blockNS.Has(ctx, cid)
	require.NoError(t, err)
	require.False(t, hasBlock, "block should be reclaimed")
	hasMarker, err := blockNS.Has(ctx, newToMergeKey(cid))
	require.NoError(t, err)
	require.False(t, hasMarker, "marker should be reclaimed")
}

func requireKept(t *testing.T, ctx context.Context, blockNS corekv.ReaderWriter, cid []byte, withMarker bool) {
	t.Helper()
	hasBlock, err := blockNS.Has(ctx, cid)
	require.NoError(t, err)
	require.True(t, hasBlock, "block should be kept")
	hasMarker, err := blockNS.Has(ctx, newToMergeKey(cid))
	require.NoError(t, err)
	require.Equal(t, withMarker, hasMarker, "marker presence should be unchanged")
}
