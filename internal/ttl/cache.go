// Copyright 2025 Democratized Data Foundation
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

// TTLCache is a generic cache with TTL based expiration. It uses the
// ttl.Wheel to effeciently manage the cache TTL expiration.
type Cache[K comparable, V any] struct {
	cache    *sync.Map
	tw       *Wheel[K]
	onExpire func(key K, value V)
}

// NewTTLCache returns a new cache which will call the onExpire func
// when a new key expires, unless its been deleted before its TTL
// expiration
func NewTTLCache[K comparable, V any](ctx context.Context, tick time.Duration, buckets int, onExpire func(key K, value V)) (*Cache[K, V], error) {
	c := &Cache[K, V]{
		cache:    &sync.Map{},
		onExpire: onExpire,
	}

	// on cache expiration we discard the txn
	var err error
	c.tw, err = NewWheel(ctx, tick, int64(buckets), func(k K) {
		vUntyped, ok := c.cache.LoadAndDelete(k)
		if !ok {
			return
		}
		vTyped := vUntyped.(V) //nolint:forcetypeassert
		c.onExpire(k, vTyped)
	})
	if err != nil {
		return nil, err
	}

	c.tw.Start()
	return c, nil
}

// Store a (key, value) pair with the corresponding TTL
func (c *Cache[K, V]) Store(key K, val V, ttl time.Duration) {
	c.cache.Store(key, val)
	c.tw.Add(key, ttl)
}

// Load a given key from the cache if it hasn't expired.
func (c *Cache[K, V]) Load(key K) (V, bool) {
	vUntyped, ok := c.cache.Load(key)
	if !ok {
		var vZero V
		return vZero, false
	}

	vTyped := vUntyped.(V) //nolint:forcetypeassert
	return vTyped, true
}

// LoadAndDelete loads a given key and deletes it from the
// cache if it hasn't expired.
func (c *Cache[K, V]) LoadAndDelete(key K) (V, bool) {
	vUntyped, ok := c.cache.LoadAndDelete(key)
	if !ok {
		var vZero V
		return vZero, false
	}

	vTyped := vUntyped.(V) //nolint:forcetypeassert
	return vTyped, true
}

// Delete deletes a given key from the cache if it hasn't
// expired.
func (c *Cache[K, V]) Delete(key K) {
	c.cache.Delete(key)
}

// UpdateTTL updates the TTL for a given key if it exists
// in the cache.
func (c *Cache[K, V]) UpdateTTL(key K, ttl time.Duration) error {
	return c.tw.UpdateTTL(key, ttl)
}

func (c *Cache[K, V]) Wheel() *Wheel[K] {
	return c.tw
}
