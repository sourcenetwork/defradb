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
	"context"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"

	ccid "github.com/sourcenetwork/defradb/internal/core/cid"
)

func newTestHeadsCID(t *testing.T, data []byte) cid.Cid {
	t.Helper()
	c, err := ccid.NewSHA256CidV1(data)
	require.NoError(t, err)
	return c
}

func TestInitHeadsCache_SetsCache(t *testing.T) {
	ctx := context.Background()

	// Before init, cache should be nil
	require.Nil(t, getHeadsCache(ctx))

	ctx = InitHeadsCache(ctx)

	cache := getHeadsCache(ctx)
	require.NotNil(t, cache)
}

func TestInitHeadsCache_Idempotent(t *testing.T) {
	ctx := context.Background()

	ctx = InitHeadsCache(ctx)
	cache1 := getHeadsCache(ctx)

	ctx = InitHeadsCache(ctx)
	cache2 := getHeadsCache(ctx)

	// Should be the exact same pointer
	require.Same(t, cache1, cache2)
}

func TestHeadsCache_SetGet(t *testing.T) {
	ctx := context.Background()
	ctx = InitHeadsCache(ctx)
	cache := getHeadsCache(ctx)
	require.NotNil(t, cache)

	c1 := newTestHeadsCID(t, []byte("head-1"))
	c2 := newTestHeadsCID(t, []byte("head-2"))
	heads := []cid.Cid{c1, c2}

	cache.set("ns1", heads, 42)

	entry := cache.get("ns1")
	require.NotNil(t, entry)
	require.Equal(t, heads, entry.heads)
	require.Equal(t, uint64(42), entry.maxHeight)
}

func TestHeadsCache_GetMiss(t *testing.T) {
	ctx := context.Background()
	ctx = InitHeadsCache(ctx)
	cache := getHeadsCache(ctx)
	require.NotNil(t, cache)

	entry := cache.get("nonexistent")
	require.Nil(t, entry)
}

func TestNewDocCreateMode(t *testing.T) {
	ctx := context.Background()

	// Not set by default
	require.False(t, IsNewDocCreateMode(ctx))

	ctx = ContextWithNewDocCreateMode(ctx)

	require.True(t, IsNewDocCreateMode(ctx))
}

func TestHeadsCache_UpdateOnWrite(t *testing.T) {
	ctx := context.Background()
	ctx = InitHeadsCache(ctx)
	cache := getHeadsCache(ctx)
	require.NotNil(t, cache)

	ns := "write-ns"
	c1 := newTestHeadsCID(t, []byte("write-1"))
	c2 := newTestHeadsCID(t, []byte("write-2"))

	// updateOnWrite on empty namespace should create entry
	cache.updateOnWrite(ns, c1, 10)

	entry := cache.get(ns)
	require.NotNil(t, entry)
	require.Len(t, entry.heads, 1)
	require.True(t, entry.heads[0].Equals(c1))
	require.Equal(t, uint64(10), entry.maxHeight)

	// updateOnWrite on existing namespace should add to heads
	cache.updateOnWrite(ns, c2, 5)

	entry = cache.get(ns)
	require.NotNil(t, entry)
	require.Len(t, entry.heads, 2)
	// maxHeight should remain 10 since 5 < 10
	require.Equal(t, uint64(10), entry.maxHeight)

	// Both CIDs should be present
	found := make(map[string]bool)
	for _, h := range entry.heads {
		found[h.String()] = true
	}
	require.True(t, found[c1.String()])
	require.True(t, found[c2.String()])
}

func TestHeadsCache_UpdateOnReplace(t *testing.T) {
	ctx := context.Background()
	ctx = InitHeadsCache(ctx)
	cache := getHeadsCache(ctx)
	require.NotNil(t, cache)

	ns := "replace-ns"
	c1 := newTestHeadsCID(t, []byte("replace-1"))
	c2 := newTestHeadsCID(t, []byte("replace-2"))
	c3 := newTestHeadsCID(t, []byte("replace-3"))

	// Set initial state with c1 and c2
	cache.set(ns, []cid.Cid{c1, c2}, 10)

	// Replace c1 with c3
	cache.updateOnReplace(ns, c1, c3, 15)

	entry := cache.get(ns)
	require.NotNil(t, entry)
	require.Len(t, entry.heads, 2)
	require.Equal(t, uint64(15), entry.maxHeight)

	// c1 should be gone, c2 and c3 should be present
	found := make(map[string]bool)
	for _, h := range entry.heads {
		found[h.String()] = true
	}
	require.False(t, found[c1.String()], "old CID should have been removed")
	require.True(t, found[c2.String()], "untouched CID should remain")
	require.True(t, found[c3.String()], "new CID should be present")
}
