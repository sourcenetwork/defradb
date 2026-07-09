// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package p2p

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// fakeClock is a deterministic clock for exercising accessCache expiry without sleeping.
type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *fakeClock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

func (c *fakeClock) time() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func newTestAccessCache(ttl time.Duration) (*accessCache, *fakeClock) {
	clock := &fakeClock{now: time.Unix(0, 0)}
	cache := newAccessCache(ttl)
	cache.now = clock.time
	return cache, clock
}

func TestAccessCache_MissBeforeStore(t *testing.T) {
	cache, _ := newTestAccessCache(time.Second)
	assert.False(t, cache.allowed("peer", "col", "doc"), "unstored triple should miss")
}

func TestAccessCache_HitAfterStore(t *testing.T) {
	cache, _ := newTestAccessCache(time.Second)
	cache.storeAllowed("peer", "col", "doc")
	assert.True(t, cache.allowed("peer", "col", "doc"), "stored grant should hit within ttl")
}

func TestAccessCache_ExpiresAfterTTL(t *testing.T) {
	cache, clock := newTestAccessCache(time.Second)
	cache.storeAllowed("peer", "col", "doc")

	clock.advance(999 * time.Millisecond)
	assert.True(t, cache.allowed("peer", "col", "doc"), "grant should still be live just before ttl")

	clock.advance(2 * time.Millisecond)
	assert.False(t, cache.allowed("peer", "col", "doc"), "grant should expire once ttl has elapsed")
}

func TestAccessCache_ScopedPerPeerCollectionAndDoc(t *testing.T) {
	cache, _ := newTestAccessCache(time.Second)
	cache.storeAllowed("peerA", "colX", "doc1")

	assert.True(t, cache.allowed("peerA", "colX", "doc1"))
	assert.False(t, cache.allowed("peerB", "colX", "doc1"), "a grant for one peer must not leak to another")
	assert.False(t, cache.allowed("peerA", "colX", "doc2"), "a grant for one doc must not leak to another")
	assert.False(t, cache.allowed("peerA", "colY", "doc1"), "a grant for one collection must not leak to another")
}

// TestAccessCache_CollectionLevelBlocksDoNotCollide guards the #4837 access-control boundary:
// collection-level (branchable) blocks resolve to an empty docID, so a grant on one branchable
// collection must not authorise another. Without the collection in the key both would share the
// ("peer", "") entry.
func TestAccessCache_CollectionLevelBlocksDoNotCollide(t *testing.T) {
	cache, _ := newTestAccessCache(time.Second)
	cache.storeAllowed("peer", "branchableA", "")

	assert.True(t, cache.allowed("peer", "branchableA", ""), "the granted collection should hit")
	assert.False(t, cache.allowed("peer", "branchableB", ""),
		"a grant on one branchable collection must not authorise another via the shared empty docID")
}

func TestAccessCache_ZeroTTLDisablesCaching(t *testing.T) {
	cache, _ := newTestAccessCache(0)
	cache.storeAllowed("peer", "col", "doc")
	assert.False(t, cache.allowed("peer", "col", "doc"), "a zero ttl should keep the cache inert")
}

// TestAccessCache_HitDoesNotExtendExpiry underpins the security argument that a revoked grant goes
// stale within one ttl window: repeated hits must not push the expiry out.
func TestAccessCache_HitDoesNotExtendExpiry(t *testing.T) {
	cache, clock := newTestAccessCache(time.Second)
	cache.storeAllowed("peer", "col", "doc")

	// Hit repeatedly while advancing toward the ttl; hits must not refresh the timer.
	for range 9 {
		clock.advance(100 * time.Millisecond)
		assert.True(t, cache.allowed("peer", "col", "doc"))
	}

	clock.advance(200 * time.Millisecond) // now past the original 1s expiry
	assert.False(t, cache.allowed("peer", "col", "doc"),
		"expiry must be measured from the store, not extended by hits")
}

// TestAccessCache_CollapsesRepeatedLookups is the mechanism at the heart of the #4837 fix: many
// block-access checks for the same (peer, collection, doc) during one DAG sync should consult the
// underlying access-control system once, not once per block. The "underlying check" is modelled as
// a counter that only runs on a cache miss, mirroring how hasAccess calls CheckDocReadAccess only
// when accessCache.allowed reports false.
func TestAccessCache_CollapsesRepeatedLookups(t *testing.T) {
	cache, _ := newTestAccessCache(time.Second)

	underlyingChecks := 0
	checkAccess := func(peerID, collectionID, docID string) bool {
		if cache.allowed(peerID, collectionID, docID) {
			return true
		}
		underlyingChecks++
		cache.storeAllowed(peerID, collectionID, docID)
		return true
	}

	const blocksInDAG = 25
	for range blocksInDAG {
		assert.True(t, checkAccess("peer", "col", "doc"))
	}

	assert.Equal(t, 1, underlyingChecks,
		"serving %d blocks of one document to one peer should cost a single access check", blocksInDAG)
}

// TestAccessCache_ConcurrentAccess exercises the cache under the -race detector to catch data races
// in the lazy-delete-on-read and store paths, since hasAccess is called concurrently by bitswap.
func TestAccessCache_ConcurrentAccess(t *testing.T) {
	cache := newAccessCache(time.Second) // real clock is fine; we only care about races here

	var wg sync.WaitGroup
	for i := range 32 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			peer := "peer"
			col := "col"
			doc := string(rune('a' + n%8))
			for range 100 {
				if !cache.allowed(peer, col, doc) {
					cache.storeAllowed(peer, col, doc)
				}
			}
		}(i)
	}
	wg.Wait()
}
