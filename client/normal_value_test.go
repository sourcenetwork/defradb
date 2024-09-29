// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import (
	"reflect"
	"testing"
	"time"

	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type nType string

const (
	BoolType     nType = "Bool"
	IntType      nType = "Int"
	FloatType    nType = "Float"
	StringType   nType = "String"
	BytesType    nType = "Bytes"
	TimeType     nType = "Time"
	DocumentType nType = "Document"

	NillableBoolType     nType = "NillableBool"
	NillableIntType      nType = "NillableInt"
	NillableFloatType    nType = "NillableFloat"
	NillableStringType   nType = "NillableString"
	NillableBytesType    nType = "NillableBytes"
	NillableTimeType     nType = "NillableTime"
	NillableDocumentType nType = "NillableDocument"

	BoolArray     nType = "BoolArray"
	IntArray      nType = "IntArray"
	FloatArray    nType = "FloatArray"
	StringArray   nType = "StringArray"
	BytesArray    nType = "BytesArray"
	TimeArray     nType = "TimeArray"
	DocumentArray nType = "DocumentArray"

	NillableBoolArray     nType = "NillableBoolArray"
	NillableIntArray      nType = "NillableIntArray"
	NillableFloatArray    nType = "NillableFloatArray"
	NillableStringArray   nType = "NillableStringArray"
	NillableBytesArray    nType = "NillableBytesArray"
	NillableTimeArray     nType = "NillableTimeArray"
	NillableDocumentArray nType = "NillableDocumentArray"

	BoolNillableArray     nType = "BoolNillableArray"
	IntNillableArray      nType = "IntNillableArray"
	FloatNillableArray    nType = "FloatNillableArray"
	StringNillableArray   nType = "StringNillableArray"
	BytesNillableArray    nType = "BytesNillableArray"
	TimeNillableArray     nType = "TimeNillableArray"
	DocumentNillableArray nType = "DocumentNillableArray"

	NillableBoolNillableArray     nType = "NillableBoolNillableArray"
	NillableIntNillableArray      nType = "NillableIntNillableArray"
	NillableFloatNillableArray    nType = "NillableFloatNillableArray"
	NillableStringNillableArray   nType = "NillableStringNillableArray"
	NillableBytesNillableArray    nType = "NillableBytesNillableArray"
	NillableTimeNillableArray     nType = "NillableTimeNillableArray"
	NillableDocumentNillableArray nType = "NillableDocumentNillableArray"
)

// extractValue takes an input of type `any` and checks if it is an `Option[T]`.
// If it is and contains a value, it returns the contained value.
// Otherwise, it returns the input itself.
func extractValue(input any) any {
	inputVal := reflect.ValueOf(input)

	// Check if the type is Option[T] by seeing if it has the HasValue and Value methods.
	hasValueMethod := inputVal.MethodByName("HasValue")
	valueMethod := inputVal.MethodByName("Value")

	if hasValueMethod.IsValid() && valueMethod.IsValid() {
		// Call HasValue to check if there's a value.
		hasValueResult := hasValueMethod.Call(nil)
		if len(hasValueResult) == 1 {
			if hasValueResult[0].Bool() {
				// Call Value to get the actual value if HasValue is true.
				valueResult := valueMethod.Call(nil)
				if len(valueResult) == 1 {
					return valueResult[0].Interface()
				}
			} else {
				// Return nil if HasValue is false.
				return nil
			}
		}
	}

	// Return the input itself if it's not an Option[T] with a value.
	return input
}

func TestNormalValue_NewValueAndTypeAssertion(t *testing.T) {
	typeAssertMap := map[nType]func(NormalValue) (any, bool){
		BoolType:     func(v NormalValue) (any, bool) { return v.Bool() },
		IntType:      func(v NormalValue) (any, bool) { return v.Int() },
		FloatType:    func(v NormalValue) (any, bool) { return v.Float() },
		StringType:   func(v NormalValue) (any, bool) { return v.String() },
		BytesType:    func(v NormalValue) (any, bool) { return v.Bytes() },
		TimeType:     func(v NormalValue) (any, bool) { return v.Time() },
		DocumentType: func(v NormalValue) (any, bool) { return v.Document() },

		NillableBoolType:     func(v NormalValue) (any, bool) { return v.NillableBool() },
		NillableIntType:      func(v NormalValue) (any, bool) { return v.NillableInt() },
		NillableFloatType:    func(v NormalValue) (any, bool) { return v.NillableFloat() },
		NillableStringType:   func(v NormalValue) (any, bool) { return v.NillableString() },
		NillableBytesType:    func(v NormalValue) (any, bool) { return v.NillableBytes() },
		NillableTimeType:     func(v NormalValue) (any, bool) { return v.NillableTime() },
		NillableDocumentType: func(v NormalValue) (any, bool) { return v.NillableDocument() },

		BoolArray:     func(v NormalValue) (any, bool) { return v.BoolArray() },
		IntArray:      func(v NormalValue) (any, bool) { return v.IntArray() },
		FloatArray:    func(v NormalValue) (any, bool) { return v.FloatArray() },
		StringArray:   func(v NormalValue) (any, bool) { return v.StringArray() },
		BytesArray:    func(v NormalValue) (any, bool) { return v.BytesArray() },
		TimeArray:     func(v NormalValue) (any, bool) { return v.TimeArray() },
		DocumentArray: func(v NormalValue) (any, bool) { return v.DocumentArray() },

		BoolNillableArray:     func(v NormalValue) (any, bool) { return v.BoolNillableArray() },
		IntNillableArray:      func(v NormalValue) (any, bool) { return v.IntNillableArray() },
		FloatNillableArray:    func(v NormalValue) (any, bool) { return v.FloatNillableArray() },
		StringNillableArray:   func(v NormalValue) (any, bool) { return v.StringNillableArray() },
		BytesNillableArray:    func(v NormalValue) (any, bool) { return v.BytesNillableArray() },
		TimeNillableArray:     func(v NormalValue) (any, bool) { return v.TimeNillableArray() },
		DocumentNillableArray: func(v NormalValue) (any, bool) { return v.DocumentNillableArray() },

		NillableBoolArray:     func(v NormalValue) (any, bool) { return v.NillableBoolArray() },
		NillableIntArray:      func(v NormalValue) (any, bool) { return v.NillableIntArray() },
		NillableFloatArray:    func(v NormalValue) (any, bool) { return v.NillableFloatArray() },
		NillableStringArray:   func(v NormalValue) (any, bool) { return v.NillableStringArray() },
		NillableBytesArray:    func(v NormalValue) (any, bool) { return v.NillableBytesArray() },
		NillableTimeArray:     func(v NormalValue) (any, bool) { return v.NillableTimeArray() },
		NillableDocumentArray: func(v NormalValue) (any, bool) { return v.NillableDocumentArray() },

		NillableBoolNillableArray:   func(v NormalValue) (any, bool) { return v.NillableBoolNillableArray() },
		NillableIntNillableArray:    func(v NormalValue) (any, bool) { return v.NillableIntNillableArray() },
		NillableFloatNillableArray:  func(v NormalValue) (any, bool) { return v.NillableFloatNillableArray() },
		NillableStringNillableArray: func(v NormalValue) (any, bool) { return v.NillableStringNillableArray() },
		NillableBytesNillableArray:  func(v NormalValue) (any, bool) { return v.NillableBytesNillableArray() },
		NillableTimeNillableArray:   func(v NormalValue) (any, bool) { return v.NillableTimeNillableArray() },
		NillableDocumentNillableArray: func(v NormalValue) (any, bool) {
			return v.NillableDocumentNillableArray()
		},
	}

	newMap := map[nType]func(any) NormalValue{
		BoolType:     func(v any) NormalValue { return NewNormalBool(v.(bool)) },
		IntType:      func(v any) NormalValue { return NewNormalInt(v.(int64)) },
		FloatType:    func(v any) NormalValue { return NewNormalFloat(v.(float64)) },
		StringType:   func(v any) NormalValue { return NewNormalString(v.(string)) },
		BytesType:    func(v any) NormalValue { return NewNormalBytes(v.([]byte)) },
		TimeType:     func(v any) NormalValue { return NewNormalTime(v.(time.Time)) },
		DocumentType: func(v any) NormalValue { return NewNormalDocument(v.(*Document)) },

		NillableBoolType:     func(v any) NormalValue { return NewNormalNillableBool(v.(immutable.Option[bool])) },
		NillableIntType:      func(v any) NormalValue { return NewNormalNillableInt(v.(immutable.Option[int64])) },
		NillableFloatType:    func(v any) NormalValue { return NewNormalNillableFloat(v.(immutable.Option[float64])) },
		NillableStringType:   func(v any) NormalValue { return NewNormalNillableString(v.(immutable.Option[string])) },
		NillableBytesType:    func(v any) NormalValue { return NewNormalNillableBytes(v.(immutable.Option[[]byte])) },
		NillableTimeType:     func(v any) NormalValue { return NewNormalNillableTime(v.(immutable.Option[time.Time])) },
		NillableDocumentType: func(v any) NormalValue { return NewNormalNillableDocument(v.(immutable.Option[*Document])) },

		BoolArray:     func(v any) NormalValue { return NewNormalBoolArray(v.([]bool)) },
		IntArray:      func(v any) NormalValue { return NewNormalIntArray(v.([]int64)) },
		FloatArray:    func(v any) NormalValue { return NewNormalFloatArray(v.([]float64)) },
		StringArray:   func(v any) NormalValue { return NewNormalStringArray(v.([]string)) },
		BytesArray:    func(v any) NormalValue { return NewNormalBytesArray(v.([][]byte)) },
		TimeArray:     func(v any) NormalValue { return NewNormalTimeArray(v.([]time.Time)) },
		DocumentArray: func(v any) NormalValue { return NewNormalDocumentArray(v.([]*Document)) },

		NillableBoolArray: func(v any) NormalValue {
			return NewNormalNillableBoolArray(v.([]immutable.Option[bool]))
		},
		NillableIntArray: func(v any) NormalValue {
			return NewNormalNillableIntArray(v.([]immutable.Option[int64]))
		},
		NillableFloatArray: func(v any) NormalValue {
			return NewNormalNillableFloatArray(v.([]immutable.Option[float64]))
		},
		NillableStringArray: func(v any) NormalValue {
			return NewNormalNillableStringArray(v.([]immutable.Option[string]))
		},
		NillableBytesArray: func(v any) NormalValue {
			return NewNormalNillableBytesArray(v.([]immutable.Option[[]byte]))
		},
		NillableTimeArray: func(v any) NormalValue {
			return NewNormalNillableTimeArray(v.([]immutable.Option[time.Time]))
		},
		NillableDocumentArray: func(v any) NormalValue {
			return NewNormalNillableDocumentArray(v.([]immutable.Option[*Document]))
		},

		BoolNillableArray: func(v any) NormalValue {
			return NewNormalBoolNillableArray(v.(immutable.Option[[]bool]))
		},
		IntNillableArray: func(v any) NormalValue {
			return NewNormalIntNillableArray(v.(immutable.Option[[]int64]))
		},
		FloatNillableArray: func(v any) NormalValue {
			return NewNormalFloatNillableArray(v.(immutable.Option[[]float64]))
		},
		StringNillableArray: func(v any) NormalValue {
			return NewNormalStringNillableArray(v.(immutable.Option[[]string]))
		},
		BytesNillableArray: func(v any) NormalValue {
			return NewNormalBytesNillableArray(v.(immutable.Option[[][]byte]))
		},
		TimeNillableArray: func(v any) NormalValue {
			return NewNormalTimeNillableArray(v.(immutable.Option[[]time.Time]))
		},
		DocumentNillableArray: func(v any) NormalValue {
			return NewNormalDocumentNillableArray(v.(immutable.Option[[]*Document]))
		},

		NillableBoolNillableArray: func(v any) NormalValue {
			return NewNormalNillableBoolNillableArray(v.(immutable.Option[[]immutable.Option[bool]]))
		},
		NillableIntNillableArray: func(v any) NormalValue {
			return NewNormalNillableIntNillableArray(v.(immutable.Option[[]immutable.Option[int64]]))
		},
		NillableFloatNillableArray: func(v any) NormalValue {
			return NewNormalNillableFloatNillableArray(v.(immutable.Option[[]immutable.Option[float64]]))
		},
		NillableStringNillableArray: func(v any) NormalValue {
			return NewNormalNillableStringNillableArray(v.(immutable.Option[[]immutable.Option[string]]))
		},
		NillableBytesNillableArray: func(v any) NormalValue {
			return NewNormalNillableBytesNillableArray(v.(immutable.Option[[]immutable.Option[[]byte]]))
		},
		NillableTimeNillableArray: func(v any) NormalValue {
			return NewNormalNillableTimeNillableArray(v.(immutable.Option[[]immutable.Option[time.Time]]))
		},
		NillableDocumentNillableArray: func(v any) NormalValue {
			return NewNormalNillableDocumentNillableArray(v.(immutable.Option[[]immutable.Option[*Document]]))
		},
	}

	tests := []struct {
		nType      nType
		input      any
		isNillable bool
		isNil      bool
		isArray    bool
	}{
		{
			nType: BoolType,
			input: true,
		},
		{
			nType: IntType,
			input: int64(1),
		},
		{
			nType: FloatType,
			input: float64(1),
		},
		{
			nType: StringType,
			input: "test",
		},
		{
			nType: BytesType,
			input: []byte{1, 2, 3},
		},
		{
			nType: TimeType,
			input: time.Now(),
		},
		{
			nType: DocumentType,
			input: &Document{},
		},
		{
			nType:      NillableBoolType,
			input:      immutable.Some(true),
			isNillable: true,
		},
		{
			nType:      NillableBoolType,
			input:      immutable.None[bool](),
			isNil:      true,
			isNillable: true,
		},
		{
			nType:      NillableIntType,
			input:      immutable.Some(int64(1)),
			isNillable: true,
		},
		{
			nType:      NillableIntType,
			input:      immutable.None[int64](),
			isNil:      true,
			isNillable: true,
		},
		{
			nType:      NillableFloatType,
			input:      immutable.Some(float64(1)),
			isNillable: true,
		},
		{
			nType:      NillableFloatType,
			input:      immutable.None[float64](),
			isNil:      true,
			isNillable: true,
		},
		{
			nType:      NillableStringType,
			input:      immutable.Some("test"),
			isNillable: true,
		},
		{
			nType:      NillableStringType,
			input:      immutable.None[string](),
			isNil:      true,
			isNillable: true,
		},
		{
			nType:      NillableBytesType,
			input:      immutable.Some([]byte{1, 2, 3}),
			isNillable: true,
		},
		{
			nType:      NillableBytesType,
			input:      immutable.None[[]byte](),
			isNil:      true,
			isNillable: true,
		},
		{
			nType:      NillableTimeType,
			input:      immutable.Some(time.Now()),
			isNillable: true,
		},
		{
			nType:      NillableTimeType,
			input:      immutable.None[time.Time](),
			isNil:      true,
			isNillable: true,
		},
		{
			nType:      NillableDocumentType,
			input:      immutable.Some(&Document{}),
			isNillable: true,
		},
		{
			nType:      NillableDocumentType,
			input:      immutable.None[*Document](),
			isNil:      true,
			isNillable: true,
		},
		{
			nType:   BoolArray,
			input:   []bool{true, false},
			isArray: true,
		},
		{
			nType:   IntArray,
			input:   []int64{1, 2, 3},
			isArray: true,
		},
		{
			nType:   FloatArray,
			input:   []float64{1, 2, 3},
			isArray: true,
		},
		{
			nType:   StringArray,
			input:   []string{"test", "test2"},
			isArray: true,
		},
		{
			nType:   BytesArray,
			input:   [][]byte{{1, 2, 3}, {4, 5, 6}},
			isArray: true,
		},
		{
			nType:   TimeArray,
			input:   []time.Time{time.Now(), time.Now()},
			isArray: true,
		},
		{
			nType:   DocumentArray,
			input:   []*Document{{}, {}},
			isArray: true,
		},
		{
			nType:   NillableBoolArray,
			input:   []immutable.Option[bool]{immutable.Some(true)},
			isArray: true,
		},
		{
			nType:   NillableIntArray,
			input:   []immutable.Option[int64]{immutable.Some(int64(1))},
			isArray: true,
		},
		{
			nType:   NillableFloatArray,
			input:   []immutable.Option[float64]{immutable.Some(float64(1))},
			isArray: true,
		},
		{
			nType:   NillableStringArray,
			input:   []immutable.Option[string]{immutable.Some("test")},
			isArray: true,
		},
		{
			nType:   NillableBytesArray,
			input:   []immutable.Option[[]byte]{immutable.Some([]byte{1, 2, 3})},
			isArray: true,
		},
		{
			nType:   NillableTimeArray,
			input:   []immutable.Option[time.Time]{immutable.Some(time.Now())},
			isArray: true,
		},
		{
			nType:   NillableDocumentArray,
			input:   []immutable.Option[*Document]{immutable.Some(&Document{})},
			isArray: true,
		},
		{
			nType:      BoolNillableArray,
			input:      immutable.Some([]bool{true, false}),
			isNillable: true,
			isArray:    true,
		},
		{
			nType:      BoolNillableArray,
			input:      immutable.None[[]bool](),
			isNillable: true,
			isNil:      true,
			isArray:    true,
		},
		{
			nType:      IntNillableArray,
			input:      immutable.Some([]int64{1, 2, 3}),
			isNillable: true,
			isArray:    true,
		},
		{
			nType:      IntNillableArray,
			input:      immutable.None[[]int64](),
			isNillable: true,
			isNil:      true,
			isArray:    true,
		},
		{
			nType:      FloatNillableArray,
			input:      immutable.Some([]float64{1, 2, 3}),
			isNillable: true,
			isArray:    true,
		},
		{
			nType:      FloatNillableArray,
			input:      immutable.None[[]float64](),
			isNillable: true,
			isNil:      true,
			isArray:    true,
		},
		{
			nType:      StringNillableArray,
			input:      immutable.Some([]string{"test", "test2"}),
			isNillable: true,
			isArray:    true,
		},
		{
			nType:      StringNillableArray,
			input:      immutable.None[[]string](),
			isNillable: true,
			isNil:      true,
			isArray:    true,
		},
		{
			nType:      BytesNillableArray,
			input:      immutable.Some([][]byte{{1, 2, 3}, {4, 5, 6}}),
			isNillable: true,
			isArray:    true,
		},
		{
			nType:      BytesNillableArray,
			input:      immutable.None[[][]byte](),
			isNillable: true,
			isNil:      true,
			isArray:    true,
		},
		{
			nType:      TimeNillableArray,
			input:      immutable.Some([]time.Time{time.Now(), time.Now()}),
			isNillable: true,
			isArray:    true,
		},
		{
			nType:      TimeNillableArray,
			input:      immutable.None[[]time.Time](),
			isNillable: true,
			isNil:      true,
			isArray:    true,
		},
		{
			nType:      DocumentNillableArray,
			input:      immutable.Some([]*Document{{}, {}}),
			isNillable: true,
			isArray:    true,
		},
		{
			nType:      DocumentNillableArray,
			input:      immutable.None[[]*Document](),
			isNillable: true,
			isNil:      true,
			isArray:    true,
		},
		{
			nType:      NillableBoolNillableArray,
			input:      immutable.Some([]immutable.Option[bool]{immutable.Some(true)}),
			isNillable: true,
			isArray:    true,
		},
		{
			nType:      NillableBoolNillableArray,
			input:      immutable.None[[]immutable.Option[bool]](),
			isNillable: true,
			isNil:      true,
			isArray:    true,
		},
		{
			nType:      NillableIntNillableArray,
			input:      immutable.Some([]immutable.Option[int64]{immutable.Some(int64(1))}),
			isNillable: true,
			isArray:    true,
		},
		{
			nType:      NillableIntNillableArray,
			input:      immutable.None[[]immutable.Option[int64]](),
			isNillable: true,
			isNil:      true,
			isArray:    true,
		},
		{
			nType:      NillableFloatNillableArray,
			input:      immutable.Some([]immutable.Option[float64]{immutable.Some(float64(1))}),
			isNillable: true,
			isArray:    true,
		},
		{
			nType:      NillableFloatNillableArray,
			input:      immutable.None[[]immutable.Option[float64]](),
			isNillable: true,
			isNil:      true,
			isArray:    true,
		},
		{
			nType:      NillableStringNillableArray,
			input:      immutable.Some([]immutable.Option[string]{immutable.Some("test")}),
			isNillable: true,
			isArray:    true,
		},
		{
			nType:      NillableStringNillableArray,
			input:      immutable.None[[]immutable.Option[string]](),
			isNillable: true,
			isNil:      true,
			isArray:    true,
		},
		{
			nType:      NillableBytesNillableArray,
			input:      immutable.Some([]immutable.Option[[]byte]{immutable.Some([]byte{1, 2, 3})}),
			isNillable: true,
			isArray:    true,
		},
		{
			nType:      NillableBytesNillableArray,
			input:      immutable.None[[]immutable.Option[[]byte]](),
			isNillable: true,
			isNil:      true,
			isArray:    true,
		},
		{
			nType:      NillableTimeNillableArray,
			input:      immutable.Some([]immutable.Option[time.Time]{immutable.Some(time.Now())}),
			isNillable: true,
			isArray:    true,
		},
		{
			nType:      NillableTimeNillableArray,
			input:      immutable.None[[]immutable.Option[time.Time]](),
			isNillable: true,
			isNil:      true,
			isArray:    true,
		},
		{
			nType:      NillableDocumentNillableArray,
			input:      immutable.Some([]immutable.Option[*Document]{immutable.Some(&Document{})}),
			isNillable: true,
			isArray:    true,
		},
	}

	for _, tt := range tests {
		tStr := string(tt.nType)
		t.Run(tStr, func(t *testing.T) {
			actual, err := NewNormalValue(tt.input)
			require.NoError(t, err)

			for nType, typeAssertFunc := range typeAssertMap {
				val, ok := typeAssertFunc(actual)
				if nType == tt.nType {
					assert.True(t, ok, tStr+"() should return true")
					assert.Equal(t, tt.input, val, tStr+"() returned unexpected value")
					newVal := newMap[nType](val)
					assert.Equal(t, actual, newVal, "New"+tStr+"() returned unexpected NormalValue")
					assert.Equal(t, extractValue(tt.input), actual.Unwrap(),
						"Unwrap() returned unexpected value for "+tStr)
				} else {
					assert.False(t, ok, string(nType)+"() should return false for "+tStr)
				}
			}

			if tt.isNillable {
				assert.True(t, actual.IsNillable(), "IsNillable() should return true for "+tStr)
			} else {
				assert.False(t, actual.IsNillable(), "IsNillable() should return false for "+tStr)
			}

			if tt.isNil {
				assert.True(t, actual.IsNil(), "IsNil() should return true for "+tStr)
			} else {
				assert.False(t, actual.IsNil(), "IsNil() should return false for "+tStr)
			}

			if tt.isArray {
				assert.True(t, actual.IsArray(), "IsArray() should return true for "+tStr)
			} else {
				assert.False(t, actual.IsArray(), "IsArray() should return false for "+tStr)
			}
		})
	}
}

func TestNormalValue_InUnknownType_ReturnError(t *testing.T) {
	_, err := NewNormalValue(struct{ name string }{})
	require.ErrorContains(t, err, errCanNotNormalizeValue)
}

func TestNormalValue_NewNormalValueFromAnyArray(t *testing.T) {
	now := time.Now()
	doc1 := &Document{}
	doc2 := &Document{}

	tests := []struct {
		name     string
		input    []any
		expected NormalValue
		err      string
	}{
		{
			name:  "nil input",
			input: nil,
			err:   errCanNotNormalizeValue,
		},
		{
			name:  "unknown element type",
			input: []any{struct{ name string }{}},
			err:   errCanNotNormalizeValue,
		},
		{
			name:  "mixed elements type",
			input: []any{1, "test", true},
			err:   errCanNotNormalizeValue,
		},
		{
			name:     "bool elements",
			input:    []any{true, false},
			expected: NewNormalBoolArray([]bool{true, false}),
		},
		{
			name:     "int elements",
			input:    []any{int64(1), int64(2)},
			expected: NewNormalIntArray([]int64{1, 2}),
		},
		{
			name:     "float elements",
			input:    []any{float64(1), float64(2)},
			expected: NewNormalFloatArray([]float64{1, 2}),
		},
		{
			name:     "string elements",
			input:    []any{"test", "test2"},
			expected: NewNormalStringArray([]string{"test", "test2"}),
		},
		{
			name:     "bytes elements",
			input:    []any{[]byte{1, 2, 3}, []byte{4, 5, 6}},
			expected: NewNormalBytesArray([][]byte{{1, 2, 3}, {4, 5, 6}}),
		},
		{
			name:     "time elements",
			input:    []any{now, now},
			expected: NewNormalTimeArray([]time.Time{now, now}),
		},
		{
			name:     "document elements",
			input:    []any{doc1, doc2},
			expected: NewNormalDocumentArray([]*Document{doc1, doc2}),
		},
		{
			name:  "bool and nil elements",
			input: []any{true, nil, false},
			expected: NewNormalNillableBoolArray(
				[]immutable.Option[bool]{immutable.Some(true), immutable.None[bool](), immutable.Some(false)},
			),
		},
		{
			name:  "int and nil elements",
			input: []any{1, nil, 2},
			expected: NewNormalNillableIntArray(
				[]immutable.Option[int64]{immutable.Some(int64(1)), immutable.None[int64](), immutable.Some(int64(2))},
			),
		},
		{
			name:  "float and nil elements",
			input: []any{1.0, nil, 2.0},
			expected: NewNormalNillableFloatArray(
				[]immutable.Option[float64]{immutable.Some(1.0), immutable.None[float64](), immutable.Some(2.0)},
			),
		},
		{
			name:  "string and nil elements",
			input: []any{"test", nil, "test2"},
			expected: NewNormalNillableStringArray(
				[]immutable.Option[string]{immutable.Some("test"), immutable.None[string](), immutable.Some("test2")},
			),
		},
		{
			name:  "bytes and nil elements",
			input: []any{[]byte{1, 2, 3}, nil, []byte{4, 5, 6}},
			expected: NewNormalNillableBytesArray(
				[]immutable.Option[[]byte]{
					immutable.Some([]byte{1, 2, 3}),
					immutable.None[[]byte](),
					immutable.Some([]byte{4, 5, 6}),
				},
			),
		},
		{
			name:  "time and nil elements",
			input: []any{now, nil, now},
			expected: NewNormalNillableTimeArray(
				[]immutable.Option[time.Time]{immutable.Some(now), immutable.None[time.Time](), immutable.Some(now)},
			),
		},
		{
			name:  "document and nil elements",
			input: []any{doc1, nil, doc2},
			expected: NewNormalNillableDocumentArray(
				[]immutable.Option[*Document]{immutable.Some(doc1), immutable.None[*Document](), immutable.Some(doc2)},
			),
		},
		{
			name: "mixed int elements",
			input: []any{int8(1), int16(2), int32(3), int64(4), int(5), uint8(6), uint16(7), uint32(8),
				uint64(9), uint(10)},
			expected: NewNormalIntArray([]int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
		},
		{
			name:     "mixed float elements",
			input:    []any{float32(1.5), float64(2.2)},
			expected: NewNormalFloatArray([]float64{1.5, 2.2}),
		},
		{
			name: "mixed number elements",
			input: []any{int8(1), int16(2), int32(3), int64(4), int(5), uint8(6), uint16(7), uint32(8),
				uint64(9), uint(10), float32(1.5), float64(2.2)},
			expected: NewNormalFloatArray([]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 1.5, 2.2}),
		},
		{
			name: "mixed int and nil elements",
			input: []any{int8(1), nil, int16(2), int32(3), int64(4), int(5), uint8(6), uint16(7), uint32(8),
				uint64(9), nil, uint(10)},
			expected: NewNormalNillableIntArray(
				[]immutable.Option[int64]{immutable.Some(int64(1)), immutable.None[int64](), immutable.Some(int64(2)),
					immutable.Some(int64(3)), immutable.Some(int64(4)), immutable.Some(int64(5)), immutable.Some(int64(6)),
					immutable.Some(int64(7)), immutable.Some(int64(8)), immutable.Some(int64(9)), immutable.None[int64](),
					immutable.Some(int64(10))},
			),
		},
		{
			name:  "mixed float and nil elements",
			input: []any{float32(1.5), nil, float64(2.2)},
			expected: NewNormalNillableFloatArray(
				[]immutable.Option[float64]{immutable.Some(1.5), immutable.None[float64](), immutable.Some(2.2)},
			),
		},
		{
			name: "mixed number and nil elements",
			input: []any{int8(1), nil, int16(2), int32(3), int64(4), int(5), uint8(6), uint16(7), uint32(8),
				uint64(9), nil, uint(10), float32(1.5), nil, float64(2.2)},
			expected: NewNormalNillableFloatArray(
				[]immutable.Option[float64]{
					immutable.Some(1.0), immutable.None[float64](), immutable.Some(2.0), immutable.Some(3.0),
					immutable.Some(4.0), immutable.Some(5.0), immutable.Some(6.0), immutable.Some(7.0),
					immutable.Some(8.0), immutable.Some(9.0), immutable.None[float64](), immutable.Some(10.0),
					immutable.Some(1.5), immutable.None[float64](), immutable.Some(2.2)},
			),
		},
	}

	for _, tt := range tests {
		tStr := string(tt.name)
		t.Run(tStr, func(t *testing.T) {
			actual, err := NewNormalValue(tt.input)
			if tt.err != "" {
				require.ErrorContains(t, err, tt.err)
				return
			}

			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestNormalValue_NewNormalInt(t *testing.T) {
	i64 := int64(2)
	v := NewNormalInt(i64)
	getInt := func(v NormalValue) int64 { i, _ := v.Int(); return i }

	assert.Equal(t, i64, getInt(v))

	v = NewNormalInt(float32(2.5))
	assert.Equal(t, i64, getInt(v))

	v = NewNormalInt(float64(2.5))
	assert.Equal(t, i64, getInt(v))

	v = NewNormalInt(int8(2))
	assert.Equal(t, i64, getInt(v))

	v = NewNormalInt(int16(2))
	assert.Equal(t, i64, getInt(v))

	v = NewNormalInt(int32(2))
	assert.Equal(t, i64, getInt(v))

	v = NewNormalInt(int(2))
	assert.Equal(t, i64, getInt(v))

	v = NewNormalInt(uint8(2))
	assert.Equal(t, i64, getInt(v))

	v = NewNormalInt(uint16(2))
	assert.Equal(t, i64, getInt(v))

	v = NewNormalInt(uint32(2))
	assert.Equal(t, i64, getInt(v))

	v = NewNormalInt(uint64(2))
	assert.Equal(t, i64, getInt(v))

	v = NewNormalInt(uint(2))
	assert.Equal(t, i64, getInt(v))
}

func TestNormalValue_NewNormalFloat(t *testing.T) {
	f64Frac := float64(2.5)
	f64 := float64(2)

	getFloat := func(v NormalValue) float64 { f, _ := v.Float(); return f }

	v := NewNormalFloat(f64Frac)
	assert.Equal(t, f64Frac, getFloat(v))

	v = NewNormalFloat(float32(2.5))
	assert.Equal(t, f64Frac, getFloat(v))

	v = NewNormalFloat(int8(2))
	assert.Equal(t, f64, getFloat(v))

	v = NewNormalFloat(int16(2))
	assert.Equal(t, f64, getFloat(v))

	v = NewNormalFloat(int32(2))
	assert.Equal(t, f64, getFloat(v))

	v = NewNormalFloat(int64(2))
	assert.Equal(t, f64, getFloat(v))

	v = NewNormalFloat(int(2))
	assert.Equal(t, f64, getFloat(v))

	v = NewNormalFloat(uint8(2))
	assert.Equal(t, f64, getFloat(v))

	v = NewNormalFloat(uint16(2))
	assert.Equal(t, f64, getFloat(v))

	v = NewNormalFloat(uint32(2))
	assert.Equal(t, f64, getFloat(v))

	v = NewNormalFloat(uint64(2))
	assert.Equal(t, f64, getFloat(v))

	v = NewNormalFloat(uint(2))
	assert.Equal(t, f64, getFloat(v))
}

func TestNormalValue_NewNormalString(t *testing.T) {
	strInput := "str"

	getString := func(v NormalValue) string { s, _ := v.String(); return s }

	v := NewNormalString(strInput)
	assert.Equal(t, strInput, getString(v))

	v = NewNormalString([]byte{'s', 't', 'r'})
	assert.Equal(t, strInput, getString(v))
}

func TestNormalValue_NewNormalBytes(t *testing.T) {
	bytesInput := []byte("str")

	getBytes := func(v NormalValue) []byte { b, _ := v.Bytes(); return b }

	v := NewNormalBytes(bytesInput)
	assert.Equal(t, bytesInput, getBytes(v))

	v = NewNormalBytes("str")
	assert.Equal(t, bytesInput, getBytes(v))
}

func TestNormalValue_NewNormalIntArray(t *testing.T) {
	i64Input := []int64{2}

	getIntArray := func(v NormalValue) []int64 { i, _ := v.IntArray(); return i }

	v := NewNormalIntArray(i64Input)
	assert.Equal(t, i64Input, getIntArray(v))

	v = NewNormalIntArray([]float32{2.5})
	assert.Equal(t, i64Input, getIntArray(v))

	v = NewNormalIntArray([]int8{2})
	assert.Equal(t, i64Input, getIntArray(v))

	v = NewNormalIntArray([]int16{2})
	assert.Equal(t, i64Input, getIntArray(v))

	v = NewNormalIntArray([]int32{2})
	assert.Equal(t, i64Input, getIntArray(v))

	v = NewNormalIntArray([]int64{2})
	assert.Equal(t, i64Input, getIntArray(v))

	v = NewNormalIntArray([]int{2})
	assert.Equal(t, i64Input, getIntArray(v))

	v = NewNormalIntArray([]uint8{2})
	assert.Equal(t, i64Input, getIntArray(v))

	v = NewNormalIntArray([]uint16{2})
	assert.Equal(t, i64Input, getIntArray(v))

	v = NewNormalIntArray([]uint32{2})
	assert.Equal(t, i64Input, getIntArray(v))

	v = NewNormalIntArray([]uint64{2})
	assert.Equal(t, i64Input, getIntArray(v))

	v = NewNormalIntArray([]uint{2})
	assert.Equal(t, i64Input, getIntArray(v))
}

func TestNormalValue_NewNormalFloatArray(t *testing.T) {
	f64InputFrac := []float64{2.5}
	f64Input := []float64{2.0}

	getFloatArray := func(v NormalValue) []float64 { f, _ := v.FloatArray(); return f }

	v := NewNormalFloatArray(f64InputFrac)
	assert.Equal(t, f64InputFrac, getFloatArray(v))

	v = NewNormalFloatArray([]float32{2.5})
	assert.Equal(t, f64InputFrac, getFloatArray(v))

	v = NewNormalFloatArray([]int8{2})
	assert.Equal(t, f64Input, getFloatArray(v))

	v = NewNormalFloatArray([]int16{2})
	assert.Equal(t, f64Input, getFloatArray(v))

	v = NewNormalFloatArray([]int32{2})
	assert.Equal(t, f64Input, getFloatArray(v))

	v = NewNormalFloatArray([]int64{2})
	assert.Equal(t, f64Input, getFloatArray(v))

	v = NewNormalFloatArray([]int{2})
	assert.Equal(t, f64Input, getFloatArray(v))

	v = NewNormalFloatArray([]uint8{2})
	assert.Equal(t, f64Input, getFloatArray(v))

	v = NewNormalFloatArray([]uint16{2})
	assert.Equal(t, f64Input, getFloatArray(v))

	v = NewNormalFloatArray([]uint32{2})
	assert.Equal(t, f64Input, getFloatArray(v))

	v = NewNormalFloatArray([]uint64{2})
	assert.Equal(t, f64Input, getFloatArray(v))

	v = NewNormalFloatArray([]uint{2})
	assert.Equal(t, f64Input, getFloatArray(v))
}

func TestNormalValue_NewNormalStringArray(t *testing.T) {
	strInput := []string{"str"}

	getStringArray := func(v NormalValue) []string { s, _ := v.StringArray(); return s }

	v := NewNormalStringArray(strInput)
	assert.Equal(t, strInput, getStringArray(v))

	v = NewNormalStringArray([][]byte{{'s', 't', 'r'}})
	assert.Equal(t, strInput, getStringArray(v))
}

func TestNormalValue_NewNormalBytesArray(t *testing.T) {
	bytesInput := [][]byte{[]byte("str")}

	getBytesArray := func(v NormalValue) [][]byte { b, _ := v.BytesArray(); return b }

	v := NewNormalBytesArray(bytesInput)
	assert.Equal(t, bytesInput, getBytesArray(v))

	v = NewNormalBytesArray([]string{"str"})
	assert.Equal(t, bytesInput, getBytesArray(v))
}

func TestNormalValue_NewNormalNillableFloatArray(t *testing.T) {
	f64InputFrac := []immutable.Option[float64]{immutable.Some(2.5)}
	f64Input := []immutable.Option[float64]{immutable.Some(2.0)}

	getNillableFloatArray := func(v NormalValue) []immutable.Option[float64] { f, _ := v.NillableFloatArray(); return f }

	v := NewNormalNillableFloatArray(f64InputFrac)
	assert.Equal(t, f64InputFrac, getNillableFloatArray(v))

	v = NewNormalNillableFloatArray([]immutable.Option[float32]{immutable.Some[float32](2.5)})
	assert.Equal(t, f64InputFrac, getNillableFloatArray(v))

	v = NewNormalNillableFloatArray([]immutable.Option[int8]{immutable.Some[int8](2)})
	assert.Equal(t, f64Input, getNillableFloatArray(v))

	v = NewNormalNillableFloatArray([]immutable.Option[int16]{immutable.Some[int16](2)})
	assert.Equal(t, f64Input, getNillableFloatArray(v))

	v = NewNormalNillableFloatArray([]immutable.Option[int32]{immutable.Some[int32](2)})
	assert.Equal(t, f64Input, getNillableFloatArray(v))

	v = NewNormalNillableFloatArray([]immutable.Option[int64]{immutable.Some[int64](2)})
	assert.Equal(t, f64Input, getNillableFloatArray(v))

	v = NewNormalNillableFloatArray([]immutable.Option[int]{immutable.Some[int](2)})
	assert.Equal(t, f64Input, getNillableFloatArray(v))

	v = NewNormalNillableFloatArray([]immutable.Option[uint8]{immutable.Some[uint8](2)})
	assert.Equal(t, f64Input, getNillableFloatArray(v))

	v = NewNormalNillableFloatArray([]immutable.Option[uint16]{immutable.Some[uint16](2)})
	assert.Equal(t, f64Input, getNillableFloatArray(v))

	v = NewNormalNillableFloatArray([]immutable.Option[uint32]{immutable.Some[uint32](2)})
	assert.Equal(t, f64Input, getNillableFloatArray(v))

	v = NewNormalNillableFloatArray([]immutable.Option[uint64]{immutable.Some[uint64](2)})
	assert.Equal(t, f64Input, getNillableFloatArray(v))

	v = NewNormalNillableFloatArray([]immutable.Option[uint]{immutable.Some[uint](2)})
	assert.Equal(t, f64Input, getNillableFloatArray(v))
}

func TestNormalValue_NewNormalNillableIntArray(t *testing.T) {
	i64Input := []immutable.Option[int64]{immutable.Some[int64](2)}

	getNillableIntArray := func(v NormalValue) []immutable.Option[int64] { i, _ := v.NillableIntArray(); return i }

	v := NewNormalNillableIntArray(i64Input)
	assert.Equal(t, i64Input, getNillableIntArray(v))

	v = NewNormalNillableIntArray([]immutable.Option[float32]{immutable.Some[float32](2.5)})
	assert.Equal(t, i64Input, getNillableIntArray(v))

	v = NewNormalNillableIntArray([]immutable.Option[float64]{immutable.Some[float64](2.5)})
	assert.Equal(t, i64Input, getNillableIntArray(v))

	v = NewNormalNillableIntArray([]immutable.Option[int8]{immutable.Some[int8](2)})
	assert.Equal(t, i64Input, getNillableIntArray(v))

	v = NewNormalNillableIntArray([]immutable.Option[int16]{immutable.Some[int16](2)})
	assert.Equal(t, i64Input, getNillableIntArray(v))

	v = NewNormalNillableIntArray([]immutable.Option[int32]{immutable.Some[int32](2)})
	assert.Equal(t, i64Input, getNillableIntArray(v))

	v = NewNormalNillableIntArray([]immutable.Option[int]{immutable.Some[int](2)})
	assert.Equal(t, i64Input, getNillableIntArray(v))

	v = NewNormalNillableIntArray([]immutable.Option[uint8]{immutable.Some[uint8](2)})
	assert.Equal(t, i64Input, getNillableIntArray(v))

	v = NewNormalNillableIntArray([]immutable.Option[uint16]{immutable.Some[uint16](2)})
	assert.Equal(t, i64Input, getNillableIntArray(v))

	v = NewNormalNillableIntArray([]immutable.Option[uint32]{immutable.Some[uint32](2)})
	assert.Equal(t, i64Input, getNillableIntArray(v))

	v = NewNormalNillableIntArray([]immutable.Option[uint64]{immutable.Some[uint64](2)})
	assert.Equal(t, i64Input, getNillableIntArray(v))

	v = NewNormalNillableIntArray([]immutable.Option[uint]{immutable.Some[uint](2)})
	assert.Equal(t, i64Input, getNillableIntArray(v))
}

func TestNormalValue_NewNormalNillableStringArray(t *testing.T) {
	strInput := []immutable.Option[string]{immutable.Some("str")}

	getNillableStringArray := func(v NormalValue) []immutable.Option[string] { s, _ := v.NillableStringArray(); return s }

	v := NewNormalNillableStringArray(strInput)
	assert.Equal(t, strInput, getNillableStringArray(v))

	v = NewNormalNillableStringArray([]immutable.Option[[]byte]{immutable.Some[[]byte]([]byte{'s', 't', 'r'})})
	assert.Equal(t, strInput, getNillableStringArray(v))
}

func TestNormalValue_NewNormalNillableBytesArray(t *testing.T) {
	bytesInput := []immutable.Option[[]byte]{immutable.Some[[]byte]([]byte("str"))}

	getNillableBytesArray := func(v NormalValue) []immutable.Option[[]byte] { b, _ := v.NillableBytesArray(); return b }

	v := NewNormalNillableBytesArray(bytesInput)
	assert.Equal(t, bytesInput, getNillableBytesArray(v))

	v = NewNormalNillableBytesArray([]immutable.Option[string]{immutable.Some("str")})
	assert.Equal(t, bytesInput, getNillableBytesArray(v))
}

func TestNormalValue_NewNormalIntArrayNillable(t *testing.T) {
	i64Input := immutable.Some([]int64{2})

	getIntNillableArray := func(v NormalValue) immutable.Option[[]int64] { i, _ := v.IntNillableArray(); return i }

	v := NewNormalIntNillableArray(i64Input)
	assert.Equal(t, i64Input, getIntNillableArray(v))

	v = NewNormalIntNillableArray(immutable.Some([]float32{2.5}))
	assert.Equal(t, i64Input, getIntNillableArray(v))

	v = NewNormalIntNillableArray(immutable.Some([]float64{2.5}))
	assert.Equal(t, i64Input, getIntNillableArray(v))

	v = NewNormalIntNillableArray(immutable.Some([]int8{2}))
	assert.Equal(t, i64Input, getIntNillableArray(v))

	v = NewNormalIntNillableArray(immutable.Some([]int16{2}))
	assert.Equal(t, i64Input, getIntNillableArray(v))

	v = NewNormalIntNillableArray(immutable.Some([]int32{2}))
	assert.Equal(t, i64Input, getIntNillableArray(v))

	v = NewNormalIntNillableArray(immutable.Some([]int{2}))
	assert.Equal(t, i64Input, getIntNillableArray(v))

	v = NewNormalIntNillableArray(immutable.Some([]uint8{2}))
	assert.Equal(t, i64Input, getIntNillableArray(v))

	v = NewNormalIntNillableArray(immutable.Some([]uint16{2}))
	assert.Equal(t, i64Input, getIntNillableArray(v))

	v = NewNormalIntNillableArray(immutable.Some([]uint32{2}))
	assert.Equal(t, i64Input, getIntNillableArray(v))

	v = NewNormalIntNillableArray(immutable.Some([]uint64{2}))
	assert.Equal(t, i64Input, getIntNillableArray(v))

	v = NewNormalIntNillableArray(immutable.Some([]uint{2}))
	assert.Equal(t, i64Input, getIntNillableArray(v))
}

func TestNormalValue_NewNormalFloatNillableArray(t *testing.T) {
	f64InputFrac := immutable.Some([]float64{2.5})
	f64Input := immutable.Some([]float64{2.0})

	getFloatNillableArray := func(v NormalValue) immutable.Option[[]float64] { f, _ := v.FloatNillableArray(); return f }

	v := NewNormalFloatNillableArray(f64InputFrac)
	assert.Equal(t, f64InputFrac, getFloatNillableArray(v))

	v = NewNormalFloatNillableArray(immutable.Some([]float32{2.5}))
	assert.Equal(t, f64InputFrac, getFloatNillableArray(v))

	v = NewNormalFloatNillableArray(immutable.Some([]int8{2}))
	assert.Equal(t, f64Input, getFloatNillableArray(v))

	v = NewNormalFloatNillableArray(immutable.Some([]int16{2}))
	assert.Equal(t, f64Input, getFloatNillableArray(v))

	v = NewNormalFloatNillableArray(immutable.Some([]int32{2}))
	assert.Equal(t, f64Input, getFloatNillableArray(v))

	v = NewNormalFloatNillableArray(immutable.Some([]int64{2}))
	assert.Equal(t, f64Input, getFloatNillableArray(v))

	v = NewNormalFloatNillableArray(immutable.Some([]int{2}))
	assert.Equal(t, f64Input, getFloatNillableArray(v))

	v = NewNormalFloatNillableArray(immutable.Some([]uint8{2}))
	assert.Equal(t, f64Input, getFloatNillableArray(v))

	v = NewNormalFloatNillableArray(immutable.Some([]uint16{2}))
	assert.Equal(t, f64Input, getFloatNillableArray(v))

	v = NewNormalFloatNillableArray(immutable.Some([]uint32{2}))
	assert.Equal(t, f64Input, getFloatNillableArray(v))

	v = NewNormalFloatNillableArray(immutable.Some([]uint64{2}))
	assert.Equal(t, f64Input, getFloatNillableArray(v))

	v = NewNormalFloatNillableArray(immutable.Some([]uint{2}))
	assert.Equal(t, f64Input, getFloatNillableArray(v))
}

func TestNormalValue_NewNormalStringNillableArray(t *testing.T) {
	strInput := immutable.Some([]string{"str"})

	getStringNillableArray := func(v NormalValue) immutable.Option[[]string] { s, _ := v.StringNillableArray(); return s }

	v := NewNormalStringNillableArray(strInput)
	assert.Equal(t, strInput, getStringNillableArray(v))

	v = NewNormalStringNillableArray(immutable.Some([][]byte{{'s', 't', 'r'}}))
	assert.Equal(t, strInput, getStringNillableArray(v))
}

func TestNormalValue_NewNormalBytesNillableArray(t *testing.T) {
	bytesInput := immutable.Some([][]byte{{'s', 't', 'r'}})

	getBytesNillableArray := func(v NormalValue) immutable.Option[[][]byte] { s, _ := v.BytesNillableArray(); return s }

	v := NewNormalBytesNillableArray(immutable.Some([]string{"str"}))
	assert.Equal(t, bytesInput, getBytesNillableArray(v))

	v = NewNormalBytesNillableArray(bytesInput)
	assert.Equal(t, bytesInput, getBytesNillableArray(v))
}

func TestNormalValue_NewNormalNillableIntNillableArray(t *testing.T) {
	i64Input := immutable.Some([]immutable.Option[int64]{immutable.Some(int64(2))})

	getNillableIntNillableArray := func(v NormalValue) immutable.Option[[]immutable.Option[int64]] {
		i, _ := v.NillableIntNillableArray()
		return i
	}

	v := NewNormalNillableIntNillableArray(i64Input)
	assert.Equal(t, i64Input, getNillableIntNillableArray(v))

	v = NewNormalNillableIntNillableArray(immutable.Some([]immutable.Option[float32]{immutable.Some(float32(2.5))}))
	assert.Equal(t, i64Input, getNillableIntNillableArray(v))

	v = NewNormalNillableIntNillableArray(immutable.Some([]immutable.Option[float64]{immutable.Some(2.5)}))
	assert.Equal(t, i64Input, getNillableIntNillableArray(v))

	v = NewNormalNillableIntNillableArray(immutable.Some([]immutable.Option[int8]{immutable.Some(int8(2))}))
	assert.Equal(t, i64Input, getNillableIntNillableArray(v))

	v = NewNormalNillableIntNillableArray(immutable.Some([]immutable.Option[int16]{immutable.Some(int16(2))}))
	assert.Equal(t, i64Input, getNillableIntNillableArray(v))

	v = NewNormalNillableIntNillableArray(immutable.Some([]immutable.Option[int32]{immutable.Some(int32(2))}))
	assert.Equal(t, i64Input, getNillableIntNillableArray(v))

	v = NewNormalNillableIntNillableArray(immutable.Some([]immutable.Option[int]{immutable.Some(int(2))}))
	assert.Equal(t, i64Input, getNillableIntNillableArray(v))

	v = NewNormalNillableIntNillableArray(immutable.Some([]immutable.Option[uint8]{immutable.Some(uint8(2))}))
	assert.Equal(t, i64Input, getNillableIntNillableArray(v))

	v = NewNormalNillableIntNillableArray(immutable.Some([]immutable.Option[uint16]{immutable.Some(uint16(2))}))
	assert.Equal(t, i64Input, getNillableIntNillableArray(v))

	v = NewNormalNillableIntNillableArray(immutable.Some([]immutable.Option[uint32]{immutable.Some(uint32(2))}))
	assert.Equal(t, i64Input, getNillableIntNillableArray(v))

	v = NewNormalNillableIntNillableArray(immutable.Some([]immutable.Option[uint64]{immutable.Some(uint64(2))}))
	assert.Equal(t, i64Input, getNillableIntNillableArray(v))

	v = NewNormalNillableIntNillableArray(immutable.Some([]immutable.Option[uint]{immutable.Some(uint(2))}))
	assert.Equal(t, i64Input, getNillableIntNillableArray(v))
}

func TestNormalValue_NewNormalNillableFloatNillableArray(t *testing.T) {
	f64InputFrac := immutable.Some([]immutable.Option[float64]{immutable.Some(2.5)})
	f64Input := immutable.Some([]immutable.Option[float64]{immutable.Some(2.0)})

	getNillableFloatNillableArray := func(v NormalValue) immutable.Option[[]immutable.Option[float64]] {
		f, _ := v.NillableFloatNillableArray()
		return f
	}

	v := NewNormalNillableFloatNillableArray(f64InputFrac)
	assert.Equal(t, f64InputFrac, getNillableFloatNillableArray(v))

	v = NewNormalNillableFloatNillableArray(immutable.Some([]immutable.Option[float32]{immutable.Some(float32(2.5))}))
	assert.Equal(t, f64InputFrac, getNillableFloatNillableArray(v))

	v = NewNormalNillableFloatNillableArray(immutable.Some([]immutable.Option[int8]{immutable.Some(int8(2))}))
	assert.Equal(t, f64Input, getNillableFloatNillableArray(v))

	v = NewNormalNillableFloatNillableArray(immutable.Some([]immutable.Option[int16]{immutable.Some(int16(2))}))
	assert.Equal(t, f64Input, getNillableFloatNillableArray(v))

	v = NewNormalNillableFloatNillableArray(immutable.Some([]immutable.Option[int32]{immutable.Some(int32(2))}))
	assert.Equal(t, f64Input, getNillableFloatNillableArray(v))

	v = NewNormalNillableFloatNillableArray(immutable.Some([]immutable.Option[int64]{immutable.Some(int64(2))}))
	assert.Equal(t, f64Input, getNillableFloatNillableArray(v))

	v = NewNormalNillableFloatNillableArray(immutable.Some([]immutable.Option[int]{immutable.Some(2)}))
	assert.Equal(t, f64Input, getNillableFloatNillableArray(v))

	v = NewNormalNillableFloatNillableArray(immutable.Some([]immutable.Option[uint8]{immutable.Some(uint8(2))}))
	assert.Equal(t, f64Input, getNillableFloatNillableArray(v))

	v = NewNormalNillableFloatNillableArray(immutable.Some([]immutable.Option[uint16]{immutable.Some(uint16(2))}))
	assert.Equal(t, f64Input, getNillableFloatNillableArray(v))

	v = NewNormalNillableFloatNillableArray(immutable.Some([]immutable.Option[uint32]{immutable.Some(uint32(2))}))
	assert.Equal(t, f64Input, getNillableFloatNillableArray(v))

	v = NewNormalNillableFloatNillableArray(immutable.Some([]immutable.Option[uint64]{immutable.Some(uint64(2))}))
	assert.Equal(t, f64Input, getNillableFloatNillableArray(v))

	v = NewNormalNillableFloatNillableArray(immutable.Some([]immutable.Option[uint]{immutable.Some(uint(2))}))
	assert.Equal(t, f64Input, getNillableFloatNillableArray(v))
}

func TestNormalValue_NewNormalNillableStringNillableArray(t *testing.T) {
	strInput := immutable.Some([]immutable.Option[string]{immutable.Some("str")})

	getNillableStringNillableArray := func(v NormalValue) immutable.Option[[]immutable.Option[string]] {
		s, _ := v.NillableStringNillableArray()
		return s
	}

	v := NewNormalNillableStringNillableArray(strInput)
	assert.Equal(t, strInput, getNillableStringNillableArray(v))

	bytesInput := immutable.Some([]immutable.Option[[]byte]{immutable.Some([]byte{'s', 't', 'r'})})
	v = NewNormalNillableStringNillableArray(bytesInput)
	assert.Equal(t, strInput, getNillableStringNillableArray(v))
}

func TestNormalValue_NewNormalNillableBytesNillableArray(t *testing.T) {
	bytesInput := immutable.Some([]immutable.Option[[]byte]{immutable.Some([]byte{'s', 't', 'r'})})

	getNillableBytesNillableArray := func(v NormalValue) immutable.Option[[]immutable.Option[[]byte]] {
		s, _ := v.NillableBytesNillableArray()
		return s
	}

	v := NewNormalNillableBytesNillableArray(bytesInput)
	assert.Equal(t, bytesInput, getNillableBytesNillableArray(v))

	strInput := immutable.Some([]immutable.Option[string]{immutable.Some("str")})
	v = NewNormalNillableBytesNillableArray(strInput)
	assert.Equal(t, bytesInput, getNillableBytesNillableArray(v))
}

func TestNormalValue_NewNormalNil(t *testing.T) {
	fieldKinds := []FieldKind{}
	for _, kind := range FieldKindStringToEnumMapping {
		fieldKinds = append(fieldKinds, kind)
	}
	fieldKinds = append(fieldKinds, NewCollectionKind(1, false))
	fieldKinds = append(fieldKinds, NewCollectionKind(1, true))

	for _, kind := range fieldKinds {
		if kind.IsNillable() {
			v, err := NewNormalNil(kind)
			require.NoError(t, err)

			assert.True(t, v.IsNil())
		} else {
			_, err := NewNormalNil(kind)
			require.Error(t, err, "field kind: "+kind.String())
		}
	}
}

func TestNormalValue_ToArrayOfNormalValues(t *testing.T) {
	now := time.Now()
	doc1 := &Document{}
	doc2 := &Document{}

	normalNil, err := NewNormalNil(FieldKind_NILLABLE_INT)
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    NormalValue
		expected []NormalValue
		err      string
	}{
		{
			name:  "nil",
			input: normalNil,
		},
		{
			name:  "not array",
			input: NewNormalInt(1),
			err:   errCanNotTurnNormalValueIntoArray,
		},
		{
			name:     "bool elements",
			input:    NewNormalBoolArray([]bool{true, false}),
			expected: []NormalValue{NewNormalBool(true), NewNormalBool(false)},
		},
		{
			name:     "int elements",
			input:    NewNormalIntArray([]int64{1, 2}),
			expected: []NormalValue{NewNormalInt(1), NewNormalInt(2)},
		},
		{
			name:     "float elements",
			input:    NewNormalFloatArray([]float64{1.0, 2.0}),
			expected: []NormalValue{NewNormalFloat(1.0), NewNormalFloat(2.0)},
		},
		{
			name:     "string elements",
			input:    NewNormalStringArray([]string{"test", "test2"}),
			expected: []NormalValue{NewNormalString("test"), NewNormalString("test2")},
		},
		{
			name:     "bytes elements",
			input:    NewNormalBytesArray([][]byte{{1, 2, 3}, {4, 5, 6}}),
			expected: []NormalValue{NewNormalBytes([]byte{1, 2, 3}), NewNormalBytes([]byte{4, 5, 6})},
		},
		{
			name:     "time elements",
			input:    NewNormalTimeArray([]time.Time{now, now}),
			expected: []NormalValue{NewNormalTime(now), NewNormalTime(now)},
		},
		{
			name:     "document elements",
			input:    NewNormalDocumentArray([]*Document{doc1, doc2}),
			expected: []NormalValue{NewNormalDocument(doc1), NewNormalDocument(doc2)},
		},
		{
			name: "nillable bool elements",
			input: NewNormalNillableBoolArray([]immutable.Option[bool]{
				immutable.Some(true), immutable.Some(false)}),
			expected: []NormalValue{
				NewNormalNillableBool(immutable.Some(true)),
				NewNormalNillableBool(immutable.Some(false)),
			},
		},
		{
			name: "nillable int elements",
			input: NewNormalNillableIntArray([]immutable.Option[int64]{
				immutable.Some(int64(1)), immutable.Some(int64(2))}),
			expected: []NormalValue{
				NewNormalNillableInt(immutable.Some(int64(1))),
				NewNormalNillableInt(immutable.Some(int64(2))),
			},
		},
		{
			name: "nillable float elements",
			input: NewNormalNillableFloatArray([]immutable.Option[float64]{
				immutable.Some(1.0), immutable.Some(2.0)}),
			expected: []NormalValue{
				NewNormalNillableFloat(immutable.Some(1.0)),
				NewNormalNillableFloat(immutable.Some(2.0)),
			},
		},
		{
			name: "nillable string elements",
			input: NewNormalNillableStringArray([]immutable.Option[string]{
				immutable.Some("test"), immutable.Some("test2")}),
			expected: []NormalValue{
				NewNormalNillableString(immutable.Some("test")),
				NewNormalNillableString(immutable.Some("test2")),
			},
		},
		{
			name: "nillable bytes elements",
			input: NewNormalNillableBytesArray([]immutable.Option[[]byte]{
				immutable.Some([]byte{1, 2, 3}), immutable.Some([]byte{4, 5, 6})}),
			expected: []NormalValue{
				NewNormalNillableBytes(immutable.Some([]byte{1, 2, 3})),
				NewNormalNillableBytes(immutable.Some([]byte{4, 5, 6})),
			},
		},
		{
			name: "nillable time elements",
			input: NewNormalNillableTimeArray([]immutable.Option[time.Time]{
				immutable.Some(now), immutable.Some(now)}),
			expected: []NormalValue{
				NewNormalNillableTime(immutable.Some(now)),
				NewNormalNillableTime(immutable.Some(now)),
			},
		},
		{
			name: "nillable document elements",
			input: NewNormalNillableDocumentArray([]immutable.Option[*Document]{
				immutable.Some(doc1), immutable.Some(doc2)}),
			expected: []NormalValue{
				NewNormalNillableDocument(immutable.Some(doc1)),
				NewNormalNillableDocument(immutable.Some(doc2)),
			},
		},
		{
			name:     "nillable array of bool elements",
			input:    NewNormalBoolNillableArray(immutable.Some([]bool{true})),
			expected: []NormalValue{NewNormalBool(true)},
		},
		{
			name:     "nillable array of int elements",
			input:    NewNormalIntNillableArray(immutable.Some([]int64{1})),
			expected: []NormalValue{NewNormalInt(1)},
		},
		{
			name:     "nillable array of float elements",
			input:    NewNormalFloatNillableArray(immutable.Some([]float64{1.0})),
			expected: []NormalValue{NewNormalFloat(1.0)},
		},
		{
			name:     "nillable array of string elements",
			input:    NewNormalStringNillableArray(immutable.Some([]string{"test"})),
			expected: []NormalValue{NewNormalString("test")},
		},
		{
			name:     "nillable array of bytes elements",
			input:    NewNormalBytesNillableArray(immutable.Some([][]byte{{1, 2, 3}})),
			expected: []NormalValue{NewNormalBytes([]byte{1, 2, 3})},
		},
		{
			name:     "nillable array of time elements",
			input:    NewNormalTimeNillableArray(immutable.Some([]time.Time{now})),
			expected: []NormalValue{NewNormalTime(now)},
		},
		{
			name:     "nillable array of document elements",
			input:    NewNormalDocumentNillableArray(immutable.Some([]*Document{doc1})),
			expected: []NormalValue{NewNormalDocument(doc1)},
		},
		{
			name: "nillable array of nillable bool elements",
			input: NewNormalNillableBoolNillableArray(
				immutable.Some([]immutable.Option[bool]{immutable.Some(true)})),
			expected: []NormalValue{NewNormalNillableBool(immutable.Some(true))},
		},
		{
			name: "nillable array of nillable int elements",
			input: NewNormalNillableIntNillableArray(
				immutable.Some([]immutable.Option[int64]{immutable.Some(int64(1))})),
			expected: []NormalValue{NewNormalNillableInt(immutable.Some(int64(1)))},
		},
		{
			name: "nillable array of nillable float elements",
			input: NewNormalNillableFloatNillableArray(
				immutable.Some([]immutable.Option[float64]{immutable.Some(1.0)})),
			expected: []NormalValue{NewNormalNillableFloat(immutable.Some(1.0))},
		},
		{
			name: "nillable array of nillable string elements",
			input: NewNormalNillableStringNillableArray(
				immutable.Some([]immutable.Option[string]{immutable.Some("test")})),
			expected: []NormalValue{NewNormalNillableString(immutable.Some("test"))},
		},
		{
			name: "nillable array of nillable bytes elements",
			input: NewNormalNillableBytesNillableArray(
				immutable.Some([]immutable.Option[[]byte]{immutable.Some([]byte{1, 2, 3})})),
			expected: []NormalValue{NewNormalNillableBytes(immutable.Some([]byte{1, 2, 3}))},
		},
		{
			name: "nillable array of nillable time elements",
			input: NewNormalNillableTimeNillableArray(
				immutable.Some([]immutable.Option[time.Time]{immutable.Some(now)})),
			expected: []NormalValue{NewNormalNillableTime(immutable.Some(now))},
		},
		{
			name: "nillable array of nillable document elements",
			input: NewNormalNillableDocumentNillableArray(
				immutable.Some([]immutable.Option[*Document]{immutable.Some(doc1)})),
			expected: []NormalValue{NewNormalNillableDocument(immutable.Some(doc1))},
		},
	}

	for _, tt := range tests {
		tStr := string(tt.name)
		t.Run(tStr, func(t *testing.T) {
			actual, err := ToArrayOfNormalValues(tt.input)
			if tt.err != "" {
				require.ErrorContains(t, err, tt.err)
				return
			}

			assert.Equal(t, tt.expected, actual)
		})
	}
}

// This test documents a bug where array values
// were not returning the correct value for IsNillable
// and were also not convertible to a normal nil kind.
func TestArrayValue_IsNillable(t *testing.T) {
	fieldKinds := []FieldKind{
		FieldKind_BOOL_ARRAY,
		FieldKind_INT_ARRAY,
		FieldKind_FLOAT_ARRAY,
		FieldKind_STRING_ARRAY,
		FieldKind_NILLABLE_BOOL_ARRAY,
		FieldKind_NILLABLE_INT_ARRAY,
		FieldKind_NILLABLE_FLOAT_ARRAY,
		FieldKind_NILLABLE_STRING_ARRAY,
	}

	for _, kind := range fieldKinds {
		assert.True(t, kind.IsNillable())

		v, err := NewNormalNil(kind)
		require.NoError(t, err)

		assert.True(t, v.IsNil())
	}
}

func TestNormalValue_IsEqual(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Hour)
	doc1 := &Document{}
	doc2 := &Document{}
	doc3 := &Document{}

	tests := []struct {
		name     string
		v1       NormalValue
		v2       NormalValue
		expected bool
	}{
		{
			name:     "void",
			v1:       NormalVoid{},
			v2:       NormalVoid{},
			expected: true,
		},
		{
			name:     "void not equal",
			v1:       NormalVoid{},
			v2:       NewNormalInt(1),
			expected: false,
		},
		{
			name:     "bool",
			v1:       NewNormalBool(true),
			v2:       NewNormalBool(true),
			expected: true,
		},
		{
			name:     "bool not equal",
			v1:       NewNormalBool(true),
			v2:       NewNormalBool(false),
			expected: false,
		},
		{
			name:     "bool different type",
			v1:       NewNormalBool(true),
			v2:       NewNormalInt(1),
			expected: false,
		},
		{
			name:     "int",
			v1:       NewNormalInt(1),
			v2:       NewNormalInt(1),
			expected: true,
		},
		{
			name:     "int not equal",
			v1:       NewNormalInt(1),
			v2:       NewNormalInt(2),
			expected: false,
		},
		{
			name:     "int different type",
			v1:       NewNormalInt(1),
			v2:       NewNormalFloat(1.0),
			expected: false,
		},
		{
			name:     "float",
			v1:       NewNormalFloat(1.0),
			v2:       NewNormalFloat(1.0),
			expected: true,
		},
		{
			name:     "float not equal",
			v1:       NewNormalFloat(1.0),
			v2:       NewNormalFloat(1.1),
			expected: false,
		},
		{
			name:     "float different type",
			v1:       NewNormalFloat(1.0),
			v2:       NewNormalString("1.0"),
			expected: false,
		},
		{
			name:     "string",
			v1:       NewNormalString("test"),
			v2:       NewNormalString("test"),
			expected: true,
		},
		{
			name:     "string not equal",
			v1:       NewNormalString("test"),
			v2:       NewNormalString("test2"),
			expected: false,
		},
		{
			name:     "string different type",
			v1:       NewNormalString("test"),
			v2:       NewNormalBytes([]byte("test")),
			expected: false,
		},
		{
			name:     "bytes",
			v1:       NewNormalBytes([]byte{1, 2, 3}),
			v2:       NewNormalBytes([]byte{1, 2, 3}),
			expected: true,
		},
		{
			name:     "bytes not equal",
			v1:       NewNormalBytes([]byte{1, 2, 3}),
			v2:       NewNormalBytes([]byte{1, 2, 4}),
			expected: false,
		},
		{
			name:     "bytes different length",
			v1:       NewNormalBytes([]byte{1, 2, 3}),
			v2:       NewNormalBytes([]byte{1, 2, 3, 4}),
			expected: false,
		},
		{
			name:     "bytes different type",
			v1:       NewNormalBytes([]byte{1, 2, 3}),
			v2:       NewNormalString("123"),
			expected: false,
		},
		{
			name:     "time",
			v1:       NewNormalTime(now),
			v2:       NewNormalTime(now),
			expected: true,
		},
		{
			name:     "time not equal",
			v1:       NewNormalTime(now),
			v2:       NewNormalTime(later),
			expected: false,
		},
		{
			name:     "time different type",
			v1:       NewNormalTime(now),
			v2:       NewNormalString(now.String()),
			expected: false,
		},
		{
			name:     "document",
			v1:       NewNormalDocument(doc1),
			v2:       NewNormalDocument(doc1),
			expected: true,
		},
		{
			name:     "document not equal",
			v1:       NewNormalDocument(doc1),
			v2:       NewNormalDocument(doc2),
			expected: false,
		},
		{
			name:     "document different type",
			v1:       NewNormalDocument(doc1),
			v2:       NewNormalString("document"),
			expected: false,
		},
		{
			name:     "bool nillable",
			v1:       NewNormalNillableBool(immutable.Some(true)),
			v2:       NewNormalNillableBool(immutable.Some(true)),
			expected: true,
		},
		{
			name:     "bool nillable not equal",
			v1:       NewNormalNillableBool(immutable.Some(true)),
			v2:       NewNormalNillableBool(immutable.Some(false)),
			expected: false,
		},
		{
			name:     "bool nillable one nil",
			v1:       NewNormalNillableBool(immutable.Some(true)),
			v2:       NewNormalNillableBool(immutable.None[bool]()),
			expected: false,
		},
		{
			name:     "bool nillable different type",
			v1:       NewNormalNillableBool(immutable.Some(true)),
			v2:       NewNormalBool(true),
			expected: false,
		},
		{
			name:     "int nillable",
			v1:       NewNormalNillableInt(immutable.Some(int64(1))),
			v2:       NewNormalNillableInt(immutable.Some(int64(1))),
			expected: true,
		},
		{
			name:     "int nillable not equal",
			v1:       NewNormalNillableInt(immutable.Some(int64(1))),
			v2:       NewNormalNillableInt(immutable.Some(int64(2))),
			expected: false,
		},
		{
			name:     "int nillable one nil",
			v1:       NewNormalNillableInt(immutable.Some(int64(1))),
			v2:       NewNormalNillableInt(immutable.None[int64]()),
			expected: false,
		},
		{
			name:     "int nillable different type",
			v1:       NewNormalNillableInt(immutable.Some(int64(1))),
			v2:       NewNormalInt(1),
			expected: false,
		},
		{
			name:     "float nillable",
			v1:       NewNormalNillableFloat(immutable.Some(1.0)),
			v2:       NewNormalNillableFloat(immutable.Some(1.0)),
			expected: true,
		},
		{
			name:     "float nillable not equal",
			v1:       NewNormalNillableFloat(immutable.Some(1.0)),
			v2:       NewNormalNillableFloat(immutable.Some(2.0)),
			expected: false,
		},
		{
			name:     "float nillable one nil",
			v1:       NewNormalNillableFloat(immutable.Some(1.0)),
			v2:       NewNormalNillableFloat(immutable.None[float64]()),
			expected: false,
		},
		{
			name:     "float nillable different type",
			v1:       NewNormalNillableFloat(immutable.Some(1.0)),
			v2:       NewNormalFloat(1.0),
			expected: false,
		},
		{
			name:     "string nillable",
			v1:       NewNormalNillableString(immutable.Some("test")),
			v2:       NewNormalNillableString(immutable.Some("test")),
			expected: true,
		},
		{
			name:     "string nillable not equal",
			v1:       NewNormalNillableString(immutable.Some("test")),
			v2:       NewNormalNillableString(immutable.Some("test2")),
			expected: false,
		},
		{
			name:     "string nillable one nil",
			v1:       NewNormalNillableString(immutable.Some("test")),
			v2:       NewNormalNillableString(immutable.None[string]()),
			expected: false,
		},
		{
			name:     "string nillable different type",
			v1:       NewNormalNillableString(immutable.Some("test")),
			v2:       NewNormalString("test"),
			expected: false,
		},
		{
			name:     "bytes nillable",
			v1:       NewNormalNillableBytes(immutable.Some([]byte{1, 2, 3})),
			v2:       NewNormalNillableBytes(immutable.Some([]byte{1, 2, 3})),
			expected: true,
		},
		{
			name:     "bytes nillable not equal",
			v1:       NewNormalNillableBytes(immutable.Some([]byte{1, 2, 3})),
			v2:       NewNormalNillableBytes(immutable.Some([]byte{1, 2, 4})),
			expected: false,
		},
		{
			name:     "bytes nillable one nil",
			v1:       NewNormalNillableBytes(immutable.Some([]byte{1, 2, 3})),
			v2:       NewNormalNillableBytes(immutable.None[[]byte]()),
			expected: false,
		},
		{
			name:     "bytes nillable different length",
			v1:       NewNormalNillableBytes(immutable.Some([]byte{1, 2, 3})),
			v2:       NewNormalNillableBytes(immutable.Some([]byte{1, 2, 3, 4})),
			expected: false,
		},
		{
			name:     "bytes nillable different type",
			v1:       NewNormalNillableBytes(immutable.Some([]byte{1, 2, 3})),
			v2:       NewNormalBytes([]byte{1, 2, 3}),
			expected: false,
		},
		{
			name:     "time nillable",
			v1:       NewNormalNillableTime(immutable.Some(now)),
			v2:       NewNormalNillableTime(immutable.Some(now)),
			expected: true,
		},
		{
			name:     "time nillable not equal",
			v1:       NewNormalNillableTime(immutable.Some(now)),
			v2:       NewNormalNillableTime(immutable.Some(later)),
			expected: false,
		},
		{
			name:     "time nillable one nil",
			v1:       NewNormalNillableTime(immutable.Some(now)),
			v2:       NewNormalNillableTime(immutable.None[time.Time]()),
			expected: false,
		},
		{
			name:     "time nillable different type",
			v1:       NewNormalNillableTime(immutable.Some(now)),
			v2:       NewNormalTime(now),
			expected: false,
		},
		{
			name:     "document nillable",
			v1:       NewNormalNillableDocument(immutable.Some(doc1)),
			v2:       NewNormalNillableDocument(immutable.Some(doc1)),
			expected: true,
		},
		{
			name:     "document nillable not equal",
			v1:       NewNormalNillableDocument(immutable.Some(doc1)),
			v2:       NewNormalNillableDocument(immutable.Some(doc2)),
			expected: false,
		},
		{
			name:     "document nillable one nil",
			v1:       NewNormalNillableDocument(immutable.Some(doc1)),
			v2:       NewNormalNillableDocument(immutable.None[*Document]()),
			expected: false,
		},
		{
			name:     "document nillable different type",
			v1:       NewNormalNillableDocument(immutable.Some(doc1)),
			v2:       NewNormalDocument(doc1),
			expected: false,
		},
		{
			name:     "bool array",
			v1:       NewNormalBoolArray([]bool{true, false}),
			v2:       NewNormalBoolArray([]bool{true, false}),
			expected: true,
		},
		{
			name:     "bool array not equal",
			v1:       NewNormalBoolArray([]bool{true, false}),
			v2:       NewNormalBoolArray([]bool{true, true}),
			expected: false,
		},
		{
			name:     "bool array different length",
			v1:       NewNormalBoolArray([]bool{true, false}),
			v2:       NewNormalBoolArray([]bool{true}),
			expected: false,
		},
		{
			name:     "bool array different type",
			v1:       NewNormalBoolArray([]bool{true, false}),
			v2:       NewNormalIntArray([]int64{1, 0}),
			expected: false,
		},
		{
			name:     "int array",
			v1:       NewNormalIntArray([]int64{1, 2}),
			v2:       NewNormalIntArray([]int64{1, 2}),
			expected: true,
		},
		{
			name:     "int array not equal",
			v1:       NewNormalIntArray([]int64{1, 2}),
			v2:       NewNormalIntArray([]int64{1, 3}),
			expected: false,
		},
		{
			name:     "int array different length",
			v1:       NewNormalIntArray([]int64{1, 2}),
			v2:       NewNormalIntArray([]int64{1}),
			expected: false,
		},
		{
			name:     "int array different type",
			v1:       NewNormalIntArray([]int64{1, 2}),
			v2:       NewNormalFloatArray([]float64{1.0, 2.0}),
			expected: false,
		},
		{
			name:     "float array",
			v1:       NewNormalFloatArray([]float64{1.0, 2.0}),
			v2:       NewNormalFloatArray([]float64{1.0, 2.0}),
			expected: true,
		},
		{
			name:     "float array not equal",
			v1:       NewNormalFloatArray([]float64{1.0, 2.0}),
			v2:       NewNormalFloatArray([]float64{1.0, 3.0}),
			expected: false,
		},
		{
			name:     "float array different length",
			v1:       NewNormalFloatArray([]float64{1.0, 2.0}),
			v2:       NewNormalFloatArray([]float64{1.0}),
			expected: false,
		},
		{
			name:     "float array different type",
			v1:       NewNormalFloatArray([]float64{1.0, 2.0}),
			v2:       NewNormalStringArray([]string{"1.0", "2.0"}),
			expected: false,
		},
		{
			name:     "string array",
			v1:       NewNormalStringArray([]string{"test", "test2"}),
			v2:       NewNormalStringArray([]string{"test", "test2"}),
			expected: true,
		},
		{
			name:     "string array not equal",
			v1:       NewNormalStringArray([]string{"test", "test2"}),
			v2:       NewNormalStringArray([]string{"test", "test3"}),
			expected: false,
		},
		{
			name:     "string array different length",
			v1:       NewNormalStringArray([]string{"test", "test2"}),
			v2:       NewNormalStringArray([]string{"test"}),
			expected: false,
		},
		{
			name:     "string array different type",
			v1:       NewNormalStringArray([]string{"test", "test2"}),
			v2:       NewNormalBytesArray([][]byte{{1, 2, 3}, {4, 5, 6}}),
			expected: false,
		},
		{
			name:     "bytes array",
			v1:       NewNormalBytesArray([][]byte{{1, 2, 3}, {4, 5, 6}}),
			v2:       NewNormalBytesArray([][]byte{{1, 2, 3}, {4, 5, 6}}),
			expected: true,
		},
		{
			name:     "bytes array not equal",
			v1:       NewNormalBytesArray([][]byte{{1, 2, 3}, {4, 5, 6}}),
			v2:       NewNormalBytesArray([][]byte{{1, 2, 3}, {4, 5, 7}}),
			expected: false,
		},
		{
			name:     "bytes array different length",
			v1:       NewNormalBytesArray([][]byte{{1, 2, 3}, {4, 5, 6}}),
			v2:       NewNormalBytesArray([][]byte{{1, 2, 3}}),
			expected: false,
		},
		{
			name:     "bytes array different type",
			v1:       NewNormalBytesArray([][]byte{{1, 2, 3}, {4, 5, 6}}),
			v2:       NewNormalStringArray([]string{"123", "456"}),
			expected: false,
		},
		{
			name:     "time array",
			v1:       NewNormalTimeArray([]time.Time{now, now}),
			v2:       NewNormalTimeArray([]time.Time{now, now}),
			expected: true,
		},
		{
			name:     "time array not equal",
			v1:       NewNormalTimeArray([]time.Time{now, now}),
			v2:       NewNormalTimeArray([]time.Time{now, later}),
			expected: false,
		},
		{
			name:     "time array different length",
			v1:       NewNormalTimeArray([]time.Time{now, now}),
			v2:       NewNormalTimeArray([]time.Time{now}),
			expected: false,
		},
		{
			name:     "time array different type",
			v1:       NewNormalTimeArray([]time.Time{now, now}),
			v2:       NewNormalStringArray([]string{now.String(), now.String()}),
			expected: false,
		},
		{
			name:     "document array",
			v1:       NewNormalDocumentArray([]*Document{doc1, doc2}),
			v2:       NewNormalDocumentArray([]*Document{doc1, doc2}),
			expected: true,
		},
		{
			name:     "document array not equal",
			v1:       NewNormalDocumentArray([]*Document{doc1, doc2}),
			v2:       NewNormalDocumentArray([]*Document{doc1, doc1}),
			expected: false,
		},
		{
			name:     "document array different length",
			v1:       NewNormalDocumentArray([]*Document{doc1, doc2}),
			v2:       NewNormalDocumentArray([]*Document{doc1}),
			expected: false,
		},
		{
			name:     "document array different type",
			v1:       NewNormalDocumentArray([]*Document{doc1, doc2}),
			v2:       NewNormalStringArray([]string{"doc1", "doc2"}),
			expected: false,
		},
		{
			name:     "bool nillable array",
			v1:       NewNormalNillableBoolArray([]immutable.Option[bool]{immutable.Some(true), immutable.Some(false)}),
			v2:       NewNormalNillableBoolArray([]immutable.Option[bool]{immutable.Some(true), immutable.Some(false)}),
			expected: true,
		},
		{
			name: "int nillable array",
			v1: NewNormalNillableIntArray(
				[]immutable.Option[int64]{immutable.Some(int64(1)), immutable.Some(int64(2))}),
			v2: NewNormalNillableIntArray(
				[]immutable.Option[int64]{immutable.Some(int64(1)), immutable.Some(int64(2))}),
			expected: true,
		},
		{
			name: "float nillable array",
			v1: NewNormalNillableFloatArray(
				[]immutable.Option[float64]{immutable.Some(1.0), immutable.Some(2.0)}),
			v2: NewNormalNillableFloatArray(
				[]immutable.Option[float64]{immutable.Some(1.0), immutable.Some(2.0)}),
			expected: true,
		},
		{
			name: "string nillable array",
			v1: NewNormalNillableStringArray(
				[]immutable.Option[string]{immutable.Some("test"), immutable.Some("test2")}),
			v2: NewNormalNillableStringArray(
				[]immutable.Option[string]{immutable.Some("test"), immutable.Some("test2")}),
			expected: true,
		},
		{
			name: "string nillable array not equal",
			v1: NewNormalNillableStringArray(
				[]immutable.Option[string]{immutable.Some("test"), immutable.Some("test2")}),
			v2: NewNormalNillableStringArray(
				[]immutable.Option[string]{immutable.Some("test"), immutable.Some("test3")}),
			expected: false,
		},
		{
			name: "string nillable array with nil",
			v1: NewNormalNillableStringArray(
				[]immutable.Option[string]{immutable.Some("test"), immutable.None[string]()}),
			v2: NewNormalNillableStringArray(
				[]immutable.Option[string]{immutable.Some("test"), immutable.None[string]()}),
			expected: true,
		},
		{
			name: "string nillable array different length",
			v1: NewNormalNillableStringArray(
				[]immutable.Option[string]{immutable.Some("test"), immutable.Some("test2")}),
			v2: NewNormalNillableStringArray(
				[]immutable.Option[string]{immutable.Some("test")}),
			expected: false,
		},
		{
			name: "string nillable array different type",
			v1: NewNormalNillableStringArray(
				[]immutable.Option[string]{immutable.Some("test"), immutable.Some("test2")}),
			v2:       NewNormalStringArray([]string{"test", "test2"}),
			expected: false,
		},
		{
			name: "bytes nillable array",
			v1: NewNormalNillableBytesArray(
				[]immutable.Option[[]byte]{immutable.Some([]byte{1, 2, 3}), immutable.Some([]byte{4, 5, 6})}),
			v2: NewNormalNillableBytesArray(
				[]immutable.Option[[]byte]{immutable.Some([]byte{1, 2, 3}), immutable.Some([]byte{4, 5, 6})}),
			expected: true,
		},
		{
			name: "bytes nillable array not equal",
			v1: NewNormalNillableBytesArray(
				[]immutable.Option[[]byte]{immutable.Some([]byte{1, 2, 3}), immutable.Some([]byte{4, 5, 6})}),
			v2: NewNormalNillableBytesArray(
				[]immutable.Option[[]byte]{immutable.Some([]byte{1, 2, 3}), immutable.Some([]byte{4, 5, 7})}),
			expected: false,
		},
		{
			name: "bytes nillable array with nil",
			v1: NewNormalNillableBytesArray(
				[]immutable.Option[[]byte]{immutable.Some([]byte{1, 2, 3}), immutable.None[[]byte]()}),
			v2: NewNormalNillableBytesArray(
				[]immutable.Option[[]byte]{immutable.Some([]byte{1, 2, 3}), immutable.None[[]byte]()}),
			expected: true,
		},
		{
			name: "bytes nillable array different length",
			v1: NewNormalNillableBytesArray(
				[]immutable.Option[[]byte]{immutable.Some([]byte{1, 2, 3}), immutable.Some([]byte{4, 5, 6})}),
			v2: NewNormalNillableBytesArray(
				[]immutable.Option[[]byte]{immutable.Some([]byte{1, 2, 3})}),
			expected: false,
		},
		{
			name: "bytes nillable array different type",
			v1: NewNormalNillableBytesArray(
				[]immutable.Option[[]byte]{immutable.Some([]byte{1, 2, 3}), immutable.Some([]byte{4, 5, 6})}),
			v2:       NewNormalBytesArray([][]byte{{1, 2, 3}, {4, 5, 6}}),
			expected: false,
		},
		{
			name: "time nillable array",
			v1: NewNormalNillableTimeArray(
				[]immutable.Option[time.Time]{immutable.Some(now), immutable.Some(now)}),
			v2: NewNormalNillableTimeArray(
				[]immutable.Option[time.Time]{immutable.Some(now), immutable.Some(now)}),
			expected: true,
		},
		{
			name: "time nillable array not equal",
			v1: NewNormalNillableTimeArray(
				[]immutable.Option[time.Time]{immutable.Some(now), immutable.Some(now)}),
			v2: NewNormalNillableTimeArray(
				[]immutable.Option[time.Time]{immutable.Some(now), immutable.Some(now.Add(time.Hour))}),
			expected: false,
		},
		{
			name: "time nillable array with nil",
			v1: NewNormalNillableTimeArray(
				[]immutable.Option[time.Time]{immutable.Some(now), immutable.None[time.Time]()}),
			v2: NewNormalNillableTimeArray(
				[]immutable.Option[time.Time]{immutable.Some(now), immutable.None[time.Time]()}),
			expected: true,
		},
		{
			name: "document nillable array",
			v1: NewNormalNillableDocumentArray(
				[]immutable.Option[*Document]{immutable.Some(doc1), immutable.Some(doc2)}),
			v2: NewNormalNillableDocumentArray(
				[]immutable.Option[*Document]{immutable.Some(doc1), immutable.Some(doc2)}),
			expected: true,
		},
		{
			name: "document nillable array not equal",
			v1: NewNormalNillableDocumentArray(
				[]immutable.Option[*Document]{immutable.Some(doc1), immutable.Some(doc2)}),
			v2: NewNormalNillableDocumentArray(
				[]immutable.Option[*Document]{immutable.Some(doc1), immutable.Some(doc3)}),
			expected: false,
		},
		{
			name: "document nillable array with nil",
			v1: NewNormalNillableDocumentArray(
				[]immutable.Option[*Document]{immutable.Some(doc1), immutable.None[*Document]()}),
			v2: NewNormalNillableDocumentArray(
				[]immutable.Option[*Document]{immutable.Some(doc1), immutable.None[*Document]()}),
			expected: true,
		},
		{
			name:     "bool nillable array",
			v1:       NewNormalBoolNillableArray(immutable.Some([]bool{true})),
			v2:       NewNormalBoolNillableArray(immutable.Some([]bool{true})),
			expected: true,
		},
		{
			name:     "bool nillable array equal",
			v1:       NewNormalNillableBoolArray([]immutable.Option[bool]{immutable.Some(true), immutable.Some(false)}),
			v2:       NewNormalNillableBoolArray([]immutable.Option[bool]{immutable.Some(true), immutable.Some(false)}),
			expected: true,
		},
		{
			name:     "bool nillable array not equal",
			v1:       NewNormalNillableBoolArray([]immutable.Option[bool]{immutable.Some(true), immutable.Some(false)}),
			v2:       NewNormalNillableBoolArray([]immutable.Option[bool]{immutable.Some(false), immutable.Some(true)}),
			expected: false,
		},
		{
			name: "bool nillable array with nil values equal",
			v1: NewNormalNillableBoolArray(
				[]immutable.Option[bool]{immutable.Some(true), immutable.None[bool]()}),
			v2: NewNormalNillableBoolArray(
				[]immutable.Option[bool]{immutable.Some(true), immutable.None[bool]()}),
			expected: true,
		},
		{
			name: "bool nillable array with nil values not equal",
			v1: NewNormalNillableBoolArray(
				[]immutable.Option[bool]{immutable.Some(true), immutable.None[bool]()}),
			v2: NewNormalNillableBoolArray(
				[]immutable.Option[bool]{immutable.None[bool](), immutable.Some(true)}),
			expected: false,
		},
		{
			name: "bool nillable array different lengths",
			v1: NewNormalNillableBoolArray(
				[]immutable.Option[bool]{immutable.Some(true), immutable.Some(false)}),
			v2:       NewNormalNillableBoolArray([]immutable.Option[bool]{immutable.Some(true)}),
			expected: false,
		},
		{
			name: "bool nillable array vs non-nillable array",
			v1: NewNormalNillableBoolArray(
				[]immutable.Option[bool]{immutable.Some(true), immutable.Some(false)}),
			v2:       NewNormalBoolArray([]bool{true, false}),
			expected: false,
		},
		{
			name:     "int nillable array",
			v1:       NewNormalIntNillableArray(immutable.Some([]int64{1})),
			v2:       NewNormalIntNillableArray(immutable.Some([]int64{1})),
			expected: true,
		},
		{
			name: "int nillable array equal",
			v1: NewNormalNillableIntArray(
				[]immutable.Option[int64]{immutable.Some(int64(1)), immutable.Some(int64(2))}),
			v2: NewNormalNillableIntArray(
				[]immutable.Option[int64]{immutable.Some(int64(1)), immutable.Some(int64(2))}),
			expected: true,
		},
		{
			name: "int nillable array not equal",
			v1: NewNormalNillableIntArray(
				[]immutable.Option[int64]{immutable.Some(int64(1)), immutable.Some(int64(2))}),
			v2: NewNormalNillableIntArray(
				[]immutable.Option[int64]{immutable.Some(int64(2)), immutable.Some(int64(1))}),
			expected: false,
		},
		{
			name: "int nillable array with nil values equal",
			v1: NewNormalNillableIntArray(
				[]immutable.Option[int64]{immutable.Some(int64(1)), immutable.None[int64]()}),
			v2: NewNormalNillableIntArray(
				[]immutable.Option[int64]{immutable.Some(int64(1)), immutable.None[int64]()}),
			expected: true,
		},
		{
			name: "int nillable array with nil values not equal",
			v1: NewNormalNillableIntArray(
				[]immutable.Option[int64]{immutable.Some(int64(1)), immutable.None[int64]()}),
			v2: NewNormalNillableIntArray(
				[]immutable.Option[int64]{immutable.None[int64](), immutable.Some(int64(1))}),
			expected: false,
		},
		{
			name: "int nillable array different lengths",
			v1: NewNormalNillableIntArray(
				[]immutable.Option[int64]{immutable.Some(int64(1)), immutable.Some(int64(2))}),
			v2:       NewNormalNillableIntArray([]immutable.Option[int64]{immutable.Some(int64(1))}),
			expected: false,
		},
		{
			name: "int nillable array vs non-nillable array",
			v1: NewNormalNillableIntArray(
				[]immutable.Option[int64]{immutable.Some(int64(1)), immutable.Some(int64(2))}),
			v2:       NewNormalIntArray([]int64{1, 2}),
			expected: false,
		},
		{
			name:     "float nillable array",
			v1:       NewNormalFloatNillableArray(immutable.Some([]float64{1.0})),
			v2:       NewNormalFloatNillableArray(immutable.Some([]float64{1.0})),
			expected: true,
		},
		{
			name: "float nillable array equal",
			v1: NewNormalNillableFloatArray(
				[]immutable.Option[float64]{immutable.Some(1.0), immutable.Some(2.0)}),
			v2: NewNormalNillableFloatArray(
				[]immutable.Option[float64]{immutable.Some(1.0), immutable.Some(2.0)}),
			expected: true,
		},
		{
			name: "float nillable array not equal",
			v1: NewNormalNillableFloatArray(
				[]immutable.Option[float64]{immutable.Some(1.0), immutable.Some(2.0)}),
			v2: NewNormalNillableFloatArray(
				[]immutable.Option[float64]{immutable.Some(2.0), immutable.Some(1.0)}),
			expected: false,
		},
		{
			name: "float nillable array with nil values equal",
			v1: NewNormalNillableFloatArray(
				[]immutable.Option[float64]{immutable.Some(1.0), immutable.None[float64]()}),
			v2: NewNormalNillableFloatArray(
				[]immutable.Option[float64]{immutable.Some(1.0), immutable.None[float64]()}),
			expected: true,
		},
		{
			name: "float nillable array with nil values not equal",
			v1: NewNormalNillableFloatArray(
				[]immutable.Option[float64]{immutable.Some(1.0), immutable.None[float64]()}),
			v2: NewNormalNillableFloatArray(
				[]immutable.Option[float64]{immutable.None[float64](), immutable.Some(1.0)}),
			expected: false,
		},
		{
			name: "float nillable array different lengths",
			v1: NewNormalNillableFloatArray(
				[]immutable.Option[float64]{immutable.Some(1.0), immutable.Some(2.0)}),
			v2:       NewNormalNillableFloatArray([]immutable.Option[float64]{immutable.Some(1.0)}),
			expected: false,
		},
		{
			name: "float nillable array vs non-nillable array",
			v1: NewNormalNillableFloatArray(
				[]immutable.Option[float64]{immutable.Some(1.0), immutable.Some(2.0)}),
			v2:       NewNormalFloatArray([]float64{1.0, 2.0}),
			expected: false,
		},
		{
			name:     "string nillable array",
			v1:       NewNormalStringNillableArray(immutable.Some([]string{"test"})),
			v2:       NewNormalStringNillableArray(immutable.Some([]string{"test"})),
			expected: true,
		},
		{
			name:     "string nillable array equal",
			v1:       NewNormalStringNillableArray(immutable.Some([]string{"test"})),
			v2:       NewNormalStringNillableArray(immutable.Some([]string{"test"})),
			expected: true,
		},
		{
			name:     "string nillable array not equal",
			v1:       NewNormalStringNillableArray(immutable.Some([]string{"test"})),
			v2:       NewNormalStringNillableArray(immutable.Some([]string{"different"})),
			expected: false,
		},
		{
			name:     "string nillable array nil vs non-nil",
			v1:       NewNormalStringNillableArray(immutable.Some([]string{"test"})),
			v2:       NewNormalStringNillableArray(immutable.None[[]string]()),
			expected: false,
		},
		{
			name:     "string nillable array different lengths",
			v1:       NewNormalStringNillableArray(immutable.Some([]string{"test", "another"})),
			v2:       NewNormalStringNillableArray(immutable.Some([]string{"test"})),
			expected: false,
		},
		{
			name:     "string nillable array empty strings",
			v1:       NewNormalStringNillableArray(immutable.Some([]string{"", ""})),
			v2:       NewNormalStringNillableArray(immutable.Some([]string{"", ""})),
			expected: true,
		},
		{
			name:     "bytes nillable array",
			v1:       NewNormalBytesNillableArray(immutable.Some([][]byte{{1, 2, 3}})),
			v2:       NewNormalBytesNillableArray(immutable.Some([][]byte{{1, 2, 3}})),
			expected: true,
		},
		{
			name:     "bytes nillable array equal",
			v1:       NewNormalBytesNillableArray(immutable.Some([][]byte{{1, 2, 3}})),
			v2:       NewNormalBytesNillableArray(immutable.Some([][]byte{{1, 2, 3}})),
			expected: true,
		},
		{
			name:     "bytes nillable array not equal",
			v1:       NewNormalBytesNillableArray(immutable.Some([][]byte{{1, 2, 3}})),
			v2:       NewNormalBytesNillableArray(immutable.Some([][]byte{{1, 2, 4}})),
			expected: false,
		},
		{
			name:     "bytes nillable array nil vs non-nil",
			v1:       NewNormalBytesNillableArray(immutable.Some([][]byte{{1, 2, 3}})),
			v2:       NewNormalBytesNillableArray(immutable.None[[][]byte]()),
			expected: false,
		},
		{
			name:     "bytes nillable array different lengths",
			v1:       NewNormalBytesNillableArray(immutable.Some([][]byte{{1, 2, 3}, {4, 5, 6}})),
			v2:       NewNormalBytesNillableArray(immutable.Some([][]byte{{1, 2, 3}})),
			expected: false,
		},
		{
			name:     "bytes nillable array empty slices",
			v1:       NewNormalBytesNillableArray(immutable.Some([][]byte{{}, {}})),
			v2:       NewNormalBytesNillableArray(immutable.Some([][]byte{{}, {}})),
			expected: true,
		},
		{
			name:     "time nillable array",
			v1:       NewNormalTimeNillableArray(immutable.Some([]time.Time{now})),
			v2:       NewNormalTimeNillableArray(immutable.Some([]time.Time{now})),
			expected: true,
		},
		{
			name:     "time nillable array equal",
			v1:       NewNormalTimeNillableArray(immutable.Some([]time.Time{now})),
			v2:       NewNormalTimeNillableArray(immutable.Some([]time.Time{now})),
			expected: true,
		},
		{
			name:     "time nillable array not equal",
			v1:       NewNormalTimeNillableArray(immutable.Some([]time.Time{now})),
			v2:       NewNormalTimeNillableArray(immutable.Some([]time.Time{now.Add(time.Hour)})),
			expected: false,
		},
		{
			name:     "time nillable array nil vs non-nil",
			v1:       NewNormalTimeNillableArray(immutable.Some([]time.Time{now})),
			v2:       NewNormalTimeNillableArray(immutable.None[[]time.Time]()),
			expected: false,
		},
		{
			name:     "time nillable array different lengths",
			v1:       NewNormalTimeNillableArray(immutable.Some([]time.Time{now, now.Add(time.Hour)})),
			v2:       NewNormalTimeNillableArray(immutable.Some([]time.Time{now})),
			expected: false,
		},
		{
			name:     "time nillable array zero times",
			v1:       NewNormalTimeNillableArray(immutable.Some([]time.Time{{}, {}})),
			v2:       NewNormalTimeNillableArray(immutable.Some([]time.Time{{}, {}})),
			expected: true,
		},
		{
			name:     "document nillable array",
			v1:       NewNormalDocumentNillableArray(immutable.Some([]*Document{doc1})),
			v2:       NewNormalDocumentNillableArray(immutable.Some([]*Document{doc1})),
			expected: true,
		},
		{
			name:     "document nillable array equal",
			v1:       NewNormalDocumentNillableArray(immutable.Some([]*Document{doc1})),
			v2:       NewNormalDocumentNillableArray(immutable.Some([]*Document{doc1})),
			expected: true,
		},
		{
			name:     "document nillable array not equal",
			v1:       NewNormalDocumentNillableArray(immutable.Some([]*Document{doc1})),
			v2:       NewNormalDocumentNillableArray(immutable.Some([]*Document{doc2})),
			expected: false,
		},
		{
			name:     "document nillable array nil vs non-nil",
			v1:       NewNormalDocumentNillableArray(immutable.Some([]*Document{doc1})),
			v2:       NewNormalDocumentNillableArray(immutable.None[[]*Document]()),
			expected: false,
		},
		{
			name:     "document nillable array different lengths",
			v1:       NewNormalDocumentNillableArray(immutable.Some([]*Document{doc1, doc2})),
			v2:       NewNormalDocumentNillableArray(immutable.Some([]*Document{doc1})),
			expected: false,
		},
		{
			name:     "document nillable array with nil documents",
			v1:       NewNormalDocumentNillableArray(immutable.Some([]*Document{nil, nil})),
			v2:       NewNormalDocumentNillableArray(immutable.Some([]*Document{nil, nil})),
			expected: true,
		},
		{
			name: "nillable bool nillable array",
			v1: NewNormalNillableBoolNillableArray(
				immutable.Some([]immutable.Option[bool]{immutable.Some(true)})),
			v2: NewNormalNillableBoolNillableArray(
				immutable.Some([]immutable.Option[bool]{immutable.Some(true)})),
			expected: true,
		},
		{
			name: "nillable bool nillable array equal",
			v1: NewNormalNillableBoolNillableArray(
				immutable.Some([]immutable.Option[bool]{immutable.Some(true)})),
			v2: NewNormalNillableBoolNillableArray(
				immutable.Some([]immutable.Option[bool]{immutable.Some(true)})),
			expected: true,
		},
		{
			name: "nillable bool nillable array not equal",
			v1: NewNormalNillableBoolNillableArray(
				immutable.Some([]immutable.Option[bool]{immutable.Some(true)})),
			v2: NewNormalNillableBoolNillableArray(
				immutable.Some([]immutable.Option[bool]{immutable.Some(false)})),
			expected: false,
		},
		{
			name: "nillable bool nillable array outer nil",
			v1: NewNormalNillableBoolNillableArray(
				immutable.Some([]immutable.Option[bool]{immutable.Some(true)})),
			v2:       NewNormalNillableBoolNillableArray(immutable.None[[]immutable.Option[bool]]()),
			expected: false,
		},
		{
			name: "nillable bool nillable array inner nil",
			v1: NewNormalNillableBoolNillableArray(
				immutable.Some([]immutable.Option[bool]{immutable.Some(true), immutable.None[bool]()})),
			v2: NewNormalNillableBoolNillableArray(
				immutable.Some([]immutable.Option[bool]{immutable.Some(true), immutable.None[bool]()})),
			expected: true,
		},
		{
			name: "nillable int nillable array",
			v1: NewNormalNillableIntNillableArray(
				immutable.Some([]immutable.Option[int64]{immutable.Some(int64(1))})),
			v2: NewNormalNillableIntNillableArray(
				immutable.Some([]immutable.Option[int64]{immutable.Some(int64(1))})),
			expected: true,
		},
		{
			name: "nillable int nillable array equal",
			v1: NewNormalNillableIntNillableArray(
				immutable.Some([]immutable.Option[int64]{immutable.Some(int64(1))})),
			v2: NewNormalNillableIntNillableArray(
				immutable.Some([]immutable.Option[int64]{immutable.Some(int64(1))})),
			expected: true,
		},
		{
			name: "nillable int nillable array not equal",
			v1: NewNormalNillableIntNillableArray(
				immutable.Some([]immutable.Option[int64]{immutable.Some(int64(1))})),
			v2: NewNormalNillableIntNillableArray(
				immutable.Some([]immutable.Option[int64]{immutable.Some(int64(2))})),
			expected: false,
		},
		{
			name: "nillable int nillable array outer nil",
			v1: NewNormalNillableIntNillableArray(
				immutable.Some([]immutable.Option[int64]{immutable.Some(int64(1))})),
			v2:       NewNormalNillableIntNillableArray(immutable.None[[]immutable.Option[int64]]()),
			expected: false,
		},
		{
			name: "nillable int nillable array inner nil",
			v1: NewNormalNillableIntNillableArray(
				immutable.Some([]immutable.Option[int64]{immutable.Some(int64(1)), immutable.None[int64]()})),
			v2: NewNormalNillableIntNillableArray(
				immutable.Some([]immutable.Option[int64]{immutable.Some(int64(1)), immutable.None[int64]()})),
			expected: true,
		},
		{
			name: "nillable float nillable array",
			v1: NewNormalNillableFloatNillableArray(
				immutable.Some([]immutable.Option[float64]{immutable.Some(1.0)})),
			v2: NewNormalNillableFloatNillableArray(
				immutable.Some([]immutable.Option[float64]{immutable.Some(1.0)})),
			expected: true,
		},
		{
			name: "nillable float nillable array equal",
			v1: NewNormalNillableFloatNillableArray(
				immutable.Some([]immutable.Option[float64]{immutable.Some(1.0)})),
			v2: NewNormalNillableFloatNillableArray(
				immutable.Some([]immutable.Option[float64]{immutable.Some(1.0)})),
			expected: true,
		},
		{
			name: "nillable float nillable array not equal",
			v1: NewNormalNillableFloatNillableArray(
				immutable.Some([]immutable.Option[float64]{immutable.Some(1.0)})),
			v2: NewNormalNillableFloatNillableArray(
				immutable.Some([]immutable.Option[float64]{immutable.Some(2.0)})),
			expected: false,
		},
		{
			name: "nillable float nillable array outer nil",
			v1: NewNormalNillableFloatNillableArray(
				immutable.Some([]immutable.Option[float64]{immutable.Some(1.0)})),
			v2:       NewNormalNillableFloatNillableArray(immutable.None[[]immutable.Option[float64]]()),
			expected: false,
		},
		{
			name: "nillable float nillable array inner nil",
			v1: NewNormalNillableFloatNillableArray(
				immutable.Some([]immutable.Option[float64]{immutable.Some(1.0), immutable.None[float64]()})),
			v2: NewNormalNillableFloatNillableArray(
				immutable.Some([]immutable.Option[float64]{immutable.Some(1.0), immutable.None[float64]()})),
			expected: true,
		},
		{
			name: "nillable string nillable array",
			v1: NewNormalNillableStringNillableArray(
				immutable.Some([]immutable.Option[string]{immutable.Some("test")})),
			v2: NewNormalNillableStringNillableArray(
				immutable.Some([]immutable.Option[string]{immutable.Some("test")})),
			expected: true,
		},
		{
			name: "nillable string nillable array equal",
			v1: NewNormalNillableStringNillableArray(
				immutable.Some([]immutable.Option[string]{immutable.Some("test")})),
			v2: NewNormalNillableStringNillableArray(
				immutable.Some([]immutable.Option[string]{immutable.Some("test")})),
			expected: true,
		},
		{
			name: "nillable string nillable array not equal",
			v1: NewNormalNillableStringNillableArray(
				immutable.Some([]immutable.Option[string]{immutable.Some("test")})),
			v2: NewNormalNillableStringNillableArray(
				immutable.Some([]immutable.Option[string]{immutable.Some("different")})),
			expected: false,
		},
		{
			name: "nillable string nillable array outer nil",
			v1: NewNormalNillableStringNillableArray(
				immutable.Some([]immutable.Option[string]{immutable.Some("test")})),
			v2:       NewNormalNillableStringNillableArray(immutable.None[[]immutable.Option[string]]()),
			expected: false,
		},
		{
			name: "nillable string nillable array inner nil",
			v1: NewNormalNillableStringNillableArray(
				immutable.Some([]immutable.Option[string]{immutable.Some("test"), immutable.None[string]()})),
			v2: NewNormalNillableStringNillableArray(
				immutable.Some([]immutable.Option[string]{immutable.Some("test"), immutable.None[string]()})),
			expected: true,
		},
		{
			name: "nillable bytes nillable array",
			v1: NewNormalNillableBytesNillableArray(
				immutable.Some([]immutable.Option[[]byte]{immutable.Some([]byte{1, 2, 3})})),
			v2: NewNormalNillableBytesNillableArray(
				immutable.Some([]immutable.Option[[]byte]{immutable.Some([]byte{1, 2, 3})})),
			expected: true,
		},
		{
			name: "nillable bytes nillable array equal",
			v1: NewNormalNillableBytesNillableArray(
				immutable.Some([]immutable.Option[[]byte]{immutable.Some([]byte{1, 2, 3})})),
			v2: NewNormalNillableBytesNillableArray(
				immutable.Some([]immutable.Option[[]byte]{immutable.Some([]byte{1, 2, 3})})),
			expected: true,
		},
		{
			name: "nillable bytes nillable array not equal",
			v1: NewNormalNillableBytesNillableArray(
				immutable.Some([]immutable.Option[[]byte]{immutable.Some([]byte{1, 2, 3})})),
			v2: NewNormalNillableBytesNillableArray(
				immutable.Some([]immutable.Option[[]byte]{immutable.Some([]byte{1, 2, 4})})),
			expected: false,
		},
		{
			name: "nillable bytes nillable array outer nil",
			v1: NewNormalNillableBytesNillableArray(
				immutable.Some([]immutable.Option[[]byte]{immutable.Some([]byte{1, 2, 3})})),
			v2:       NewNormalNillableBytesNillableArray(immutable.None[[]immutable.Option[[]byte]]()),
			expected: false,
		},
		{
			name: "nillable bytes nillable array inner nil",
			v1: NewNormalNillableBytesNillableArray(
				immutable.Some([]immutable.Option[[]byte]{immutable.Some([]byte{1, 2, 3}), immutable.None[[]byte]()})),
			v2: NewNormalNillableBytesNillableArray(
				immutable.Some([]immutable.Option[[]byte]{immutable.Some([]byte{1, 2, 3}), immutable.None[[]byte]()})),
			expected: true,
		},
		{
			name: "nillable time nillable array",
			v1: NewNormalNillableTimeNillableArray(
				immutable.Some([]immutable.Option[time.Time]{immutable.Some(now)})),
			v2: NewNormalNillableTimeNillableArray(
				immutable.Some([]immutable.Option[time.Time]{immutable.Some(now)})),
			expected: true,
		},
		{
			name: "nillable time nillable array equal",
			v1: NewNormalNillableTimeNillableArray(
				immutable.Some([]immutable.Option[time.Time]{immutable.Some(now)})),
			v2: NewNormalNillableTimeNillableArray(
				immutable.Some([]immutable.Option[time.Time]{immutable.Some(now)})),
			expected: true,
		},
		{
			name: "nillable time nillable array not equal",
			v1: NewNormalNillableTimeNillableArray(
				immutable.Some([]immutable.Option[time.Time]{immutable.Some(now)})),
			v2: NewNormalNillableTimeNillableArray(
				immutable.Some([]immutable.Option[time.Time]{immutable.Some(now.Add(time.Hour))})),
			expected: false,
		},
		{
			name: "nillable time nillable array outer nil",
			v1: NewNormalNillableTimeNillableArray(
				immutable.Some([]immutable.Option[time.Time]{immutable.Some(now)})),
			v2:       NewNormalNillableTimeNillableArray(immutable.None[[]immutable.Option[time.Time]]()),
			expected: false,
		},
		{
			name: "nillable time nillable array inner nil",
			v1: NewNormalNillableTimeNillableArray(
				immutable.Some([]immutable.Option[time.Time]{immutable.Some(now), immutable.None[time.Time]()})),
			v2: NewNormalNillableTimeNillableArray(
				immutable.Some([]immutable.Option[time.Time]{immutable.Some(now), immutable.None[time.Time]()})),
			expected: true,
		},
		{
			name: "nillable document nillable array",
			v1: NewNormalNillableDocumentNillableArray(
				immutable.Some([]immutable.Option[*Document]{immutable.Some(doc1)})),
			v2: NewNormalNillableDocumentNillableArray(
				immutable.Some([]immutable.Option[*Document]{immutable.Some(doc1)})),
			expected: true,
		},
		{
			name: "nillable document nillable array equal",
			v1: NewNormalNillableDocumentNillableArray(
				immutable.Some([]immutable.Option[*Document]{immutable.Some(doc1)})),
			v2: NewNormalNillableDocumentNillableArray(
				immutable.Some([]immutable.Option[*Document]{immutable.Some(doc1)})),
			expected: true,
		},
		{
			name: "nillable document nillable array not equal",
			v1: NewNormalNillableDocumentNillableArray(
				immutable.Some([]immutable.Option[*Document]{immutable.Some(doc1)})),
			v2: NewNormalNillableDocumentNillableArray(
				immutable.Some([]immutable.Option[*Document]{immutable.Some(doc2)})),
			expected: false,
		},
		{
			name: "nillable document nillable array outer nil",
			v1: NewNormalNillableDocumentNillableArray(
				immutable.Some([]immutable.Option[*Document]{immutable.Some(doc1)})),
			v2:       NewNormalNillableDocumentNillableArray(immutable.None[[]immutable.Option[*Document]]()),
			expected: false,
		},
		{
			name: "nillable document nillable array inner nil",
			v1: NewNormalNillableDocumentNillableArray(
				immutable.Some([]immutable.Option[*Document]{immutable.Some(doc1), immutable.None[*Document]()})),
			v2: NewNormalNillableDocumentNillableArray(
				immutable.Some([]immutable.Option[*Document]{immutable.Some(doc1), immutable.None[*Document]()})),
			expected: true,
		},
	}

	for _, tt := range tests {
		tStr := string(tt.name)
		t.Run(tStr, func(t *testing.T) {
			actual := tt.v1.IsEqual(tt.v2)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
