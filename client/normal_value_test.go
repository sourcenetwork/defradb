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
		NilType:     func(v any) NormalValue { return NewNilNormalValue() },
		BoolType:    func(v any) NormalValue { return NewBoolNormalValue(v.(bool)) },
		IntType:     func(v any) NormalValue { return NewIntNormalValue(v.(int64)) },
		FloatType:   func(v any) NormalValue { return NewFloatNormalValue(v.(float64)) },
		StringType:  func(v any) NormalValue { return NewStringNormalValue(v.(string)) },
		BytesType:   func(v any) NormalValue { return NewBytesNormalValue(v.([]byte)) },
		TimeType:    func(v any) NormalValue { return NewTimeNormalValue(v.(time.Time)) },
		BoolArray:   func(v any) NormalValue { return NewBoolArrayNormalValue(v.([]bool)) },
		IntArray:    func(v any) NormalValue { return NewIntArrayNormalValue(v.([]int64)) },
		FloatArray:  func(v any) NormalValue { return NewFloatArrayNormalValue(v.([]float64)) },
		StringArray: func(v any) NormalValue { return NewStringArrayNormalValue(v.([]string)) },
		BytesArray:  func(v any) NormalValue { return NewBytesArrayNormalValue(v.([][]byte)) },
		TimeArray:   func(v any) NormalValue { return NewTimeArrayNormalValue(v.([]time.Time)) },
		NillableBoolArray: func(v any) NormalValue {
			return NewNillableBoolArrayNormalValue(v.([]immutable.Option[bool]))
		},
		NillableIntArray: func(v any) NormalValue {
			return NewNillableIntArrayNormalValue(v.([]immutable.Option[int64]))
		},
		NillableFloatArray: func(v any) NormalValue {
			return NewNillableFloatArrayNormalValue(v.([]immutable.Option[float64]))
		},
		NillableStringArray: func(v any) NormalValue {
			return NewNillableStringArrayNormalValue(v.([]immutable.Option[string]))
		},
		NillableBytesArray: func(v any) NormalValue {
			return NewNillableBytesArrayNormalValue(v.([]immutable.Option[[]byte]))
		},
		NillableTimeArray: func(v any) NormalValue {
			return NewNillableTimeArrayNormalValue(v.([]immutable.Option[time.Time]))
		},
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

func TestNormalValue_NewIntNormalValue(t *testing.T) {
	i64 := int64(2)
	v := NewIntNormalValue(i64)
	assert.Equal(t, i64, v.Int())

	v = NewIntNormalValue(float32(2.5))
	assert.Equal(t, i64, v.Int())

	v = NewIntNormalValue(float64(2.5))
	assert.Equal(t, i64, v.Int())

	v = NewIntNormalValue(int8(2))
	assert.Equal(t, i64, v.Int())

	v = NewIntNormalValue(int16(2))
	assert.Equal(t, i64, v.Int())

	v = NewIntNormalValue(int32(2))
	assert.Equal(t, i64, v.Int())

	v = NewIntNormalValue(int(2))
	assert.Equal(t, i64, v.Int())

	v = NewIntNormalValue(uint8(2))
	assert.Equal(t, i64, v.Int())

	v = NewIntNormalValue(uint16(2))
	assert.Equal(t, i64, v.Int())

	v = NewIntNormalValue(uint32(2))
	assert.Equal(t, i64, v.Int())

	v = NewIntNormalValue(uint64(2))
	assert.Equal(t, i64, v.Int())

	v = NewIntNormalValue(uint(2))
	assert.Equal(t, i64, v.Int())
}

func TestNormalValue_NewFloatNormalValue(t *testing.T) {
	f64Frac := float64(2.5)
	f64 := float64(2)

	v := NewFloatNormalValue(f64Frac)
	assert.Equal(t, f64Frac, v.Float())

	v = NewFloatNormalValue(float32(2.5))
	assert.Equal(t, f64Frac, v.Float())

	v = NewFloatNormalValue(int8(2))
	assert.Equal(t, f64, v.Float())

	v = NewFloatNormalValue(int16(2))
	assert.Equal(t, f64, v.Float())

	v = NewFloatNormalValue(int32(2))
	assert.Equal(t, f64, v.Float())

	v = NewFloatNormalValue(int64(2))
	assert.Equal(t, f64, v.Float())

	v = NewFloatNormalValue(int(2))
	assert.Equal(t, f64, v.Float())

	v = NewFloatNormalValue(uint8(2))
	assert.Equal(t, f64, v.Float())

	v = NewFloatNormalValue(uint16(2))
	assert.Equal(t, f64, v.Float())

	v = NewFloatNormalValue(uint32(2))
	assert.Equal(t, f64, v.Float())

	v = NewFloatNormalValue(uint64(2))
	assert.Equal(t, f64, v.Float())

	v = NewFloatNormalValue(uint(2))
	assert.Equal(t, f64, v.Float())
}

func TestNormalValue_NewStringNormalValue(t *testing.T) {
	strInput := "str"

	v := NewStringNormalValue(strInput)
	assert.Equal(t, strInput, v.String())

	v = NewStringNormalValue([]byte{'s', 't', 'r'})
	assert.Equal(t, strInput, v.String())
}

func TestNormalValue_NewBytesNormalValue(t *testing.T) {
	bytesInput := []byte("str")

	v := NewBytesNormalValue(bytesInput)
	assert.Equal(t, bytesInput, v.Bytes())

	v = NewBytesNormalValue("str")
	assert.Equal(t, bytesInput, v.Bytes())
}

func TestNormalValue_NewIntArrayNormalValue(t *testing.T) {
	i64Input := []int64{2}

	v := NewIntArrayNormalValue(i64Input)
	assert.Equal(t, i64Input, v.IntArray())

	v = NewIntArrayNormalValue([]float32{2.5})
	assert.Equal(t, i64Input, v.IntArray())

	v = NewIntArrayNormalValue([]int8{2})
	assert.Equal(t, i64Input, v.IntArray())

	v = NewIntArrayNormalValue([]int16{2})
	assert.Equal(t, i64Input, v.IntArray())

	v = NewIntArrayNormalValue([]int32{2})
	assert.Equal(t, i64Input, v.IntArray())

	v = NewIntArrayNormalValue([]int64{2})
	assert.Equal(t, i64Input, v.IntArray())

	v = NewIntArrayNormalValue([]int{2})
	assert.Equal(t, i64Input, v.IntArray())

	v = NewIntArrayNormalValue([]uint8{2})
	assert.Equal(t, i64Input, v.IntArray())

	v = NewIntArrayNormalValue([]uint16{2})
	assert.Equal(t, i64Input, v.IntArray())

	v = NewIntArrayNormalValue([]uint32{2})
	assert.Equal(t, i64Input, v.IntArray())

	v = NewIntArrayNormalValue([]uint64{2})
	assert.Equal(t, i64Input, v.IntArray())

	v = NewIntArrayNormalValue([]uint{2})
	assert.Equal(t, i64Input, v.IntArray())
}

func TestNormalValue_NewFloatArrayNormalValue(t *testing.T) {
	f64InputFrac := []float64{2.5}
	f64Input := []float64{2.0}

	v := NewFloatArrayNormalValue(f64InputFrac)
	assert.Equal(t, f64InputFrac, v.FloatArray())

	v = NewFloatArrayNormalValue([]float32{2.5})
	assert.Equal(t, f64InputFrac, v.FloatArray())

	v = NewFloatArrayNormalValue([]int8{2})
	assert.Equal(t, f64Input, v.FloatArray())

	v = NewFloatArrayNormalValue([]int16{2})
	assert.Equal(t, f64Input, v.FloatArray())

	v = NewFloatArrayNormalValue([]int32{2})
	assert.Equal(t, f64Input, v.FloatArray())

	v = NewFloatArrayNormalValue([]int64{2})
	assert.Equal(t, f64Input, v.FloatArray())

	v = NewFloatArrayNormalValue([]int{2})
	assert.Equal(t, f64Input, v.FloatArray())

	v = NewFloatArrayNormalValue([]uint8{2})
	assert.Equal(t, f64Input, v.FloatArray())

	v = NewFloatArrayNormalValue([]uint16{2})
	assert.Equal(t, f64Input, v.FloatArray())

	v = NewFloatArrayNormalValue([]uint32{2})
	assert.Equal(t, f64Input, v.FloatArray())

	v = NewFloatArrayNormalValue([]uint64{2})
	assert.Equal(t, f64Input, v.FloatArray())

	v = NewFloatArrayNormalValue([]uint{2})
	assert.Equal(t, f64Input, v.FloatArray())
}

func TestNormalValue_NewStringArrayNormalValue(t *testing.T) {
	strInput := []string{"str"}

	v := NewStringArrayNormalValue(strInput)
	assert.Equal(t, strInput, v.StringArray())

	v = NewStringArrayNormalValue([][]byte{{'s', 't', 'r'}})
	assert.Equal(t, strInput, v.StringArray())
}

func TestNormalValue_NewBytesArrayNormalValue(t *testing.T) {
	bytesInput := [][]byte{[]byte("str")}

	v := NewBytesArrayNormalValue(bytesInput)
	assert.Equal(t, bytesInput, v.BytesArray())

	v = NewBytesArrayNormalValue([]string{"str"})
	assert.Equal(t, bytesInput, v.BytesArray())
}

func TestNormalValue_NewNillableFloatArrayNormalValue(t *testing.T) {
	f64InputFrac := []immutable.Option[float64]{immutable.Some(2.5)}
	f64Input := []immutable.Option[float64]{immutable.Some(2.0)}

	v := NewNillableFloatArrayNormalValue(f64InputFrac)
	assert.Equal(t, f64InputFrac, v.NillableFloatArray())

	v = NewNillableFloatArrayNormalValue([]immutable.Option[float32]{immutable.Some[float32](2.5)})
	assert.Equal(t, f64InputFrac, v.NillableFloatArray())

	v = NewNillableFloatArrayNormalValue([]immutable.Option[int8]{immutable.Some[int8](2)})
	assert.Equal(t, f64Input, v.NillableFloatArray())

	v = NewNillableFloatArrayNormalValue([]immutable.Option[int16]{immutable.Some[int16](2)})
	assert.Equal(t, f64Input, v.NillableFloatArray())

	v = NewNillableFloatArrayNormalValue([]immutable.Option[int32]{immutable.Some[int32](2)})
	assert.Equal(t, f64Input, v.NillableFloatArray())

	v = NewNillableFloatArrayNormalValue([]immutable.Option[int64]{immutable.Some[int64](2)})
	assert.Equal(t, f64Input, v.NillableFloatArray())

	v = NewNillableFloatArrayNormalValue([]immutable.Option[int]{immutable.Some[int](2)})
	assert.Equal(t, f64Input, v.NillableFloatArray())

	v = NewNillableFloatArrayNormalValue([]immutable.Option[uint8]{immutable.Some[uint8](2)})
	assert.Equal(t, f64Input, v.NillableFloatArray())

	v = NewNillableFloatArrayNormalValue([]immutable.Option[uint16]{immutable.Some[uint16](2)})
	assert.Equal(t, f64Input, v.NillableFloatArray())

	v = NewNillableFloatArrayNormalValue([]immutable.Option[uint32]{immutable.Some[uint32](2)})
	assert.Equal(t, f64Input, v.NillableFloatArray())

	v = NewNillableFloatArrayNormalValue([]immutable.Option[uint64]{immutable.Some[uint64](2)})
	assert.Equal(t, f64Input, v.NillableFloatArray())

	v = NewNillableFloatArrayNormalValue([]immutable.Option[uint]{immutable.Some[uint](2)})
	assert.Equal(t, f64Input, v.NillableFloatArray())
}

func TestNormalValue_NewNillableIntArrayNormalValue(t *testing.T) {
	i64Input := []immutable.Option[int64]{immutable.Some[int64](2)}

	v := NewNillableIntArrayNormalValue(i64Input)
	assert.Equal(t, i64Input, v.NillableIntArray())

	v = NewNillableIntArrayNormalValue([]immutable.Option[float32]{immutable.Some[float32](2.5)})
	assert.Equal(t, i64Input, v.NillableIntArray())

	v = NewNillableIntArrayNormalValue([]immutable.Option[float64]{immutable.Some[float64](2.5)})
	assert.Equal(t, i64Input, v.NillableIntArray())

	v = NewNillableIntArrayNormalValue([]immutable.Option[int8]{immutable.Some[int8](2)})
	assert.Equal(t, i64Input, v.NillableIntArray())

	v = NewNillableIntArrayNormalValue([]immutable.Option[int16]{immutable.Some[int16](2)})
	assert.Equal(t, i64Input, v.NillableIntArray())

	v = NewNillableIntArrayNormalValue([]immutable.Option[int32]{immutable.Some[int32](2)})
	assert.Equal(t, i64Input, v.NillableIntArray())

	v = NewNillableIntArrayNormalValue([]immutable.Option[int]{immutable.Some[int](2)})
	assert.Equal(t, i64Input, v.NillableIntArray())

	v = NewNillableIntArrayNormalValue([]immutable.Option[uint8]{immutable.Some[uint8](2)})
	assert.Equal(t, i64Input, v.NillableIntArray())

	v = NewNillableIntArrayNormalValue([]immutable.Option[uint16]{immutable.Some[uint16](2)})
	assert.Equal(t, i64Input, v.NillableIntArray())

	v = NewNillableIntArrayNormalValue([]immutable.Option[uint32]{immutable.Some[uint32](2)})
	assert.Equal(t, i64Input, v.NillableIntArray())

	v = NewNillableIntArrayNormalValue([]immutable.Option[uint64]{immutable.Some[uint64](2)})
	assert.Equal(t, i64Input, v.NillableIntArray())

	v = NewNillableIntArrayNormalValue([]immutable.Option[uint]{immutable.Some[uint](2)})
	assert.Equal(t, i64Input, v.NillableIntArray())
}

func TestNormalValue_NewNillableStringArrayNormalValue(t *testing.T) {
	strInput := []immutable.Option[string]{immutable.Some("str")}

	v := NewNillableStringArrayNormalValue(strInput)
	assert.Equal(t, strInput, v.NillableStringArray())

	v = NewNillableStringArrayNormalValue([]immutable.Option[[]byte]{immutable.Some[[]byte]([]byte{'s', 't', 'r'})})
	assert.Equal(t, strInput, v.NillableStringArray())
}

func TestNormalValue_NewNillableBytesArrayNormalValue(t *testing.T) {
	bytesInput := []immutable.Option[[]byte]{immutable.Some[[]byte]([]byte("str"))}

	v := NewNillableBytesArrayNormalValue(bytesInput)
	assert.Equal(t, bytesInput, v.NillableBytesArray())

	v = NewNillableBytesArrayNormalValue([]immutable.Option[string]{immutable.Some("str")})
	assert.Equal(t, bytesInput, v.NillableBytesArray())
}
