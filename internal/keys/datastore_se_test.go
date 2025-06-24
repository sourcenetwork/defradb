// Copyright 2025 Democratized Data Foundation
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
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatastoreSE_ToString(t *testing.T) {
	tests := []struct {
		name     string
		key      DatastoreSE
		expected string
	}{
		{
			name:     "empty key",
			key:      DatastoreSE{},
			expected: "/se",
		},
		{
			name: "collection only",
			key: DatastoreSE{
				CollectionID: "col123",
			},
			expected: "/se/col123",
		},
		{
			name: "collection and index",
			key: DatastoreSE{
				CollectionID: "col123",
				IndexID:      "idx456",
			},
			expected: "/se/col123/idx456",
		},
		{
			name: "collection, index, and search tag",
			key: DatastoreSE{
				CollectionID: "col123",
				IndexID:      "idx456",
				SearchTag:    []byte{0x01, 0x02, 0x03},
			},
			expected: "/se/col123/idx456/010203",
		},
		{
			name: "full key with all fields",
			key: DatastoreSE{
				CollectionID: "col123",
				IndexID:      "idx456",
				SearchTag:    []byte{0x01, 0x02, 0x03},
				DocID:        "doc789",
			},
			expected: "/se/col123/idx456/010203/doc789",
		},
		{
			name: "skip index when no collection",
			key: DatastoreSE{
				IndexID:   "idx456",
				SearchTag: []byte{0x01, 0x02, 0x03},
				DocID:     "doc789",
			},
			expected: "/se",
		},
		{
			name: "skip search tag when no index",
			key: DatastoreSE{
				CollectionID: "col123",
				SearchTag:    []byte{0x01, 0x02, 0x03},
				DocID:        "doc789",
			},
			expected: "/se/col123",
		},
		{
			name: "skip doc id when no search tag",
			key: DatastoreSE{
				CollectionID: "col123",
				IndexID:      "idx456",
				DocID:        "doc789",
			},
			expected: "/se/col123/idx456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.key.ToString()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDatastoreSE_Bytes(t *testing.T) {
	key := DatastoreSE{
		CollectionID: "col123",
		IndexID:      "idx456",
		SearchTag:    []byte{0x01, 0x02, 0x03},
		DocID:        "doc789",
	}

	expected := []byte("/se/col123/idx456/010203/doc789")
	result := key.Bytes()
	assert.Equal(t, expected, result)
}

func TestDatastoreSE_ToDS(t *testing.T) {
	key := DatastoreSE{
		CollectionID: "col123",
		IndexID:      "idx456",
		SearchTag:    []byte{0x01, 0x02, 0x03},
		DocID:        "doc789",
	}

	dsKey := key.ToDS()
	assert.Equal(t, "/se/col123/idx456/010203/doc789", dsKey.String())
}

func TestNewDatastoreSEFromString(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    DatastoreSE
		expectError bool
		errorMsg    string
	}{
		{
			name:  "full valid key",
			input: "/se/col123/idx456/010203/doc789",
			expected: DatastoreSE{
				CollectionID: "col123",
				IndexID:      "idx456",
				SearchTag:    []byte{0x01, 0x02, 0x03},
				DocID:        "doc789",
			},
			expectError: false,
		},
		{
			name:  "key with only collection",
			input: "/se/col123",
			expected: DatastoreSE{
				CollectionID: "col123",
			},
			expectError: false,
		},
		{
			name:  "key with collection and index",
			input: "/se/col123/idx456",
			expected: DatastoreSE{
				CollectionID: "col123",
				IndexID:      "idx456",
			},
			expectError: false,
		},
		{
			name:  "key with collection, index, and search tag",
			input: "/se/col123/idx456/010203",
			expected: DatastoreSE{
				CollectionID: "col123",
				IndexID:      "idx456",
				SearchTag:    []byte{0x01, 0x02, 0x03},
			},
			expectError: false,
		},
		{
			name:        "invalid prefix",
			input:       "/notse/col123",
			expected:    DatastoreSE{},
			expectError: true,
			errorMsg:    "invalid SE key format",
		},
		{
			name:        "empty string",
			input:       "",
			expected:    DatastoreSE{},
			expectError: true,
			errorMsg:    "invalid SE key format",
		},
		{
			name:        "only slash",
			input:       "/",
			expected:    DatastoreSE{},
			expectError: true,
			errorMsg:    "invalid SE key format",
		},
		{
			name:        "invalid hex in search tag",
			input:       "/se/col123/idx456/xyz/doc789",
			expected:    DatastoreSE{},
			expectError: true,
			errorMsg:    "failed to decode search tag",
		},
		{
			name:  "minimum valid key",
			input: "/se",
			expected: DatastoreSE{
				CollectionID: "",
				IndexID:      "",
				SearchTag:    nil,
				DocID:        "",
			},
			expectError: false,
		},
		{
			name:  "complex search tag",
			input: "/se/col123/idx456/" + hex.EncodeToString([]byte("complex search tag")),
			expected: DatastoreSE{
				CollectionID: "col123",
				IndexID:      "idx456",
				SearchTag:    []byte("complex search tag"),
			},
			expectError: false,
		},
		{
			name:  "key with empty components",
			input: "/se///",
			expected: DatastoreSE{
				CollectionID: "",
				IndexID:      "",
				SearchTag:    []byte{},
			},
			expectError: false,
		},
		{
			name:  "key with trailing slash",
			input: "/se/col123/idx456/010203/doc789/",
			expected: DatastoreSE{
				CollectionID: "col123",
				IndexID:      "idx456",
				SearchTag:    []byte{0x01, 0x02, 0x03},
				DocID:        "doc789",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewDatastoreSEFromString(tt.input)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected.CollectionID, result.CollectionID)
				assert.Equal(t, tt.expected.IndexID, result.IndexID)
				assert.Equal(t, tt.expected.SearchTag, result.SearchTag)
				assert.Equal(t, tt.expected.DocID, result.DocID)
			}
		})
	}
}
