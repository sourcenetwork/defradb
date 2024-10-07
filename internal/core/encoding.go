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
	"time"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/encoding"
)

// NormalizeFieldValue takes a field value and description and converts it to the
// standardized Defra Go type.
func NormalizeFieldValue(fieldDesc client.FieldDefinition, val any) (any, error) {
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

		case client.FieldKind_NILLABLE_JSON:
			return convertToJSON(fieldDesc.Name, val)
		}
	} else { // CBOR often encodes values typed as floats as ints
		switch fieldDesc.Kind {
		case client.FieldKind_NILLABLE_FLOAT:
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
		case client.FieldKind_NILLABLE_INT:
			switch v := val.(type) {
			case float64:
				return int64(v), nil
			case int64:
				return int64(v), nil
			case int:
				return int64(v), nil
			case uint64:
				return int64(v), nil
			case uint:
				return int64(v), nil
			}
		case client.FieldKind_NILLABLE_DATETIME:
			switch v := val.(type) {
			case string:
				return time.Parse(time.RFC3339, v)
			}
		case client.FieldKind_NILLABLE_BOOL:
			switch v := val.(type) {
			case int64:
				return v != 0, nil
			}
		case client.FieldKind_NILLABLE_STRING:
			switch v := val.(type) {
			case []byte:
				return string(v), nil
			}
		case client.FieldKind_NILLABLE_JSON:
			return convertToJSON(fieldDesc.Name, val)
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

// convertToJSON converts the given value to a valid JSON value.
//
// When maps are decoded, they are of type map[any]any, and need to
// be converted to map[string]any. All other values are valid JSON.
func convertToJSON(propertyName string, untypedValue any) (any, error) {
	switch t := untypedValue.(type) {
	case map[any]any:
		resultValue := make(map[string]any)
		for k, v := range t {
			key, ok := k.(string)
			if !ok {
				return nil, client.NewErrUnexpectedType[string](propertyName, k)
			}
			val, err := convertToJSON(fmt.Sprintf("%s.%s", propertyName, key), v)
			if err != nil {
				return nil, err
			}
			resultValue[key] = val
		}
		return resultValue, nil

	case []any:
		resultValue := make([]any, len(t))
		for i, v := range t {
			val, err := convertToJSON(fmt.Sprintf("%s[%d]", propertyName, i), v)
			if err != nil {
				return nil, err
			}
			resultValue[i] = val
		}
		return resultValue, nil

	default:
		return untypedValue, nil
	}
}

// DecodeIndexDataStoreKey decodes a IndexDataStoreKey from bytes.
// It expects the input bytes is in the following format:
//
// /[CollectionID]/[IndexID]/[FieldValue](/[FieldValue]...)
//
// Where [CollectionID] and [IndexID] are integers
//
// All values of the fields are converted to standardized Defra Go type
// according to fields description.
func DecodeIndexDataStoreKey(
	data []byte,
	indexDesc *client.IndexDescription,
	fields []client.FieldDefinition,
) (IndexDataStoreKey, error) {
	if len(data) == 0 {
		return IndexDataStoreKey{}, ErrEmptyKey
	}

	if data[0] != '/' {
		return IndexDataStoreKey{}, ErrInvalidKey
	}
	data = data[1:]

	data, colID, err := encoding.DecodeUvarintAscending(data)
	if err != nil {
		return IndexDataStoreKey{}, err
	}

	key := IndexDataStoreKey{CollectionID: uint32(colID)}

	if data[0] != '/' {
		return IndexDataStoreKey{}, ErrInvalidKey
	}
	data = data[1:]

	data, indID, err := encoding.DecodeUvarintAscending(data)
	if err != nil {
		return IndexDataStoreKey{}, err
	}
	key.IndexID = uint32(indID)

	if len(data) == 0 {
		return key, nil
	}

	for len(data) > 0 {
		if data[0] != '/' {
			return IndexDataStoreKey{}, ErrInvalidKey
		}
		data = data[1:]

		i := len(key.Fields)
		descending := false
		var kind client.FieldKind = client.FieldKind_DocID
		// If the key has more values encoded then fields on the index description, the last
		// value must be the docID and we treat it as a string.
		if i < len(indexDesc.Fields) {
			descending = indexDesc.Fields[i].Descending
			kind = fields[i].Kind
		} else if i > len(indexDesc.Fields) {
			return IndexDataStoreKey{}, ErrInvalidKey
		}

		if kind != nil && kind.IsArray() {
			if arrKind, ok := kind.(client.ScalarArrayKind); ok {
				kind = arrKind.SubKind()
			}
		}

		var val client.NormalValue
		data, val, err = encoding.DecodeFieldValue(data, descending, kind)
		if err != nil {
			return IndexDataStoreKey{}, err
		}

		key.Fields = append(key.Fields, IndexedField{Value: val, Descending: descending})
	}

	return key, nil
}

// EncodeIndexDataStoreKey encodes a IndexDataStoreKey to bytes to be stored as a key
// for secondary indexes.
func EncodeIndexDataStoreKey(key *IndexDataStoreKey) []byte {
	if key.CollectionID == 0 {
		return []byte{}
	}

	b := encoding.EncodeUvarintAscending([]byte{'/'}, uint64(key.CollectionID))

	if key.IndexID == 0 {
		return b
	}
	b = append(b, '/')
	b = encoding.EncodeUvarintAscending(b, uint64(key.IndexID))

	for _, field := range key.Fields {
		b = append(b, '/')
		b = encoding.EncodeFieldValue(b, field.Value, field.Descending)
	}

	return b
}

// DecodeDataStoreKey decodes a store key into a [DataStoreKey].
func DecodeDataStoreKey(data []byte) (DataStoreKey, error) {
	if len(data) == 0 {
		return DataStoreKey{}, ErrEmptyKey
	}

	if data[0] != '/' {
		return DataStoreKey{}, ErrInvalidKey
	}
	data = data[1:]

	data, colRootID, err := encoding.DecodeUvarintAscending(data)
	if err != nil {
		return DataStoreKey{}, err
	}

	var instanceType InstanceType
	if len(data) > 1 {
		if data[0] == '/' {
			data = data[1:]
		}
		instanceType = InstanceType(data[0])
		data = data[1:]
	}

	const docKeyLength int = 40
	var docID string
	if len(data) > docKeyLength {
		if data[0] == '/' {
			data = data[1:]
		}
		docID = string(data[:docKeyLength])
		data = data[docKeyLength:]
	}

	var fieldID string
	if len(data) > 1 {
		if data[0] == '/' {
			data = data[1:]
		}
		// Todo: This should be encoded/decoded properly in
		// https://github.com/sourcenetwork/defradb/issues/2818
		fieldID = string(data)
	}

	return DataStoreKey{
		CollectionRootID: uint32(colRootID),
		InstanceType:     (instanceType),
		DocID:            docID,
		FieldID:          fieldID,
	}, nil
}

// EncodeDataStoreKey encodes a [*DataStoreKey] to a byte array suitable for sorting in the store.
func EncodeDataStoreKey(key *DataStoreKey) []byte {
	var result []byte

	if key.CollectionRootID != 0 {
		result = encoding.EncodeUvarintAscending([]byte{'/'}, uint64(key.CollectionRootID))
	}

	if key.InstanceType != "" {
		result = append(result, '/')
		result = append(result, []byte(string(key.InstanceType))...)
	}

	if key.DocID != "" {
		result = append(result, '/')
		result = append(result, []byte(key.DocID)...)
	}

	if key.FieldID != "" {
		result = append(result, '/')
		// Todo: This should be encoded/decoded properly in
		// https://github.com/sourcenetwork/defradb/issues/2818
		result = append(result, []byte(key.FieldID)...)
	}

	return result
}
