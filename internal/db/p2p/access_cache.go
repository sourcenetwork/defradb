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
// collectionID is in the key because a collection-level (branchable) block has an empty docID and
// its access is decided per collection; without it, a grant on one branchable collection would
// authorize another under the shared ("", peer) key.
//
// Keying on peerID rather than the resolved identity is safe only while a peer's verified identity
// stays pinned per-id (see p.peerIdentities in hasAccess): the cache is grant-only, so a stale key
// can never wrongly deny a peer. If per-id identities ever become mutable, fold an identity
// fingerprint into the key.
type accessCacheKey struct {
	peerID       string
	collectionID string
	docID        string
}

// accessCache memoizes positive read-access decisions so that serving a document's DAG to a peer
// costs one access-control round-trip instead of one per block.
//
// Only grants are cached. A denial usually means the grant has not propagated yet, so caching it
// would lock the peer out until it expired even after the grant lands; re-checking a denial is
// cheap next to the sync failure caching one would cause.
//
// A hit does not extend expiry, so with a short ttl a revocation takes effect within one ttl
// window: the worst case is a briefly stale grant, never a stale denial.
type accessCache struct {
	mu  sync.Mutex
	ttl time.Duration
	// now is a seam for tests to advance time without sleeping. Defaults to time.Now.
	now     func() time.Time
	entries map[accessCacheKey]time.Time
	// order lets eviction find expired entries without scanning the map: every entry shares one
	// ttl, so insertion order is expiry order and the expired ones sit at the front. A re-stored
	// key appends a new slot and leaves a stale one behind, told apart by its expiry no longer
	// matching the map.
	order []orderEntry
}

type orderEntry struct {
	key    accessCacheKey
	expiry time.Time
}

// evictBatch caps the work one store does so a burst that all expires at once can't hold the lock
// for a long scan. Each subsequent store clears another batch, so growth stays bounded.
const evictBatch = 64

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
	// A read only reclaims the key it looks up, so a key never read again would leak. Evict here,
	// on the map's only growth path, to keep that work off the hot read path.
	c.evictExpired()
	expiry := c.now().Add(c.ttl)
	c.entries[key] = expiry
	c.order = append(c.order, orderEntry{key: key, expiry: expiry})
}

// evictExpired drops expired entries from the front of the insertion order, up to evictBatch, and
// stops at the first live one. Caller must hold c.mu.
func (c *accessCache) evictExpired() {
	now := c.now()
	evicted := 0
	for len(c.order) > 0 && evicted < evictBatch {
		front := c.order[0]
		mapExpiry, ok := c.entries[front.key]
		switch {
		case ok && mapExpiry.Equal(front.expiry) && now.Before(mapExpiry):
			// Oldest live slot is fresh, so nothing behind it can be expired.
			return
		case ok && mapExpiry.Equal(front.expiry):
			delete(c.entries, front.key)
			evicted++
		}
		// A slot the map no longer matches was superseded by a re-store; drop just the slot.
		c.order = c.order[1:]
	}
}
