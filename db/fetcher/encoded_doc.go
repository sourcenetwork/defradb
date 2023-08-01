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
	"github.com/bits-and-blooms/bitset"
	"github.com/fxamacker/cbor/v2"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
)

type EncodedDocument interface {
	// Key returns the key of the document
	Key() []byte
	SchemaVersionID() string
	// Status returns the document status.
	//
	// For example, whether it is deleted or active.
	Status() client.DocumentStatus
	// Properties returns a copy of the decoded property values mapped by their field
	// description.
	Properties(onlyFilterProps bool) (map[client.FieldDescription]any, error)
	// Reset re-initializes the EncodedDocument object.
	Reset()
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
func (e encProperty) Decode() (any, error) {
	var val any
	err := cbor.Unmarshal(e.Raw, &val)
	if err != nil {
		return nil, err
	}

	return core.DecodeFieldValue(e.Desc, val)
}

// @todo: Implement Encoded Document type
type encodedDocument struct {
	key                  []byte
	schemaVersionID      string
	status               client.DocumentStatus
	properties           map[client.FieldDescription]*encProperty
	decodedpropertyCache map[client.FieldDescription]any

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

func (encdoc *encodedDocument) SchemaVersionID() string {
	return encdoc.schemaVersionID
}

func (encdoc *encodedDocument) Status() client.DocumentStatus {
	return encdoc.status
}

// Reset re-initializes the EncodedDocument object.
func (encdoc *encodedDocument) Reset() {
	encdoc.properties = make(map[client.FieldDescription]*encProperty, 0)
	encdoc.key = nil
	encdoc.filterSet = nil
	encdoc.selectSet = nil
	encdoc.schemaVersionID = ""
	encdoc.status = 0
	encdoc.decodedpropertyCache = nil
}

// Decode returns a properly decoded document object
func Decode(encdoc EncodedDocument) (*client.Document, error) {
	key, err := client.NewDocKeyFromString(string(encdoc.Key()))
	if err != nil {
		return nil, err
	}

	doc := client.NewDocWithKey(key)
	properties, err := encdoc.Properties(false)
	if err != nil {
		return nil, err
	}

	for desc, val := range properties {
		err = doc.SetAs(desc.Name, val, desc.Typ)
		if err != nil {
			return nil, err
		}
	}

	doc.SchemaVersionID = encdoc.SchemaVersionID()

	return doc, nil
}

// DecodeToDoc returns a decoded document as a
// map of field/value pairs
func DecodeToDoc(encdoc EncodedDocument, mapping *core.DocumentMapping, filter bool) (core.Doc, error) {
	doc := mapping.NewDoc()
	doc.SetKey(string(encdoc.Key()))

	properties, err := encdoc.Properties(filter)
	if err != nil {
		return core.Doc{}, err
	}

	for desc, value := range properties {
		doc.Fields[desc.ID] = value
	}

	doc.SchemaVersionID = encdoc.SchemaVersionID()
	doc.Status = encdoc.Status()

	return doc, nil
}

func (encdoc *encodedDocument) Properties(onlyFilterProps bool) (map[client.FieldDescription]any, error) {
	result := map[client.FieldDescription]any{}
	if encdoc.decodedpropertyCache == nil {
		encdoc.decodedpropertyCache = map[client.FieldDescription]any{}
	}

	for _, prop := range encdoc.properties {
		// used cached decoded fields
		cachedValue := encdoc.decodedpropertyCache[prop.Desc]
		if cachedValue != nil {
			result[prop.Desc] = cachedValue
			continue
		}

		// only get filter fields if filter=true
		if onlyFilterProps && !prop.IsFilter {
			continue
		}

		val, err := prop.Decode()
		if err != nil {
			return nil, err
		}

		// cache value
		encdoc.decodedpropertyCache[prop.Desc] = val
		result[prop.Desc] = val
	}

	return result, nil
}
