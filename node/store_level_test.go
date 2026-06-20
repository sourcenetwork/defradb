// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package node

import (
	"context"
	"os"
	"path/filepath"
	"runtime/pprof"
	"runtime/trace"
	"strings"
	"testing"
	"time"

	"github.com/sourcenetwork/corekv/leveldb"
	"github.com/stretchr/testify/require"
)

// These tests characterize the goleveldb transaction model that DefraDB's leveldb backend
// inherits, and which underlies the deadlock in https://github.com/sourcenetwork/defradb/issues/4959.
//
// The corekv leveldb package documents the contract:
//
//	"Only one transaction can be opened at a time. Subsequent call to Write and OpenTransaction
//	 will be blocked until in-flight transaction is committed or discarded."
//
// DefraDB's Truncate/RefreshView flow violates this by performing txn-free writes (the
// action-execution status markers) while an outer transaction is still open, and by allowing two
// transactions to be open concurrently. These tests pin the underlying behavior so that, when the
// backend or the flow is changed to resolve #4959, the assumptions here are revisited explicitly.
//
// The waits are timing-based by necessity (proving an operation blocks), but generously bounded.
//
// # Visualizing the deadlock
//
// Set DEFRA_TRACE=1 to capture a Go execution trace per test (written to $DEFRA_TRACE_DIR, or the
// OS temp dir by default). Goroutines are labelled via runtime/trace tasks+regions and
// runtime/pprof labels so the blocking goroutine is easy to spot. View with:
//
//	DEFRA_TRACE=1 go test ./node/ -run TestStoreLevel_TxnFreeWriteBlocksWhileTxnOpen -v
//	go tool trace <printed-path>          # or: gotraceui <printed-path>
//
// In go tool trace, see "User-defined tasks" / "Goroutine analysis" — the txn-free-writer goroutine
// sits blocked in its region for the whole probe window while T1 is held open.

const levelTxnBlockProbe = 500 * time.Millisecond

// startTrace begins an execution trace for the test when DEFRA_TRACE is set, and returns the root
// context to attach goroutine labels/regions to. When tracing is disabled it is a cheap no-op so
// normal/CI runs are unaffected.
func startTrace(t *testing.T) context.Context {
	t.Helper()
	if os.Getenv("DEFRA_TRACE") == "" {
		return context.Background()
	}

	dir := os.Getenv("DEFRA_TRACE_DIR")
	if dir == "" {
		dir = os.TempDir()
	}
	require.NoError(t, os.MkdirAll(dir, 0o755))
	path := filepath.Join(dir, "defra-leveldb-"+strings.ReplaceAll(t.Name(), "/", "_")+".trace")
	// Log an absolute path: `go test` runs with the working directory set to the package dir, so a
	// relative DEFRA_TRACE_DIR would otherwise print a path that doesn't resolve from the repo root.
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}

	f, err := os.Create(path)
	require.NoError(t, err)
	require.NoError(t, trace.Start(f))

	t.Cleanup(func() {
		trace.Stop()
		_ = f.Close()
		t.Logf("execution trace written: %s", path)
		t.Logf("view with: go tool trace %s   (or: gotraceui %s)", path, path)
	})
	return context.Background()
}

// labelGoroutine tags the *current* goroutine with an "actor" pprof label (e.g. "T1") so it is
// identifiable in goroutine profiles and in go tool trace's goroutine analysis.
func labelGoroutine(ctx context.Context, actor string) {
	pprof.SetGoroutineLabels(pprof.WithLabels(ctx, pprof.Labels("actor", actor)))
}

// goLabeled spawns fn on a new goroutine, labelled both with an "actor" runtime/pprof label (visible
// in goroutine profiles) and a runtime/trace region nested under the test's task (visible in
// go tool trace). The labelling is what makes the blocked goroutine readable in the visualization.
func goLabeled(ctx context.Context, actor, region string, fn func()) {
	go func() {
		pprof.Do(ctx, pprof.Labels("actor", actor), func(ctx context.Context) {
			trace.WithRegion(ctx, region, fn)
		})
	}()
}

// Manifestation 1 primitive: a txn-free write blocks while a transaction is open. This is the
// single-goroutine self-deadlock behind Truncate/RefreshView (action.Register's txn-free
// Systemstore().Set, written via context.TODO(), blocks behind the flow's own open transaction).
func TestStoreLevel_TxnFreeWriteBlocksWhileTxnOpen(t *testing.T) {
	ctx, task := trace.NewTask(startTrace(t), "manifestation-1: self-deadlock")
	defer task.End()

	labelGoroutine(ctx, "T1") // this goroutine holds the open transaction

	store, err := leveldb.NewDatastore(t.TempDir(), nil)
	require.NoError(t, err)
	defer store.Close() //nolint:errcheck

	// Open a transaction and intentionally leave it open (holds the exclusive write lock).
	txn := store.NewTxn(false)
	trace.Log(ctx, "T1", "opened — holds leveldb write lock")

	// The competing actor performs a txn-free write (no transaction of its own). Labelled "T2"
	// in the trace for readability, though it is the absence of a txn that makes it block on T1.
	done := make(chan error, 1)
	goLabeled(ctx, "T2", "T2: txn-free Set — blocks while T1 open", func() {
		done <- store.Set(context.Background(), []byte("k"), []byte("v"))
	})

	select {
	case err := <-done:
		t.Fatalf("txn-free Set returned (err=%v) while a txn was open; expected it to block", err)
	case <-time.After(levelTxnBlockProbe):
		// Expected: the write is blocked behind the open transaction.
		trace.Log(ctx, "T2", "still blocked after probe window")
	}

	// Releasing the transaction must unblock the pending write.
	trace.Log(ctx, "T1", "Discard — releasing write lock (T2 can now proceed)")
	txn.Discard()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("txn-free Set did not unblock after txn Discard")
	}
}

// Manifestation 2 primitive: a second transaction cannot be opened while the first is still open.
// This is the concurrent-transaction deadlock behind the "..._Deadlocks..." collection-version tests.
func TestStoreLevel_SecondTxnBlocksWhileFirstOpen(t *testing.T) {
	ctx, task := trace.NewTask(startTrace(t), "manifestation-2: adversarial txns")
	defer task.End()

	labelGoroutine(ctx, "T1") // this goroutine holds the first transaction

	store, err := leveldb.NewDatastore(t.TempDir(), nil)
	require.NoError(t, err)
	defer store.Close() //nolint:errcheck

	txn1 := store.NewTxn(false)
	trace.Log(ctx, "T1", "opened — occupies the single txn slot")

	done := make(chan struct{}, 1)
	goLabeled(ctx, "T2", "T2: NewTxn/OpenTransaction — blocks while T1 open", func() {
		txn2 := store.NewTxn(false) // OpenTransaction blocks while txn1 is open
		txn2.Discard()
		done <- struct{}{}
	})

	select {
	case <-done:
		t.Fatal("second NewTxn returned while the first txn was open; expected it to block")
	case <-time.After(levelTxnBlockProbe):
		// Expected: the second OpenTransaction is blocked.
		trace.Log(ctx, "T2", "OpenTransaction still blocked after probe window")
	}

	trace.Log(ctx, "T1", "Discard — releasing the txn slot (T2 can now open)")
	txn1.Discard()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("second NewTxn did not unblock after the first txn Discard")
	}
}

// Reads are not blocked by an open transaction (only writes and OpenTransaction are). This is why
// getStatus's Systemstore().Get succeeds and the first call to actually hang is the subsequent Set.
func TestStoreLevel_TxnFreeReadNotBlockedByOpenTxn(t *testing.T) {
	ctx, task := trace.NewTask(startTrace(t), "control: reads are not blocked")
	defer task.End()

	labelGoroutine(ctx, "T1") // this goroutine holds the open transaction

	store, err := leveldb.NewDatastore(t.TempDir(), nil)
	require.NoError(t, err)
	defer store.Close() //nolint:errcheck

	txn := store.NewTxn(false)
	defer txn.Discard()
	trace.Log(ctx, "T1", "opened — holds leveldb write lock")

	done := make(chan struct{}, 1)
	goLabeled(ctx, "T2", "T2: txn-free Get — succeeds despite T1 open", func() {
		_, _ = store.Get(context.Background(), []byte("missing"))
		done <- struct{}{}
	})

	select {
	case <-done:
		// Expected: the read completes despite the open transaction.
	case <-time.After(levelTxnBlockProbe):
		t.Fatal("txn-free Get blocked while a txn was open; expected reads not to block")
	}
}
