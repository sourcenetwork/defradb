// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokens(t *testing.T) {
	tests := []struct {
		name     string
		val      string
		expected map[string]int
	}{
		{
			name:     "empty string",
			val:      "",
			expected: map[string]int{},
		},
		{
			name:     "lowercased",
			val:      "The Quick Fox",
			expected: map[string]int{"the": 1, "quick": 1, "fox": 1},
		},
		{
			name:     "repeated term is counted",
			val:      "the cat sat on the mat",
			expected: map[string]int{"the": 2, "cat": 1, "sat": 1, "on": 1, "mat": 1},
		},
		{
			name:     "split on anything that is not a letter or a digit",
			val:      "full-text search, done_well!",
			expected: map[string]int{"full": 1, "text": 1, "search": 1, "done": 1, "well": 1},
		},
		{
			name:     "digits are kept",
			val:      "bm25 ranks 10 documents",
			expected: map[string]int{"bm25": 1, "ranks": 1, "10": 1, "documents": 1},
		},
		{
			name:     "single-character tokens are dropped",
			val:      "a database is a store",
			expected: map[string]int{"database": 1, "is": 1, "store": 1},
		},
		{
			// A single-character token is dropped by character count, not by byte count, so a
			// one-rune multi-byte token goes the same way an ASCII one does.
			name:     "single multi-byte character is dropped",
			val:      "é étude",
			expected: map[string]int{"étude": 1},
		},
		{
			name:     "unicode is lowercased and kept whole",
			val:      "Ünicode ÉTUDE",
			expected: map[string]int{"ünicode": 1, "étude": 1},
		},
		{
			name:     "punctuation only",
			val:      "--- ... !!!",
			expected: map[string]int{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, Tokens(test.val))
		})
	}
}
