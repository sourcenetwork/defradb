package document

import (
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

	// encoding meta data
	encoding base.DataEncoding
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
