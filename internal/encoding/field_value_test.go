// Copyright 2024 Democratized Data Foundation
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
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

func TestEncodeDecodeFieldValue(t *testing.T) {
	normalNil, err := client.NewNormalNil(client.FieldKind_NILLABLE_INT)
	require.NoError(t, err)

	tests := []struct {
		name               string
		inputVal           client.NormalValue
		expectedBytes      []byte
		expectedBytesDesc  []byte
		expectedDecodedVal any
	}{
		{
			name:               "nil",
			inputVal:           normalNil,
			expectedBytes:      EncodeNullAscending(nil),
			expectedBytesDesc:  EncodeNullDescending(nil),
			expectedDecodedVal: normalNil,
		},
		{
			name:               "bool true",
			inputVal:           client.NewNormalBool(true),
			expectedBytes:      EncodeBoolAscending(nil, true),
			expectedBytesDesc:  EncodeBoolDescending(nil, true),
			expectedDecodedVal: client.NewNormalBool(true),
		},
		{
			name:               "bool false",
			inputVal:           client.NewNormalBool(false),
			expectedBytes:      EncodeBoolAscending(nil, false),
			expectedBytesDesc:  EncodeBoolDescending(nil, false),
			expectedDecodedVal: client.NewNormalBool(false),
		},
		{
			name:               "int",
			inputVal:           client.NewNormalInt(55),
			expectedBytes:      EncodeVarintAscending(nil, 55),
			expectedBytesDesc:  EncodeVarintDescending(nil, 55),
			expectedDecodedVal: client.NewNormalInt(55),
		},
		{
			name:               "float",
			inputVal:           client.NewNormalFloat(0.2),
			expectedBytes:      EncodeFloatAscending(nil, 0.2),
			expectedBytesDesc:  EncodeFloatDescending(nil, 0.2),
			expectedDecodedVal: client.NewNormalFloat(0.2),
		},
		{
			name:               "string",
			inputVal:           client.NewNormalString("str"),
			expectedBytes:      EncodeBytesAscending(nil, []byte("str")),
			expectedBytesDesc:  EncodeBytesDescending(nil, []byte("str")),
			expectedDecodedVal: client.NewNormalString("str"),
		},
	}

	for _, tt := range tests {
		for _, descending := range []bool{false, true} {
			label := " (ascending)"
			if descending {
				label = " (descending)"
			}
			t.Run(tt.name+label, func(t *testing.T) {
				encoded := EncodeFieldValue(nil, tt.inputVal, descending)
				expectedBytes := tt.expectedBytes
				if descending {
					expectedBytes = tt.expectedBytesDesc
				}
				if !reflect.DeepEqual(encoded, expectedBytes) {
					t.Errorf("EncodeFieldValue() = %v, want %v", encoded, expectedBytes)
				}

				_, decodedFieldVal, err := DecodeFieldValue(encoded, descending, client.FieldKind_NILLABLE_INT)
				assert.NoError(t, err)
				if !reflect.DeepEqual(decodedFieldVal, tt.expectedDecodedVal) {
					t.Errorf("DecodeFieldValue() = %v, want %v", decodedFieldVal, tt.expectedDecodedVal)
				}
			})
		}
	}
}

func TestDecodeInvalidFieldValue(t *testing.T) {
	tests := []struct {
		name           string
		inputBytes     []byte
		inputBytesDesc []byte
	}{
		{
			name:           "invalid int value",
			inputBytes:     []byte{IntMax, 2},
			inputBytesDesc: []byte{^byte(IntMax), 2},
		},
		{
			name:           "invalid float value",
			inputBytes:     []byte{floatPos, 2},
			inputBytesDesc: []byte{floatPos, 2},
		},
		{
			name:           "invalid bytes value",
			inputBytes:     []byte{bytesMarker, 2},
			inputBytesDesc: []byte{bytesMarker, 2},
		},
		{
			name:           "invalid data",
			inputBytes:     []byte{IntMin - 1, 2},
			inputBytesDesc: []byte{^byte(IntMin - 1), 2},
		},
	}

	for _, tt := range tests {
		for _, descending := range []bool{false, true} {
			label := " (ascending)"
			if descending {
				label = " (descending)"
			}
			t.Run(tt.name+label, func(t *testing.T) {
				inputBytes := tt.inputBytes
				if descending {
					inputBytes = tt.inputBytesDesc
				}
				_, _, err := DecodeFieldValue(inputBytes, descending, client.FieldKind_NILLABLE_INT)
				assert.ErrorIs(t, err, ErrCanNotDecodeFieldValue)
			})
		}
	}
}
