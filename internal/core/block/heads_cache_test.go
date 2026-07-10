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
	"bytes"
	"context"
	"sort"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/memory"

	ccid "github.com/sourcenetwork/defradb/internal/core/cid"
	"github.com/sourcenetwork/defradb/internal/keys"
)

func newTestHeadsCID(t *testing.T, data []byte) cid.Cid {
	t.Helper()
	c, err := ccid.NewSHA256CidV1(data)
	require.NoError(t, err)
	return c
}

// sortedCIDs returns cids ordered by their bytes, the ordering the cache and
// List are expected to keep heads in.
func sortedCIDs(cids ...cid.Cid) []cid.Cid {
	out := make([]cid.Cid, len(cids))
	copy(out, cids)
	sort.Slice(out, func(i, j int) bool {
		return bytes.Compare(out[i].Bytes(), out[j].Bytes()) < 0
	})
	return out
}

func TestInitHeadsCache_SetsCache(t *testing.T) {
	ctx := context.Background()
	require.Nil(t, getHeadsCache(ctx))

	ctx = InitHeadsCache(ctx)
	require.NotNil(t, getHeadsCache(ctx))
}

func TestInitHeadsCache_Idempotent(t *testing.T) {
	ctx := context.Background()

	ctx = InitHeadsCache(ctx)
	cache1 := getHeadsCache(ctx)

	ctx = InitHeadsCache(ctx)
	cache2 := getHeadsCache(ctx)

	// A second init on the same context must reuse the cache, not replace it, so
	// heads accumulated earlier in the transaction survive.
	require.Same(t, cache1, cache2)
}

func TestNewDocCreateMode(t *testing.T) {
	ctx := context.Background()
	require.False(t, IsNewDocCreateMode(ctx))

	ctx = ContextWithNewDocCreateMode(ctx)
	require.True(t, IsNewDocCreateMode(ctx))
}

func TestHeadsCache_SetGet(t *testing.T) {
	ctx := InitHeadsCache(context.Background())
	cache := getHeadsCache(ctx)

	heads := []cid.Cid{newTestHeadsCID(t, []byte("head-1")), newTestHeadsCID(t, []byte("head-2"))}
	cache.set("ns", heads, 42)

	entry := cache.get("ns")
	require.NotNil(t, entry)
	require.Equal(t, heads, entry.heads)
	require.Equal(t, uint64(42), entry.maxHeight)

	require.Nil(t, cache.get("missing"))
}

func TestHeadsCache_UpdateOnWrite(t *testing.T) {
	ctx := InitHeadsCache(context.Background())
	cache := getHeadsCache(ctx)

	ns := "ns"
	c1 := newTestHeadsCID(t, []byte("write-1"))
	c2 := newTestHeadsCID(t, []byte("write-2"))

	cache.updateOnWrite(ns, c1, 10)
	entry := cache.get(ns)
	require.Len(t, entry.heads, 1)
	require.True(t, entry.heads[0].Equals(c1))
	require.Equal(t, uint64(10), entry.maxHeight)

	// A lower height must not lower maxHeight, and the heads stay ordered by bytes.
	cache.updateOnWrite(ns, c2, 5)
	entry = cache.get(ns)
	require.Equal(t, uint64(10), entry.maxHeight)
	require.Equal(t, sortedCIDs(c1, c2), entry.heads)
}

func TestHeadsCache_UpdateOnReplace(t *testing.T) {
	ctx := InitHeadsCache(context.Background())
	cache := getHeadsCache(ctx)

	// Byte-sorted CIDs. Start with the outer three and replace the lowest with the
	// middle one, so the new CID must be inserted between the two survivors rather
	// than appended. The replace height is below the entry's, so maxHeight must hold.
	cids := sortedCIDs(
		newTestHeadsCID(t, []byte("r1")),
		newTestHeadsCID(t, []byte("r2")),
		newTestHeadsCID(t, []byte("r3")),
		newTestHeadsCID(t, []byte("r4")),
	)
	ns := "ns"
	cache.set(ns, []cid.Cid{cids[0], cids[1], cids[3]}, 10)

	cache.updateOnReplace(ns, cids[0], cids[2], 8)

	entry := cache.get(ns)
	require.Equal(t, []cid.Cid{cids[1], cids[2], cids[3]}, entry.heads)
	require.Equal(t, uint64(10), entry.maxHeight)
}

// TestHeadsList_NewDocCreateMode_SkipsScan proves the fast path: in new-doc create
// mode List reports no heads without scanning, even when one exists in the store.
func TestHeadsList_NewDocCreateMode_SkipsScan(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	hs := NewHeadSet(store, keys.HeadstoreDocKey{}.WithDocShortID(1).WithFieldID("1"))

	require.NoError(t, hs.Write(ctx, newTestHeadsCID(t, []byte("existing")), 1))

	ctx = ContextWithNewDocCreateMode(ctx)
	heads, height, err := hs.List(ctx)
	require.NoError(t, err)
	require.Nil(t, heads)
	require.Equal(t, uint64(0), height)
}

// TestHeadsList_NewDocCreateMode_ReadsCollectionHead proves the fast path is
// limited to document heads: a collection head is still read in create mode, so a
// branchable collection's chain stays intact across documents.
func TestHeadsList_NewDocCreateMode_ReadsCollectionHead(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	hs := NewHeadSet(store, keys.NewHeadstoreColKey(1))

	c := newTestHeadsCID(t, []byte("collection-head"))
	require.NoError(t, hs.Write(ctx, c, 1))

	ctx = ContextWithNewDocCreateMode(ctx)
	heads, height, err := hs.List(ctx)
	require.NoError(t, err)
	require.Len(t, heads, 1)
	require.True(t, heads[0].Equals(c))
	require.Equal(t, uint64(1), height)
}

// TestHeadsList_ServesFromCache proves a second List returns the cached scan
// result rather than re-reading the store.
func TestHeadsList_ServesFromCache(t *testing.T) {
	ctx := InitHeadsCache(context.Background())
	store := memory.NewDatastore(context.Background())
	hs := NewHeadSet(store, keys.HeadstoreDocKey{}.WithDocShortID(1).WithFieldID("1"))

	c := newTestHeadsCID(t, []byte("head"))
	// writeHead skips the cache, so only a scan surfaces the head on the first List.
	require.NoError(t, hs.writeHead(ctx, c, 3))

	heads, height, err := hs.List(ctx)
	require.NoError(t, err)
	require.Len(t, heads, 1)
	require.True(t, heads[0].Equals(c))
	require.Equal(t, uint64(3), height)

	// Remove the head from the store; the cached result must still be returned.
	require.NoError(t, store.Delete(ctx, hs.key(c).Bytes()))
	heads, height, err = hs.List(ctx)
	require.NoError(t, err)
	require.Len(t, heads, 1)
	require.True(t, heads[0].Equals(c))
	require.Equal(t, uint64(3), height)
}

// TestHeadsList_CreateFlow mirrors what a document create actually does with the
// cache and create mode: the first List fast-paths to no heads and caches that,
// a Write records the new head, and a second List in the same context returns it
// from the cache. This is the path the write-time optimization relies on.
func TestHeadsList_CreateFlow(t *testing.T) {
	ctx := InitHeadsCache(ContextWithNewDocCreateMode(context.Background()))
	store := memory.NewDatastore(context.Background())
	hs := NewHeadSet(store, keys.HeadstoreDocKey{}.WithDocShortID(1).WithFieldID("1"))

	heads, _, err := hs.List(ctx)
	require.NoError(t, err)
	require.Nil(t, heads)

	c := newTestHeadsCID(t, []byte("head"))
	require.NoError(t, hs.Write(ctx, c, 1))

	heads, height, err := hs.List(ctx)
	require.NoError(t, err)
	require.Equal(t, []cid.Cid{c}, heads)
	require.Equal(t, uint64(1), height)
}

// TestHeadsReplace_KeepsCacheConsistent verifies Replace updates the cache and
// store together: a List after a Replace returns the new head, not the old one.
func TestHeadsReplace_KeepsCacheConsistent(t *testing.T) {
	ctx := InitHeadsCache(context.Background())
	store := memory.NewDatastore(context.Background())
	hs := NewHeadSet(store, keys.HeadstoreDocKey{}.WithDocShortID(1).WithFieldID("1"))

	old := newTestHeadsCID(t, []byte("old"))
	require.NoError(t, hs.Write(ctx, old, 1))

	updated := newTestHeadsCID(t, []byte("new"))
	require.NoError(t, hs.Replace(ctx, old, updated, 2))

	heads, height, err := hs.List(ctx)
	require.NoError(t, err)
	require.Equal(t, []cid.Cid{updated}, heads)
	require.Equal(t, uint64(2), height)
}
