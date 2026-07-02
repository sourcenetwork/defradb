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
	"sync"

	"github.com/ipfs/go-cid"
)

const defaultMergedCIDCacheSize = 4096

// MergedCIDCache is a bounded FIFO set that tracks which CIDs have been confirmed
// as merged in the current process.  It eliminates redundant IsMerged store reads
// for CIDs already processed within the same session.
//
// Eviction is FIFO: when the cache is full the oldest entry is removed.
// Thread-safe.
type MergedCIDCache struct {
	mu      sync.RWMutex
	entries map[string]struct{}
	order   []string // insertion order for FIFO eviction
	cap     int
}

// NewMergedCIDCache creates a new cache with the given capacity.
// If capacity ≤ 0, defaultMergedCIDCacheSize (4096) is used.
func NewMergedCIDCache(capacity int) *MergedCIDCache {
	if capacity <= 0 {
		capacity = defaultMergedCIDCacheSize
	}
	return &MergedCIDCache{
		entries: make(map[string]struct{}, capacity),
		order:   make([]string, 0, capacity),
		cap:     capacity,
	}
}

// Add marks c as merged. No-ops if c is already cached.
func (mc *MergedCIDCache) Add(c cid.Cid) {
	key := string(c.Bytes())
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if _, ok := mc.entries[key]; ok {
		return
	}
	if len(mc.entries) >= mc.cap {
		oldest := mc.order[0]
		mc.order = mc.order[1:]
		delete(mc.entries, oldest)
	}
	mc.entries[key] = struct{}{}
	mc.order = append(mc.order, key)
}

// Has reports whether c is known to be merged.
func (mc *MergedCIDCache) Has(c cid.Cid) bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	_, ok := mc.entries[string(c.Bytes())]
	return ok
}

// BatchAdd marks all cids as merged.
func (mc *MergedCIDCache) BatchAdd(cids []cid.Cid) {
	for _, c := range cids {
		mc.Add(c)
	}
}
