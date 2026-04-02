// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package coreblock

import (
	"fmt"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"

	ccid "github.com/sourcenetwork/defradb/internal/core/cid"
)

func newTestCID(t *testing.T, data []byte) cid.Cid {
	t.Helper()
	c, err := ccid.NewSHA256CidV1(data)
	require.NoError(t, err)
	return c
}

func TestBlockCache_PutGet_RoundTrip(t *testing.T) {
	cache, err := NewBlockCache(128)
	require.NoError(t, err)

	c := newTestCID(t, []byte("block-data-1"))
	data := []byte("hello world")

	cache.Put(c, data)

	got, ok := cache.Get(c)
	require.True(t, ok, "expected cache hit")
	require.Equal(t, data, got)
}

func TestBlockCache_Get_Miss(t *testing.T) {
	cache, err := NewBlockCache(128)
	require.NoError(t, err)

	c := newTestCID(t, []byte("nonexistent"))

	got, ok := cache.Get(c)
	require.False(t, ok, "expected cache miss")
	require.Nil(t, got)
}

func TestBlockCache_DisableEnable(t *testing.T) {
	// Save original state and restore after test
	originalDisabled := globalBlockCache.disabled.Load()
	t.Cleanup(func() {
		globalBlockCache.disabled.Store(originalDisabled)
	})

	global := GetGlobalBlockCache()

	c := newTestCID(t, []byte("disable-test"))
	data := []byte("some data")

	// Disable the global cache
	DisableGlobalBlockCache()

	global.Put(c, data)
	got, ok := global.Get(c)
	require.False(t, ok, "expected cache miss when disabled")
	require.Nil(t, got)

	// Re-enable the global cache
	EnableGlobalBlockCache()

	global.Put(c, data)
	got, ok = global.Get(c)
	require.True(t, ok, "expected cache hit after re-enable")
	require.Equal(t, data, got)
}

func TestBlockCache_Size(t *testing.T) {
	cache, err := NewBlockCache(128)
	require.NoError(t, err)

	n := 5
	for i := 0; i < n; i++ {
		c := newTestCID(t, []byte(fmt.Sprintf("size-test-%d", i)))
		cache.Put(c, []byte(fmt.Sprintf("data-%d", i)))
	}

	require.Equal(t, n, cache.Size())
}

func TestBlockCache_Clear(t *testing.T) {
	cache, err := NewBlockCache(128)
	require.NoError(t, err)

	cids := make([]cid.Cid, 3)
	for i := 0; i < 3; i++ {
		c := newTestCID(t, []byte(fmt.Sprintf("clear-test-%d", i)))
		cids[i] = c
		cache.Put(c, []byte(fmt.Sprintf("data-%d", i)))
	}

	require.Equal(t, 3, cache.Size())

	cache.Clear()

	require.Equal(t, 0, cache.Size())

	for _, c := range cids {
		got, ok := cache.Get(c)
		require.False(t, ok, "expected cache miss after Clear")
		require.Nil(t, got)
	}
}
