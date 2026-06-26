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

// Cache stores values with idle-time expiration that can be suspended by
// leasing a value while it is in use.
//
// Entries expire once they have been idle for their TTL, but a caller may
// Acquire a lease to keep a value alive while it is in use: a value with one or
// more active leases is never expired, and once its final lease is released its
// idle timer restarts from the value's full TTL.
type Cache[K comparable, V any] struct {
	mu       sync.Mutex
	items    map[K]*cacheItem[V]
	nextSeq  uint64
	wheel    *Wheel[cacheWheelKey[K]]
	onExpire func(K, V)
}

type cacheItem[V any] struct {
	value  V
	ttl    time.Duration
	active int
	seq    uint64
}

// cacheWheelKey pairs a cache key with the sequence of the item it was
// scheduled for so a stale wheel expiration cannot evict a newer item.
type cacheWheelKey[K comparable] struct {
	key K
	seq uint64
}

// Lease is a hold on a cached value that keeps it from expiring until released.
// It is returned by Cache.Acquire.
type Lease[K comparable, V any] struct {
	cache *Cache[K, V]
	key   K
	value V
	once  sync.Once
}

// NewCache returns a cache that expires idle values using a timing wheel.
func NewCache[K comparable, V any](
	ctx context.Context,
	tick time.Duration,
	buckets int,
	onExpire func(K, V),
) (*Cache[K, V], error) {
	c := &Cache[K, V]{
		items:    make(map[K]*cacheItem[V]),
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

// Store stores val at key with an idle expiration of ttl, replacing any
// existing value and discarding its lease state.
func (c *Cache[K, V]) Store(key K, val V, ttl time.Duration) error {
	if err := c.wheel.ValidateTTL(ttl); err != nil {
		return err
	}

	c.mu.Lock()
	old, hadOld := c.items[key]
	c.nextSeq++
	item := &cacheItem[V]{
		value: val,
		ttl:   ttl,
		seq:   c.nextSeq,
	}
	c.items[key] = item
	c.mu.Unlock()

	if hadOld {
		c.wheel.Delete(cacheWheelKey[K]{key: key, seq: old.seq})
	}
	return c.wheel.Add(cacheWheelKey[K]{key: key, seq: item.seq}, ttl)
}

// Load returns the value at key, if still cached, without affecting its
// expiration or lease state.
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

// Acquire returns a lease on the value at key, suspending its expiration until
// the lease is released. It reports false if key is not cached.
func (c *Cache[K, V]) Acquire(key K) (*Lease[K, V], bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.items[key]
	if !ok {
		return nil, false
	}
	// The first lease removes the value from the wheel so it cannot expire
	// while in use. Bumping the sequence ensures any expiration already in
	// flight for the old schedule is ignored once the lease is released.
	if item.active == 0 {
		c.wheel.Delete(cacheWheelKey[K]{key: key, seq: item.seq})
		c.nextSeq++
		item.seq = c.nextSeq
	}
	item.active++
	return &Lease[K, V]{
		cache: c,
		key:   key,
		value: item.value,
	}, true
}

// LoadAndDelete removes key and returns its value, if still cached. A value
// removed this way is never expired by the cache, even while leases are held;
// the caller assumes ownership of it.
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

// Delete removes key from the cache without expiring its value.
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	item, ok := c.items[key]
	if ok {
		delete(c.items, key)
	}
	c.mu.Unlock()

	c.wheel.Delete(cacheWheelKey[K]{key: key, seq: item.seq})
}

// UpdateTTL updates the idle TTL for key if key is still cached. If key is
// idle, it also resets the current expiration timer. If key is leased, the new
// TTL is used when the final lease is released.
// It returns false if key was no longer cached.
func (c *Cache[K, V]) UpdateTTL(key K, ttl time.Duration) (bool, error) {
	if err := c.wheel.ValidateTTL(ttl); err != nil {
		return false, err
	}

	c.mu.Lock()
	item, ok := c.items[key]
	if !ok {
		c.mu.Unlock()
		return false, nil
	}
	item.ttl = ttl
	if item.active > 0 {
		c.mu.Unlock()
		return true, nil
	}

	updated, err := c.wheel.UpdateTTL(cacheWheelKey[K]{key: key, seq: item.seq}, ttl)
	if err != nil {
		c.mu.Unlock()
		return false, err
	}
	if updated {
		c.mu.Unlock()
		return true, nil
	}

	c.nextSeq++
	item.seq = c.nextSeq
	err = c.wheel.Add(cacheWheelKey[K]{key: key, seq: item.seq}, ttl)
	c.mu.Unlock()
	if err != nil {
		return false, err
	}
	return true, nil
}

// ValidateTTL returns an error if ttl cannot be used by this cache.
func (c *Cache[K, V]) ValidateTTL(ttl time.Duration) error {
	return c.wheel.ValidateTTL(ttl)
}

// Stop stops the cache expiration wheel.
func (c *Cache[K, V]) Stop() {
	c.wheel.Stop()
}

// Value returns the leased value.
func (l *Lease[K, V]) Value() V {
	return l.value
}

// Release releases the lease. When the final lease on a value is released its
// idle expiration timer restarts from the value's full TTL. Release is
// idempotent.
func (l *Lease[K, V]) Release() {
	l.once.Do(func() {
		l.cache.release(l.key)
	})
}

func (c *Cache[K, V]) release(key K) {
	c.mu.Lock()
	item, ok := c.items[key]
	if !ok || item.active == 0 {
		c.mu.Unlock()
		return
	}
	item.active--
	if item.active > 0 {
		c.mu.Unlock()
		return
	}
	// The final lease was released; restart the idle timer from the full TTL.
	c.nextSeq++
	item.seq = c.nextSeq

	// Add only fails for a TTL the wheel cannot hold, which Store already
	// validated, so this is unreachable in practice. Expire the value if it
	// ever happens so the resource is not leaked while untracked.
	err := c.wheel.Add(cacheWheelKey[K]{key: key, seq: item.seq}, item.ttl)
	if err != nil {
		delete(c.items, key)
	}
	value := item.value
	c.mu.Unlock()

	if err != nil && c.onExpire != nil {
		c.onExpire(key, value)
	}
}

func (c *Cache[K, V]) expire(wheelKey cacheWheelKey[K]) {
	c.mu.Lock()
	item, ok := c.items[wheelKey.key]
	if !ok || item.seq != wheelKey.seq || item.active > 0 {
		c.mu.Unlock()
		return
	}
	delete(c.items, wheelKey.key)
	c.mu.Unlock()

	if c.onExpire != nil {
		c.onExpire(wheelKey.key, item.value)
	}
}
