// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package encoding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeBoolAscending(t *testing.T) {
	tests := []struct {
		name     string
		input    bool
		expected []byte
	}{
		{
			name:     "true value",
			input:    true,
			expected: []byte{trueMarker},
		},
		{
			name:     "false value",
			input:    false,
			expected: []byte{falseMarker},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeBoolAscending(nil, tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEncodeBoolDescending(t *testing.T) {
	tests := []struct {
		name     string
		input    bool
		expected []byte
	}{
		{
			name:     "true value",
			input:    true,
			expected: []byte{falseMarker}, // inverted due to descending order
		},
		{
			name:     "false value",
			input:    false,
			expected: []byte{trueMarker}, // inverted due to descending order
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeBoolDescending(nil, tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDecodeBoolAscending(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expected    bool
		expectedErr error
		remaining   []byte
	}{
		{
			name:        "decode true",
			input:       []byte{trueMarker},
			expected:    true,
			expectedErr: nil,
			remaining:   []byte{},
		},
		{
			name:        "decode false",
			input:       []byte{falseMarker},
			expected:    false,
			expectedErr: nil,
			remaining:   []byte{},
		},
		{
			name:        "invalid marker",
			input:       []byte{0x99},
			expected:    false,
			expectedErr: NewErrMarkersNotFound([]byte{0x99}, falseMarker, trueMarker),
			remaining:   []byte{0x99},
		},
		{
			name:        "with remaining bytes",
			input:       []byte{trueMarker, 0x01, 0x02},
			expected:    true,
			expectedErr: nil,
			remaining:   []byte{0x01, 0x02},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remaining, result, err := DecodeBoolAscending(tt.input)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			assert.Equal(t, tt.remaining, remaining)
		})
	}
}

func TestDecodeBoolDescending(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expected    bool
		expectedErr error
		remaining   []byte
	}{
		{
			name:        "decode true",
			input:       []byte{falseMarker}, // inverted due to descending order
			expected:    true,
			expectedErr: nil,
			remaining:   []byte{},
		},
		{
			name:        "decode false",
			input:       []byte{trueMarker}, // inverted due to descending order
			expected:    false,
			expectedErr: nil,
			remaining:   []byte{},
		},
		{
			name:        "invalid marker",
			input:       []byte{0x99},
			expected:    false,
			expectedErr: NewErrMarkersNotFound([]byte{0x99}, falseMarker, trueMarker),
			remaining:   []byte{0x99},
		},
		{
			name:        "with remaining bytes",
			input:       []byte{falseMarker, 0x01, 0x02},
			expected:    true,
			expectedErr: nil,
			remaining:   []byte{0x01, 0x02},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remaining, result, err := DecodeBoolDescending(tt.input)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			assert.Equal(t, tt.remaining, remaining)
		})
	}
}
