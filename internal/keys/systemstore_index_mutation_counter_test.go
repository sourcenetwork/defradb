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

// TestIndexMutationCounterKey_Format pins the wire format. A write reads this key by exact bytes, so a
// change to the prefix or segment silently strands the counter and reopens the stale-handle race.
// ToString, Bytes and ToDS must agree.
func TestIndexMutationCounterKey_Format(t *testing.T) {
	key := NewIndexMutationCounterKey(3)

	want := INDEX_MUTATION_COUNTER + "/3"
	assert.Equal(t, want, key.ToString())
	assert.Equal(t, []byte(want), key.Bytes())
	assert.Equal(t, key.ToString(), key.ToDS().String())

	assert.Equal(t, uint32(3), key.CollectionShortID)
}

// TestIndexMutationCounterKey_Distinct checks each collection short ID gets its own key, so a bump on
// one collection cannot conflict a write on another.
func TestIndexMutationCounterKey_Distinct(t *testing.T) {
	cases := []uint32{0, 1, 2, 4294967295} // includes max uint32
	seen := make(map[string]uint32, len(cases))
	for _, shortID := range cases {
		s := NewIndexMutationCounterKey(shortID).ToString()
		if prev, dup := seen[s]; dup {
			t.Fatalf("short ids %d and %d produced the same key %q", prev, shortID, s)
		}
		seen[s] = shortID
	}
}
