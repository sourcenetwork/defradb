// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package document

import (
	"fmt"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/document/key"

	"github.com/fxamacker/cbor/v2"
)

type EPTuple []EncProperty

// EncProperty is an encoded property of a EncodedDocument
type EncProperty struct {
	Desc base.FieldDescription
	Raw  []byte

	// // encoding meta data
	// encoding base.DataEncoding
}

// Decode returns the decoded value and CRDT type for the given property.
func (e EncProperty) Decode() (core.CType, interface{}, error) {
	ctype := core.CType(e.Raw[0])
	buf := e.Raw[1:]
	var val interface{}
	err := cbor.Unmarshal(buf, &val)
	if err != nil {
		return ctype, nil, err
	}

	if array, isArray := val.([]interface{}); isArray {
		var ok bool
		switch e.Desc.Kind {
		case base.FieldKind_BOOL_ARRAY:
			boolArray := make([]bool, len(array))
			for i, untypedValue := range array {
				boolArray[i], ok = untypedValue.(bool)
				if !ok {
					return ctype, nil, fmt.Errorf("Could not convert type: %T, value: %v to bool.", untypedValue, untypedValue)
				}
			}
			val = boolArray
		case base.FieldKind_INT_ARRAY:
			intArray := make([]int64, len(array))
			for i, untypedValue := range array {
				switch value := untypedValue.(type) {
				case uint64:
					intArray[i] = int64(value)
				case int64:
					intArray[i] = value
				case float64:
					intArray[i] = int64(value)
				default:
					return ctype, nil, fmt.Errorf("Could not convert type: %T, value: %v to int64.", untypedValue, untypedValue)
				}
			}
			val = intArray
		case base.FieldKind_FLOAT_ARRAY:
			floatArray := make([]float64, len(array))
			for i, untypedValue := range array {
				floatArray[i], ok = untypedValue.(float64)
				if !ok {
					return ctype, nil, fmt.Errorf("Could not convert type: %T, value: %v to float64.", untypedValue, untypedValue)
				}
			}
			val = floatArray
		case base.FieldKind_STRING_ARRAY:
			stringArray := make([]string, len(array))
			for i, untypedValue := range array {
				stringArray[i], ok = untypedValue.(string)
				if !ok {
					return ctype, nil, fmt.Errorf("Could not convert type: %T, value: %v to string.", untypedValue, untypedValue)
				}
			}
			val = stringArray
		}
	}

	return ctype, val, nil
}

// @todo: Implement Encoded Document type
type EncodedDocument struct {
	Key        []byte
	Schema     *base.SchemaDescription
	Properties map[base.FieldDescription]*EncProperty
}

// Reset re-initalizes the EncodedDocument object.
func (encdoc *EncodedDocument) Reset() {
	encdoc.Properties = make(map[base.FieldDescription]*EncProperty)
	encdoc.Key = nil
}

// Decode returns a properly decoded document object
func (encdoc *EncodedDocument) Decode() (*Document, error) {
	key, err := key.NewFromString(string(encdoc.Key))
	if err != nil {
		return nil, err
	}
	doc := NewWithKey(key)
	if encdoc.Schema != nil {
		doc.schema = *encdoc.Schema
	}
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

// DecodeToMap returns a decoded document as a
// map of field/value pairs
func (encdoc *EncodedDocument) DecodeToMap() (map[string]interface{}, error) {
	doc := make(map[string]interface{})
	doc["_key"] = string(encdoc.Key)
	for fieldDesc, prop := range encdoc.Properties {
		_, val, err := prop.Decode()
		if err != nil {
			return nil, err
		}
		doc[fieldDesc.Name] = val
	}
	return doc, nil
}
