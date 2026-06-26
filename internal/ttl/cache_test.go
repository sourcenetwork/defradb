// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package ttl

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCacheExpiresUnleasedValue(t *testing.T) {
	expired := make(chan int, 1)
	cache, err := NewCache(context.Background(), 10*time.Millisecond, 20, func(_ string, value int) {
		expired <- value
	})
	require.NoError(t, err)
	defer cache.Stop()

	require.NoError(t, cache.Store("key", 1, 30*time.Millisecond))

	require.Eventually(t, func() bool {
		select {
		case value := <-expired:
			return value == 1
		default:
			return false
		}
	}, time.Second, 5*time.Millisecond)

	_, ok := cache.Load("key")
	require.False(t, ok)
}

func TestCacheActiveLeasePreventsExpiration(t *testing.T) {
	expired := make(chan struct{}, 1)
	cache, err := NewCache(context.Background(), 10*time.Millisecond, 20, func(_ string, _ int) {
		expired <- struct{}{}
	})
	require.NoError(t, err)
	defer cache.Stop()

	require.NoError(t, cache.Store("key", 1, 30*time.Millisecond))

	lease, ok := cache.Acquire("key")
	require.True(t, ok)
	require.Equal(t, 1, lease.Value())

	// The held lease keeps the value alive well beyond its TTL.
	require.Never(t, func() bool {
		select {
		case <-expired:
			return true
		default:
			return false
		}
	}, 90*time.Millisecond, 5*time.Millisecond)

	lease.Release()

	// Releasing the final lease restarts the idle timer, so it expires.
	require.Eventually(t, func() bool {
		select {
		case <-expired:
			return true
		default:
			return false
		}
	}, time.Second, 5*time.Millisecond)
}

func TestCacheConcurrentLeasesExpireAfterFinalRelease(t *testing.T) {
	expired := make(chan struct{}, 1)
	cache, err := NewCache(context.Background(), 10*time.Millisecond, 20, func(_ string, _ int) {
		expired <- struct{}{}
	})
	require.NoError(t, err)
	defer cache.Stop()

	require.NoError(t, cache.Store("key", 1, 30*time.Millisecond))

	first, ok := cache.Acquire("key")
	require.True(t, ok)
	second, ok := cache.Acquire("key")
	require.True(t, ok)

	first.Release()

	// One lease remains, so the value must not expire.
	require.Never(t, func() bool {
		select {
		case <-expired:
			return true
		default:
			return false
		}
	}, 90*time.Millisecond, 5*time.Millisecond)

	second.Release()

	require.Eventually(t, func() bool {
		select {
		case <-expired:
			return true
		default:
			return false
		}
	}, time.Second, 5*time.Millisecond)
}

func TestCacheReleaseIsIdempotent(t *testing.T) {
	expired := make(chan struct{}, 1)
	cache, err := NewCache(context.Background(), 10*time.Millisecond, 20, func(_ string, _ int) {
		expired <- struct{}{}
	})
	require.NoError(t, err)
	defer cache.Stop()

	require.NoError(t, cache.Store("key", 1, 30*time.Millisecond))

	first, ok := cache.Acquire("key")
	require.True(t, ok)
	second, ok := cache.Acquire("key")
	require.True(t, ok)

	// Releasing the same lease repeatedly must only drop one hold, leaving the
	// value alive while the second lease is still held.
	first.Release()
	first.Release()

	require.Never(t, func() bool {
		select {
		case <-expired:
			return true
		default:
			return false
		}
	}, 90*time.Millisecond, 5*time.Millisecond)

	second.Release()

	require.Eventually(t, func() bool {
		select {
		case <-expired:
			return true
		default:
			return false
		}
	}, time.Second, 5*time.Millisecond)
}

func TestCacheLoadAndDeletePreventsExpiration(t *testing.T) {
	expired := make(chan struct{}, 1)
	cache, err := NewCache(context.Background(), 10*time.Millisecond, 20, func(_ string, _ int) {
		expired <- struct{}{}
	})
	require.NoError(t, err)
	defer cache.Stop()

	require.NoError(t, cache.Store("key", 1, 30*time.Millisecond))

	lease, ok := cache.Acquire("key")
	require.True(t, ok)

	value, ok := cache.LoadAndDelete("key")
	require.True(t, ok)
	require.Equal(t, 1, value)

	// Releasing a lease on a removed value must not reschedule or expire it.
	lease.Release()

	require.Never(t, func() bool {
		select {
		case <-expired:
			return true
		default:
			return false
		}
	}, 90*time.Millisecond, 5*time.Millisecond)

	_, ok = cache.Load("key")
	require.False(t, ok)
}

func TestCacheDeletePreventsExpiration(t *testing.T) {
	expired := make(chan struct{}, 1)
	cache, err := NewCache(context.Background(), 10*time.Millisecond, 20, func(_ string, _ int) {
		expired <- struct{}{}
	})
	require.NoError(t, err)
	defer cache.Stop()

	require.NoError(t, cache.Store("key", 1, 10*time.Millisecond))
	cache.Delete("key")

	require.Never(t, func() bool {
		select {
		case <-expired:
			return true
		default:
			return false
		}
	}, 50*time.Millisecond, 5*time.Millisecond)
}

func TestCacheInvalidTTLDoesNotStoreValue(t *testing.T) {
	cache, err := NewCache(context.Background(), 10*time.Millisecond, 20, func(_ string, _ int) {})
	require.NoError(t, err)
	defer cache.Stop()

	require.ErrorIs(t, cache.Store("key", 1, -time.Millisecond), ErrNegativeTTL)

	_, ok := cache.Load("key")
	require.False(t, ok)
}

func TestCacheOverwriteDoesNotExpireNewValue(t *testing.T) {
	expired := make(chan int, 2)
	cache, err := NewCache(context.Background(), 10*time.Millisecond, 20, func(_ string, value int) {
		expired <- value
	})
	require.NoError(t, err)
	defer cache.Stop()

	require.NoError(t, cache.Store("key", 1, 20*time.Millisecond))
	time.Sleep(5 * time.Millisecond)
	require.NoError(t, cache.Store("key", 2, 100*time.Millisecond))

	require.Never(t, func() bool {
		select {
		case value := <-expired:
			return value == 1
		default:
			return false
		}
	}, 50*time.Millisecond, 5*time.Millisecond)

	value, ok := cache.Load("key")
	require.True(t, ok)
	require.Equal(t, 2, value)
}

func TestCacheLoadDoesNotLeaseValue(t *testing.T) {
	expired := make(chan struct{}, 1)
	cache, err := NewCache(context.Background(), 10*time.Millisecond, 20, func(_ string, _ int) {
		expired <- struct{}{}
	})
	require.NoError(t, err)
	defer cache.Stop()

	require.NoError(t, cache.Store("key", 1, 30*time.Millisecond))

	// A plain Load is a peek: it neither leases the value nor extends its TTL.
	value, ok := cache.Load("key")
	require.True(t, ok)
	require.Equal(t, 1, value)

	require.Eventually(t, func() bool {
		select {
		case <-expired:
			return true
		default:
			return false
		}
	}, time.Second, 5*time.Millisecond)
}

func TestCacheUpdateTTLRefreshesExpiration(t *testing.T) {
	expired := make(chan struct{}, 1)
	cache, err := NewCache(context.Background(), 10*time.Millisecond, 20, func(_ string, _ int) {
		expired <- struct{}{}
	})
	require.NoError(t, err)
	defer cache.Stop()

	require.NoError(t, cache.Store("key", 1, 30*time.Millisecond))
	time.Sleep(15 * time.Millisecond)

	refreshed, err := cache.UpdateTTL("key", 80*time.Millisecond)
	require.NoError(t, err)
	require.True(t, refreshed)

	require.Never(t, func() bool {
		select {
		case <-expired:
			return true
		default:
			return false
		}
	}, 40*time.Millisecond, 5*time.Millisecond)

	require.Eventually(t, func() bool {
		select {
		case <-expired:
			return true
		default:
			return false
		}
	}, 150*time.Millisecond, 5*time.Millisecond)
}

func TestCacheUpdateTTLWhileLeasedAppliesAfterRelease(t *testing.T) {
	expired := make(chan struct{}, 1)
	cache, err := NewCache(context.Background(), 10*time.Millisecond, 20, func(_ string, _ int) {
		expired <- struct{}{}
	})
	require.NoError(t, err)
	defer cache.Stop()

	require.NoError(t, cache.Store("key", 1, 30*time.Millisecond))
	lease, ok := cache.Acquire("key")
	require.True(t, ok)

	refreshed, err := cache.UpdateTTL("key", 80*time.Millisecond)
	require.NoError(t, err)
	require.True(t, refreshed)

	lease.Release()

	require.Never(t, func() bool {
		select {
		case <-expired:
			return true
		default:
			return false
		}
	}, 40*time.Millisecond, 5*time.Millisecond)

	require.Eventually(t, func() bool {
		select {
		case <-expired:
			return true
		default:
			return false
		}
	}, 150*time.Millisecond, 5*time.Millisecond)
}

func TestCacheUpdateTTLReturnsFalseForMissingKey(t *testing.T) {
	cache, err := NewCache(context.Background(), 10*time.Millisecond, 20, func(_ string, _ int) {})
	require.NoError(t, err)
	defer cache.Stop()

	refreshed, err := cache.UpdateTTL("key", 80*time.Millisecond)

	require.NoError(t, err)
	require.False(t, refreshed)
}
