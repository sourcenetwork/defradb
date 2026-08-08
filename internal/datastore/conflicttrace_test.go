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
	"testing"

	"github.com/stretchr/testify/require"
)

// The point of the trace is to separate a key many transactions share from keys each
// transaction owns, so ranking has to be by distinct transactions and not write volume.
func TestKeyTracerRanksSharedKeyAboveBusyPrivateKey(t *testing.T) {
	tr := newKeyTracer(true)

	// One key touched once by each of 5 transactions.
	for txn := uint64(1); txn <= 5; txn++ {
		tr.record(txn, []byte("shared"))
	}
	// A key written far more often, but only ever by one transaction.
	for i := 0; i < 50; i++ {
		tr.record(99, []byte("private"))
	}

	entries, dropped := tr.snapshot()
	require.Equal(t, 0, dropped)
	require.Len(t, entries, 2)
	require.Equal(t, "shared", entries[0].key)
	require.Equal(t, 5, entries[0].txns)
	require.Equal(t, "private", entries[1].key)
	require.Equal(t, 50, entries[1].writes)
	require.Equal(t, 1, entries[1].txns)
}

// The window is per interval, so a report must not carry counts into the next one.
func TestKeyTracerSnapshotClearsWindow(t *testing.T) {
	tr := newKeyTracer(true)
	tr.record(1, []byte("k"))
	_, _ = tr.snapshot()

	entries, _ := tr.snapshot()
	require.Empty(t, entries)
}

// A disabled tracer must not accumulate, so the default build carries no cost.
func TestKeyTracerDisabledRecordsNothing(t *testing.T) {
	tr := newKeyTracer(false)
	tr.record(1, []byte("k"))

	entries, _ := tr.snapshot()
	require.Empty(t, entries)
}

func TestStoreName(t *testing.T) {
	require.Equal(t, "system", storeName(string([]byte{systemStoreKey})+"x"))
	require.Equal(t, "data", storeName(string([]byte{dataStoreKey})+"x"))
	require.Equal(t, "block/to-merge", storeName(string([]byte{blockStoreKey, toMergeIndexPrefix})+"x"))
	require.Equal(t, "block", storeName(string([]byte{blockStoreKey})+"x"))
}
