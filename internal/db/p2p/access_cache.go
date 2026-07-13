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

// accessCacheKey identifies a cached read-access decision.
//
// The collection is part of the key because a collection-level (branchable) block resolves to an
// empty docID, and access to such a block is decided per collection. Without the collection in the
// key, a grant on one branchable collection would be reused for another under the shared ("", peer)
// key and wrongly authorise it.
//
// The peer is keyed by its libp2p id rather than its resolved identity. This is safe only because a
// peer's verified identity is pinned per-id (see p.peerIdentities in hasAccess) and the cache is
// grant-only: a stale positive can never wrongly deny a peer, and a peer's id→identity binding does
// not change within a session. If per-id identities ever become refreshable, this key must fold in
// an identity fingerprint.
type accessCacheKey struct {
	peerID       string
	collectionID string
	docID        string
}

// accessCache memoizes positive read-access decisions for a short period so that serving every
// block of a document's DAG to a peer does not require a separate access-control round-trip per
// block.
//
// Only positive (allowed) decisions are stored. A denial is never cached: on the subscription
// path a denial usually means the access grant has not propagated yet, and caching it would keep
// a peer locked out until the entry expired even after the grant lands. Re-checking a denial is
// cheap relative to the sync failure that caching one would cause.
//
// Entries expire after ttl. Because only grants are cached, the ttl is short, and a cache hit does
// not extend expiry, the worst case after a revocation is that a peer retains read access for at
// most one ttl window — a briefly stale grant, never a stale denial. Expiry is lazy (stale entries
// are dropped when their key is read), with an opportunistic sweep on write to bound growth from
// keys that are never read again.
type accessCache struct {
	mu  sync.Mutex
	ttl time.Duration
	// now is injectable so tests can advance time deterministically. Defaults to time.Now.
	now     func() time.Time
	entries map[accessCacheKey]time.Time // key -> expiry time
}

// sweepThreshold is the entry count above which a store triggers an opportunistic sweep of expired
// entries. It only bounds growth; correctness does not depend on it.
const sweepThreshold = 1024

// newAccessCache returns a cache whose positive entries live for ttl. A ttl of zero disables
// caching (every lookup misses), which keeps the cache inert if it is ever misconfigured.
func newAccessCache(ttl time.Duration) *accessCache {
	return &accessCache{
		ttl:     ttl,
		now:     time.Now,
		entries: make(map[accessCacheKey]time.Time),
	}
}

// allowed reports whether a positive access decision for the (peer, collection, document) triple is
// cached and still fresh.
func (c *accessCache) allowed(peerID, collectionID, docID string) bool {
	if c.ttl <= 0 {
		return false
	}
	key := accessCacheKey{peerID: peerID, collectionID: collectionID, docID: docID}

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

// storeAllowed records that the peer has read access to the document in the collection, valid for
// ttl.
func (c *accessCache) storeAllowed(peerID, collectionID, docID string) {
	if c.ttl <= 0 {
		return
	}
	key := accessCacheKey{peerID: peerID, collectionID: collectionID, docID: docID}

	c.mu.Lock()
	defer c.mu.Unlock()
	// Reads keep their own key fresh (see allowed), but a key that is never read again would leak.
	// Store is the only place the map grows, so bound it here with a bulk sweep rather than scanning
	// the whole map on every (hot-path) read.
	if len(c.entries) >= sweepThreshold {
		c.sweepExpired()
	}
	c.entries[key] = c.now().Add(c.ttl)
}

// sweepExpired removes all expired entries. The caller must hold c.mu.
func (c *accessCache) sweepExpired() {
	now := c.now()
	for k, expiry := range c.entries {
		if !now.Before(expiry) {
			delete(c.entries, k)
		}
	}
}
