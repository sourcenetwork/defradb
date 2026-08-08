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

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/memory"
	"github.com/stretchr/testify/require"
)

// Only a key read by one transaction and written by another can cause a conflict, so a
// key that is never read, or never written, must not outrank one that is both. Between
// two keys that are both, the one with the smaller side larger is the more contended.
func TestKeyTracerRanksByContendedSide(t *testing.T) {
	tr := newKeyTracer(true)

	// Both read and written, but only ever by one transaction on the read side.
	tr.record(1, []byte("lopsided"), false)
	for txn := uint64(100); txn < 160; txn++ {
		tr.record(txn, []byte("lopsided"), true)
	}
	// Fewer touches overall, but contended from both sides.
	for txn := uint64(1); txn <= 5; txn++ {
		tr.record(txn, []byte("contended"), false)
		tr.record(txn+50, []byte("contended"), true)
	}
	// Written by many, never read: cannot conflict, so it is not reported at all.
	for txn := uint64(200); txn < 240; txn++ {
		tr.record(txn, []byte("write-only"), true)
	}
	// Read by many, never written: cannot conflict either.
	for txn := uint64(300); txn < 340; txn++ {
		tr.record(txn, []byte("read-only"), false)
	}

	entries, dropped := tr.snapshot()
	require.Equal(t, 0, dropped)

	require.Equal(t, "contended", entries[0].key, "the key contended from both sides must rank first")
	require.Equal(t, 5, entries[0].readers)
	require.Equal(t, 5, entries[0].writers)

	// lopsided has 60 writers against contended's 5, so ranking on writers alone would
	// put it first.
	require.Equal(t, "lopsided", entries[1].key)
	require.Equal(t, 60, entries[1].writers)

	byKey := map[string]traceEntry{}
	for _, e := range entries {
		byKey[e.key] = e
	}
	require.NotContains(t, byKey, "write-only", "a key nothing reads cannot conflict")
	require.Equal(t, 0, byKey["read-only"].writers)
}

// The window is per interval, so a report must not carry counts into the next one.
func TestKeyTracerSnapshotClearsWindow(t *testing.T) {
	tr := newKeyTracer(true)
	tr.record(1, []byte("k"), false)
	tr.record(2, []byte("k"), true)
	_, _ = tr.snapshot()

	entries, _ := tr.snapshot()
	require.Empty(t, entries)
}

// A disabled tracer must not accumulate, so the default build carries no cost.
func TestKeyTracerDisabledRecordsNothing(t *testing.T) {
	tr := newKeyTracer(false)
	tr.record(1, []byte("k"), false)
	tr.record(1, []byte("k"), true)

	entries, _ := tr.snapshot()
	require.Empty(t, entries)
}

func TestStoreName(t *testing.T) {
	require.Equal(t, "system", storeName(string([]byte{systemStoreKey})+"x"))
	require.Equal(t, "data", storeName(string([]byte{dataStoreKey})+"x"))
	require.Equal(t, "block/to-merge", storeName(string([]byte{blockStoreKey, toMergeIndexPrefix})+"x"))
	require.Equal(t, "block", storeName(string([]byte{blockStoreKey})+"x"))
}

// Keys reached by scanning join the read set just like a direct Get, so a conflict
// caused by a scan has to be visible in the trace.
func TestTracedIteratorRecordsScannedKeys(t *testing.T) {
	conflictTracer = newKeyTracer(true)
	t.Cleanup(func() { conflictTracer = newKeyTracer(false) })

	store := memory.NewDatastore(context.Background())
	require.NoError(t, store.Set(context.Background(), []byte("a"), []byte("1")))
	require.NoError(t, store.Set(context.Background(), []byte("b"), []byte("2")))

	traced := traceWrites(store, 7)
	iter, err := traced.Iterator(context.Background(), corekv.IterOptions{})
	require.NoError(t, err)
	for {
		ok, err := iter.Next()
		require.NoError(t, err)
		if !ok {
			break
		}
	}
	require.NoError(t, iter.Close())

	entries, _ := conflictTracer.snapshot()
	scanned := map[string]bool{}
	for _, e := range entries {
		scanned[e.key] = true
	}
	require.True(t, scanned["a"], "scanned key not recorded as a read")
	require.True(t, scanned["b"], "scanned key not recorded as a read")
}
