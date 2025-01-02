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
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
)

func TestEncodeDecodeFieldValue(t *testing.T) {
	normalNil, err := client.NewNormalNil(client.FieldKind_NILLABLE_INT)
	require.NoError(t, err)

	// Create test JSON values
	simpleJSON, err := client.NewJSON("simple string")
	require.NoError(t, err)
	normalSimpleJSON := client.NewNormalJSON(simpleJSON)

	numberJSON, err := client.NewJSON(42.5)
	require.NoError(t, err)
	normalNumberJSON := client.NewNormalJSON(numberJSON)

	boolJSON, err := client.NewJSON(true)
	require.NoError(t, err)
	normalBoolJSON := client.NewNormalJSON(boolJSON)

	nullJSON, err := client.NewJSON(nil)
	require.NoError(t, err)
	normalNullJSON := client.NewNormalJSON(nullJSON)

	date := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

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
			name:               "nillable bool true",
			inputVal:           client.NewNormalNillableBool(immutable.Some(true)),
			expectedBytes:      EncodeBoolAscending(nil, true),
			expectedBytesDesc:  EncodeBoolDescending(nil, true),
			expectedDecodedVal: client.NewNormalBool(true),
		},
		{
			name:               "nillable bool false",
			inputVal:           client.NewNormalNillableBool(immutable.Some(false)),
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
			name:               "nillable int",
			inputVal:           client.NewNormalNillableInt(immutable.Some(55)),
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
			name:               "nillable float",
			inputVal:           client.NewNormalNillableFloat(immutable.Some(0.2)),
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
		{
			name:               "nillable string",
			inputVal:           client.NewNormalNillableString(immutable.Some("str")),
			expectedBytes:      EncodeBytesAscending(nil, []byte("str")),
			expectedBytesDesc:  EncodeBytesDescending(nil, []byte("str")),
			expectedDecodedVal: client.NewNormalString("str"),
		},
		{
			name:               "time",
			inputVal:           client.NewNormalTime(date),
			expectedBytes:      EncodeTimeAscending(nil, date),
			expectedBytesDesc:  EncodeTimeDescending(nil, date),
			expectedDecodedVal: client.NewNormalTime(date),
		},
		{
			name:               "nillable time",
			inputVal:           client.NewNormalNillableTime(immutable.Some(date)),
			expectedBytes:      EncodeTimeAscending(nil, date),
			expectedBytesDesc:  EncodeTimeDescending(nil, date),
			expectedDecodedVal: client.NewNormalTime(date),
		},
		{
			name:               "json string",
			inputVal:           normalSimpleJSON,
			expectedBytes:      EncodeJSONAscending(nil, simpleJSON),
			expectedBytesDesc:  EncodeJSONDescending(nil, simpleJSON),
			expectedDecodedVal: normalSimpleJSON,
		},
		{
			name:               "json number",
			inputVal:           normalNumberJSON,
			expectedBytes:      EncodeJSONAscending(nil, numberJSON),
			expectedBytesDesc:  EncodeJSONDescending(nil, numberJSON),
			expectedDecodedVal: normalNumberJSON,
		},
		{
			name:               "json bool",
			inputVal:           normalBoolJSON,
			expectedBytes:      EncodeJSONAscending(nil, boolJSON),
			expectedBytesDesc:  EncodeJSONDescending(nil, boolJSON),
			expectedDecodedVal: normalBoolJSON,
		},
		{
			name:               "json null",
			inputVal:           normalNullJSON,
			expectedBytes:      EncodeJSONAscending(nil, nullJSON),
			expectedBytesDesc:  EncodeJSONDescending(nil, nullJSON),
			expectedDecodedVal: normalNullJSON,
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
		{
			name:           "invalid json value",
			inputBytes:     []byte{jsonMarker, 0xFF},
			inputBytesDesc: []byte{jsonMarker, 0xFF},
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
