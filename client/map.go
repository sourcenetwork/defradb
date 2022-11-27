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

// KV is a key value struct used as input to Map
type KV[K comparable, V any] struct {
	key K
	val V

	item *item[*KV[K, V]]
}

// NewKV returns a new KV set with key and value
func NewKV[K comparable, V any](key K, val V) KV[K, V] {
	return KV[K, V]{key, val, nil}
}

// Map implemnents an ordered map
type Map[K comparable, V any] struct {
	values map[K]*KV[K, V]
	list   *list[*KV[K, V]]
}

// NewMap return a pointer to a new Map set with the given KVs
func NewMap[K comparable, V any](kvs ...KV[K, V]) *Map[K, V] {
	m := &Map[K, V]{
		values: make(map[K]*KV[K, V]),
		list:   newList[*KV[K, V]](),
	}
	for _, kv := range kvs {
		m.Set(kv.key, kv.val)
	}
	return m
}

// Clear removes all values form the ordered map
func (m *Map[K, V]) Clear() {
	for kv := m.Start(); kv != nil; kv = kv.Next() {
		delete(m.values, kv.key)
	}
	m.list.Init()
}

// Copy returns a copy of the original map
func (m *Map[K, V]) Copy() *Map[K, V] {
	newMap := NewMap[K, V]()
	for kv := m.Start(); kv != nil; kv = kv.Next() {
		newMap.Set(kv.key, kv.val)
	}
	return newMap
}

// Delete removes the items from the map according to the given keys
func (m *Map[K, V]) Delete(key K) (V, bool) {
	if kv, exists := m.values[key]; exists {
		m.list.delete(kv.item)
		delete(m.values, key)
		return kv.val, true
	}
	var zero V
	return zero, false
}

// DeleteIndex removes the items from the map at the given index
func (m *Map[K, V]) DeleteIndex(i int) {
	kv := m.list.deleteIndex(i)
	if kv != nil {
		delete(m.values, kv.key)
	}
}

// Get returns a map item
func (m *Map[K, V]) Get(key K) (V, bool) {
	if kv, exists := m.values[key]; exists {
		return kv.val, true
	}
	var zero V
	return zero, false
}

// GetIndex returns a map item at the given index
func (m *Map[K, V]) GetIndex(i int) (V, bool) {
	kv := m.list.index(i)
	if kv == nil {
		var zero V
		return zero, false
	}
	return kv.val, true
}

// Keys returns the list of keys of the ordered map
func (m *Map[K, V]) Keys() []K {
	keys := []K{}
	for kv := m.Start(); kv != nil; kv = kv.Next() {
		keys = append(keys, kv.key)
	}
	return keys
}

// Values returns the list of values of the ordered map
func (m *Map[K, V]) Values() []V {
	values := []V{}
	for kv := m.Start(); kv != nil; kv = kv.Next() {
		values = append(values, kv.val)
	}
	return values
}

// Len returns the number of items in the map
func (m *Map[K, V]) Len() int {
	return len(m.values)
}

// Set adds or modifies values in the map according to the given KVs
func (m *Map[K, V]) Set(key K, val V) {
	if kv, exists := m.values[key]; exists {
		kv.val = val
		return
	}

	kv := &KV[K, V]{
		key: key,
		val: val,
	}
	kv.item = m.list.append(kv)
	m.values[key] = kv
}

// From creates a new ordered map with items represented by the given keys.
func (m *Map[K, V]) From(keys []K) []*KV[K, V] {
	kvList := make([]*KV[K, V], 0, len(keys))
	for _, key := range keys {
		if kv, exists := m.values[key]; exists {
			kvList = append(kvList, kv)
		}
	}
	return kvList
}

// Start returns the first item in the ordered map.
func (m *Map[K, V]) Start() *KV[K, V] {
	return m.list.root.next.value
}

// Next returns the following item in the ordered map.
// If the next item is the root, we return nil.
func (kv *KV[K, V]) Next() *KV[K, V] {
	if i := kv.item.next; kv.item.list != nil && i != &kv.item.list.root {
		return i.value
	}
	return nil
}

type item[T any] struct {
	prev, next *item[T]

	list *list[T]

	value T
}
type list[T any] struct {
	root item[T]
}

func (l *list[T]) Init() *list[T] {
	l.root.prev = &l.root
	l.root.next = &l.root
	return l
}

func newList[T any]() *list[T] {
	return new(list[T]).Init()
}

func (l *list[T]) append(val T) *item[T] {
	i := &item[T]{
		value: val,
		prev:  l.root.prev,
		next:  l.root.prev.next,
		list:  l,
	}
	i.prev.next = i
	i.next.prev = i
	return i
}

func (l *list[T]) delete(i *item[T]) {
	i.prev.next = i.next
	i.next.prev = i.prev
	i.next = nil
	i.prev = nil
	i.list = nil
}

func (l *list[T]) deleteIndex(i int) T {
	next := l.root.next
	j := 0
	for {
		if next == &l.root {
			return next.value
		}
		if i == j {
			next.prev.next = next.next
			next.next.prev = next.prev
			next.next = nil
			next.prev = nil
			next.list = nil
			return next.value
		}
		next = next.next
		j++
	}
}

func (l *list[T]) index(i int) T {
	next := l.root.next
	var zero T
	j := 0
	for {
		if next == &l.root {
			return zero
		}
		if i == j {
			return next.value
		}
		next = next.next
		j++
	}
}
