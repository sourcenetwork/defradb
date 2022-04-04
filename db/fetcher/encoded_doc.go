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
func (e encProperty) Decode() (client.CType, interface{}, error) {
	ctype := client.CType(e.Raw[0])
	buf := e.Raw[1:]
	// fmt.Println("decode...", e.Desc.Name, buf)
	var val interface{}
	err := cbor.Unmarshal(buf, &val)
	if err != nil {
		return ctype, nil, err
	}

	if array, isArray := val.([]interface{}); isArray {
		var ok bool
		switch e.Desc.Kind {
		case client.FieldKind_BOOL_ARRAY:
			boolArray := make([]bool, len(array))
			for i, untypedValue := range array {
				boolArray[i], ok = untypedValue.(bool)
				if !ok {
					return ctype, nil, fmt.Errorf("Could not convert type: %T, value: %v to bool.", untypedValue, untypedValue)
				}
			}
			val = boolArray
		case client.FieldKind_INT_ARRAY:
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
		case client.FieldKind_FLOAT_ARRAY:
			floatArray := make([]float64, len(array))
			for i, untypedValue := range array {
				floatArray[i], ok = untypedValue.(float64)
				if !ok {
					return ctype, nil, fmt.Errorf("Could not convert type: %T, value: %v to float64.", untypedValue, untypedValue)
				}
			}
			val = floatArray
		case client.FieldKind_STRING_ARRAY:
			stringArray := make([]string, len(array))
			for i, untypedValue := range array {
				stringArray[i], ok = untypedValue.(string)
				if !ok {
					return ctype, nil, fmt.Errorf("Could not convert type: %T, value: %v to string.", untypedValue, untypedValue)
				}
			}
			val = stringArray
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

	// fmt.Println("decoded:", val)
	return ctype, val, nil
}

// @todo: Implement Encoded Document type
type encodedDocument struct {
	Key        []byte
	Properties map[client.FieldID]*encProperty
}

// Reset re-initializes the EncodedDocument object.
func (encdoc *encodedDocument) Reset() {
	encdoc.Properties = make(map[client.FieldID]*encProperty)
	encdoc.Key = nil
}

// Decode returns a properly decoded document object
func (encdoc *encodedDocument) Decode() (*client.Document, error) {
	key, err := client.NewDocKeyFromString(string(encdoc.Key))
	if err != nil {
		return nil, err
	}
	doc := client.NewDocWithKey(key)
	for _, prop := range encdoc.Properties {
		ctype, val, err := prop.Decode()
		if err != nil {
			return nil, err
		}
		err = doc.SetAs(prop.Desc.Name, val, ctype)
		if err != nil {
			return nil, err
		}
	}

	return doc, nil
}

// DecodeToMap returns a decoded document as a
// map of field/value pairs
func (encdoc *encodedDocument) DecodeToMap() (map[string]interface{}, error) {
	doc := make(map[string]interface{})
	doc["_key"] = string(encdoc.Key)
	for _, prop := range encdoc.Properties {
		_, val, err := prop.Decode()
		if err != nil {
			fmt.Println("err5")
			return nil, err
		}
		doc[prop.Desc.Name] = val
	}
	return doc, nil
}
