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

	"github.com/bits-and-blooms/bitset"
	"github.com/fxamacker/cbor/v2"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
)

type EncodedDocument interface {
	// Key returns the key of the document
	Key() []byte
	// Reset re-initializes the EncodedDocument object.
	Reset()
	// Decode returns a properly decoded document object
	Decode() (*client.Document, error)
	// DecodeToDoc returns a decoded document as a
	// map of field/value pairs
	DecodeToDoc() (core.Doc, error)
}

type EPTuple []encProperty

// EncProperty is an encoded property of a EncodedDocument
type encProperty struct {
	Desc client.FieldDescription
	Raw  []byte

	// Filter flag to determine if this flag
	// is needed for eager filter evaluation
	IsFilter bool

	// // encoding meta data
	// encoding base.DataEncoding
}

// Decode returns the decoded value and CRDT type for the given property.
func (e encProperty) Decode() (client.CType, any, error) {
	var val any
	err := cbor.Unmarshal(e.Raw, &val)
	if err != nil {
		return client.NONE_CRDT, nil, err
	}

	if array, isArray := val.([]any); isArray {
		var ok bool
		switch e.Desc.Kind {
		case client.FieldKind_BOOL_ARRAY:
			boolArray := make([]bool, len(array))
			for i, untypedValue := range array {
				boolArray[i], ok = untypedValue.(bool)
				if !ok {
					return client.NONE_CRDT, nil, client.NewErrUnexpectedType[bool](e.Desc.Name, untypedValue)
				}
			}
			val = boolArray

		case client.FieldKind_NILLABLE_BOOL_ARRAY:
			val, err = convertNillableArray[bool](e.Desc.Name, array)
			if err != nil {
				return client.NONE_CRDT, nil, err
			}

		case client.FieldKind_INT_ARRAY:
			intArray := make([]int64, len(array))
			for i, untypedValue := range array {
				intArray[i], err = convertToInt(fmt.Sprintf("%s[%v]", e.Desc.Name, i), untypedValue)
				if err != nil {
					return client.NONE_CRDT, nil, err
				}
			}
			val = intArray

		case client.FieldKind_NILLABLE_INT_ARRAY:
			val, err = convertNillableArrayWithConverter(e.Desc.Name, array, convertToInt)
			if err != nil {
				return client.NONE_CRDT, nil, err
			}

		case client.FieldKind_FLOAT_ARRAY:
			floatArray := make([]float64, len(array))
			for i, untypedValue := range array {
				floatArray[i], ok = untypedValue.(float64)
				if !ok {
					return client.NONE_CRDT, nil, client.NewErrUnexpectedType[float64](e.Desc.Name, untypedValue)
				}
			}
			val = floatArray

		case client.FieldKind_NILLABLE_FLOAT_ARRAY:
			val, err = convertNillableArray[float64](e.Desc.Name, array)
			if err != nil {
				return client.NONE_CRDT, nil, err
			}

		case client.FieldKind_STRING_ARRAY:
			stringArray := make([]string, len(array))
			for i, untypedValue := range array {
				stringArray[i], ok = untypedValue.(string)
				if !ok {
					return client.NONE_CRDT, nil, client.NewErrUnexpectedType[string](e.Desc.Name, untypedValue)
				}
			}
			val = stringArray

		case client.FieldKind_NILLABLE_STRING_ARRAY:
			val, err = convertNillableArray[string](e.Desc.Name, array)
			if err != nil {
				return client.NONE_CRDT, nil, err
			}
		}
	} else { // CBOR often encodes values typed as floats as ints
		switch e.Desc.Kind {
		case client.FieldKind_FLOAT:
			switch v := val.(type) {
			case int64:
				return client.NONE_CRDT, float64(v), nil
			case int:
				return client.NONE_CRDT, float64(v), nil
			case uint64:
				return client.NONE_CRDT, float64(v), nil
			case uint:
				return client.NONE_CRDT, float64(v), nil
			}
		}
	}

	return e.Desc.Typ, val, nil
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

// @todo: Implement Encoded Document type
type encodedDocument struct {
	mapping *core.DocumentMapping
	doc     *core.Doc

	key        []byte
	Properties []*encProperty

	// tracking bitsets
	// A value of 1 indicates a required field
	// 0 means we we ignore the field
	// we update the bitsets as we collect values
	// by clearing the bit for the FieldID
	filterSet *bitset.BitSet // filter fields
	selectSet *bitset.BitSet // select fields

}

var _ EncodedDocument = (*encodedDocument)(nil)

func (encdoc *encodedDocument) Key() []byte {
	return encdoc.key
}

// Reset re-initializes the EncodedDocument object.
func (encdoc *encodedDocument) Reset() {
	encdoc.Properties = make([]*encProperty, 0)
	encdoc.key = nil
	if encdoc.mapping != nil {
		doc := encdoc.mapping.NewDoc()
		encdoc.doc = &doc
	}
	encdoc.filterSet = nil
	encdoc.selectSet = nil
}

// Decode returns a properly decoded document object
func (encdoc *encodedDocument) Decode() (*client.Document, error) {
	key, err := client.NewDocKeyFromString(string(encdoc.key))
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

// DecodeToDoc returns a decoded document as a
// map of field/value pairs
func (encdoc *encodedDocument) DecodeToDoc() (core.Doc, error) {
	return encdoc.decodeToDoc(false)
}

func (encdoc *encodedDocument) decodeToDocForFilter() (core.Doc, error) {
	return encdoc.decodeToDoc(true)
}

func (encdoc *encodedDocument) decodeToDoc(filter bool) (core.Doc, error) {
	if encdoc.mapping == nil {
		return core.Doc{}, ErrMissingMapper
	}
	if encdoc.doc == nil {
		doc := encdoc.mapping.NewDoc()
		encdoc.doc = &doc
	}
	encdoc.doc.SetKey(string(encdoc.key))
	for _, prop := range encdoc.Properties {
		if encdoc.doc.Fields[prop.Desc.ID] != nil { // used cached decoded fields
			continue
		}
		if filter && !prop.IsFilter { // only get filter fields if filter=true
			continue
		}
		_, val, err := prop.Decode()
		if err != nil {
			return core.Doc{}, err
		}
		encdoc.doc.Fields[prop.Desc.ID] = val
	}
	return *encdoc.doc, nil
}
