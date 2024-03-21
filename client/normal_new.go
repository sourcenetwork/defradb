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
)

// NewNormalValue creates a new NormalValue from the given value.
// It will normalize all known types that can be converted to normal ones.
// For example, is the given type is `[]int32`, it will be converted to `[]int64`.
// If the given value is of type `[]any` is will to through every element and try to convert it
// minimal common type and normalizes it.
// For examples, the following conversions will be made:
//   - `[]any{int32(1), int64(2)}` -> `[]int64{1, 2}`.
//   - `[]any{int32(1), int64(2), float32(1.5)}` -> `[]float64{1.0, 2.0, 1.5}`.
//   - `[]any{int32(1), nil}` -> `[]immutable.Option[int64]{immutable.Some(1), immutable.None[int64]()}`.
//
// This function will not check if the given value is `nil`. To normalize a `nil` value use the
// `NewNormalNil` function.
func NewNormalValue(val any) (NormalValue, error) {
	switch v := val.(type) {
	case bool:
		return NewNormalBool(v), nil
	case int8:
		return newNormalInt(int64(v)), nil
	case int16:
		return newNormalInt(int64(v)), nil
	case int32:
		return newNormalInt(int64(v)), nil
	case int64:
		return newNormalInt(v), nil
	case int:
		return newNormalInt(int64(v)), nil
	case uint8:
		return newNormalInt(int64(v)), nil
	case uint16:
		return newNormalInt(int64(v)), nil
	case uint32:
		return newNormalInt(int64(v)), nil
	case uint64:
		return newNormalInt(int64(v)), nil
	case uint:
		return newNormalInt(int64(v)), nil
	case float32:
		return newNormalFloat(float64(v)), nil
	case float64:
		return newNormalFloat(v), nil
	case string:
		return NewNormalString(v), nil
	case []byte:
		return NewNormalBytes(v), nil
	case time.Time:
		return NewNormalTime(v), nil
	case *Document:
		return NewNormalDocument(v), nil

	case immutable.Option[bool]:
		return NewNormalNillableBool(v), nil
	case immutable.Option[int8]:
		return NewNormalNillableInt(v), nil
	case immutable.Option[int16]:
		return NewNormalNillableInt(v), nil
	case immutable.Option[int32]:
		return NewNormalNillableInt(v), nil
	case immutable.Option[int64]:
		return NewNormalNillableInt(v), nil
	case immutable.Option[int]:
		return NewNormalNillableInt(v), nil
	case immutable.Option[uint8]:
		return NewNormalNillableInt(v), nil
	case immutable.Option[uint16]:
		return NewNormalNillableInt(v), nil
	case immutable.Option[uint32]:
		return NewNormalNillableInt(v), nil
	case immutable.Option[uint64]:
		return NewNormalNillableInt(v), nil
	case immutable.Option[uint]:
		return NewNormalNillableInt(v), nil
	case immutable.Option[float32]:
		return NewNormalNillableFloat(v), nil
	case immutable.Option[float64]:
		return NewNormalNillableFloat(v), nil
	case immutable.Option[string]:
		return NewNormalNillableString(v), nil
	case immutable.Option[[]byte]:
		return NewNormalNillableBytes(v), nil
	case immutable.Option[time.Time]:
		return NewNormalNillableTime(v), nil
	case immutable.Option[*Document]:
		return NewNormalNillableDocument(v), nil

	case []bool:
		return NewNormalBoolArray(v), nil
	case []int8:
		return NewNormalIntArray(v), nil
	case []int16:
		return NewNormalIntArray(v), nil
	case []int32:
		return NewNormalIntArray(v), nil
	case []int64:
		return NewNormalIntArray(v), nil
	case []int:
		return NewNormalIntArray(v), nil
	case []uint16:
		return NewNormalIntArray(v), nil
	case []uint32:
		return NewNormalIntArray(v), nil
	case []uint64:
		return NewNormalIntArray(v), nil
	case []uint:
		return NewNormalIntArray(v), nil
	case []float32:
		return NewNormalFloatArray(v), nil
	case []float64:
		return NewNormalFloatArray(v), nil
	case []string:
		return NewNormalStringArray(v), nil
	case [][]byte:
		return NewNormalBytesArray(v), nil
	case []time.Time:
		return NewNormalTimeArray(v), nil
	case []*Document:
		return NewNormalDocumentArray(v), nil

	case []immutable.Option[bool]:
		return NewNormalNillableBoolArray(v), nil
	case []immutable.Option[int8]:
		return NewNormalNillableIntArray(v), nil
	case []immutable.Option[int16]:
		return NewNormalNillableIntArray(v), nil
	case []immutable.Option[int32]:
		return NewNormalNillableIntArray(v), nil
	case []immutable.Option[int64]:
		return NewNormalNillableIntArray(v), nil
	case []immutable.Option[int]:
		return NewNormalNillableIntArray(v), nil
	case []immutable.Option[uint8]:
		return NewNormalNillableIntArray(v), nil
	case []immutable.Option[uint16]:
		return NewNormalNillableIntArray(v), nil
	case []immutable.Option[uint32]:
		return NewNormalNillableIntArray(v), nil
	case []immutable.Option[uint64]:
		return NewNormalNillableIntArray(v), nil
	case []immutable.Option[uint]:
		return NewNormalNillableIntArray(v), nil
	case []immutable.Option[float32]:
		return NewNormalNillableFloatArray(v), nil
	case []immutable.Option[float64]:
		return NewNormalNillableFloatArray(v), nil
	case []immutable.Option[string]:
		return NewNormalNillableStringArray(v), nil
	case []immutable.Option[[]byte]:
		return NewNormalNillableBytesArray(v), nil
	case []immutable.Option[time.Time]:
		return NewNormalNillableTimeArray(v), nil
	case []immutable.Option[*Document]:
		return NewNormalNillableDocumentArray(v), nil

	case immutable.Option[[]bool]:
		return NewNormalBoolNillableArray(v), nil
	case immutable.Option[[]int8]:
		return NewNormalIntNillableArray(v), nil
	case immutable.Option[[]int16]:
		return NewNormalIntNillableArray(v), nil
	case immutable.Option[[]int32]:
		return NewNormalIntNillableArray(v), nil
	case immutable.Option[[]int64]:
		return NewNormalIntNillableArray(v), nil
	case immutable.Option[[]int]:
		return NewNormalIntNillableArray(v), nil
	case immutable.Option[[]uint16]:
		return NewNormalIntNillableArray(v), nil
	case immutable.Option[[]uint32]:
		return NewNormalIntNillableArray(v), nil
	case immutable.Option[[]uint64]:
		return NewNormalIntNillableArray(v), nil
	case immutable.Option[[]uint]:
		return NewNormalIntNillableArray(v), nil
	case immutable.Option[[]float32]:
		return NewNormalFloatNillableArray(v), nil
	case immutable.Option[[]float64]:
		return NewNormalFloatNillableArray(v), nil
	case immutable.Option[[]string]:
		return NewNormalStringNillableArray(v), nil
	case immutable.Option[[][]byte]:
		return NewNormalBytesNillableArray(v), nil
	case immutable.Option[[]time.Time]:
		return NewNormalTimeNillableArray(v), nil
	case immutable.Option[[]*Document]:
		return NewNormalDocumentNillableArray(v), nil

	case immutable.Option[[]immutable.Option[bool]]:
		return NewNormalNillableBoolNillableArray(v), nil
	case immutable.Option[[]immutable.Option[int8]]:
		return NewNormalNillableIntNillableArray(v), nil
	case immutable.Option[[]immutable.Option[int16]]:
		return NewNormalNillableIntNillableArray(v), nil
	case immutable.Option[[]immutable.Option[int32]]:
		return NewNormalNillableIntNillableArray(v), nil
	case immutable.Option[[]immutable.Option[int64]]:
		return NewNormalNillableIntNillableArray(v), nil
	case immutable.Option[[]immutable.Option[int]]:
		return NewNormalNillableIntNillableArray(v), nil
	case immutable.Option[[]immutable.Option[uint8]]:
		return NewNormalNillableIntNillableArray(v), nil
	case immutable.Option[[]immutable.Option[uint16]]:
		return NewNormalNillableIntNillableArray(v), nil
	case immutable.Option[[]immutable.Option[uint32]]:
		return NewNormalNillableIntNillableArray(v), nil
	case immutable.Option[[]immutable.Option[uint64]]:
		return NewNormalNillableIntNillableArray(v), nil
	case immutable.Option[[]immutable.Option[uint]]:
		return NewNormalNillableIntNillableArray(v), nil
	case immutable.Option[[]immutable.Option[float32]]:
		return NewNormalNillableFloatNillableArray(v), nil
	case immutable.Option[[]immutable.Option[float64]]:
		return NewNormalNillableFloatNillableArray(v), nil
	case immutable.Option[[]immutable.Option[string]]:
		return NewNormalNillableStringNillableArray(v), nil
	case immutable.Option[[]immutable.Option[[]byte]]:
		return NewNormalNillableBytesNillableArray(v), nil
	case immutable.Option[[]immutable.Option[time.Time]]:
		return NewNormalNillableTimeNillableArray(v), nil
	case immutable.Option[[]immutable.Option[*Document]]:
		return NewNormalNillableDocumentNillableArray(v), nil

	case []any:
		if len(v) == 0 {
			return nil, NewCanNotNormalizeValue(val)
		}
		first, err := NewNormalValue(v[0])
		if err != nil {
			return nil, err
		}
		if _, ok := first.Bool(); ok {
			return convertAnyArrToTypedArr[bool](v, NewNormalBoolArray, NewNormalNillableBoolArray)
		}
		if _, ok := first.Int(); ok {
			return convertAnyArrToIntOrFloatArr(v)
		}
		if _, ok := first.Float(); ok {
			return convertAnyArrToFloatArr(v)
		}
		if _, ok := first.String(); ok {
			return convertAnyArrToTypedArr[string](v, NewNormalStringArray, NewNormalNillableStringArray)
		}
		if _, ok := first.Bytes(); ok {
			return convertAnyArrToTypedArr[[]byte](v, NewNormalBytesArray, NewNormalNillableBytesArray)
		}
		if _, ok := first.Time(); ok {
			return convertAnyArrToTypedArr[time.Time](v, NewNormalTimeArray, NewNormalNillableTimeArray)
		}
		if _, ok := first.Document(); ok {
			return convertAnyArrToTypedArr[*Document](v, NewNormalDocumentArray, NewNormalNillableDocumentArray)
		}
	}
	return nil, NewCanNotNormalizeValue(val)
}

func convertAnyArrToIntOrFloatArr(arr []any) (NormalValue, error) {
	result := make([]int64, len(arr))
	for i := range arr {
		if arr[i] == nil {
			return convertAnyArrToNillableIntOrFloatArr(arr)
		}
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
			return nil, NewCanNotNormalizeValue(arr)
		}
	}
	return NewNormalIntArray(result), nil
}

func convertAnyArrToNillableIntOrFloatArr(arr []any) (NormalValue, error) {
	result := make([]immutable.Option[int64], len(arr))
	for i := range arr {
		if arr[i] == nil {
			result[i] = immutable.None[int64]()
			continue
		}
		var intVal int64
		switch v := arr[i].(type) {
		case int64:
			intVal = v
		case float64, float32:
			return convertAnyArrToFloatArr(arr)
		case int8:
			intVal = int64(v)
		case int16:
			intVal = int64(v)
		case int32:
			intVal = int64(v)
		case int:
			intVal = int64(v)
		case uint8:
			intVal = int64(v)
		case uint16:
			intVal = int64(v)
		case uint32:
			intVal = int64(v)
		case uint64:
			intVal = int64(v)
		case uint:
			intVal = int64(v)
		default:
			return nil, NewCanNotNormalizeValue(arr)
		}
		result[i] = immutable.Some(intVal)
	}
	return NewNormalNillableIntArray(result), nil
}

func convertAnyArrToFloatArr(arr []any) (NormalValue, error) {
	result := make([]float64, len(arr))
	for i := range arr {
		if arr[i] == nil {
			return convertAnyArrToNillableFloatArr(arr)
		}

		var floatVal float64
		switch v := arr[i].(type) {
		case float64:
			floatVal = v
		case float32:
			floatVal = float64(v)
		case int8:
			floatVal = float64(v)
		case int16:
			floatVal = float64(v)
		case int32:
			floatVal = float64(v)
		case int64:
			floatVal = float64(v)
		case int:
			floatVal = float64(v)
		case uint8:
			floatVal = float64(v)
		case uint16:
			floatVal = float64(v)
		case uint32:
			floatVal = float64(v)
		case uint64:
			floatVal = float64(v)
		case uint:
			floatVal = float64(v)
		default:
			return nil, NewCanNotNormalizeValue(arr)
		}
		result[i] = floatVal
	}
	return NewNormalFloatArray(result), nil
}

func convertAnyArrToNillableFloatArr(arr []any) (NormalValue, error) {
	result := make([]immutable.Option[float64], len(arr))
	for i := range arr {
		if arr[i] == nil {
			result[i] = immutable.None[float64]()
			continue
		}
		var floatVal float64
		switch v := arr[i].(type) {
		case float64:
			floatVal = v
		case float32:
			floatVal = float64(v)
		case int8:
			floatVal = float64(v)
		case int16:
			floatVal = float64(v)
		case int32:
			floatVal = float64(v)
		case int64:
			floatVal = float64(v)
		case int:
			floatVal = float64(v)
		case uint8:
			floatVal = float64(v)
		case uint16:
			floatVal = float64(v)
		case uint32:
			floatVal = float64(v)
		case uint64:
			floatVal = float64(v)
		case uint:
			floatVal = float64(v)
		default:
			return nil, NewCanNotNormalizeValue(arr)
		}
		result[i] = immutable.Some(floatVal)
	}
	return NewNormalNillableFloatArray(result), nil
}

func convertAnyArrToTypedArr[T any](
	arr []any,
	newNormalArr func([]T) NormalValue,
	newNormalNillableArr func([]immutable.Option[T]) NormalValue,
) (NormalValue, error) {
	result := make([]T, len(arr))
	for i := range arr {
		if arr[i] == nil {
			return convertAnyArrToNillableTypedArr[T](arr, newNormalNillableArr)
		}
		if v, ok := arr[i].(T); ok {
			result[i] = v
		} else {
			return nil, NewCanNotNormalizeValue(arr)
		}
	}
	return newNormalArr(result), nil
}

func convertAnyArrToNillableTypedArr[T any](
	arr []any,
	newNormalNillableArr func([]immutable.Option[T]) NormalValue,
) (NormalValue, error) {
	result := make([]immutable.Option[T], len(arr))
	for i := range arr {
		if arr[i] == nil {
			result[i] = immutable.None[T]()
			continue
		}
		if v, ok := arr[i].(T); ok {
			result[i] = immutable.Some(v)
		} else {
			return nil, NewCanNotNormalizeValue(arr)
		}
	}
	return newNormalNillableArr(result), nil
}
