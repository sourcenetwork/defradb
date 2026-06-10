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

func TestIndexStateKey_ToString(t *testing.T) {
	key := NewIndexStateKey("col1", 42)
	expected := INDEX_STATE + "/col1/42"
	assert.Equal(t, expected, key.ToString())
}

func TestIndexStateKey_Bytes(t *testing.T) {
	key := NewIndexStateKey("col1", 42)
	expected := INDEX_STATE + "/col1/42"
	assert.Equal(t, []byte(expected), key.Bytes())
}

func TestIndexStateKey_ToDS(t *testing.T) {
	key := NewIndexStateKey("col1", 42)
	assert.Equal(t, key.ToString(), key.ToDS().String())
}

func TestNewIndexStateKeyFromString(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    IndexStateKey
		expectError bool
		errorMsg    string
	}{
		{
			name:  "valid key round-trips back to struct",
			input: INDEX_STATE + "/col1/42",
			expected: IndexStateKey{
				CollectionID: "col1",
				IndexID:      42,
			},
		},
		{
			name:  "indexID zero",
			input: INDEX_STATE + "/mycollection/0",
			expected: IndexStateKey{
				CollectionID: "mycollection",
				IndexID:      0,
			},
		},
		{
			name:        "invalid prefix",
			input:       "/wrong/prefix/col1/42",
			expectError: true,
			errorMsg:    "invalid key string",
		},
		{
			name:        "missing indexID component",
			input:       INDEX_STATE + "/col1",
			expectError: true,
			errorMsg:    "invalid key string",
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
			errorMsg:    "invalid key string",
		},
		{
			name:        "non-numeric indexID",
			input:       INDEX_STATE + "/col1/notanumber",
			expectError: true,
			errorMsg:    "invalid key string",
		},
		{
			name:        "too many components",
			input:       INDEX_STATE + "/col1/42/extra",
			expectError: true,
			errorMsg:    "invalid key string",
		},
		{
			name:        "empty collectionID",
			input:       INDEX_STATE + "//42",
			expectError: true,
			errorMsg:    "invalid key string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewIndexStateKeyFromString(tt.input)
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected.CollectionID, result.CollectionID)
				assert.Equal(t, tt.expected.IndexID, result.IndexID)
			}
		})
	}
}

func TestIndexStateKey_RoundTrip(t *testing.T) {
	original := NewIndexStateKey("testcollection", 99)
	parsed, err := NewIndexStateKeyFromString(original.ToString())
	require.NoError(t, err)
	assert.Equal(t, original.CollectionID, parsed.CollectionID)
	assert.Equal(t, original.IndexID, parsed.IndexID)
}

func TestNewIndexStateKeyPrefix(t *testing.T) {
	prefix := NewIndexStateKeyPrefix()
	assert.Equal(t, []byte(INDEX_STATE+"/"), prefix)
}

func TestNewIndexStateCollectionPrefix(t *testing.T) {
	prefix := NewIndexStateCollectionPrefix("col1")
	assert.Equal(t, []byte(INDEX_STATE+"/col1/"), prefix)
}
