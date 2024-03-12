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
	"testing"
	"time"

	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/assert"
)

type nType string

const (
	NilType             nType = "Nil"
	BoolType            nType = "Bool"
	IntType             nType = "Int"
	FloatType           nType = "Float"
	StringType          nType = "String"
	BytesType           nType = "Bytes"
	TimeType            nType = "Time"
	BoolArray           nType = "BoolArray"
	IntArray            nType = "IntArray"
	FloatArray          nType = "FloatArray"
	StringArray         nType = "StringArray"
	BytesArray          nType = "BytesArray"
	TimeArray           nType = "TimeArray"
	NillableBoolArray   nType = "NillableBoolArray"
	NillableIntArray    nType = "NillableIntArray"
	NillableFloatArray  nType = "NillableFloatArray"
	NillableStringArray nType = "NillableStringArray"
	NillableBytesArray  nType = "NillableBytesArray"
	NillableTimeArray   nType = "NillableTimeArray"
)

func TestNormalValue_New_Is_Value(t *testing.T) {
	isMap := map[nType]func(NormalValue) bool{
		NilType:             func(v NormalValue) bool { return v.IsNil() },
		BoolType:            func(v NormalValue) bool { return v.IsBool() },
		IntType:             func(v NormalValue) bool { return v.IsInt() },
		FloatType:           func(v NormalValue) bool { return v.IsFloat() },
		StringType:          func(v NormalValue) bool { return v.IsString() },
		BytesType:           func(v NormalValue) bool { return v.IsBytes() },
		TimeType:            func(v NormalValue) bool { return v.IsTime() },
		BoolArray:           func(v NormalValue) bool { return v.IsBoolArray() },
		IntArray:            func(v NormalValue) bool { return v.IsIntArray() },
		FloatArray:          func(v NormalValue) bool { return v.IsFloatArray() },
		StringArray:         func(v NormalValue) bool { return v.IsStringArray() },
		BytesArray:          func(v NormalValue) bool { return v.IsBytesArray() },
		TimeArray:           func(v NormalValue) bool { return v.IsTimeArray() },
		NillableBoolArray:   func(v NormalValue) bool { return v.IsNillableBoolArray() },
		NillableIntArray:    func(v NormalValue) bool { return v.IsNillableIntArray() },
		NillableFloatArray:  func(v NormalValue) bool { return v.IsNillableFloatArray() },
		NillableStringArray: func(v NormalValue) bool { return v.IsNillableStringArray() },
		NillableBytesArray:  func(v NormalValue) bool { return v.IsNillableBytesArray() },
		NillableTimeArray:   func(v NormalValue) bool { return v.IsNillableTimeArray() },
	}

	getValMap := map[nType]func(NormalValue) any{
		NilType:             func(v NormalValue) any { return nil },
		BoolType:            func(v NormalValue) any { return v.Bool() },
		IntType:             func(v NormalValue) any { return v.Int() },
		FloatType:           func(v NormalValue) any { return v.Float() },
		StringType:          func(v NormalValue) any { return v.String() },
		BytesType:           func(v NormalValue) any { return v.Bytes() },
		TimeType:            func(v NormalValue) any { return v.Time() },
		BoolArray:           func(v NormalValue) any { return v.BoolArray() },
		IntArray:            func(v NormalValue) any { return v.IntArray() },
		FloatArray:          func(v NormalValue) any { return v.FloatArray() },
		StringArray:         func(v NormalValue) any { return v.StringArray() },
		BytesArray:          func(v NormalValue) any { return v.BytesArray() },
		TimeArray:           func(v NormalValue) any { return v.TimeArray() },
		NillableBoolArray:   func(v NormalValue) any { return v.NillableBoolArray() },
		NillableIntArray:    func(v NormalValue) any { return v.NillableIntArray() },
		NillableFloatArray:  func(v NormalValue) any { return v.NillableFloatArray() },
		NillableStringArray: func(v NormalValue) any { return v.NillableStringArray() },
		NillableBytesArray:  func(v NormalValue) any { return v.NillableBytesArray() },
		NillableTimeArray:   func(v NormalValue) any { return v.NillableTimeArray() },
	}

	newMap := map[nType]func(any) NormalValue{
		NilType:             func(v any) NormalValue { return NewNilNormalValue() },
		BoolType:            func(v any) NormalValue { return NewBoolNormalValue(v.(bool)) },
		IntType:             func(v any) NormalValue { return NewIntNormalValue(v.(int64)) },
		FloatType:           func(v any) NormalValue { return NewFloatNormalValue(v.(float64)) },
		StringType:          func(v any) NormalValue { return NewStringNormalValue(v.(string)) },
		BytesType:           func(v any) NormalValue { return NewBytesNormalValue(v.([]byte)) },
		TimeType:            func(v any) NormalValue { return NewTimeNormalValue(v.(time.Time)) },
		BoolArray:           func(v any) NormalValue { return NewBoolArrayNormalValue(v.([]bool)) },
		IntArray:            func(v any) NormalValue { return NewIntArrayNormalValue(v.([]int64)) },
		FloatArray:          func(v any) NormalValue { return NewFloatArrayNormalValue(v.([]float64)) },
		StringArray:         func(v any) NormalValue { return NewStringArrayNormalValue(v.([]string)) },
		BytesArray:          func(v any) NormalValue { return NewBytesArrayNormalValue(v.([][]byte)) },
		TimeArray:           func(v any) NormalValue { return NewTimeArrayNormalValue(v.([]time.Time)) },
		NillableBoolArray:   func(v any) NormalValue { return NewNillableBoolArrayNormalValue(v.([]immutable.Option[bool])) },
		NillableIntArray:    func(v any) NormalValue { return NewNillableIntArrayNormalValue(v.([]immutable.Option[int64])) },
		NillableFloatArray:  func(v any) NormalValue { return NewNillableFloatArrayNormalValue(v.([]immutable.Option[float64])) },
		NillableStringArray: func(v any) NormalValue { return NewNillableStringArrayNormalValue(v.([]immutable.Option[string])) },
		NillableBytesArray:  func(v any) NormalValue { return NewNillableBytesArrayNormalValue(v.([]immutable.Option[[]byte])) },
		NillableTimeArray:   func(v any) NormalValue { return NewNillableTimeArrayNormalValue(v.([]immutable.Option[time.Time])) },
	}

	tests := []struct {
		nType              nType
		input              any
		isNillableArray    bool
		isNotNillableArray bool
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
			nType:              BoolArray,
			input:              []bool{true, false},
			isNotNillableArray: true,
		},
		{
			nType:              IntArray,
			input:              []int64{1, 2, 3},
			isNotNillableArray: true,
		},
		{
			nType:              FloatArray,
			input:              []float64{1, 2, 3},
			isNotNillableArray: true,
		},
		{
			nType:              StringArray,
			input:              []string{"test", "test2"},
			isNotNillableArray: true,
		},
		{
			nType:              BytesArray,
			input:              [][]byte{{1, 2, 3}, {4, 5, 6}},
			isNotNillableArray: true,
		},
		{
			nType:              TimeArray,
			input:              []time.Time{time.Now(), time.Now()},
			isNotNillableArray: true,
		},
		{
			nType:           NillableBoolArray,
			input:           []immutable.Option[bool]{immutable.Some(true)},
			isNillableArray: true,
		},
		{
			nType:           NillableIntArray,
			input:           []immutable.Option[int64]{immutable.Some(int64(1))},
			isNillableArray: true,
		},
		{
			nType:           NillableFloatArray,
			input:           []immutable.Option[float64]{immutable.Some(float64(1))},
			isNillableArray: true,
		},
		{
			nType:           NillableStringArray,
			input:           []immutable.Option[string]{immutable.Some("test")},
			isNillableArray: true,
		},
		{
			nType:           NillableBytesArray,
			input:           []immutable.Option[[]byte]{immutable.Some([]byte{1, 2, 3})},
			isNillableArray: true,
		},
		{
			nType:           NillableTimeArray,
			input:           []immutable.Option[time.Time]{immutable.Some(time.Now())},
			isNillableArray: true,
		},
	}

	for _, tt := range tests {
		tStr := string(tt.nType)
		t.Run(tStr, func(t *testing.T) {
			actual := NormalValue{value: tt.input}
			for nType, isFunc := range isMap {
				if nType == tt.nType {
					assert.True(t, isFunc(actual), "Is"+tStr+"() should return true")
					val := getValMap[nType](actual)
					assert.Equal(t, tt.input, val, tStr+"() returned unexpected value")
					newVal := newMap[nType](val)
					assert.Equal(t, actual, newVal, "New"+tStr+"() returned unexpected NormalValue")
				} else {
					assert.False(t, isFunc(actual), "Is"+string(nType)+"() should return false for "+tStr)
				}
			}
			if tt.isNillableArray {
				assert.True(t, actual.IsNillableArray(), "IsNillableArray() should return true for "+tStr)
				assert.False(t, actual.IsArray(), "IsArray() should return false for "+tStr)
			}
			if tt.isNotNillableArray {
				assert.False(t, actual.IsNillableArray(), "IsNillableArray() should return false for "+tStr)
				assert.True(t, actual.IsArray(), "IsArray() should return true for "+tStr)
			}
			if tt.isNillableArray || tt.isNotNillableArray {
				assert.True(t, actual.IsAnyArray(), "IsAnyArray() should return true for "+tStr)
			}
		})
	}
}
