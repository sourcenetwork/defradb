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

	"github.com/sourcenetwork/defradb/client"
)

func TestEncodeDecodeFieldValue(t *testing.T) {
	tests := []struct {
		name              string
		inputVal          *client.FieldValue
		expectedBytes     []byte
		expectedBytesDesc []byte
		expectErr         bool
	}{
		{
			name:              "nil bool",
			inputVal:          client.NewFieldValue(client.NONE_CRDT, nil, client.FieldKind_NILLABLE_BOOL),
			expectedBytes:     EncodeNullAscending(nil),
			expectedBytesDesc: EncodeNullDescending(nil),
		},
		{
			name:              "nil int",
			inputVal:          client.NewFieldValue(client.NONE_CRDT, nil, client.FieldKind_NILLABLE_INT),
			expectedBytes:     EncodeNullAscending(nil),
			expectedBytesDesc: EncodeNullDescending(nil),
		},
		{
			name:              "nil float",
			inputVal:          client.NewFieldValue(client.NONE_CRDT, nil, client.FieldKind_NILLABLE_FLOAT),
			expectedBytes:     EncodeNullAscending(nil),
			expectedBytesDesc: EncodeNullDescending(nil),
		},
		{
			name:              "nil string",
			inputVal:          client.NewFieldValue(client.NONE_CRDT, nil, client.FieldKind_NILLABLE_STRING),
			expectedBytes:     EncodeNullAscending(nil),
			expectedBytesDesc: EncodeNullDescending(nil),
		},
		{
			name:      "invalid bool",
			inputVal:  client.NewFieldValue(client.NONE_CRDT, "str", client.FieldKind_NILLABLE_BOOL),
			expectErr: true,
		},
		{
			name:      "invalid int",
			inputVal:  client.NewFieldValue(client.NONE_CRDT, "str", client.FieldKind_NILLABLE_INT),
			expectErr: true,
		},
		{
			name:      "invalid float",
			inputVal:  client.NewFieldValue(client.NONE_CRDT, "str", client.FieldKind_NILLABLE_FLOAT),
			expectErr: true,
		},
		{
			name:      "invalid string",
			inputVal:  client.NewFieldValue(client.NONE_CRDT, 666, client.FieldKind_NILLABLE_STRING),
			expectErr: true,
		},
		{
			name:      "invalid docID",
			inputVal:  client.NewFieldValue(client.NONE_CRDT, 666, client.FieldKind_DocID),
			expectErr: true,
		},
		{
			name:              "bool true",
			inputVal:          client.NewFieldValue(client.NONE_CRDT, true, client.FieldKind_NILLABLE_BOOL),
			expectedBytes:     EncodeVarintAscending(nil, 1),
			expectedBytesDesc: EncodeVarintDescending(nil, 1),
		},
		{
			name:              "bool false",
			inputVal:          client.NewFieldValue(client.NONE_CRDT, false, client.FieldKind_NILLABLE_BOOL),
			expectedBytes:     EncodeVarintAscending(nil, 0),
			expectedBytesDesc: EncodeVarintDescending(nil, 0),
		},
		{
			name:              "int",
			inputVal:          client.NewFieldValue(client.NONE_CRDT, int64(55), client.FieldKind_NILLABLE_INT),
			expectedBytes:     EncodeVarintAscending(nil, 55),
			expectedBytesDesc: EncodeVarintDescending(nil, 55),
		},
		{
			name:              "float",
			inputVal:          client.NewFieldValue(client.NONE_CRDT, 0.2, client.FieldKind_NILLABLE_FLOAT),
			expectedBytes:     EncodeFloatAscending(nil, 0.2),
			expectedBytesDesc: EncodeFloatDescending(nil, 0.2),
		},
		{
			name:              "string",
			inputVal:          client.NewFieldValue(client.NONE_CRDT, "str", client.FieldKind_NILLABLE_STRING),
			expectedBytes:     EncodeBytesAscending(nil, []byte("str")),
			expectedBytesDesc: EncodeBytesDescending(nil, []byte("str")),
		},
		{
			name:              "docID",
			inputVal:          client.NewFieldValue(client.NONE_CRDT, "str", client.FieldKind_DocID),
			expectedBytes:     EncodeBytesAscending(nil, []byte("str")),
			expectedBytesDesc: EncodeBytesDescending(nil, []byte("str")),
		},
	}

	for _, tt := range tests {
		for _, descending := range []bool{false, true} {
			label := " (ascending)"
			if descending {
				label = " (descending)"
			}
			t.Run(tt.name+label, func(t *testing.T) {
				encoded, err := EncodeFieldValue(nil, tt.inputVal, descending)
				if tt.expectErr {
					if err == nil {
						t.Errorf("EncodeFieldValue() error = %v, wantErr %v", err, tt.expectErr)
					}
					return
				}
				expectedBytes := tt.expectedBytes
				if descending {
					expectedBytes = tt.expectedBytesDesc
				}
				if !reflect.DeepEqual(encoded, expectedBytes) {
					t.Errorf("EncodeFieldValue() = %v, want %v", encoded, expectedBytes)
				}

				_, decodedFieldVal, err := DecodeFieldValue(encoded, tt.inputVal.Kind(), descending)
				if (err != nil) != tt.expectErr {
					t.Errorf("DecodeFieldValue() error = %v, wantErr %v", err, tt.expectErr)
				}
				if !reflect.DeepEqual(decodedFieldVal, tt.inputVal) {
					t.Errorf("DecodeFieldValue() = %v, want %v", decodedFieldVal, tt.inputVal)
				}
			})
		}
	}
}

func TestDecodeInvalidFieldValue(t *testing.T) {
	tests := []struct {
		name           string
		kind           client.FieldKind
		inputBytes     []byte
		inputBytesDesc []byte
	}{
		{
			name:           "bool > 1",
			inputBytes:     EncodeUvarintAscending(nil, 2),
			inputBytesDesc: EncodeUvarintDescending(nil, 2),
			kind:           client.FieldKind_NILLABLE_BOOL,
		},
		{
			name:           "bool < 0",
			inputBytes:     EncodeVarintAscending(nil, -1),
			inputBytesDesc: EncodeVarintDescending(nil, -1),
			kind:           client.FieldKind_NILLABLE_BOOL,
		},
		{
			name:           "wrong kind for bytes value",
			inputBytes:     EncodeBytesAscending(nil, []byte{1, 2, 3}),
			inputBytesDesc: EncodeBytesDescending(nil, []byte{1, 2, 3}),
			kind:           client.FieldKind_NILLABLE_INT,
		},
		{
			name:           "wrong kind for int value",
			inputBytes:     EncodeUvarintAscending(nil, 3),
			inputBytesDesc: EncodeUvarintDescending(nil, 3),
			kind:           client.FieldKind_NILLABLE_FLOAT,
		},
		{
			name:           "wrong kind for float value",
			inputBytes:     EncodeFloatAscending(nil, 0.2),
			inputBytesDesc: EncodeFloatDescending(nil, 0.2),
			kind:           client.FieldKind_NILLABLE_INT,
		},
		{
			name:           "invalid int value",
			inputBytes:     []byte{IntMax, 2},
			inputBytesDesc: []byte{IntMax, 2},
			kind:           client.FieldKind_NILLABLE_INT,
		},
		{
			name:           "invalid float value",
			inputBytes:     []byte{floatPos, 2},
			inputBytesDesc: []byte{floatPos, 2},
			kind:           client.FieldKind_NILLABLE_FLOAT,
		},
		{
			name:           "invalid bytes value",
			inputBytes:     []byte{bytesMarker, 2},
			inputBytesDesc: []byte{bytesMarker, 2},
			kind:           client.FieldKind_NILLABLE_STRING,
		},
		{
			name:           "nil value for not-nillable kind",
			inputBytes:     EncodeNullAscending(nil),
			inputBytesDesc: EncodeNullDescending(nil),
			kind:           client.FieldKind_DocID,
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
				_, _, err := DecodeFieldValue(inputBytes, tt.kind, descending)
				assert.ErrorIs(t, err, ErrCanNotDecodeFieldValue)
			})
		}
	}
}
