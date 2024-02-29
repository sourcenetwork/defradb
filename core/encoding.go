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
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/encoding"
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

		i := len(key.fields)
		descending := false
		if i < len(indexDesc.Fields) {
			descending = indexDesc.Fields[i].Descending
		} else if i > len(indexDesc.Fields) {
			return IndexDataStoreKey{}, ErrInvalidKey
		}

		var val any
		data, val, err = encoding.DecodeFieldValue(data, descending)
		if err != nil {
			return IndexDataStoreKey{}, err
		}

		key.fields = append(key.fields, IndexedField{Value: val, Descending: descending})
	}

	err = normalizeIndexDataStoreKeyValues(&key, fields)
	return key, err
}

// normalizeIndexDataStoreKeyValues converts all field values  to standardized
// Defra Go type according to fields description.
func normalizeIndexDataStoreKeyValues(key *IndexDataStoreKey, fields []client.FieldDefinition) error {
	for i := range key.fields {
		if key.fields[i].Value == nil {
			continue
		}
		var err error
		var val any
		if i == len(key.fields)-1 && len(key.fields)-len(fields) == 1 {
			bytes, ok := key.fields[i].Value.([]byte)
			if !ok {
				return client.NewErrUnexpectedType[[]byte](request.DocIDArgName, key.fields[i].Value)
			}
			val = string(bytes)
		} else {
			val, err = NormalizeFieldValue(fields[i], key.fields[i].Value)
		}
		if err != nil {
			return err
		}
		key.fields[i].Value = val
	}
	return nil
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

	for _, field := range key.fields {
		b = append(b, '/')
		b = encoding.EncodeFieldValue(b, field.Value, field.Descending)
	}

	return b
}
