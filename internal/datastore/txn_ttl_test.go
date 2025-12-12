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

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/internal/ttl"
)

func newTestTTLCache(ctx context.Context, t *testing.T, onExpire func(uint64, Txn)) *ttl.Cache[uint64, Txn] {
	ctx, cancel := context.WithCancel(ctx)
	cache, err := ttl.NewTTLCache(ctx, 10*time.Millisecond, 500, onExpire)
	require.NoError(t, err)
	t.Cleanup(func() {
		cancel()
	})
	return cache
}

func TestNewTTLTxnFrom(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	cache := newTestTTLCache(ctx, t, func(id uint64, txn Txn) {
		txn.Discard()
	})

	txn, err := NewTTLTxnFrom(rootstore, cache, 1, false, 500*time.Millisecond, immutable.None[int]())
	require.NoError(t, err)

	// Store in cache to enable TTL tracking
	cache.Store(txn.ID(), txn, 500*time.Millisecond)

	err = txn.Commit()
	require.NoError(t, err)
}

func TestTTLTxnFrom_OnSuccess(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	cache := newTestTTLCache(ctx, t, func(id uint64, txn Txn) {
		txn.Discard()
	})

	txn, err := NewTTLTxnFrom(rootstore, cache, 1, false, 500*time.Millisecond, immutable.None[int]())
	require.NoError(t, err)

	cache.Store(txn.ID(), txn, 500*time.Millisecond)

	runOnSuccessTest(t, txn)
}

func TestTTLTxnFrom_OnSuccessAsync(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	cache := newTestTTLCache(ctx, t, func(id uint64, txn Txn) {
		txn.Discard()
	})

	txn, err := NewTTLTxnFrom(rootstore, cache, 1, false, 500*time.Millisecond, immutable.None[int]())
	require.NoError(t, err)

	cache.Store(txn.ID(), txn, 500*time.Millisecond)

	runOnSuccessAsyncTest(t, txn)
}

func TestTTLTxnFrom_OnError(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	cache := newTestTTLCache(ctx, t, func(id uint64, txn Txn) {
		txn.Discard()
	})

	txn, err := NewTTLTxnFrom(rootstore, cache, 1, false, 500*time.Millisecond, immutable.None[int]())
	require.NoError(t, err)

	cache.Store(txn.ID(), txn, 500*time.Millisecond)

	err = rootstore.Close()
	require.NoError(t, err)

	runOnErrorTest(t, txn)
}

func TestTTLTxnFrom_OnErrorAsync(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	cache := newTestTTLCache(ctx, t, func(id uint64, txn Txn) {
		txn.Discard()
	})

	txn, err := NewTTLTxnFrom(rootstore, cache, 1, false, 500*time.Millisecond, immutable.None[int]())
	require.NoError(t, err)

	cache.Store(txn.ID(), txn, 500*time.Millisecond)

	err = rootstore.Close()
	require.NoError(t, err)

	runOnErrorAsyncTest(t, txn)
}

func TestTTLTxnFrom_OnDiscard(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	cache := newTestTTLCache(ctx, t, func(id uint64, txn Txn) {
		txn.Discard()
	})

	txn, err := NewTTLTxnFrom(rootstore, cache, 1, false, 500*time.Millisecond, immutable.None[int]())
	require.NoError(t, err)

	cache.Store(txn.ID(), txn, 500*time.Millisecond)

	runOnDiscardTest(t, txn)
}

func TestTTLTxnFrom_OnDiscardAsync(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	cache := newTestTTLCache(ctx, t, func(id uint64, txn Txn) {
		txn.Discard()
	})

	txn, err := NewTTLTxnFrom(rootstore, cache, 1, false, 500*time.Millisecond, immutable.None[int]())
	require.NoError(t, err)

	cache.Store(txn.ID(), txn, 500*time.Millisecond)

	runOnDiscardAsyncTest(t, txn)
}

func TestTTLTxnFrom_Commit_RemovesFromCache(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	expireCalled := false
	var mu sync.Mutex
	cache := newTestTTLCache(ctx, t, func(id uint64, txn Txn) {
		txn.Discard()
		mu.Lock()
		expireCalled = true
		mu.Unlock()
	})

	txn, err := NewTTLTxnFrom(rootstore, cache, 1, false, 100*time.Millisecond, immutable.None[int]())
	require.NoError(t, err)

	cache.Store(txn.ID(), txn, 100*time.Millisecond)

	err = txn.Commit()
	require.NoError(t, err)

	// Verify txn was removed from cache
	_, ok := cache.Load(txn.ID())
	require.False(t, ok, "txn should be removed from cache after commit")

	// Wait for potential expiration
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	called := expireCalled
	mu.Unlock()

	require.False(t, called, "expire should not be called after commit")
}

func TestTTLTxnFrom_Discard_RemovesFromCache(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	expireCalled := false
	var mu sync.Mutex
	cache := newTestTTLCache(ctx, t, func(id uint64, txn Txn) {
		txn.Discard()
		mu.Lock()
		expireCalled = true
		mu.Unlock()
	})

	txn, err := NewTTLTxnFrom(rootstore, cache, 1, false, 100*time.Millisecond, immutable.None[int]())
	require.NoError(t, err)

	cache.Store(txn.ID(), txn, 100*time.Millisecond)

	txn.Discard()

	// Verify txn was removed from cache
	_, ok := cache.Load(txn.ID())
	require.False(t, ok, "txn should be removed from cache after discard")

	// Wait for potential expiration
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	called := expireCalled
	mu.Unlock()

	require.False(t, called, "expire should not be called after discard")
}

func TestTTLTxnFrom_Expiration_CallsOnExpire(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	expired := make(chan uint64, 1)
	cache := newTestTTLCache(ctx, t, func(id uint64, txn Txn) {
		txn.Discard()
		expired <- id
	})

	txn, err := NewTTLTxnFrom(rootstore, cache, 42, false, 50*time.Millisecond, immutable.None[int]())
	require.NoError(t, err)

	cache.Store(txn.ID(), txn, 50*time.Millisecond)

	select {
	case id := <-expired:
		require.Equal(t, uint64(42), id)
	case <-time.After(70 * time.Millisecond):
		t.Fatal("expected txn to expire within timeout")
	}
}

func TestTTLTxnFrom_InvalidTTL_ReturnsError(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	cache := newTestTTLCache(ctx, t, func(id uint64, txn Txn) {
		txn.Discard()
	})

	txn, err := NewTTLTxnFrom(rootstore, cache, 1, false, 500*time.Millisecond, immutable.None[int]())
	require.NoError(t, err)

	// Storing with invalid TTL should error (negative TTL)
	err = cache.UpdateTTL(txn.ID(), -100*time.Millisecond)
	require.Error(t, err)

	// Clean up
	txn.Discard()
}

func TestTTLTxnFrom_MultipleTxns_IndependentExpiration(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	var expired []uint64
	var mu sync.Mutex
	cache := newTestTTLCache(ctx, t, func(id uint64, txn Txn) {
		txn.Discard()
		mu.Lock()
		expired = append(expired, id)
		mu.Unlock()
	})

	// Create two transactions with different TTLs
	txn1, err := NewTTLTxnFrom(rootstore, cache, 1, false, 50*time.Millisecond, immutable.None[int]())
	require.NoError(t, err)
	cache.Store(txn1.ID(), txn1, 50*time.Millisecond)

	txn2, err := NewTTLTxnFrom(rootstore, cache, 2, false, 500*time.Millisecond, immutable.None[int]())
	require.NoError(t, err)
	cache.Store(txn2.ID(), txn2, 500*time.Millisecond)

	// Wait for txn1 to expire but not txn2
	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	txn1Expired := false
	for _, id := range expired {
		if id == 1 {
			txn1Expired = true
		}
	}
	mu.Unlock()

	require.True(t, txn1Expired, "txn1 should have expired")

	// txn2 should still be in cache
	_, ok := cache.Load(txn2.ID())
	require.True(t, ok, "txn2 should still be in cache")

	// Commit txn2 to clean up
	err = txn2.Commit()
	require.NoError(t, err)
}

// Transaction semantics tests: These tests verify actual transaction behavior
// including data persistence and expiration semantics.

func TestTTLTxnCache_Expiration_DiscardsTransaction(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	expired := make(chan uint64, 1)
	cache := newTestTTLCache(ctx, t, func(id uint64, txn Txn) {
		txn.Discard()
		expired <- id
	})

	txn, err := NewTTLTxnFrom(rootstore, cache, 1, false, 50*time.Millisecond, immutable.None[int]())
	require.NoError(t, err)

	cache.Store(txn.ID(), txn, 50*time.Millisecond)

	// Write some data to the transaction
	err = txn.Datastore().Set(ctx, []byte("key1"), []byte("value1"))
	require.NoError(t, err)

	// Verify data is visible within the transaction
	val, err := txn.Datastore().Get(ctx, []byte("key1"))
	require.NoError(t, err)
	require.Equal(t, []byte("value1"), val)

	// Wait for expiration
	select {
	case <-expired:
		// Transaction was discarded
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected txn to expire within timeout")
	}

	// Transaction should no longer be in cache
	_, ok := cache.Load(txn.ID())
	require.False(t, ok, "transaction should be removed from cache after expiration")
}

func TestTTLTxnCache_Commit_PersistsData(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	expireCalled := false
	var mu sync.Mutex
	cache := newTestTTLCache(ctx, t, func(id uint64, txn Txn) {
		txn.Discard()
		mu.Lock()
		expireCalled = true
		mu.Unlock()
	})

	txn, err := NewTTLTxnFrom(rootstore, cache, 1, false, 500*time.Millisecond, immutable.None[int]())
	require.NoError(t, err)

	cache.Store(txn.ID(), txn, 500*time.Millisecond)

	// Write some data
	err = txn.Datastore().Set(ctx, []byte("key1"), []byte("value1"))
	require.NoError(t, err)

	// Commit the transaction
	err = txn.Commit()
	require.NoError(t, err)

	// Verify transaction is removed from cache
	_, ok := cache.Load(txn.ID())
	require.False(t, ok, "transaction should be removed from cache after commit")

	// Verify data was persisted - create a new read txn to check
	readTxn := NewTxnFrom(rootstore, 2, true, immutable.None[int]())
	defer readTxn.Discard()

	val, err := readTxn.Datastore().Get(ctx, []byte("key1"))
	require.NoError(t, err)
	require.Equal(t, []byte("value1"), val)

	// Wait to ensure onExpire was not called
	time.Sleep(200 * time.Millisecond)
	mu.Lock()
	called := expireCalled
	mu.Unlock()
	require.False(t, called, "onExpire should not be called after commit")
}

func TestTTLTxnCache_Discard_DoesNotPersistData(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	cache := newTestTTLCache(ctx, t, func(id uint64, txn Txn) {
		txn.Discard()
	})

	txn, err := NewTTLTxnFrom(rootstore, cache, 1, false, 500*time.Millisecond, immutable.None[int]())
	require.NoError(t, err)

	cache.Store(txn.ID(), txn, 500*time.Millisecond)

	// Write some data
	err = txn.Datastore().Set(ctx, []byte("key1"), []byte("value1"))
	require.NoError(t, err)

	// Discard the transaction
	txn.Discard()

	// Verify transaction is removed from cache
	_, ok := cache.Load(txn.ID())
	require.False(t, ok, "transaction should be removed from cache after discard")

	// Verify data was NOT persisted - create a new read txn to check
	readTxn := NewTxnFrom(rootstore, 2, true, immutable.None[int]())
	defer readTxn.Discard()

	_, err = readTxn.Datastore().Get(ctx, []byte("key1"))
	require.Error(t, err, "data should not be persisted after discard")
}

func TestTTLTxnCache_Expiration_UncommittedDataNotPersisted(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	expired := make(chan struct{}, 1)
	cache := newTestTTLCache(ctx, t, func(id uint64, txn Txn) {
		txn.Discard()
		expired <- struct{}{}
	})

	txn, err := NewTTLTxnFrom(rootstore, cache, 1, false, 50*time.Millisecond, immutable.None[int]())
	require.NoError(t, err)

	cache.Store(txn.ID(), txn, 50*time.Millisecond)

	// Write data but don't commit
	err = txn.Datastore().Set(ctx, []byte("uncommitted_key"), []byte("uncommitted_value"))
	require.NoError(t, err)

	// Wait for expiration
	select {
	case <-expired:
		// Transaction was discarded
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected transaction to expire within timeout")
	}

	// Verify uncommitted data was NOT persisted
	readTxn := NewTxnFrom(rootstore, 2, true, immutable.None[int]())
	defer readTxn.Discard()

	_, err = readTxn.Datastore().Get(ctx, []byte("uncommitted_key"))
	require.Error(t, err, "uncommitted data should not be persisted after expiration")
}

func TestTTLTxnCache_MultipleTxns_OnlyExpiredDiscarded(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	var discardedIDs []uint64
	var mu sync.Mutex

	cache := newTestTTLCache(ctx, t, func(id uint64, txn Txn) {
		txn.Discard()
		mu.Lock()
		discardedIDs = append(discardedIDs, id)
		mu.Unlock()
	})

	// Create fast-expiring transaction
	txn1, err := NewTTLTxnFrom(rootstore, cache, 1, false, 50*time.Millisecond, immutable.None[int]())
	require.NoError(t, err)
	cache.Store(txn1.ID(), txn1, 50*time.Millisecond)

	// Write data to txn1
	err = txn1.Datastore().Set(ctx, []byte("txn1_key"), []byte("txn1_value"))
	require.NoError(t, err)

	// Create slow-expiring transaction
	txn2, err := NewTTLTxnFrom(rootstore, cache, 2, false, 500*time.Millisecond, immutable.None[int]())
	require.NoError(t, err)
	cache.Store(txn2.ID(), txn2, 500*time.Millisecond)

	// Write data to txn2
	err = txn2.Datastore().Set(ctx, []byte("txn2_key"), []byte("txn2_value"))
	require.NoError(t, err)

	// Wait for txn1 to expire
	time.Sleep(150 * time.Millisecond)

	// Verify only txn1 was discarded
	mu.Lock()
	txn1Discarded := false
	txn2Discarded := false
	for _, id := range discardedIDs {
		if id == 1 {
			txn1Discarded = true
		}
		if id == 2 {
			txn2Discarded = true
		}
	}
	mu.Unlock()

	require.True(t, txn1Discarded, "txn1 should have been discarded")
	require.False(t, txn2Discarded, "txn2 should NOT have been discarded yet")

	// txn2 should still be loadable and functional
	loadedTxn2, ok := cache.Load(txn2.ID())
	require.True(t, ok, "txn2 should still be in cache")

	// Commit txn2 to persist its data
	err = loadedTxn2.Commit()
	require.NoError(t, err)

	// Verify txn2's data was persisted
	readTxn := NewTxnFrom(rootstore, 3, true, immutable.None[int]())
	defer readTxn.Discard()

	val, err := readTxn.Datastore().Get(ctx, []byte("txn2_key"))
	require.NoError(t, err)
	require.Equal(t, []byte("txn2_value"), val)

	// Verify txn1's data was NOT persisted
	_, err = readTxn.Datastore().Get(ctx, []byte("txn1_key"))
	require.Error(t, err, "txn1's data should not be persisted")
}
