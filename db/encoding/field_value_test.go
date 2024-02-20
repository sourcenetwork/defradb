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

	"github.com/sourcenetwork/defradb/client"
	"github.com/stretchr/testify/assert"
)

func TestEncodeDecodeFieldValue(t *testing.T) {
	tests := []struct {
		name      string
		inputVal  *client.FieldValue
		wantBytes []byte
		wantErr   bool
	}{
		{
			name:      "nil bool",
			inputVal:  client.NewFieldValue(client.NONE_CRDT, nil, client.FieldKind_NILLABLE_BOOL),
			wantBytes: EncodeNullAscending(nil),
			wantErr:   false,
		},
		{
			name:      "nil int",
			inputVal:  client.NewFieldValue(client.NONE_CRDT, nil, client.FieldKind_NILLABLE_INT),
			wantBytes: EncodeNullAscending(nil),
			wantErr:   false,
		},
		{
			name:      "nil float",
			inputVal:  client.NewFieldValue(client.NONE_CRDT, nil, client.FieldKind_NILLABLE_FLOAT),
			wantBytes: EncodeNullAscending(nil),
		},
		{
			name:      "nil string",
			inputVal:  client.NewFieldValue(client.NONE_CRDT, nil, client.FieldKind_NILLABLE_STRING),
			wantBytes: EncodeNullAscending(nil),
		},
		{
			name:     "invalid bool",
			inputVal: client.NewFieldValue(client.NONE_CRDT, "str", client.FieldKind_NILLABLE_BOOL),
			wantErr:  true,
		},
		{
			name:     "invalid int",
			inputVal: client.NewFieldValue(client.NONE_CRDT, "str", client.FieldKind_NILLABLE_INT),
			wantErr:  true,
		},
		{
			name:     "invalid float",
			inputVal: client.NewFieldValue(client.NONE_CRDT, "str", client.FieldKind_NILLABLE_FLOAT),
			wantErr:  true,
		},
		{
			name:     "invalid string",
			inputVal: client.NewFieldValue(client.NONE_CRDT, 666, client.FieldKind_NILLABLE_STRING),
			wantErr:  true,
		},
		{
			name:     "invalid docID",
			inputVal: client.NewFieldValue(client.NONE_CRDT, 666, client.FieldKind_DocID),
			wantErr:  true,
		},
		{
			name:      "bool true",
			inputVal:  client.NewFieldValue(client.NONE_CRDT, true, client.FieldKind_NILLABLE_BOOL),
			wantBytes: EncodeVarintAscending(nil, 1),
		},
		{
			name:      "bool false",
			inputVal:  client.NewFieldValue(client.NONE_CRDT, false, client.FieldKind_NILLABLE_BOOL),
			wantBytes: EncodeVarintAscending(nil, 0),
		},
		{
			name:      "int",
			inputVal:  client.NewFieldValue(client.NONE_CRDT, int64(55), client.FieldKind_NILLABLE_INT),
			wantBytes: EncodeVarintAscending(nil, 55),
		},
		{
			name:      "float",
			inputVal:  client.NewFieldValue(client.NONE_CRDT, 0.2, client.FieldKind_NILLABLE_FLOAT),
			wantBytes: EncodeFloatAscending(nil, 0.2),
		},
		{
			name:      "string",
			inputVal:  client.NewFieldValue(client.NONE_CRDT, "str", client.FieldKind_NILLABLE_STRING),
			wantBytes: EncodeBytesAscending(nil, []byte("str")),
		},
		{
			name:      "docID",
			inputVal:  client.NewFieldValue(client.NONE_CRDT, "str", client.FieldKind_DocID),
			wantBytes: EncodeBytesAscending(nil, []byte("str")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := EncodeFieldValue(nil, tt.inputVal)
			if tt.wantErr {
				if err == nil {
					t.Errorf("EncodeFieldValue() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if !reflect.DeepEqual(encoded, tt.wantBytes) {
				t.Errorf("EncodeFieldValue() = %v, want %v", encoded, tt.wantBytes)
			}

			_, decodedVal, err := DecodeFieldValue(encoded, tt.inputVal.Kind())
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeFieldValue() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(decodedVal, tt.inputVal) {
				t.Errorf("DecodeFieldValue() = %v, want %v", decodedVal, tt.inputVal)
			}
		})
	}
}

func TestDecodeInvalidFieldValue(t *testing.T) {
	tests := []struct {
		name       string
		kind       client.FieldKind
		inputBytes []byte
	}{
		{
			name:       "bool > 1",
			inputBytes: EncodeUvarintAscending(nil, 2),
			kind:       client.FieldKind_NILLABLE_BOOL,
		},
		{
			name:       "bool < 0",
			inputBytes: EncodeVarintAscending(nil, -1),
			kind:       client.FieldKind_NILLABLE_BOOL,
		},
		{
			name:       "wrong kind for bytes value",
			inputBytes: EncodeBytesAscending(nil, []byte{1, 2, 3}),
			kind:       client.FieldKind_NILLABLE_INT,
		},
		{
			name:       "wrong kind for int value",
			inputBytes: EncodeUvarintAscending(nil, 3),
			kind:       client.FieldKind_NILLABLE_FLOAT,
		},
		{
			name:       "wrong kind for float value",
			inputBytes: EncodeFloatAscending(nil, 0.2),
			kind:       client.FieldKind_NILLABLE_INT,
		},
		{
			name:       "invalid int value",
			inputBytes: []byte{IntMax, 2},
			kind:       client.FieldKind_NILLABLE_INT,
		},
		{
			name:       "invalid float value",
			inputBytes: []byte{floatPos, 2},
			kind:       client.FieldKind_NILLABLE_FLOAT,
		},
		{
			name:       "invalid bytes value",
			inputBytes: []byte{bytesMarker, 2},
			kind:       client.FieldKind_NILLABLE_STRING,
		},
		{
			name:       "nil value for not-nillable kind",
			inputBytes: EncodeNullAscending(nil),
			kind:       client.FieldKind_DocID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := DecodeFieldValue(tt.inputBytes, tt.kind)
			assert.ErrorIs(t, err, ErrCanNotDecodeFieldValue)
		})
	}
}
