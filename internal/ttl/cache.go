// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package ttl

import (
	"context"
	"sync"
	"time"
)

// Cache stores values with idle-time expiration.
type Cache[K comparable, V any] struct {
	mu       sync.Mutex
	items    map[K]cacheItem[V]
	nextSeq  uint64
	wheel    *Wheel[cacheWheelKey[K]]
	onExpire func(K, V)
}

type cacheItem[V any] struct {
	value V
	seq   uint64
}

type cacheWheelKey[K comparable] struct {
	key K
	seq uint64
}

// NewCache returns a cache that expires entries using a timing wheel.
func NewCache[K comparable, V any](
	ctx context.Context,
	tick time.Duration,
	buckets int,
	onExpire func(K, V),
) (*Cache[K, V], error) {
	c := &Cache[K, V]{
		items:    make(map[K]cacheItem[V]),
		onExpire: onExpire,
	}
	wheel, err := NewWheel(ctx, tick, int64(buckets), c.expire)
	if err != nil {
		return nil, err
	}
	c.wheel = wheel
	c.wheel.Start()
	return c, nil
}

// Store stores val at key until ttl elapses.
func (c *Cache[K, V]) Store(key K, val V, ttl time.Duration) error {
	if err := c.wheel.validTTL(ttl); err != nil {
		return err
	}

	c.mu.Lock()
	old, hadOld := c.items[key]
	c.nextSeq++
	item := cacheItem[V]{
		value: val,
		seq:   c.nextSeq,
	}
	c.items[key] = item
	c.mu.Unlock()

	if hadOld {
		c.wheel.Delete(cacheWheelKey[K]{key: key, seq: old.seq})
	}
	return c.wheel.Add(cacheWheelKey[K]{key: key, seq: item.seq}, ttl)
}

// Load returns a value if key is still cached.
func (c *Cache[K, V]) Load(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.items[key]
	if !ok {
		var zero V
		return zero, false
	}
	return item.value, true
}

// LoadAndDelete loads and removes key if it is still cached.
func (c *Cache[K, V]) LoadAndDelete(key K) (V, bool) {
	c.mu.Lock()
	item, ok := c.items[key]
	if ok {
		delete(c.items, key)
	}
	c.mu.Unlock()

	if !ok {
		var zero V
		return zero, false
	}
	c.wheel.Delete(cacheWheelKey[K]{key: key, seq: item.seq})
	return item.value, true
}

// Delete removes key from the cache.
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	item, ok := c.items[key]
	if ok {
		delete(c.items, key)
	}
	c.mu.Unlock()

	if ok {
		c.wheel.Delete(cacheWheelKey[K]{key: key, seq: item.seq})
	}
}

// UpdateTTL resets the expiration time for key if key is still cached.
func (c *Cache[K, V]) UpdateTTL(key K, ttl time.Duration) error {
	if err := c.wheel.validTTL(ttl); err != nil {
		return err
	}

	c.mu.Lock()
	item, ok := c.items[key]
	c.mu.Unlock()
	if !ok {
		return nil
	}
	return c.wheel.UpdateTTL(cacheWheelKey[K]{key: key, seq: item.seq}, ttl)
}

// ValidateTTL returns an error if ttl cannot be used by this cache.
func (c *Cache[K, V]) ValidateTTL(ttl time.Duration) error {
	return c.wheel.validTTL(ttl)
}

// Stop stops the cache expiration wheel.
func (c *Cache[K, V]) Stop() {
	c.wheel.Stop()
}

func (c *Cache[K, V]) expire(wheelKey cacheWheelKey[K]) {
	c.mu.Lock()
	item, ok := c.items[wheelKey.key]
	if !ok || item.seq != wheelKey.seq {
		c.mu.Unlock()
		return
	}
	delete(c.items, wheelKey.key)
	c.mu.Unlock()

	if c.onExpire != nil {
		c.onExpire(wheelKey.key, item.value)
	}
}
