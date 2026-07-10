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
	"sync"

	cid "github.com/ipfs/go-cid"
)

type headsCacheKey struct{}
type newDocCreateModeKey struct{}

// ContextWithNewDocCreateMode marks the context as creating a new document. A new
// document has no existing heads, so the head lookup can skip the headstore scan.
func ContextWithNewDocCreateMode(ctx context.Context) context.Context {
	return context.WithValue(ctx, newDocCreateModeKey{}, true)
}

// IsNewDocCreateMode reports whether the context is creating a new document.
func IsNewDocCreateMode(ctx context.Context) bool {
	v, _ := ctx.Value(newDocCreateModeKey{}).(bool)
	return v
}

// headsCache holds the heads of each Merkle-CRDT namespace touched within a
// transaction. It lets repeated writes in one batch reuse the heads instead of
// re-scanning the headstore, whose iterator cost grows with the batch size.
type headsCache struct {
	mu    sync.RWMutex
	cache map[string]*headsCacheEntry
}

type headsCacheEntry struct {
	heads     []cid.Cid
	maxHeight uint64
}

// InitHeadsCache attaches a heads cache to the context if one is not already
// present. The cache must be scoped to a single transaction so heads never leak
// between transactions.
func InitHeadsCache(ctx context.Context) context.Context {
	if getHeadsCache(ctx) != nil {
		return ctx
	}
	return context.WithValue(ctx, headsCacheKey{}, &headsCache{
		cache: make(map[string]*headsCacheEntry),
	})
}

func getHeadsCache(ctx context.Context) *headsCache {
	cache, _ := ctx.Value(headsCacheKey{}).(*headsCache)
	return cache
}

func (c *headsCache) get(namespace string) *headsCacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cache[namespace]
}

func (c *headsCache) set(namespace string, heads []cid.Cid, maxHeight uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[namespace] = &headsCacheEntry{
		heads:     heads,
		maxHeight: maxHeight,
	}
}

// updateOnWrite records a newly written head, keeping the cached heads ordered by
// CID bytes so they match the ordering List returns from a scan.
func (c *headsCache) updateOnWrite(namespace string, newCid cid.Cid, height uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := c.cache[namespace]
	if entry == nil {
		c.cache[namespace] = &headsCacheEntry{
			heads:     []cid.Cid{newCid},
			maxHeight: height,
		}
		return
	}

	newHeads := make([]cid.Cid, 0, len(entry.heads)+1)
	inserted := false
	newCidBytes := newCid.Bytes()
	for _, h := range entry.heads {
		if !inserted && bytes.Compare(newCidBytes, h.Bytes()) < 0 {
			newHeads = append(newHeads, newCid)
			inserted = true
		}
		newHeads = append(newHeads, h)
	}
	if !inserted {
		newHeads = append(newHeads, newCid)
	}

	c.cache[namespace] = &headsCacheEntry{
		heads:     newHeads,
		maxHeight: max(entry.maxHeight, height),
	}
}

// updateOnReplace swaps oldCid for newCid, keeping the cached heads in CID-byte
// order to match List. maxHeight is never lowered: a stale-low height could
// collide with an existing block's priority, while a high one is harmless.
func (c *headsCache) updateOnReplace(namespace string, oldCid, newCid cid.Cid, height uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := c.cache[namespace]
	if entry == nil {
		return
	}

	newHeads := make([]cid.Cid, 0, len(entry.heads))
	inserted := false
	newCidBytes := newCid.Bytes()
	for _, h := range entry.heads {
		if h.Equals(oldCid) {
			continue
		}
		if !inserted && bytes.Compare(newCidBytes, h.Bytes()) < 0 {
			newHeads = append(newHeads, newCid)
			inserted = true
		}
		newHeads = append(newHeads, h)
	}
	if !inserted {
		newHeads = append(newHeads, newCid)
	}

	c.cache[namespace] = &headsCacheEntry{
		heads:     newHeads,
		maxHeight: max(entry.maxHeight, height),
	}
}
