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
	"github.com/sourcenetwork/defradb/internal/core"
)

type EncodedDocument interface {
	// ID returns the ID of the document
	ID() []byte

	SchemaVersionID() string

	// Status returns the document status.
	//
	// For example, whether it is deleted or active.
	Status() client.DocumentStatus

	// Properties returns a copy of the decoded property values mapped by their field
	// description.
	Properties(onlyFilterProps bool) (map[client.FieldDefinition]any, error)

	// Reset re-initializes the EncodedDocument object.
	Reset()
}

type EPTuple []encProperty

// EncProperty is an encoded property of a EncodedDocument
type encProperty struct {
	Desc client.FieldDefinition
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

	return core.NormalizeFieldValue(e.Desc, val)
}

// @todo: Implement Encoded Document type
type encodedDocument struct {
	id                   []byte
	schemaVersionID      string
	status               client.DocumentStatus
	properties           map[client.FieldDefinition]*encProperty
	decodedPropertyCache map[client.FieldDefinition]any

	// tracking bitsets
	// A value of 1 indicates a required field
	// 0 means we we ignore the field
	// we update the bitsets as we collect values
	// by clearing the bit for the FieldID
	filterSet *bitset.BitSet // filter fields
	selectSet *bitset.BitSet // select fields
}

var _ EncodedDocument = (*encodedDocument)(nil)

func (encdoc *encodedDocument) ID() []byte {
	return encdoc.id
}

func (encdoc *encodedDocument) SchemaVersionID() string {
	return encdoc.schemaVersionID
}

func (encdoc *encodedDocument) Status() client.DocumentStatus {
	return encdoc.status
}

// Reset re-initializes the EncodedDocument object.
func (encdoc *encodedDocument) Reset() {
	encdoc.properties = make(map[client.FieldDefinition]*encProperty, 0)
	encdoc.id = nil
	encdoc.filterSet = nil
	encdoc.selectSet = nil
	encdoc.schemaVersionID = ""
	encdoc.status = 0
	encdoc.decodedPropertyCache = nil
}

// Decode returns a properly decoded document object
func Decode(encdoc EncodedDocument, collectionDefinition client.CollectionDefinition) (*client.Document, error) {
	docID, err := client.NewDocIDFromString(string(encdoc.ID()))
	if err != nil {
		return nil, err
	}

	doc := client.NewDocWithID(docID, collectionDefinition)
	properties, err := encdoc.Properties(false)
	if err != nil {
		return nil, err
	}

	for desc, val := range properties {
		err = doc.Set(desc.Name, val)
		if err != nil {
			return nil, err
		}
	}

	// client.Document tracks which fields have been set ('dirtied'), here we
	// are simply decoding a clean document and the dirty flag is an artifact
	// of the current client.Document interface.
	doc.Clean()

	return doc, nil
}

// MergeProperties merges the properties of the given document into this document.
// Existing fields of the current document are overwritten.
func (encdoc *encodedDocument) MergeProperties(other EncodedDocument) {
	otherEncDoc, ok := other.(*encodedDocument)
	if !ok {
		return
	}
	for field, prop := range otherEncDoc.properties {
		encdoc.properties[field] = prop
	}
	if other.ID() != nil {
		encdoc.id = other.ID()
	}
	if other.SchemaVersionID() != "" {
		encdoc.schemaVersionID = other.SchemaVersionID()
	}
}

// DecodeToDoc returns a decoded document as a
// map of field/value pairs
func DecodeToDoc(encdoc EncodedDocument, mapping *core.DocumentMapping, filter bool) (core.Doc, error) {
	doc := mapping.NewDoc()
	doc.SetID(string(encdoc.ID()))

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

func (encdoc *encodedDocument) Properties(onlyFilterProps bool) (map[client.FieldDefinition]any, error) {
	result := map[client.FieldDefinition]any{}
	if encdoc.decodedPropertyCache == nil {
		encdoc.decodedPropertyCache = map[client.FieldDefinition]any{}
	}

	for _, prop := range encdoc.properties {
		// only get filter fields if filter=true
		if onlyFilterProps && !prop.IsFilter {
			continue
		}

		// used cached decoded fields
		cachedValue := encdoc.decodedPropertyCache[prop.Desc]
		if cachedValue != nil {
			result[prop.Desc] = cachedValue
			continue
		}

		val, err := prop.Decode()
		if err != nil {
			return nil, err
		}

		// cache value
		encdoc.decodedPropertyCache[prop.Desc] = val
		result[prop.Desc] = val
	}

	return result, nil
}
