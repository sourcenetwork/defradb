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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFullTextIndexKey_PartsAndEpochsAreDisjoint(t *testing.T) {
	posting := NewFullTextPostingKey(1, 2, 3, "cat", 4)
	length := NewFullTextLengthKey(1, 2, 3, 4)
	totals := NewFullTextTotalsKey(1, 2, 3)
	otherEpoch := NewFullTextPostingKey(1, 2, 4, "cat", 4)
	prefix := NewFullTextPostingKey(1, 2, 3, "cat", 0)

	assert.NotEqual(t, posting.Bytes(), length.Bytes())
	assert.NotEqual(t, length.Bytes(), totals.Bytes())
	assert.NotEqual(t, posting.Bytes(), otherEpoch.Bytes())
	assert.True(t, bytes.HasPrefix(posting.Bytes(), prefix.Bytes()))
}

func TestFullTextIndexKey_TermEncodingDoesNotCollideWithSeparators(t *testing.T) {
	a := NewFullTextPostingKey(1, 2, 3, "a/b", 4)
	b := NewFullTextPostingKey(1, 2, 3, "a", 4)
	assert.NotEqual(t, a.Bytes(), b.Bytes())
}

func TestFullTextIndexKey_AllFamiliesShareCanonicalIndexEpochPrefix(t *testing.T) {
	canonicalKey := NewIndexDataStoreKey(123, 456, 7, nil)
	canonical := canonicalKey.Bytes()
	keys := []FullTextIndexKey{
		NewFullTextPostingKey(123, 456, 7, "", 0),
		NewFullTextPostingKey(123, 456, 7, "cat/dog", 89),
		NewFullTextLengthKey(123, 456, 7, 89),
		NewFullTextTotalsKey(123, 456, 7),
	}

	for _, key := range keys {
		encoded := key.Bytes()
		assert.Greater(t, len(encoded), len(canonical))
		assert.Equal(t, canonical, encoded[:len(canonical)])
		assert.Equal(t, byte('/'), encoded[len(canonical)])
	}
}
