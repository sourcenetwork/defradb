// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package core

import (
	"fmt"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
)

// DecodeFieldValue takes a field value and description and converts it to the
// standardized Defra Go type.
func DecodeFieldValue(fieldDesc client.FieldDescription, val any) (any, error) {
	if val == nil {
		return nil, nil
	}

	var err error
	if array, isArray := val.([]any); isArray {
		var ok bool
		switch fieldDesc.Kind {
		case client.FieldKind_BOOL_ARRAY:
			boolArray := make([]bool, len(array))
			for i, untypedValue := range array {
				boolArray[i], ok = untypedValue.(bool)
				if !ok {
					return nil, client.NewErrUnexpectedType[bool](fieldDesc.Name, untypedValue)
				}
			}
			val = boolArray

		case client.FieldKind_NILLABLE_BOOL_ARRAY:
			val, err = convertNillableArray[bool](fieldDesc.Name, array)
			if err != nil {
				return nil, err
			}

		case client.FieldKind_INT_ARRAY:
			intArray := make([]int64, len(array))
			for i, untypedValue := range array {
				intArray[i], err = convertToInt(fmt.Sprintf("%s[%v]", fieldDesc.Name, i), untypedValue)
				if err != nil {
					return nil, err
				}
			}
			val = intArray

		case client.FieldKind_NILLABLE_INT_ARRAY:
			val, err = convertNillableArrayWithConverter(fieldDesc.Name, array, convertToInt)
			if err != nil {
				return nil, err
			}

		case client.FieldKind_FLOAT_ARRAY:
			floatArray := make([]float64, len(array))
			for i, untypedValue := range array {
				floatArray[i], ok = untypedValue.(float64)
				if !ok {
					return nil, client.NewErrUnexpectedType[float64](fieldDesc.Name, untypedValue)
				}
			}
			val = floatArray

		case client.FieldKind_NILLABLE_FLOAT_ARRAY:
			val, err = convertNillableArray[float64](fieldDesc.Name, array)
			if err != nil {
				return nil, err
			}

		case client.FieldKind_STRING_ARRAY:
			stringArray := make([]string, len(array))
			for i, untypedValue := range array {
				stringArray[i], ok = untypedValue.(string)
				if !ok {
					return nil, client.NewErrUnexpectedType[string](fieldDesc.Name, untypedValue)
				}
			}
			val = stringArray

		case client.FieldKind_NILLABLE_STRING_ARRAY:
			val, err = convertNillableArray[string](fieldDesc.Name, array)
			if err != nil {
				return nil, err
			}
		}
	} else { // CBOR often encodes values typed as floats as ints
		switch fieldDesc.Kind {
		case client.FieldKind_FLOAT:
			switch v := val.(type) {
			case int64:
				return float64(v), nil
			case int:
				return float64(v), nil
			case uint64:
				return float64(v), nil
			case uint:
				return float64(v), nil
			}
		case client.FieldKind_INT:
			switch v := val.(type) {
			case float64:
				if v >= 0 {
					return uint64(v), nil
				}
				return int64(v), nil
			}
		}
	}

	return val, nil
}

func convertNillableArray[T any](propertyName string, items []any) ([]immutable.Option[T], error) {
	resultArray := make([]immutable.Option[T], len(items))
	for i, untypedValue := range items {
		if untypedValue == nil {
			resultArray[i] = immutable.None[T]()
			continue
		}
		value, ok := untypedValue.(T)
		if !ok {
			return nil, client.NewErrUnexpectedType[T](fmt.Sprintf("%s[%v]", propertyName, i), untypedValue)
		}
		resultArray[i] = immutable.Some(value)
	}
	return resultArray, nil
}

func convertNillableArrayWithConverter[TOut any](
	propertyName string,
	items []any,
	converter func(propertyName string, in any) (TOut, error),
) ([]immutable.Option[TOut], error) {
	resultArray := make([]immutable.Option[TOut], len(items))
	for i, untypedValue := range items {
		if untypedValue == nil {
			resultArray[i] = immutable.None[TOut]()
			continue
		}
		value, err := converter(fmt.Sprintf("%s[%v]", propertyName, i), untypedValue)
		if err != nil {
			return nil, err
		}
		resultArray[i] = immutable.Some(value)
	}
	return resultArray, nil
}

func convertToInt(propertyName string, untypedValue any) (int64, error) {
	switch value := untypedValue.(type) {
	case uint64:
		return int64(value), nil
	case int64:
		return value, nil
	case float64:
		return int64(value), nil
	default:
		return 0, client.NewErrUnexpectedType[string](propertyName, untypedValue)
	}
}
