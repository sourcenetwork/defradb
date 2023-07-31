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
	mapping *core.DocumentMapping
	doc     *core.Doc

	key             []byte
	schemaVersionID string
	Properties      map[client.FieldDescription]*encProperty

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

// Reset re-initializes the EncodedDocument object.
func (encdoc *encodedDocument) Reset() {
	encdoc.Properties = make(map[client.FieldDescription]*encProperty, 0)
	encdoc.key = nil
	if encdoc.mapping != nil {
		doc := encdoc.mapping.NewDoc()
		encdoc.doc = &doc
	}
	encdoc.filterSet = nil
	encdoc.selectSet = nil
	encdoc.schemaVersionID = ""
}

// Decode returns a properly decoded document object
func (encdoc *encodedDocument) Decode() (*client.Document, error) {
	key, err := client.NewDocKeyFromString(string(encdoc.key))
	if err != nil {
		return nil, err
	}
	doc := client.NewDocWithKey(key)
	for _, prop := range encdoc.Properties {
		val, err := prop.Decode()
		if err != nil {
			return nil, err
		}
		err = doc.SetAs(prop.Desc.Name, val, prop.Desc.Typ)
		if err != nil {
			return nil, err
		}
	}

	doc.SchemaVersionID = encdoc.SchemaVersionID()

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
		val, err := prop.Decode()
		if err != nil {
			return core.Doc{}, err
		}
		encdoc.doc.Fields[prop.Desc.ID] = val
	}

	encdoc.doc.SchemaVersionID = encdoc.SchemaVersionID()
	return *encdoc.doc, nil
}
