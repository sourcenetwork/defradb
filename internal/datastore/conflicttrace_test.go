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

// blame is what the trace exists to answer: of everything a rejected transaction read,
// which key did another transaction write first.
func TestTracerBlamesTheKeyAnotherTxnWrote(t *testing.T) {
	tr := newTracer(true)

	// The loser reads two keys.
	tr.begin(1, false)
	tr.record(1, []byte("contended"), false)
	tr.record(1, []byte("untouched"), false)

	// The winner writes one of them and commits first.
	tr.begin(2, false)
	tr.record(2, []byte("contended"), true)
	tr.record(2, []byte("something-else"), true)
	tr.commit(2, false)

	tr.mu.Lock()
	blamed, checked := tr.blameLocked(tr.open[1])
	tr.mu.Unlock()

	require.Equal(t, []string{"contended"}, blamed)
	require.Equal(t, 1, checked)
}

// A commit that landed before the transaction started cannot have caused its conflict,
// so it must not be blamed for one.
func TestTracerIgnoresCommitsFromBeforeTheTxnStarted(t *testing.T) {
	tr := newTracer(true)

	tr.begin(1, false)
	tr.record(1, []byte("contended"), true)
	tr.commit(1, false)

	// Starts after that commit, and reads the same key.
	tr.begin(2, false)
	tr.record(2, []byte("contended"), false)

	tr.mu.Lock()
	blamed, checked := tr.blameLocked(tr.open[2])
	tr.mu.Unlock()

	require.Empty(t, blamed)
	require.Equal(t, 0, checked)
}

// Writing a key is not enough to conflict on it; the rejected transaction has to have
// read it.
func TestTracerDoesNotBlameAKeyTheTxnOnlyWrote(t *testing.T) {
	tr := newTracer(true)

	tr.begin(1, false)
	tr.record(1, []byte("shared"), true)

	tr.begin(2, false)
	tr.record(2, []byte("shared"), true)
	tr.commit(2, false)

	tr.mu.Lock()
	blamed, _ := tr.blameLocked(tr.open[1])
	tr.mu.Unlock()

	require.Empty(t, blamed)
}

// A rejected transaction wrote nothing, so it must not join the history and be blamed
// for someone else's conflict.
func TestTracerConflictedTxnLeavesNoHistory(t *testing.T) {
	tr := newTracer(true)

	tr.begin(1, false)
	tr.record(1, []byte("k"), true)
	tr.commit(1, true)

	tr.mu.Lock()
	open, history := len(tr.open), len(tr.history)
	tr.mu.Unlock()

	require.Equal(t, 0, open)
	require.Equal(t, 0, history)
}

// An abandoned transaction has to be released, or its keys are held until the process
// exits.
func TestTracerDiscardReleasesTxn(t *testing.T) {
	tr := newTracer(true)

	tr.begin(1, false)
	tr.record(1, []byte("k"), false)
	tr.discard(1)

	tr.mu.Lock()
	open := len(tr.open)
	tr.mu.Unlock()

	require.Equal(t, 0, open)
}

// The history is bounded, so a long run cannot grow it without limit.
func TestTracerHistoryIsBounded(t *testing.T) {
	tr := newTracer(true)

	for i := uint64(1); i <= traceHistory+50; i++ {
		tr.begin(i, false)
		tr.record(i, []byte("k"), true)
		tr.commit(i, false)
	}

	tr.mu.Lock()
	history := len(tr.history)
	tr.mu.Unlock()

	require.Equal(t, traceHistory, history)
}

// A disabled tracer must not accumulate, so the default build carries no cost.
func TestTracerDisabledRecordsNothing(t *testing.T) {
	tr := newTracer(false)

	tr.begin(1, false)
	tr.record(1, []byte("k"), false)
	tr.commit(1, false)

	tr.mu.Lock()
	open, history := len(tr.open), len(tr.history)
	tr.mu.Unlock()

	require.Equal(t, 0, open)
	require.Equal(t, 0, history)
}

// Keys reached by scanning join the read set just like a direct Get, so a conflict
// caused by a scan has to be attributable.
func TestTracedIteratorRecordsScannedKeys(t *testing.T) {
	conflictTracer = newTracer(true)
	t.Cleanup(func() { conflictTracer = newTracer(false) })

	store := memory.NewDatastore(context.Background())
	require.NoError(t, store.Set(context.Background(), []byte("a"), []byte("1")))

	conflictTracer.begin(7, false)
	traced := traceKeys(store, 7, false)
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

	conflictTracer.mu.Lock()
	_, scanned := conflictTracer.open[7].reads["a"]
	conflictTracer.mu.Unlock()

	require.True(t, scanned, "scanned key was not recorded as a read")
}

func TestStoreName(t *testing.T) {
	require.Equal(t, "system", storeName(string([]byte{systemStoreKey})+"x"))
	require.Equal(t, "data", storeName(string([]byte{dataStoreKey})+"x"))
	require.Equal(t, "block/to-merge", storeName(string([]byte{blockStoreKey, toMergeIndexPrefix})+"x"))
	require.Equal(t, "block", storeName(string([]byte{blockStoreKey})+"x"))
}

// A read-only transaction cannot be rejected for conflict, so tracking one would only
// cost memory.
func TestTracerSkipsReadOnlyTxn(t *testing.T) {
	tr := newTracer(true)

	tr.begin(1, true)
	tr.record(1, []byte("k"), false)

	tr.mu.Lock()
	open := len(tr.open)
	tr.mu.Unlock()

	require.Equal(t, 0, open)
}
