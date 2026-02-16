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
	"sync/atomic"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/ipfs/go-cid"
)

const defaultCacheSize = 1_000_000

// BlockCache wraps an LRU cache for recently written blocks.
type BlockCache struct {
	cache    *lru.Cache[string, []byte]
	disabled atomic.Bool
}

// Global block cache instance
var globalBlockCache = mustNewBlockCache(defaultCacheSize)

func mustNewBlockCache(size int) *BlockCache {
	cache, err := lru.New[string, []byte](size)
	if err != nil {
		panic(err)
	}
	return &BlockCache{cache: cache}
}

// NewBlockCache creates a new block cache with the given max size.
func NewBlockCache(maxSize int) (*BlockCache, error) {
	cache, err := lru.New[string, []byte](maxSize)
	if err != nil {
		return nil, err
	}
	return &BlockCache{cache: cache}, nil
}

// GetGlobalBlockCache returns the global block cache instance.
func GetGlobalBlockCache() *BlockCache {
	return globalBlockCache
}

// SetGlobalBlockCacheSize creates a new global cache with the given size.
func SetGlobalBlockCacheSize(size int) {
	cache, err := lru.New[string, []byte](size)
	if err != nil {
		return
	}
	globalBlockCache.cache = cache
}

// DisableGlobalBlockCache disables the global block cache.
func DisableGlobalBlockCache() {
	globalBlockCache.disabled.Store(true)
	globalBlockCache.cache.Purge()
}

// EnableGlobalBlockCache enables the global block cache.
func EnableGlobalBlockCache() {
	globalBlockCache.disabled.Store(false)
}

// Put adds a block to the cache.
func (bc *BlockCache) Put(c cid.Cid, data []byte) {
	if bc.disabled.Load() {
		return
	}
	bc.cache.Add(c.String(), data)
}

// Get retrieves a block from the cache.
func (bc *BlockCache) Get(c cid.Cid) ([]byte, bool) {
	if bc.disabled.Load() {
		return nil, false
	}
	return bc.cache.Get(c.String())
}

// Size returns the current number of entries in the cache.
func (bc *BlockCache) Size() int {
	return bc.cache.Len()
}

// Clear removes all entries from the cache.
func (bc *BlockCache) Clear() {
	bc.cache.Purge()
}
