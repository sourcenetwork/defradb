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
func NewTTLCache[K comparable, V any](tick time.Duration, buckets int, onExpire func(key K, value V)) *Cache[K, V] {
	tc := &Cache[K, V]{
		cache:    &sync.Map{},
		onExpire: onExpire,
	}

	// on cache expiration we discard the txn
	tc.tw = NewWheel(tick, buckets, func(k K) {
		vUntyped, ok := tc.cache.LoadAndDelete(k)
		if !ok {
			return
		}
		vTyped := vUntyped.(V)
		tc.onExpire(k, vTyped)
	})
	tc.tw.Start()
	return tc
}

// Store a (key, value) pair with the corresponding TTL
func (tc *Cache[K, V]) Store(key K, val V, ttl time.Duration) {
	tc.cache.Store(key, val)
	tc.tw.Add(key, ttl)
}

// Load a given key from the cache if it hasn't expired.
func (tc *Cache[K, V]) Load(key K) (V, bool) {
	vUntyped, ok := tc.cache.Load(key)
	if !ok {
		var vZero V
		return vZero, false
	}

	vTyped := vUntyped.(V)
	return vTyped, true
}

// LoadAndDelete loads a given key and deletes it from the
// cache if it hasn't expired.
func (tc *Cache[K, V]) LoadAndDelete(key K) (V, bool) {
	vUntyped, ok := tc.cache.LoadAndDelete(key)
	if !ok {
		var vZero V
		return vZero, false
	}

	vTyped := vUntyped.(V)
	return vTyped, true
}

// Delete deletes a given key from the cache if it hasn't
// expired.
func (tc *Cache[K, V]) Delete(key K) {
	tc.cache.Delete(key)
}
