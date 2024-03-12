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
	"time"

	"github.com/sourcenetwork/immutable"
	"golang.org/x/exp/constraints"
)

type NormalValue struct {
	value any
}

func (val NormalValue) Value() any {
	return val.value
}

func (val NormalValue) IsNil() bool {
	return val.value == nil
}

func (val NormalValue) IsBool() bool {
	_, ok := val.value.(bool)
	return ok
}

func (val NormalValue) IsInt() bool {
	_, ok := val.value.(int64)
	return ok
}

func (val NormalValue) IsFloat() bool {
	_, ok := val.value.(float64)
	return ok
}

func (val NormalValue) IsString() bool {
	_, ok := val.value.(string)
	return ok
}

func (val NormalValue) IsBytes() bool {
	_, ok := val.value.([]byte)
	return ok
}

func (val NormalValue) IsTime() bool {
	_, ok := val.value.(time.Time)
	return ok
}

func (val NormalValue) IsBoolArray() bool {
	_, ok := val.value.([]bool)
	return ok
}

func (val NormalValue) IsIntArray() bool {
	_, ok := val.value.([]int64)
	return ok
}

func (val NormalValue) IsFloatArray() bool {
	_, ok := val.value.([]float64)
	return ok
}

func (val NormalValue) IsStringArray() bool {
	_, ok := val.value.([]string)
	return ok
}

func (val NormalValue) IsBytesArray() bool {
	_, ok := val.value.([][]byte)
	return ok
}

func (val NormalValue) IsTimeArray() bool {
	_, ok := val.value.([]time.Time)
	return ok
}

func (val NormalValue) IsNillableBoolArray() bool {
	_, ok := val.value.([]immutable.Option[bool])
	return ok
}

func (val NormalValue) IsNillableIntArray() bool {
	_, ok := val.value.([]immutable.Option[int64])
	return ok
}

func (val NormalValue) IsNillableFloatArray() bool {
	_, ok := val.value.([]immutable.Option[float64])
	return ok
}

func (val NormalValue) IsNillableStringArray() bool {
	_, ok := val.value.([]immutable.Option[string])
	return ok
}

func (val NormalValue) IsNillableBytesArray() bool {
	_, ok := val.value.([]immutable.Option[[]byte])
	return ok
}

func (val NormalValue) IsNillableTimeArray() bool {
	_, ok := val.value.([]immutable.Option[time.Time])
	return ok
}

func (val NormalValue) IsArray() bool {
	switch val.value.(type) {
	case []bool, []int64, []float64, []string, [][]byte, []time.Time:
		return true
	default:
		return false
	}
}

func (val NormalValue) IsNillableArray() bool {
	switch val.value.(type) {
	case []immutable.Option[bool],
		[]immutable.Option[int64],
		[]immutable.Option[float64],
		[]immutable.Option[string],
		[]immutable.Option[[]byte],
		[]immutable.Option[time.Time]:
		return true
	default:
		return false
	}
}

func (val NormalValue) IsAnyArray() bool {
	return val.IsArray() || val.IsNillableArray()
}

func (val NormalValue) Bool() bool {
	return val.value.(bool)
}

func (val NormalValue) Int() int64 {
	return val.value.(int64)
}

func (val NormalValue) Float() float64 {
	return val.value.(float64)
}

func (val NormalValue) String() string {
	return val.value.(string)
}

func (val NormalValue) Bytes() []byte {
	return val.value.([]byte)
}

func (val NormalValue) Time() time.Time {
	return val.value.(time.Time)
}

func (val NormalValue) Array() []any {
	switch v := val.value.(type) {
	case []bool:
		return toAnyArray(v)
	case []int64:
		return toAnyArray(v)
	case []float64:
		return toAnyArray(v)
	case []string:
		return toAnyArray(v)
	case [][]byte:
		return toAnyArray(v)
	case []time.Time:
		return toAnyArray(v)
	}
	return nil
}

func (val NormalValue) ArrayOfNormalValues() []NormalValue {
	switch v := val.value.(type) {
	case []bool:
		return toNormalArray(v, NewBoolNormalValue)
	case []string:
		return toNormalArray(v, NewStringNormalValue)
	case []int64:
		return toNormalArray(v, NewIntNormalValue)
	case []float64:
		return toNormalArray(v, NewFloatNormalValue)
	case []time.Time:
		return toNormalArray(v, NewTimeNormalValue)
	case [][]byte:
		return toNormalArray(v, NewBytesNormalValue)
	}
	return nil
}

func toNormalArray[T any](val []T, f func(T) NormalValue) []NormalValue {
	res := make([]NormalValue, len(val))
	for i := range val {
		res[i] = f(val[i])
	}
	return res
}

func (val NormalValue) NillableArray() []immutable.Option[any] {
	switch v := val.value.(type) {
	case []immutable.Option[bool]:
		return toAnyNillableArray(v)
	case []immutable.Option[string]:
		return toAnyNillableArray(v)
	case []immutable.Option[int64]:
		return toAnyNillableArray(v)
	case []immutable.Option[float64]:
		return toAnyNillableArray(v)
	case []immutable.Option[time.Time]:
		return toAnyNillableArray(v)
	case []immutable.Option[[]byte]:
		return toAnyNillableArray(v)
	}
	return nil
}

func toAnyArray[T any](val []T) []any {
	res := make([]any, len(val))
	for i := range val {
		res[i] = val[i]
	}
	return res
}

func toAnyNillableArray[T any](val []immutable.Option[T]) []immutable.Option[any] {
	res := make([]immutable.Option[any], len(val))
	for i := range val {
		if val[i].HasValue() {
			res[i] = immutable.Some[any](val[i].Value())
		} else {
			res[i] = immutable.None[any]()
		}
	}
	return res
}

func (val NormalValue) BoolArray() []bool {
	return val.value.([]bool)
}

func (val NormalValue) IntArray() []int64 {
	return val.value.([]int64)
}

func (val NormalValue) FloatArray() []float64 {
	return val.value.([]float64)
}

func (val NormalValue) StringArray() []string {
	return val.value.([]string)
}

func (val NormalValue) BytesArray() [][]byte {
	return val.value.([][]byte)
}

func (val NormalValue) TimeArray() []time.Time {
	return val.value.([]time.Time)
}

func (val NormalValue) NillableBoolArray() []immutable.Option[bool] {
	return val.value.([]immutable.Option[bool])
}

func (val NormalValue) NillableIntArray() []immutable.Option[int64] {
	return val.value.([]immutable.Option[int64])
}

func (val NormalValue) NillableFloatArray() []immutable.Option[float64] {
	return val.value.([]immutable.Option[float64])
}

func (val NormalValue) NillableStringArray() []immutable.Option[string] {
	return val.value.([]immutable.Option[string])
}

func (val NormalValue) NillableBytesArray() []immutable.Option[[]byte] {
	return val.value.([]immutable.Option[[]byte])
}

func (val NormalValue) NillableTimeArray() []immutable.Option[time.Time] {
	return val.value.([]immutable.Option[time.Time])
}

func NewNormalValue(val any) (NormalValue, error) {
	if val == nil {
		return NormalValue{}, nil
	}
	switch v := val.(type) {
	case bool, int64, float64, string, []byte, time.Time, []bool, []int64, []float64, []string, [][]byte,
		[]immutable.Option[bool], []immutable.Option[int64], []immutable.Option[float64], []immutable.Option[string],
		[]immutable.Option[[]byte], []immutable.Option[time.Time]:
		return NormalValue{value: val}, nil
	case int8:
		return NormalValue{int64(v)}, nil
	case int16:
		return NormalValue{int64(v)}, nil
	case int32:
		return NormalValue{int64(v)}, nil
	case int:
		return NormalValue{int64(v)}, nil
	case uint8:
		return NormalValue{int64(v)}, nil
	case uint16:
		return NormalValue{int64(v)}, nil
	case uint32:
		return NormalValue{int64(v)}, nil
	case uint64:
		return NormalValue{int64(v)}, nil
	case uint:
		return NormalValue{int64(v)}, nil
	case float32:
		return NormalValue{value: float64(v)}, nil
	case []int8:
		return NewIntArrayNormalValue(v), nil
	case []int16:
		return NewIntArrayNormalValue(v), nil
	case []int32:
		return NewIntArrayNormalValue(v), nil
	case []int:
		return NewIntArrayNormalValue(v), nil
	case []uint16:
		return NewIntArrayNormalValue(v), nil
	case []uint32:
		return NewIntArrayNormalValue(v), nil
	case []uint64:
		return NewIntArrayNormalValue(v), nil
	case []uint:
		return NewIntArrayNormalValue(v), nil
	case []float32:
		return NewFloatArrayNormalValue(v), nil
	case []immutable.Option[int8]:
		return NewNillableIntArrayNormalValue(v), nil
	case []immutable.Option[int16]:
		return NewNillableIntArrayNormalValue(v), nil
	case []immutable.Option[int32]:
		return NewNillableIntArrayNormalValue(v), nil
	case []immutable.Option[int]:
		return NewNillableIntArrayNormalValue(v), nil
	case []immutable.Option[uint8]:
		return NewNillableIntArrayNormalValue(v), nil
	case []immutable.Option[uint16]:
		return NewNillableIntArrayNormalValue(v), nil
	case []immutable.Option[uint32]:
		return NewNillableIntArrayNormalValue(v), nil
	case []immutable.Option[uint64]:
		return NewNillableIntArrayNormalValue(v), nil
	case []immutable.Option[uint]:
		return NewNillableIntArrayNormalValue(v), nil
	case []immutable.Option[float32]:
		return NewNillableFloatArrayNormalValue(v), nil
	case []any:
		if len(v) == 0 {
			return NormalValue{}, NewCanNotNormalizeValue(val)
		}
		first, err := NewNormalValue(v[0])
		if err != nil {
			return NormalValue{}, err
		}
		switch {
		case first.IsBool():
			return convertAnyArrToTypedArr[bool](v)
		case first.IsInt():
			return convertAnyArrToIntOrFloatArr(v)
		case first.IsFloat():
			return convertAnyArrToFloatArr(v)
		case first.IsString():
			return convertAnyArrToTypedArr[string](v)
		case first.IsBytes():
			return convertAnyArrToTypedArr[[]byte](v)
		case first.IsTime():
			return convertAnyArrToTypedArr[time.Time](v)
		}
	}
	return NormalValue{}, NewCanNotNormalizeValue(val)
}

func convertAnyArrToIntOrFloatArr(arr []any) (NormalValue, error) {
	result := make([]int64, len(arr))
	for i := range arr {
		switch v := arr[i].(type) {
		case int64:
			result[i] = v
		case float64, float32:
			return convertAnyArrToFloatArr(arr)
		case int8:
			result[i] = int64(v)
		case int16:
			result[i] = int64(v)
		case int32:
			result[i] = int64(v)
		case int:
			result[i] = int64(v)
		case uint8:
			result[i] = int64(v)
		case uint16:
			result[i] = int64(v)
		case uint32:
			result[i] = int64(v)
		case uint64:
			result[i] = int64(v)
		case uint:
			result[i] = int64(v)
		default:
			return NormalValue{}, NewCanNotNormalizeValue(arr)
		}
	}
	return NormalValue{value: result}, nil
}

func convertAnyArrToFloatArr(arr []any) (NormalValue, error) {
	result := make([]float64, len(arr))
	for i := range arr {
		if v, ok := arr[i].(float64); ok {
			result[i] = v
		} else if v, ok := arr[i].(float32); ok {
			result[i] = float64(v)
		} else {
			return NormalValue{}, NewCanNotNormalizeValue(arr)
		}
	}
	return NormalValue{value: result}, nil
}

func convertAnyArrToTypedArr[T any](arr []any) (NormalValue, error) {
	result := make([]T, len(arr))
	for i := range arr {
		if v, ok := arr[i].(T); ok {
			result[i] = v
		} else {
			return NormalValue{}, NewCanNotNormalizeValue(arr)
		}
	}
	return NormalValue{value: result}, nil
}

func NewNilNormalValue() NormalValue {
	return NormalValue{}
}

func NewBoolNormalValue(val bool) NormalValue {
	return NormalValue{value: val}
}

func NewIntNormalValue[T constraints.Integer | constraints.Float](val T) NormalValue {
	return NormalValue{value: int64(val)}
}

func NewFloatNormalValue[T constraints.Integer | constraints.Float](val T) NormalValue {
	return NormalValue{value: float64(val)}
}

func NewStringNormalValue[T string | []byte](val T) NormalValue {
	return NormalValue{value: string(val)}
}

func NewBytesNormalValue[T string | []byte](val T) NormalValue {
	return NormalValue{value: []byte(val)}
}

func NewTimeNormalValue(val time.Time) NormalValue {
	return NormalValue{value: val}
}

func NewBoolArrayNormalValue(val []bool) NormalValue {
	return NormalValue{value: val}
}

func NewIntArrayNormalValue[T constraints.Integer | constraints.Float](val []T) NormalValue {
	return normalizeNumberArr[int64](val)
}

func NewFloatArrayNormalValue[T constraints.Integer | constraints.Float](val []T) NormalValue {
	return normalizeNumberArr[float64](val)
}

func NewStringArrayNormalValue[T string | []byte](val []T) NormalValue {
	return normalizeCharsArr[string](val)
}

func NewBytesArrayNormalValue[T string | []byte](val []T) NormalValue {
	return normalizeCharsArr[[]byte](val)
}

func NewTimeArrayNormalValue(val []time.Time) NormalValue {
	return NormalValue{value: val}
}

func NewNillableBoolArrayNormalValue(val []immutable.Option[bool]) NormalValue {
	return NormalValue{value: val}
}

func NewNillableIntArrayNormalValue[T constraints.Integer | constraints.Float](val []immutable.Option[T]) NormalValue {
	return normalizeNillableNumberArr[int64](val)
}

func NewNillableFloatArrayNormalValue[T constraints.Integer | constraints.Float](
	val []immutable.Option[T],
) NormalValue {
	return normalizeNillableNumberArr[float64](val)
}

func NewNillableStringArrayNormalValue[T string | []byte](val []immutable.Option[T]) NormalValue {
	return normalizeNillableCharsArr[string](val)
}

func NewNillableBytesArrayNormalValue[T string | []byte](val []immutable.Option[T]) NormalValue {
	return normalizeNillableCharsArr[[]byte](val)
}

func NewNillableTimeArrayNormalValue(val []immutable.Option[time.Time]) NormalValue {
	return NormalValue{value: val}
}

func normalizeCharsArr[R string | []byte, T string | []byte](val []T) NormalValue {
	var v any = val
	if resultChars, ok := v.(R); ok {
		return NormalValue{value: resultChars}
	}
	resultChars := make([]R, len(val))
	for i, v := range val {
		resultChars[i] = R(v)
	}
	return NormalValue{value: resultChars}
}

func normalizeNillableCharsArr[R string | []byte, T string | []byte](val []immutable.Option[T]) NormalValue {
	var v any = val
	if resultChars, ok := v.(R); ok {
		return NormalValue{value: resultChars}
	}
	resultChars := make([]immutable.Option[R], len(val))
	for i, v := range val {
		if v.HasValue() {
			resultChars[i] = immutable.Some(R(v.Value()))
		} else {
			resultChars[i] = immutable.None[R]()
		}
	}
	return NormalValue{value: resultChars}
}

func normalizeNumberArr[R constraints.Integer | constraints.Float, T constraints.Integer | constraints.Float](
	val []T,
) NormalValue {
	var v any = val
	if numArr, ok := v.([]R); ok {
		return NormalValue{value: numArr}
	}
	numArr := make([]R, len(val))
	for i, v := range val {
		numArr[i] = R(v)
	}
	return NormalValue{value: numArr}
}

func normalizeNillableNumberArr[R constraints.Integer | constraints.Float, T constraints.Integer | constraints.Float](
	val []immutable.Option[T],
) NormalValue {
	var v any = val
	if numArr, ok := v.([]R); ok {
		return NormalValue{value: numArr}
	}
	numArr := make([]immutable.Option[R], len(val))
	for i, v := range val {
		if v.HasValue() {
			numArr[i] = immutable.Some(R(v.Value()))
		} else {
			numArr[i] = immutable.None[R]()
		}
	}
	return NormalValue{value: numArr}
}
