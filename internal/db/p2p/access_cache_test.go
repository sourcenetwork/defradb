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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// fakeClock is a deterministic clock for exercising accessCache expiry without sleeping.
type fakeClock struct {
	now time.Time
}

func (c *fakeClock) advance(d time.Duration) { c.now = c.now.Add(d) }
func (c *fakeClock) time() time.Time         { return c.now }

func newTestAccessCache(ttl time.Duration) (*accessCache, *fakeClock) {
	clock := &fakeClock{now: time.Unix(0, 0)}
	cache := newAccessCache(ttl)
	cache.now = clock.time
	return cache, clock
}

func TestAccessCache_MissBeforeStore(t *testing.T) {
	cache, _ := newTestAccessCache(time.Second)
	assert.False(t, cache.allowed("peer", "doc"), "unstored pair should miss")
}

func TestAccessCache_HitAfterStore(t *testing.T) {
	cache, _ := newTestAccessCache(time.Second)
	cache.storeAllowed("peer", "doc")
	assert.True(t, cache.allowed("peer", "doc"), "stored grant should hit within ttl")
}

func TestAccessCache_ExpiresAfterTTL(t *testing.T) {
	cache, clock := newTestAccessCache(time.Second)
	cache.storeAllowed("peer", "doc")

	clock.advance(999 * time.Millisecond)
	assert.True(t, cache.allowed("peer", "doc"), "grant should still be live just before ttl")

	clock.advance(2 * time.Millisecond)
	assert.False(t, cache.allowed("peer", "doc"), "grant should expire once ttl has elapsed")
}

func TestAccessCache_ScopedPerPeerAndDoc(t *testing.T) {
	cache, _ := newTestAccessCache(time.Second)
	cache.storeAllowed("peerA", "doc1")

	assert.True(t, cache.allowed("peerA", "doc1"))
	assert.False(t, cache.allowed("peerB", "doc1"), "a grant for one peer must not leak to another")
	assert.False(t, cache.allowed("peerA", "doc2"), "a grant for one doc must not leak to another")
}

func TestAccessCache_ZeroTTLDisablesCaching(t *testing.T) {
	cache, _ := newTestAccessCache(0)
	cache.storeAllowed("peer", "doc")
	assert.False(t, cache.allowed("peer", "doc"), "a zero ttl should keep the cache inert")
}

// TestAccessCache_CollapsesRepeatedLookups is the mechanism at the heart of the #4837 fix: many
// block-access checks for the same (peer, doc) during one DAG sync should consult the underlying
// access-control system once, not once per block. Here the "underlying check" is modelled as a
// counter that only runs on a cache miss, mirroring how hasAccess calls CheckDocReadAccess only
// when accessCache.allowed reports false.
func TestAccessCache_CollapsesRepeatedLookups(t *testing.T) {
	cache, _ := newTestAccessCache(time.Second)

	underlyingChecks := 0
	checkAccess := func(peerID, docID string) bool {
		if cache.allowed(peerID, docID) {
			return true
		}
		underlyingChecks++
		cache.storeAllowed(peerID, docID)
		return true
	}

	const blocksInDAG = 25
	for range blocksInDAG {
		assert.True(t, checkAccess("peer", "doc"))
	}

	assert.Equal(t, 1, underlyingChecks,
		"serving %d blocks of one document to one peer should cost a single access check", blocksInDAG)
}
