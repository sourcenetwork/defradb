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

func TestCacheExpiresStoredValue(t *testing.T) {
	expired := make(chan int, 1)
	cache, err := NewCache(context.Background(), 10*time.Millisecond, 10, func(_ string, value int) {
		expired <- value
	})
	require.NoError(t, err)
	defer cache.Stop()

	require.NoError(t, cache.Store("key", 1, 10*time.Millisecond))

	require.Eventually(t, func() bool {
		select {
		case value := <-expired:
			return value == 1
		default:
			return false
		}
	}, 200*time.Millisecond, 5*time.Millisecond)

	_, ok := cache.Load("key")
	require.False(t, ok)
}

func TestCacheDeletePreventsExpiration(t *testing.T) {
	expired := make(chan struct{}, 1)
	cache, err := NewCache(context.Background(), 10*time.Millisecond, 10, func(_ string, _ int) {
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
	cache, err := NewCache(context.Background(), 10*time.Millisecond, 10, func(_ string, _ int) {})
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

func TestCacheUpdateTTLReturnsFalseForMissingKey(t *testing.T) {
	cache, err := NewCache(context.Background(), 10*time.Millisecond, 20, func(_ string, _ int) {})
	require.NoError(t, err)
	defer cache.Stop()

	refreshed, err := cache.UpdateTTL("key", 80*time.Millisecond)

	require.NoError(t, err)
	require.False(t, refreshed)
}
