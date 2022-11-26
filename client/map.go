// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import (
	"github.com/sourcenetwork/immutable/enumerable"
)

// KV is a key value struct used as input to Map
type KV[K comparable, V any] struct {
	key K
	val V
}

// NewKV returns a new KV set with key and value
func NewKV[K comparable, V any](key K, val V) KV[K, V] {
	return KV[K, V]{key, val}
}

// Map implemnents an ordered map
type Map[K comparable, V any] struct {
	values map[K]V
	order  []K
}

// NewMap return a pointer to a new Map set with the given KVs
func NewMap[K comparable, V any](kvs ...KV[K, V]) *Map[K, V] {
	m := &Map[K, V]{
		values: make(map[K]V),
		order:  make([]K, 0, len(kvs)),
	}
	m.Set(kvs...)
	return m
}

// Delete removes the items from the map according to the given keys
func (m *Map[K, V]) Delete(keys ...K) {
	for _, key := range keys {
		if _, ok := m.values[key]; ok {
			delete(m.values, key)
			for i, k := range m.order {
				if k == key {
					copy(m.order[i:], m.order[i+1:])
					var zero K
					m.order[len(m.order)-1] = zero
					m.order = m.order[:len(m.order)-1]
				}
			}
		}
	}
}

// DeleteIndex removes the items from the map at the given index
func (m *Map[K, V]) DeleteIndex(i int) {
	if _, ok := m.values[m.order[i]]; ok {
		delete(m.values, m.order[i])
		copy(m.order[i:], m.order[i+1:])
		var zero K
		m.order[len(m.order)-1] = zero
		m.order = m.order[:len(m.order)-1]
	}
}

// Get returns a map item
func (m *Map[K, V]) Get(key K) V {
	return m.values[key]
}

// GetIndex returns a map item at the given index
func (m *Map[K, V]) GetIndex(i int) V {
	return m.values[m.order[i]]
}

// Len returns the number of items in the map
func (m *Map[K, V]) Len() int {
	return len(m.order)
}

// Set adds or modifies values in the map according to the given KVs
func (m *Map[K, V]) Set(kvs ...KV[K, V]) {
	for _, kv := range kvs {
		if _, ok := m.values[kv.key]; !ok {
			m.values[kv.key] = kv.val
			m.order = append(m.order, kv.key)
		} else {
			m.values[kv.key] = kv.val
		}
	}
}

type mapIterator[K comparable, V any] struct {
	m            *Map[K, V]
	currentIndex int
	maxIndex     int
}

// New creates an `Enumerable` from the given Map.
func (m *Map[K, V]) Iter() enumerable.Enumerable[V] {
	return &mapIterator[K, V]{
		m:            m,
		currentIndex: -1,
		maxIndex:     len(m.order) - 1,
	}
}

func (mi *mapIterator[K, V]) Next() (bool, error) {
	if mi.currentIndex == mi.maxIndex {
		return false, nil
	}
	mi.currentIndex += 1
	return true, nil
}

func (mi *mapIterator[K, V]) Value() (V, error) {
	return mi.m.values[mi.m.order[mi.currentIndex]], nil
}

func (mi *mapIterator[K, V]) Reset() {
	mi.currentIndex = -1
}
