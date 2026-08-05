// Copyright 2024 Democratized Data Foundation
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
	"crypto/rand"
	"testing"
	"time"

	badgerds "github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/utils"
)

func TestSetBadgerInMemory(t *testing.T) {
	opts := utils.NewOptions(options.Node().Store().SetBadgerInMemory(true).Node())
	assert.Equal(t, true, opts.Store.BadgerInMemory)
}

func TestSetBadgerFileSize(t *testing.T) {
	opts := utils.NewOptions(options.Node().Store().SetBadgerFileSize(int64(5 << 30)).Node())
	assert.Equal(t, int64(5<<30), opts.Store.BadgerFileSize)
}

func TestSetBadgerEncryptionKey(t *testing.T) {
	encryptionKey := make([]byte, 32)
	_, err := rand.Read(encryptionKey)
	require.NoError(t, err)

	opts := utils.NewOptions(options.Node().Store().SetBadgerEncryptionKey(encryptionKey).Node())
	assert.Equal(t, encryptionKey, opts.Store.BadgerEncryptionKey)
}

func TestNewStoreBadgerGCWrapping(t *testing.T) {
	ctx := context.Background()

	inMem, _, err := NewStore(ctx, options.NodeStore().
		SetType(options.NodeBadgerStore).SetBadgerInMemory(true).SetBadgerFileSize(1<<20))
	require.NoError(t, err)
	defer inMem.Close()
	_, wrapped := inMem.(*badgerStore)
	assert.False(t, wrapped, "in-memory store must not run value log GC")

	persistent, _, err := NewStore(ctx, options.NodeStore().
		SetType(options.NodeBadgerStore).SetPath(t.TempDir()).SetBadgerFileSize(1<<20))
	require.NoError(t, err)
	defer persistent.Close()
	_, wrapped = persistent.(*badgerStore)
	assert.True(t, wrapped, "persistent store must run value log GC")
}

func TestBadgerStoreCloseStopsGCAndIsIdempotent(t *testing.T) {
	store, err := newBadgerStore(t.TempDir(), badgerds.DefaultOptions(""))
	require.NoError(t, err)

	// Close waits for the background goroutines, so it returning is the proof they
	// stopped; the timeout turns a stuck goroutine into a failure instead of a hang.
	done := make(chan struct{})
	go func() {
		_ = store.Close()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("Close did not stop the background goroutines")
	}

	// A second Close must not panic on the already-closed stop channel.
	assert.NotPanics(t, func() { _ = store.Close() })
}

func TestBadgerStoreReclaimValueLogTerminates(t *testing.T) {
	store, err := newBadgerStore(t.TempDir(), badgerds.DefaultOptions(""))
	require.NoError(t, err)
	defer store.Close()

	// With no value log file eligible for GC, RunValueLogGC returns ErrNoRewrite on
	// the first call, so reclaimValueLog must return rather than spin on the error.
	done := make(chan struct{})
	go func() {
		store.reclaimValueLog(context.Background())
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("reclaimValueLog did not terminate")
	}
}

func TestDurationFromEnvFallsBackWhenUnsetOrInvalid(t *testing.T) {
	const fallback = 7 * time.Minute

	require.Equal(t, fallback, durationFromEnv(envOrphanBlockTTL, fallback))

	// A malformed or non-positive value falls back rather than failing the store open.
	for _, val := range []string{"", "not-a-duration", "0", "-5m"} {
		t.Setenv(envOrphanBlockTTL, val)
		require.Equal(t, fallback, durationFromEnv(envOrphanBlockTTL, fallback), val)
	}

	t.Setenv(envOrphanBlockTTL, "90s")
	require.Equal(t, 90*time.Second, durationFromEnv(envOrphanBlockTTL, fallback))
}

func TestBoolFromEnv(t *testing.T) {
	require.False(t, boolFromEnv(envOrphanGCDisabled))

	for _, val := range []string{"true", "1", "TRUE", "T"} {
		t.Setenv(envOrphanGCDisabled, val)
		require.True(t, boolFromEnv(envOrphanGCDisabled), val)
	}

	// The variable names a boolean, so "false" must read as false rather than as set.
	for _, val := range []string{"false", "0", "FALSE", "F"} {
		t.Setenv(envOrphanGCDisabled, val)
		require.False(t, boolFromEnv(envOrphanGCDisabled), val)
	}

	for _, val := range []string{"", "yes", "off", "maybe"} {
		t.Setenv(envOrphanGCDisabled, val)
		require.False(t, boolFromEnv(envOrphanGCDisabled), val)
	}
}

func TestNewBadgerStoreResolvesOrphanGCDisabled(t *testing.T) {
	for val, want := range map[string]bool{"true": true, "false": false} {
		t.Run(val, func(t *testing.T) {
			t.Setenv(envOrphanGCDisabled, val)

			dir := t.TempDir()
			store, err := newBadgerStore(dir, badgerds.DefaultOptions(dir))
			require.NoError(t, err)
			defer func() { require.NoError(t, store.Close()) }()

			require.Equal(t, want, store.orphanGCDisabled)
		})
	}
}

func TestNewBadgerStoreAppliesEnvOverrides(t *testing.T) {
	t.Setenv(envOrphanGCInterval, "3m")
	t.Setenv(envOrphanBlockTTL, "45m")

	dir := t.TempDir()
	store, err := newBadgerStore(dir, badgerds.DefaultOptions(dir))
	require.NoError(t, err)
	defer func() { require.NoError(t, store.Close()) }()

	require.Equal(t, 3*time.Minute, store.orphanGCInterval)
	require.Equal(t, 45*time.Minute, store.orphanBlockTTL)
}

func TestNewBadgerStoreDefaultsOrphanGCSettings(t *testing.T) {
	dir := t.TempDir()
	store, err := newBadgerStore(dir, badgerds.DefaultOptions(dir))
	require.NoError(t, err)
	defer func() { require.NoError(t, store.Close()) }()

	require.Equal(t, defaultOrphanBlockGCInterval, store.orphanGCInterval)
	require.Equal(t, defaultOrphanBlockTTL, store.orphanBlockTTL)

	// Pinned by value: the TTL decides when a block is deleted, so changing it should
	// take a deliberate test update rather than passing silently.
	require.Equal(t, 5*time.Minute, defaultOrphanBlockGCInterval)
	require.Equal(t, 30*time.Minute, defaultOrphanBlockTTL)
}

// Close waits on the maintenance goroutines, so a miscounted WaitGroup when the sweep is
// disabled would hang here rather than fail an assertion.
func TestNewBadgerStoreClosesCleanlyWithOrphanGCDisabled(t *testing.T) {
	t.Setenv(envOrphanGCDisabled, "1")

	dir := t.TempDir()
	store, err := newBadgerStore(dir, badgerds.DefaultOptions(dir))
	require.NoError(t, err)

	done := make(chan error, 1)
	go func() { done <- store.Close() }()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("Close hung with the orphan sweep disabled")
	}
}
