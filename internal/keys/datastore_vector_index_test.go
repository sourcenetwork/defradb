// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package keys

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVectorIndexKey_MetaKey_ProducesExpectedString(t *testing.T) {
	key := NewVectorMetaKey(1, 2, 3)
	s := key.ToString()
	require.True(t, strings.HasPrefix(s, "/"))
	require.True(t, strings.HasSuffix(s, "m"))
}

func TestVectorIndexKey_NodeKey_ProducesExpectedString(t *testing.T) {
	key := NewVectorNodeKey(1, 2, 3, 7)
	prefix := NewVectorNodePrefix(1, 2, 3)

	s := key.ToString()
	require.True(t, strings.HasPrefix(s, "/"))
	// A node key extends the node prefix with a "/" plus the encoded node id.
	require.True(t, strings.HasPrefix(s, prefix.ToString()))
	require.Greater(t, len(s), len(prefix.ToString()))
}

func TestVectorIndexKey_MetaAndNodeKeys_ProduceDistinctBytes(t *testing.T) {
	meta := NewVectorMetaKey(1, 2, 3)
	node := NewVectorNodeKey(1, 2, 3, 7)

	assert.NotEqual(t, meta.Bytes(), node.Bytes())
	assert.NotEqual(t, meta.ToString(), node.ToString())
}

func TestVectorIndexKey_NodeKeysSameCollectionIndexEpoch_ShareNodePrefix(t *testing.T) {
	prefix := NewVectorNodePrefix(1, 2, 3)
	node1 := NewVectorNodeKey(1, 2, 3, 7)
	node2 := NewVectorNodeKey(1, 2, 3, 42)

	assert.True(t, bytes.HasPrefix(node1.Bytes(), prefix.Bytes()))
	assert.True(t, bytes.HasPrefix(node2.Bytes(), prefix.Bytes()))
}

func TestVectorIndexKey_NodePrefix_DoesNotContainMetaKey(t *testing.T) {
	prefix := NewVectorNodePrefix(1, 2, 3)
	meta := NewVectorMetaKey(1, 2, 3)

	assert.False(t, bytes.HasPrefix(meta.Bytes(), prefix.Bytes()))
}

func TestVectorIndexKey_DifferentEpochs_ProduceDisjointNodePrefixes(t *testing.T) {
	prefix1 := NewVectorNodePrefix(1, 2, 3)
	prefix2 := NewVectorNodePrefix(1, 2, 4)

	assert.False(t, bytes.HasPrefix(prefix1.Bytes(), prefix2.Bytes()))
	assert.False(t, bytes.HasPrefix(prefix2.Bytes(), prefix1.Bytes()))

	node1 := NewVectorNodeKey(1, 2, 3, 7)
	node2 := NewVectorNodeKey(1, 2, 4, 7)
	assert.False(t, bytes.HasPrefix(node1.Bytes(), prefix2.Bytes()))
	assert.False(t, bytes.HasPrefix(node2.Bytes(), prefix1.Bytes()))
}

func TestVectorIndexKey_DifferentIndexIDs_ProduceDisjointNodePrefixes(t *testing.T) {
	prefix1 := NewVectorNodePrefix(1, 2, 3)
	prefix2 := NewVectorNodePrefix(1, 5, 3)

	assert.False(t, bytes.HasPrefix(prefix1.Bytes(), prefix2.Bytes()))
	assert.False(t, bytes.HasPrefix(prefix2.Bytes(), prefix1.Bytes()))
}

func TestVectorIndexKey_DifferentCollections_ProduceDisjointNodePrefixes(t *testing.T) {
	prefix1 := NewVectorNodePrefix(1, 2, 3)
	prefix2 := NewVectorNodePrefix(9, 2, 3)

	assert.False(t, bytes.HasPrefix(prefix1.Bytes(), prefix2.Bytes()))
	assert.False(t, bytes.HasPrefix(prefix2.Bytes(), prefix1.Bytes()))
}

func TestVectorIndexKey_NodeKeysOrdering_SortsByNodeID(t *testing.T) {
	node1 := NewVectorNodeKey(1, 2, 3, 1)
	node2 := NewVectorNodeKey(1, 2, 3, 2)
	node10 := NewVectorNodeKey(1, 2, 3, 10)

	assert.True(t, bytes.Compare(node1.Bytes(), node2.Bytes()) < 0)
	assert.True(t, bytes.Compare(node2.Bytes(), node10.Bytes()) < 0)
}

func TestVectorIndexKey_ToDS_MatchesToString(t *testing.T) {
	key := NewVectorMetaKey(1, 2, 3)
	assert.Equal(t, key.ToString(), key.ToDS().String())
}

func TestVectorIndexKey_GetCollectionShortID_ReturnsCollectionShortID(t *testing.T) {
	key := NewVectorNodeKey(11, 2, 3, 7)
	assert.Equal(t, uint32(11), key.GetCollectionShortID())
}

func TestVectorIndexKey_PrefixEnd_ExcludesExactPrefixMatch(t *testing.T) {
	prefix := NewVectorNodePrefix(1, 2, 3)
	node := NewVectorNodeKey(1, 2, 3, 7)

	end := prefix.PrefixEnd()
	require.NotNil(t, end)

	// The node key falls strictly between the prefix and its PrefixEnd.
	assert.True(t, bytes.Compare(prefix.Bytes(), node.Bytes()) < 0)
	assert.True(t, bytes.Compare(node.Bytes(), end.Bytes()) < 0)
}
