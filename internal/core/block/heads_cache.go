// Copyright 2025 Democratized Data Foundation
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
	"encoding/binary"
	"sort"
	"sync"

	cid "github.com/ipfs/go-cid"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/internal/keys"
)

type headsCacheKey struct{}

// headsCache caches the result of heads.List() calls.
type headsCache struct {
	mu    sync.RWMutex
	cache map[string]*headsCacheEntry
}

type headsCacheEntry struct {
	heads     []cid.Cid
	maxHeight uint64
}

// InitHeadsCache initializes the context with a heads cache.
func InitHeadsCache(ctx context.Context) context.Context {
	return context.WithValue(ctx, headsCacheKey{}, &headsCache{
		cache: make(map[string]*headsCacheEntry),
	})
}

// getHeadsCache retrieves the heads cache from the context.
func getHeadsCache(ctx context.Context) *headsCache {
	cache, _ := ctx.Value(headsCacheKey{}).(*headsCache)
	return cache
}

// get retrieves cached heads for the given namespace.
func (c *headsCache) get(namespace string) *headsCacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cache[namespace]
}

// set stores heads in the cache for the given namespace.
func (c *headsCache) set(namespace string, heads []cid.Cid, maxHeight uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[namespace] = &headsCacheEntry{
		heads:     heads,
		maxHeight: maxHeight,
	}
}

// invalidate removes the cached entry for the given namespace.
func (c *headsCache) invalidate(namespace string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, namespace)
}

// updateOnWrite updates the cache when a new head is written.
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
		hBytes := h.Bytes()
		if !inserted && bytes.Compare(newCidBytes, hBytes) < 0 {
			newHeads = append(newHeads, newCid)
			inserted = true
		}
		newHeads = append(newHeads, h)
	}
	if !inserted {
		newHeads = append(newHeads, newCid)
	}

	newHeight := height
	if entry.maxHeight > height {
		newHeight = entry.maxHeight
	}

	c.cache[namespace] = &headsCacheEntry{
		heads:     newHeads,
		maxHeight: newHeight,
	}
}

// updateOnReplace updates the cache when a head is replaced.
func (c *headsCache) updateOnReplace(namespace string, oldCid, newCid cid.Cid, height uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := c.cache[namespace]
	if entry == nil {
		return
	}

	newHeads := make([]cid.Cid, 0, len(entry.heads))
	newCidBytes := newCid.Bytes()
	inserted := false

	for _, h := range entry.heads {
		if h.Equals(oldCid) {
			continue
		}
		hBytes := h.Bytes()
		if !inserted && bytes.Compare(newCidBytes, hBytes) < 0 {
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
		maxHeight: height,
	}
}

// invalidateByPrefix removes all cached entries that start with the given prefix.
func (c *headsCache) invalidateByPrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key := range c.cache {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.cache, key)
		}
	}
}

// PrefetchDocHeads fetches all heads for a document and populates the cache.
func PrefetchDocHeads(ctx context.Context, store corekv.ReaderWriter, docID string) error {
	cache := getHeadsCache(ctx)
	if cache == nil {
		return nil
	}

	// Create prefix for all heads of this document: /d/[DocID]/
	docPrefix := keys.HeadstoreDocKey{DocID: docID}
	prefixBytes := docPrefix.Bytes()

	iter, err := store.Iterator(ctx, corekv.IterOptions{
		Prefix: prefixBytes,
	})
	if err != nil {
		return err
	}
	defer iter.Close()

	// Map from full namespace key to heads
	headsMap := make(map[string]*headsCacheEntry)

	for {
		hasNext, err := iter.Next()
		if err != nil {
			return err
		}
		if !hasNext {
			break
		}

		headKey, err := keys.NewHeadstoreDocKey(string(iter.Key()))
		if err != nil {
			return err
		}

		value, err := iter.Value()
		if err != nil {
			return err
		}

		height, n := binary.Uvarint(value)
		if n <= 0 {
			return ErrDecodingHeight
		}

		// Build the namespace key for this field
		fieldNamespace := keys.HeadstoreDocKey{
			DocID:   headKey.DocID,
			FieldID: headKey.FieldID,
		}
		namespaceKey := string(fieldNamespace.Bytes())

		entry, exists := headsMap[namespaceKey]
		if !exists {
			entry = &headsCacheEntry{
				heads:     make([]cid.Cid, 0, 1),
				maxHeight: 0,
			}
			headsMap[namespaceKey] = entry
		}

		entry.heads = append(entry.heads, headKey.Cid)
		if height > entry.maxHeight {
			entry.maxHeight = height
		}
	}

	// Sort heads in each entry and populate cache
	cache.mu.Lock()
	defer cache.mu.Unlock()

	for namespaceKey, entry := range headsMap {
		if len(entry.heads) > 1 {
			sort.Slice(entry.heads, func(i, j int) bool {
				ci := entry.heads[i].Bytes()
				cj := entry.heads[j].Bytes()
				return bytes.Compare(ci, cj) < 0
			})
		}
		cache.cache[namespaceKey] = entry
	}

	return nil
}
