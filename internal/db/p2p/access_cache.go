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
	"time"
)

// accessCacheKey identifies a cached read-access decision for a (peer, document) pair.
type accessCacheKey struct {
	peerID string
	docID  string
}

// accessCache memoizes positive read-access decisions for (peer, document) pairs for a short
// period so that serving every block of a document's DAG to a peer does not require a separate
// access-control round-trip per block.
//
// Only positive (allowed) decisions are stored. A denial is never cached: on the subscription
// path a denial usually means the access grant has not propagated yet, and caching it would keep
// a peer locked out until the entry expired even after the grant lands. Re-checking a denial is
// cheap relative to the sync failure that caching one would cause.
//
// Entries expire after ttl. Because only grants are cached and the ttl is short, the worst case
// after a revocation is that a peer retains read access for at most one ttl window — a briefly
// stale grant, never a stale denial. Expiry is lazy: stale entries are dropped on read.
type accessCache struct {
	mu  sync.Mutex
	ttl time.Duration
	// now is injectable so tests can advance time deterministically. Defaults to time.Now.
	now     func() time.Time
	entries map[accessCacheKey]time.Time // key -> expiry time
}

// newAccessCache returns a cache whose positive entries live for ttl. A ttl of zero disables
// caching (every lookup misses), which keeps the cache inert if it is ever misconfigured.
func newAccessCache(ttl time.Duration) *accessCache {
	return &accessCache{
		ttl:     ttl,
		now:     time.Now,
		entries: make(map[accessCacheKey]time.Time),
	}
}

// allowed reports whether a positive access decision for the pair is cached and still fresh.
func (c *accessCache) allowed(peerID, docID string) bool {
	if c.ttl <= 0 {
		return false
	}
	key := accessCacheKey{peerID: peerID, docID: docID}

	c.mu.Lock()
	defer c.mu.Unlock()

	expiry, ok := c.entries[key]
	if !ok {
		return false
	}
	if !c.now().Before(expiry) {
		// Expired; drop it lazily.
		delete(c.entries, key)
		return false
	}
	return true
}

// storeAllowed records that the peer has read access to the document, valid for ttl.
func (c *accessCache) storeAllowed(peerID, docID string) {
	if c.ttl <= 0 {
		return
	}
	key := accessCacheKey{peerID: peerID, docID: docID}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = c.now().Add(c.ttl)
}
