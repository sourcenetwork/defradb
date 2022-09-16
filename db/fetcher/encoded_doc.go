// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package fetcher

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
)

type EPTuple []encProperty

// EncProperty is an encoded property of a EncodedDocument
type encProperty struct {
	Desc client.FieldDescription
	Raw  []byte

	// // encoding meta data
	// encoding base.DataEncoding
}

// Decode returns the decoded value and CRDT type for the given property.
func (e encProperty) Decode() (client.CType, any, error) {
	ctype := client.CType(e.Raw[0])
	buf := e.Raw[1:]
	var val any
	err := cbor.Unmarshal(buf, &val)
	if err != nil {
		return ctype, nil, err
	}

	if array, isArray := val.([]any); isArray {
		var ok bool
		switch e.Desc.Kind {
		case client.FieldKind_BOOL_ARRAY:
			boolArray := make([]bool, len(array))
			for i, untypedValue := range array {
				boolArray[i], ok = untypedValue.(bool)
				if !ok {
					return ctype, nil, fmt.Errorf(
						"Could not convert type: %T, value: %v to bool.",
						untypedValue,
						untypedValue,
					)
				}
			}
			val = boolArray

		case client.FieldKind_NILLABLE_BOOL_ARRAY:
			val, err = convertNillableArray[bool](array)
			if err != nil {
				return ctype, nil, err
			}

		case client.FieldKind_INT_ARRAY:
			intArray := make([]int64, len(array))
			for i, untypedValue := range array {
				intArray[i], err = convertToInt(untypedValue)
				if err != nil {
					return ctype, nil, err
				}
			}
			val = intArray

		case client.FieldKind_NILLABLE_INT_ARRAY:
			val, err = convertNillableArrayWithConverter(array, convertToInt)
			if err != nil {
				return ctype, nil, err
			}

		case client.FieldKind_FLOAT_ARRAY:
			floatArray := make([]float64, len(array))
			for i, untypedValue := range array {
				floatArray[i], ok = untypedValue.(float64)
				if !ok {
					return ctype, nil, fmt.Errorf(
						"Could not convert type: %T, value: %v to float64.",
						untypedValue,
						untypedValue,
					)
				}
			}
			val = floatArray

		case client.FieldKind_NILLABLE_FLOAT_ARRAY:
			val, err = convertNillableArray[float64](array)
			if err != nil {
				return ctype, nil, err
			}

		case client.FieldKind_STRING_ARRAY:
			stringArray := make([]string, len(array))
			for i, untypedValue := range array {
				stringArray[i], ok = untypedValue.(string)
				if !ok {
					return ctype, nil, fmt.Errorf(
						"Could not convert type: %T, value: %v to string.",
						untypedValue,
						untypedValue,
					)
				}
			}
			val = stringArray

		case client.FieldKind_NILLABLE_STRING_ARRAY:
			val, err = convertNillableArray[string](array)
			if err != nil {
				return ctype, nil, err
			}
		}
	} else { // CBOR often encodes values typed as floats as ints
		switch e.Desc.Kind {
		case client.FieldKind_FLOAT:
			switch v := val.(type) {
			case int64:
				return ctype, float64(v), nil
			case int:
				return ctype, float64(v), nil
			case uint64:
				return ctype, float64(v), nil
			case uint:
				return ctype, float64(v), nil
			}
		}
	}

	return ctype, val, nil
}

func convertNillableArray[T any](items []any) ([]client.Option[T], error) {
	resultArray := make([]client.Option[T], len(items))
	for i, untypedValue := range items {
		if untypedValue == nil {
			resultArray[i] = client.None[T]()
			continue
		}
		value, ok := untypedValue.(T)
		if !ok {
			return nil, fmt.Errorf(
				"Could not convert type: %T, value: %v to %T.",
				untypedValue,
				untypedValue,
				*new(T),
			)
		}
		resultArray[i] = client.Some(value)
	}
	return resultArray, nil
}

func convertNillableArrayWithConverter[TOut any](
	items []any,
	converter func(in any) (TOut, error),
) ([]client.Option[TOut], error) {
	resultArray := make([]client.Option[TOut], len(items))
	for i, untypedValue := range items {
		if untypedValue == nil {
			resultArray[i] = client.None[TOut]()
			continue
		}
		value, err := converter(untypedValue)
		if err != nil {
			return nil, err
		}
		resultArray[i] = client.Some(value)
	}
	return resultArray, nil
}

func convertToInt(untypedValue any) (int64, error) {
	switch value := untypedValue.(type) {
	case uint64:
		return int64(value), nil
	case int64:
		return value, nil
	case float64:
		return int64(value), nil
	default:
		return 0, fmt.Errorf(
			"Could not convert type: %T, value: %v to int64.",
			untypedValue,
			untypedValue,
		)
	}
}

// @todo: Implement Encoded Document type
type encodedDocument struct {
	Key        []byte
	Properties map[client.FieldDescription]*encProperty
}

// Reset re-initializes the EncodedDocument object.
func (encdoc *encodedDocument) Reset() {
	encdoc.Properties = make(map[client.FieldDescription]*encProperty)
	encdoc.Key = nil
}

// Decode returns a properly decoded document object
func (encdoc *encodedDocument) Decode() (*client.Document, error) {
	key, err := client.NewDocKeyFromString(string(encdoc.Key))
	if err != nil {
		return nil, err
	}
	doc := client.NewDocWithKey(key)
	for fieldDesc, prop := range encdoc.Properties {
		ctype, val, err := prop.Decode()
		if err != nil {
			return nil, err
		}
		err = doc.SetAs(fieldDesc.Name, val, ctype)
		if err != nil {
			return nil, err
		}
	}

	return doc, nil
}

// DecodeToDoc returns a decoded document as a
// map of field/value pairs
func (encdoc *encodedDocument) DecodeToDoc(mapping *core.DocumentMapping) (core.Doc, error) {
	doc := mapping.NewDoc()

	doc.SetKey(string(encdoc.Key))
	for fieldDesc, prop := range encdoc.Properties {
		_, val, err := prop.Decode()
		if err != nil {
			return core.Doc{}, err
		}
		doc.Fields[fieldDesc.ID] = val
	}
	return doc, nil
}
