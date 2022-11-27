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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapOperations(t *testing.T) {
	m := NewMap(
		NewKV("test1", 1),
		NewKV("test2", 2),
		NewKV("test3", 3),
		NewKV("test4", 4),
	)
	v, exists := m.GetIndex(3)
	assert.True(t, exists)
	assert.Equal(t, 4, v)
	v, exists = m.Get("test1")
	assert.True(t, exists)
	assert.Equal(t, 1, v)
	v, exists = m.Get("test3")
	assert.True(t, exists)
	assert.Equal(t, 3, v)
	assert.Equal(t, 4, m.Len())

	m.Delete("test3")
	v, exists = m.GetIndex(2)
	assert.True(t, exists)
	assert.Equal(t, 4, v)
	_, exists = m.Get("test3")
	assert.False(t, exists)
	assert.Equal(t, 3, m.Len())

	m.Set("test5", 5)
	v, exists = m.GetIndex(3)
	assert.True(t, exists)
	assert.Equal(t, 5, v)
	v, exists = m.Get("test5")
	assert.True(t, exists)
	assert.Equal(t, 5, v)
	assert.Equal(t, 4, m.Len())

	m.Set("test5", 6)
	v, exists = m.GetIndex(3)
	assert.True(t, exists)
	assert.Equal(t, 6, v)
	v, exists = m.Get("test5")
	assert.True(t, exists)
	assert.Equal(t, 6, v)
	assert.Equal(t, 4, m.Len())

	m.DeleteIndex(0)
	v, exists = m.GetIndex(2)
	assert.True(t, exists)
	assert.Equal(t, 6, v)
	_, exists = m.Get("test1")
	assert.False(t, exists)
	assert.Equal(t, 3, m.Len())
}

func TestMapIteratorOperations(t *testing.T) {
	m := NewMap(
		NewKV("test1", 1),
		NewKV("test2", 2),
		NewKV("test3", 3),
		NewKV("test4", 4),
	)

	i := 1
	for kv := m.Start(); kv != nil; kv = kv.Next() {
		assert.Equal(t, i, kv.val)
		i++
	}
}

func TestMapFromListOfKeys(t *testing.T) {
	m := NewMap(
		NewKV("test1", 1),
		NewKV("test2", 2),
		NewKV("test3", 3),
		NewKV("test4", 4),
	)

	m2 := m.From([]string{"test2", "test3"})

	assert.Equal(t, 2, len(m2))
	assert.Equal(t, 2, m2[0].val)
	assert.Equal(t, 3, m2[1].val)
}

func TestEmptyMapStartIsNil(t *testing.T) {
	m := NewMap[string, int]()

	assert.Nil(t, m.Start())
}

func TestMapGetIndexOutOfBounds(t *testing.T) {
	m := NewMap(
		NewKV("test1", 1),
	)
	_, exists := m.GetIndex(1)
	assert.False(t, exists)
}

func TestMaDeleteIndexOutOfBounds(t *testing.T) {
	m := NewMap(
		NewKV("test1", 1),
	)

	m.DeleteIndex(1)

	assert.Equal(t, 1, m.Len())
}
