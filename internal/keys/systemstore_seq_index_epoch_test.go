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
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIndexEpochSequenceKey_Format pins the wire format of the epoch sequence key. The format is
// load-bearing: a change to the prefix or the segment order silently moves epochs to a different
// keyspace, making every epoch read miss. ToString, Bytes and ToDS must agree.
func TestIndexEpochSequenceKey_Format(t *testing.T) {
	key := NewIndexEpochSequenceKey(3, 7)

	want := INDEX_EPOCH_SEQ + "/3/7"
	assert.Equal(t, want, key.ToString())
	assert.Equal(t, []byte(want), key.Bytes())
	assert.Equal(t, key.ToString(), key.ToDS().String())

	assert.Equal(t, uint32(3), key.CollectionShortID)
	assert.Equal(t, uint32(7), key.IndexID)
}
