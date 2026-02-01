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
	"github.com/stretchr/testify/require"
)

func TestSystemStoreFieldIDKey_Bytes(t *testing.T) {
	tests := []struct {
		name     string
		key      FieldID
		expected string
	}{
		{
			name:     "empty key",
			key:      FieldID{},
			expected: FIELD_SHORT_ID,
		},
		{
			name: "collection only",
			key: FieldID{
				CollectionShortID: 1,
			},
			expected: FIELD_SHORT_ID + "/\x89",
		},
		{
			name: "collection and index",
			key: FieldID{
				CollectionShortID: 1,
				FieldID:           "idx456",
			},
			expected: FIELD_SHORT_ID + "/\x89/idx456",
		},
		{
			name: "collection, index, and search tag",
			key: FieldID{
				CollectionShortID: 10,
				FieldID:           "idx456",
			},
			expected: FIELD_SHORT_ID + "/\x92/idx456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.key.Bytes()
			assert.Equal(t, tt.expected, string(result))
		})
	}
}

func TestNewFieldIDFromBytes(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    FieldID
		expectError bool
		errorMsg    string
	}{
		{
			name:  "only prefix",
			input: FIELD_SHORT_ID,
			expected: FieldID{
				CollectionShortID: 0,
				FieldID:           "",
			},
			expectError: false,
		},
		{
			name:  "full valid key",
			input: FIELD_SHORT_ID + "/\x89/idx456",
			expected: FieldID{
				CollectionShortID: 1,
				FieldID:           "idx456",
			},
			expectError: false,
		},
		{
			name:  "key with only collection",
			input: FIELD_SHORT_ID + "/\x89",
			expected: FieldID{
				CollectionShortID: 1,
			},
			expectError: false,
		},
		{
			name:        "invalid prefix",
			input:       "/notfieldshorID/\x89",
			expected:    FieldID{},
			expectError: true,
			errorMsg:    "invalid key string",
		},
		{
			name:        "empty string",
			input:       "",
			expected:    FieldID{},
			expectError: true,
			errorMsg:    "invalid key string",
		},
		{
			name:        "only slash",
			input:       "/",
			expected:    FieldID{},
			expectError: true,
			errorMsg:    "invalid key string",
		},
		{
			name:  "key with empty components",
			input: FIELD_SHORT_ID + "//",
			expected: FieldID{
				CollectionShortID: 0,
				FieldID:           "",
			},
			expectError: true,
			errorMsg:    "insufficient bytes to decode buffer",
		},
		{
			name:  "key with trailing slash",
			input: FIELD_SHORT_ID + "/\x89/idx456/",
			expected: FieldID{
				CollectionShortID: 1,
				FieldID:           "idx456",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewFieldIDFromBytes([]byte(tt.input))
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected.CollectionShortID, result.CollectionShortID)
				assert.Equal(t, tt.expected.FieldID, result.FieldID)
			}
		})
	}
}
