// Copyright 2025 Democratized Data Foundation
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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTTLCache_ValidParameters(t *testing.T) {
	ctx := context.Background()
	cache, err := NewTTLCache(ctx, 100*time.Millisecond, 10, func(k string, v int) {})

	require.NoError(t, err)
	assert.NotNil(t, cache)

	cache.tw.Stop()
}

func TestNewTTLCache_InvalidParameters_ReturnsError(t *testing.T) {
	ctx := context.Background()
	cache, err := NewTTLCache(ctx, 0, 10, func(k string, v int) {})

	require.Error(t, err)
	assert.Nil(t, cache)
}

func TestCache_StoreAndLoad(t *testing.T) {
	ctx := context.Background()
	cache, err := NewTTLCache(ctx, 100*time.Millisecond, 10, func(k string, v int) {})
	require.NoError(t, err)
	defer cache.tw.Stop()

	cache.Store("key1", 42, 500*time.Millisecond)

	val, ok := cache.Load("key1")
	assert.True(t, ok)
	assert.Equal(t, 42, val)
}

func TestCache_StoreAndDelete(t *testing.T) {
	ctx := context.Background()
	cache, err := NewTTLCache(ctx, 100*time.Millisecond, 10, func(k string, v int) {})
	require.NoError(t, err)
	defer cache.tw.Stop()

	cache.Store("key1", 42, 500*time.Millisecond)
	cache.Delete("key1")

	val, ok := cache.Load("key1")
	assert.False(t, ok)
	assert.Equal(t, 0, val)
}

func TestCache_LoadAndDelete(t *testing.T) {
	ctx := context.Background()
	cache, err := NewTTLCache(ctx, 100*time.Millisecond, 10, func(k string, v int) {})
	require.NoError(t, err)
	defer cache.tw.Stop()

	cache.Store("key1", 42, 500*time.Millisecond)

	val, ok := cache.LoadAndDelete("key1")
	assert.True(t, ok)
	assert.Equal(t, 42, val)

	// Should be gone now
	val, ok = cache.Load("key1")
	assert.False(t, ok)
	assert.Equal(t, 0, val)
}

func TestCache_Expiration(t *testing.T) {
	ctx := context.Background()
	expired := make(chan string, 1)

	cache, err := NewTTLCache(ctx, 50*time.Millisecond, 10, func(k string, v int) {
		expired <- k
	})
	require.NoError(t, err)
	defer cache.tw.Stop()

	cache.Store("key1", 42, 50*time.Millisecond)

	select {
	case key := <-expired:
		assert.Equal(t, "key1", key)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected key to expire within timeout")
	}

	// Key should be gone from cache
	_, ok := cache.Load("key1")
	assert.False(t, ok)
}

func TestCache_Load_NonExistentKey_ReturnsZeroValue(t *testing.T) {
	ctx := context.Background()
	cache, err := NewTTLCache(ctx, 100*time.Millisecond, 10, func(k string, v int) {})
	require.NoError(t, err)
	defer cache.tw.Stop()

	val, ok := cache.Load("nonexistent")
	assert.False(t, ok)
	assert.Equal(t, 0, val) // Zero value for int
}

func TestCache_Load_NonExistentKey_ReturnsZeroValueStruct(t *testing.T) {
	type TestStruct struct {
		Name  string
		Value int
	}

	ctx := context.Background()
	cache, err := NewTTLCache(ctx, 100*time.Millisecond, 10, func(k string, v TestStruct) {})
	require.NoError(t, err)
	defer cache.tw.Stop()

	val, ok := cache.Load("nonexistent")
	assert.False(t, ok)
	assert.Equal(t, TestStruct{}, val) // Zero value for struct
}

func TestCache_LoadAndDelete_NonExistentKey_ReturnsZeroValue(t *testing.T) {
	ctx := context.Background()
	cache, err := NewTTLCache(ctx, 100*time.Millisecond, 10, func(k string, v int) {})
	require.NoError(t, err)
	defer cache.tw.Stop()

	val, ok := cache.LoadAndDelete("nonexistent")
	assert.False(t, ok)
	assert.Equal(t, 0, val)
}

func TestCache_Delete_NonExistentKey_NoPanic(t *testing.T) {
	ctx := context.Background()
	cache, err := NewTTLCache(ctx, 100*time.Millisecond, 10, func(k string, v int) {})
	require.NoError(t, err)
	defer cache.tw.Stop()

	// Should not panic
	cache.Delete("nonexistent")
}

func TestCache_OnExpire_CalledWithCorrectKeyValue(t *testing.T) {
	ctx := context.Background()
	type expiredItem struct {
		key   string
		value int
	}
	expired := make(chan expiredItem, 1)

	cache, err := NewTTLCache(ctx, 50*time.Millisecond, 10, func(k string, v int) {
		expired <- expiredItem{k, v}
	})
	require.NoError(t, err)
	defer cache.tw.Stop()

	cache.Store("mykey", 123, 50*time.Millisecond)

	select {
	case item := <-expired:
		assert.Equal(t, "mykey", item.key)
		assert.Equal(t, 123, item.value)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected key to expire within timeout")
	}
}

func TestCache_Store_OverwriteExistingKey(t *testing.T) {
	ctx := context.Background()
	cache, err := NewTTLCache(ctx, 100*time.Millisecond, 10, func(k string, v int) {})
	require.NoError(t, err)
	defer cache.tw.Stop()

	cache.Store("key1", 42, 500*time.Millisecond)
	cache.Store("key1", 100, 500*time.Millisecond)

	val, ok := cache.Load("key1")
	assert.True(t, ok)
	assert.Equal(t, 100, val)
}

func TestCache_DeleteBeforeExpiration_PreventsOnExpire(t *testing.T) {
	ctx := context.Background()
	expireCalled := false
	var mu sync.Mutex

	cache, err := NewTTLCache(ctx, 50*time.Millisecond, 10, func(k string, v int) {
		mu.Lock()
		expireCalled = true
		mu.Unlock()
	})
	require.NoError(t, err)
	defer cache.tw.Stop()

	cache.Store("key1", 42, 100*time.Millisecond)
	cache.Delete("key1")

	// Wait for potential expiration
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	called := expireCalled
	mu.Unlock()

	assert.False(t, called)
}

func TestCache_LoadAndDelete_PreventsOnExpire(t *testing.T) {
	ctx := context.Background()
	expireCalled := false
	var mu sync.Mutex

	cache, err := NewTTLCache(ctx, 50*time.Millisecond, 10, func(k string, v int) {
		mu.Lock()
		expireCalled = true
		mu.Unlock()
	})
	require.NoError(t, err)
	defer cache.tw.Stop()

	cache.Store("key1", 42, 100*time.Millisecond)
	cache.LoadAndDelete("key1")

	// Wait for potential expiration
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	called := expireCalled
	mu.Unlock()

	assert.False(t, called)
}

func TestCache_MultipleEntries_IndependentExpiration(t *testing.T) {
	ctx := context.Background()
	var expired []string
	var mu sync.Mutex

	cache, err := NewTTLCache(ctx, 50*time.Millisecond, 20, func(k string, v int) {
		mu.Lock()
		expired = append(expired, k)
		mu.Unlock()
	})
	require.NoError(t, err)
	defer cache.tw.Stop()

	cache.Store("fast", 1, 50*time.Millisecond)
	cache.Store("slow", 2, 500*time.Millisecond)

	// Wait for fast to expire but not slow (give extra time for timing variance)
	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	fastExpired := false
	for _, k := range expired {
		if k == "fast" {
			fastExpired = true
		}
	}
	mu.Unlock()

	assert.True(t, fastExpired)

	// Slow should still be loadable
	val, ok := cache.Load("slow")
	assert.True(t, ok)
	assert.Equal(t, 2, val)
}

func TestCache_WithPointerValues(t *testing.T) {
	ctx := context.Background()
	type Data struct {
		Value int
	}

	cache, err := NewTTLCache(ctx, 100*time.Millisecond, 10, func(k string, v *Data) {})
	require.NoError(t, err)
	defer cache.tw.Stop()

	original := &Data{Value: 42}
	cache.Store("key1", original, 500*time.Millisecond)

	loaded, ok := cache.Load("key1")
	assert.True(t, ok)
	assert.Same(t, original, loaded)
	assert.Equal(t, 42, loaded.Value)
}

func TestCache_WithIntegerKeys(t *testing.T) {
	ctx := context.Background()
	cache, err := NewTTLCache(ctx, 100*time.Millisecond, 10, func(k int, v string) {})
	require.NoError(t, err)
	defer cache.tw.Stop()

	cache.Store(42, "answer", 500*time.Millisecond)

	val, ok := cache.Load(42)
	assert.True(t, ok)
	assert.Equal(t, "answer", val)
}
