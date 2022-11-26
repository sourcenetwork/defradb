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
	"github.com/stretchr/testify/require"
)

func TestMapOperations(t *testing.T) {
	m := NewMap(
		NewKV("test1", 1),
		NewKV("test2", 2),
		NewKV("test3", 3),
		NewKV("test4", 4),
	)
	assert.Equal(t, 4, m.GetIndex(3))
	assert.Equal(t, 1, m.Get("test1"))
	assert.Equal(t, 3, m.Get("test3"))
	assert.Equal(t, 4, m.Len())

	m.Delete("test3")
	assert.Equal(t, 4, m.GetIndex(2))
	assert.Equal(t, 0, m.Get("test3"))
	assert.Equal(t, 3, m.Len())

	m.Set(NewKV("test5", 5))
	assert.Equal(t, 5, m.GetIndex(3))
	assert.Equal(t, 5, m.Get("test5"))
	assert.Equal(t, 4, m.Len())

	m.Set(NewKV("test5", 6))
	assert.Equal(t, 6, m.GetIndex(3))
	assert.Equal(t, 6, m.Get("test5"))
	assert.Equal(t, 4, m.Len())

	m.DeleteIndex(0)
	assert.Equal(t, 6, m.GetIndex(2))
	assert.Equal(t, 0, m.Get("test1"))
	assert.Equal(t, 3, m.Len())
}

func TestMapIteratorOperations(t *testing.T) {
	m := NewMap(
		NewKV("test1", 1),
		NewKV("test2", 2),
		NewKV("test3", 3),
		NewKV("test4", 4),
	)

	iter := m.Iter()

	i := 1
	for {
		hasValue, err := iter.Next()
		require.NoError(t, err)
		if !hasValue {
			break
		}
		val, err := iter.Value()
		require.NoError(t, err)
		assert.Equal(t, i, val)
		i++
	}

	iter.Reset()
}
