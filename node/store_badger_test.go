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
