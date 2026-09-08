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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIndexStaleEpochKey_Format pins the wire format. The format is load-bearing: the sweep scans by
// prefix and parses the key back, so a change to the prefix or segment order silently strands
// markers. ToString, Bytes and ToDS must agree.
func TestIndexStaleEpochKey_Format(t *testing.T) {
	key := NewIndexStaleEpochKey(3, 7)

	want := INDEX_STALE_EPOCH + "/3/7"
	assert.Equal(t, want, key.ToString())
	assert.Equal(t, []byte(want), key.Bytes())
	assert.Equal(t, key.ToString(), key.ToDS().String())

	assert.Equal(t, uint32(3), key.CollectionShortID)
	assert.Equal(t, uint32(7), key.IndexID)
}

// TestIndexStaleEpochPrefix checks the prefix is exactly INDEX_STALE_EPOCH and is itself a prefix of
// every key, so a prefix scan finds all markers.
func TestIndexStaleEpochPrefix(t *testing.T) {
	assert.Equal(t, INDEX_STALE_EPOCH, IndexStaleEpochPrefix())
	assert.Contains(t, NewIndexStaleEpochKey(1, 2).ToString(), IndexStaleEpochPrefix())
}

// TestIndexStaleEpochKey_RoundTrip checks parsing a key produced by ToString yields the original,
// across boundary values.
func TestIndexStaleEpochKey_RoundTrip(t *testing.T) {
	cases := []IndexStaleEpochKey{
		{CollectionShortID: 0, IndexID: 0},
		{CollectionShortID: 1, IndexID: 2},
		{CollectionShortID: 4294967295, IndexID: 4294967295}, // max uint32
	}
	for _, want := range cases {
		got, err := NewIndexStaleEpochKeyFromString(want.ToString())
		require.NoError(t, err)
		assert.Equal(t, want, got)
	}
}

// TestIndexStaleEpochKey_TrailingSlashSibling checks a sibling key that shares the prefix bytes but
// not the separator is not mis-parsed as a marker. The prefix is INDEX_STALE_EPOCH with no trailing
// slash, so a prefix scan also yields siblings like "/index/staleepochX" and "/index/staleepochs/1/2";
// the parser must reject them rather than return a bogus key pointing the sweep at the wrong index.
func TestIndexStaleEpochKey_TrailingSlashSibling(t *testing.T) {
	siblings := []string{
		INDEX_STALE_EPOCH + "X",         // no separator: not a real marker
		INDEX_STALE_EPOCH + "X/3/7",     // sibling namespace that happens to share the prefix bytes
		INDEX_STALE_EPOCH + "s/3/7",     // pluralized sibling
		INDEX_STALE_EPOCH + "extra/3/7", // longer sibling name
	}
	for _, key := range siblings {
		t.Run(key, func(t *testing.T) {
			// The sibling shares the prefix bytes, so a prefix scan would surface it.
			require.True(t, strings.HasPrefix(key, IndexStaleEpochPrefix()),
				"sibling must share the prefix bytes for this test to be meaningful")
			// But it must not parse as a valid marker.
			_, err := NewIndexStaleEpochKeyFromString(key)
			require.Error(t, err, "sibling key %q must not parse as a stale-epoch marker", key)
		})
	}
}

// TestIndexStaleEpochKey_FromStringErrors checks the parser rejects malformed keys instead of
// returning a bogus key that would point the sweep at the wrong index.
func TestIndexStaleEpochKey_FromStringErrors(t *testing.T) {
	cases := []struct {
		name string
		key  string
	}{
		{"empty", ""},
		{"prefix only", INDEX_STALE_EPOCH},
		{"missing index segment", INDEX_STALE_EPOCH + "/3"},
		{"too many segments", INDEX_STALE_EPOCH + "/3/7/9"},
		{"non-numeric short id", INDEX_STALE_EPOCH + "/x/7"},
		{"non-numeric index id", INDEX_STALE_EPOCH + "/3/y"},
		{"negative short id", INDEX_STALE_EPOCH + "/-1/7"},
		{"overflow short id", INDEX_STALE_EPOCH + "/4294967296/7"}, // uint32 max + 1
		{"empty segment", INDEX_STALE_EPOCH + "//7"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewIndexStaleEpochKeyFromString(tc.key)
			require.Error(t, err)
		})
	}
}
