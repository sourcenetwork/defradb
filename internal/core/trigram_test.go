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

func TestTrigrams(t *testing.T) {
	tests := []struct {
		name     string
		val      string
		expected []string
	}{
		{
			name:     "empty string",
			val:      "",
			expected: nil,
		},
		{
			name:     "shorter than three bytes",
			val:      "hi",
			expected: nil,
		},
		{
			name:     "exactly three bytes",
			val:      "abc",
			expected: []string{"abc"},
		},
		{
			name:     "overlapping windows",
			val:      "abcd",
			expected: []string{"abc", "bcd"},
		},
		{
			name:     "lowercased",
			val:      "ABCD",
			expected: []string{"abc", "bcd"},
		},
		{
			name:     "repeated trigram appears once",
			val:      "banana",
			expected: []string{"ban", "ana", "nan"},
		},
		{
			// The windows are bytes, so a two-byte rune is split across them. The candidates
			// this produces are still correct, since the pattern match is re-run afterwards.
			name:     "unicode is windowed by byte",
			val:      "héllo",
			expected: []string{"h\xc3\xa9", "\xc3\xa9l", "\xa9ll", "llo"},
		},
		{
			// A single three-byte rune is exactly one trigram.
			name:     "three byte rune",
			val:      "€",
			expected: []string{"\xe2\x82\xac"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, Trigrams(test.val))
		})
	}
}
