// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package se

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/encoding"
)

func TestEncodeFieldValueForSE(t *testing.T) {
	tests := []struct {
		name     string
		value    client.NormalValue
		expected string // We'll check if the encoded bytes contain expected patterns
	}{
		{
			name:     "string value",
			value:    client.NewNormalString([]byte("hello")),
			expected: "hello",
		},
		{
			name:     "int value",
			value:    client.NewNormalInt(42),
			expected: "", // Int encoding is binary, not human-readable
		},
		{
			name:     "float64 value",
			value:    client.NewNormalFloat64(3.14),
			expected: "", // Float encoding is binary, not human-readable
		},
		{
			name:     "bool true",
			value:    client.NewNormalBool(true),
			expected: "", // Bool encoding is binary
		},
		{
			name:     "bool false",
			value:    client.NewNormalBool(false),
			expected: "", // Bool encoding is binary
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := encoding.EncodeFieldValue(nil, tt.value, false)
			require.NotNil(t, encoded)
			
			// For strings, we can check if the encoded value contains the expected text
			if tt.expected != "" {
				assert.True(t, bytes.Contains(encoded, []byte(tt.expected)))
			}
			
			// Ensure we can decode back
			if str, ok := tt.value.String(); ok {
				_, decoded, err := encoding.DecodeFieldValue(encoded, false, client.FieldKind_NILLABLE_STRING)
				require.NoError(t, err)
				decodedStr, ok := decoded.String()
				require.True(t, ok)
				assert.Equal(t, str, decodedStr)
			}
		})
	}
}

func TestNilValueHandling(t *testing.T) {
	// Test that nil values are properly handled
	nilVal, err := client.NewNormalNil(client.FieldKind_NILLABLE_STRING)
	require.NoError(t, err)
	
	assert.True(t, nilVal.IsNil())
	
	// Encoding a nil value should work
	encoded := encoding.EncodeFieldValue(nil, nilVal, false)
	require.NotNil(t, encoded)
	
	// Decoding should return a nil value
	_, decoded, err := encoding.DecodeFieldValue(encoded, false, client.FieldKind_NILLABLE_STRING)
	require.NoError(t, err)
	assert.True(t, decoded.IsNil())
}