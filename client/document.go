// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
	"github.com/sourcenetwork/defradb/client/request"
)

// This is the main implementation starting point for accessing the internal Document API
// which provides API access to the various operations available for Documents, i.e. CRUD.
//
// Documents in this case refer to the core database type of DefraDB which is a
// "NoSQL Document Datastore".
//
// This section is not concerned with the outer request layer used to interact with the
// Document API, but instead is solely concerned with carrying out the internal API
// operations.
//
// Note: These actions on the outside are deceivingly simple, but require a number
// of complex interactions with the underlying KV Datastore, as well as the
// Merkle CRDT semantics.

// Document is a generalized struct referring to a stored document in the database.
//
// It *can* have a reference to a enforced schema, which is enforced at the time
// of an operation.
//
// Documents are similar to JSON Objects stored in MongoDB, which are collections
// of Fields and Values.
//
// Fields are Key names that point to values.
// Values are literal or complex objects such as strings, integers, or sub documents (objects).
//
// Note: Documents represent the serialized state of the underlying MerkleCRDTs
//
// @todo: Extract Document into a Interface
// @body: A document interface can be implemented by both a TypedDocument and a
// UnTypedDocument, which use a schema and schemaless approach respectively.
type Document struct {
	key DocKey
	// SchemaVersionID holds the id of the schema version that this document is
	// currently at.
	//
	// Migrating the document will update this value to the output version of the
	// migration.
	SchemaVersionID string
	fields          map[string]Field
	values          map[Field]Value
	head            cid.Cid
	mu              sync.RWMutex
	// marks if document has unsaved changes
	isDirty bool
}

// NewDocWithKey creates a new Document with a specified key.
func NewDocWithKey(key DocKey) *Document {
	doc := newEmptyDoc()
	doc.key = key
	return doc
}

func newEmptyDoc() *Document {
	return &Document{
		fields: make(map[string]Field),
		values: make(map[Field]Value),
	}
}

// NewDocFromMap creates a new Document from a data map.
func NewDocFromMap(data map[string]any) (*Document, error) {
	var err error
	doc := &Document{
		fields: make(map[string]Field),
		values: make(map[Field]Value),
	}

	// check if document contains special _key field
	k, hasKey := data["_key"]
	if hasKey {
		delete(data, "_key") // remove the key so it isn't parsed further
		kstr, ok := k.(string)
		if !ok {
			return nil, NewErrUnexpectedType[string]("data[_key]", k)
		}
		if doc.key, err = NewDocKeyFromString(kstr); err != nil {
			return nil, err
		}
	}

	err = doc.setAndParseObjectType(data)
	if err != nil {
		return nil, err
	}

	// if no key was specified, then we assume it doesn't exist and we generate, and set it.
	if !hasKey {
		doc.generateAndSetDockey()
	}

	return doc, nil
}

// NewFromJSON creates a new instance of a Document from a raw JSON object byte array.
func NewDocFromJSON(obj []byte) (*Document, error) {
	data := make(map[string]any)
	err := json.Unmarshal(obj, &data)
	if err != nil {
		return nil, err
	}

	return NewDocFromMap(data)
}

// Head returns the current head CID of the document.
func (doc *Document) Head() cid.Cid {
	doc.mu.RLock()
	defer doc.mu.RUnlock()
	return doc.head
}

// SetHead sets the current head CID of the document.
func (doc *Document) SetHead(head cid.Cid) {
	doc.mu.Lock()
	defer doc.mu.Unlock()
	doc.head = head
}

// Key returns the generated DocKey for this document.
func (doc *Document) Key() DocKey {
	// Reading without a read-lock as we assume the DocKey is immutable
	return doc.key
}

// Get returns the raw value for a given field.
// Since Documents are objects with potentially sub objects a supplied field string can be of the
// form "A/B/C", where field A is an object containing a object B which has a field C.
// If no matching field exists then return an empty interface and an error.
func (doc *Document) Get(field string) (any, error) {
	val, err := doc.GetValue(field)
	if err != nil {
		return nil, err
	}
	return val.Value(), nil
}

// GetValue given a field as a string, return the Value type.
func (doc *Document) GetValue(field string) (Value, error) {
	doc.mu.RLock()
	defer doc.mu.RUnlock()
	path, subPaths, hasSubPaths := parseFieldPath(field)
	f, exists := doc.fields[path]
	if !exists {
		return nil, NewErrFieldNotExist(path)
	}

	val, err := doc.GetValueWithField(f)
	if err != nil {
		return nil, err
	}

	if !hasSubPaths {
		return val, nil
	} else if hasSubPaths && !val.IsDocument() {
		return nil, ErrFieldNotObject
	} else {
		return val.Value().(*Document).GetValue(subPaths)
	}
}

// GetValueWithField gets the Value type from a given Field type
func (doc *Document) GetValueWithField(f Field) (Value, error) {
	doc.mu.RLock()
	defer doc.mu.RUnlock()
	v, exists := doc.values[f]
	if !exists {
		return nil, NewErrFieldNotExist(f.Name())
	}
	return v, nil
}

// SetWithJSON sets all the fields of a document using the provided
// JSON Merge Patch object. Note: fields indicated as nil in the Merge
// Patch are to be deleted
// @todo: Handle sub documents for SetWithJSON
func (doc *Document) SetWithJSON(patch []byte) error {
	var patchObj map[string]any
	err := json.Unmarshal(patch, &patchObj)
	if err != nil {
		return err
	}

	for k, v := range patchObj {
		if v == nil {
			err = doc.Delete(k)
		} else {
			err = doc.Set(k, v)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// Set the value of a field.
func (doc *Document) Set(field string, value any) error {
	return doc.setAndParseType(field, value)
}

// SetAs is the same as set, but you can manually set the CRDT type.
func (doc *Document) SetAs(field string, value any, t CType) error {
	return doc.setCBOR(t, field, value)
}

// Delete removes a field, and marks it to be deleted on the following db.Update() call.
func (doc *Document) Delete(fields ...string) error {
	doc.mu.Lock()
	defer doc.mu.Unlock()
	for _, f := range fields {
		field, exists := doc.fields[f]
		if !exists {
			return NewErrFieldNotExist(f)
		}
		doc.values[field].Delete()
	}
	return nil
}

func (doc *Document) set(t CType, field string, value Value) error {
	doc.mu.Lock()
	defer doc.mu.Unlock()
	var f Field
	if v, exists := doc.fields[field]; exists {
		f = v
	} else {
		f = doc.newField(t, field)
		doc.fields[field] = f
	}
	doc.values[f] = value
	doc.isDirty = true
	return nil
}

func (doc *Document) setCBOR(t CType, field string, val any) error {
	value := newCBORValue(t, val)
	return doc.set(t, field, value)
}

func (doc *Document) setObject(t CType, field string, val *Document) error {
	value := newValue(t, val)
	return doc.set(t, field, &value)
}

// @todo: Update with document schemas
func (doc *Document) setAndParseType(field string, value any) error {
	if value == nil {
		return nil
	}

	switch val := value.(type) {
	// int (any number)
	case int:
		err := doc.setCBOR(LWW_REGISTER, field, int64(val))
		if err != nil {
			return err
		}
	case float64:
		// case int64:

		// Check if its actually a float or just an int
		if float64(int64(val)) == val { //int
			err := doc.setCBOR(LWW_REGISTER, field, int64(val))
			if err != nil {
				return err
			}
		} else { //float
			err := doc.setCBOR(LWW_REGISTER, field, val)
			if err != nil {
				return err
			}
		}

	// string, bool, and more
	case string, bool, []any:
		err := doc.setCBOR(LWW_REGISTER, field, val)
		if err != nil {
			return err
		}

	// sub object, recurse down.
	// @TODO: Object Definitions
	// You can use an object as a way to override defaults
	// and types for JSON literals.
	// Eg.
	// Instead of { "Timestamp": 123 }
	//			- which is parsed as an int
	// Use { "Timestamp" : { "_Type": "uint64", "_Value": 123 } }
	//			- Which is parsed as an uint64
	case map[string]any:
		subDoc := newEmptyDoc()
		err := subDoc.setAndParseObjectType(val)
		if err != nil {
			return err
		}

		err = doc.setObject(OBJECT, field, subDoc)
		if err != nil {
			return err
		}

	default:
		return NewErrUnhandledType(field, val)
	}
	return nil
}

func (doc *Document) setAndParseObjectType(value map[string]any) error {
	for k, v := range value {
		err := doc.setAndParseType(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

// Fields gets the document fields as a map.
func (doc *Document) Fields() map[string]Field {
	doc.mu.RLock()
	defer doc.mu.RUnlock()
	return doc.fields
}

// Values gets the document values as a map.
func (doc *Document) Values() map[Field]Value {
	doc.mu.RLock()
	defer doc.mu.RUnlock()
	return doc.values
}

// Bytes returns the document as a serialzed byte array using CBOR encoding.
func (doc *Document) Bytes() ([]byte, error) {
	docMap, err := doc.toMap()
	if err != nil {
		return nil, err
	}

	// Important: CanonicalEncOptions ensures consistent serialization of
	// indeterministic datastructures, like Go Maps
	em, err := cbor.CanonicalEncOptions().EncMode()
	if err != nil {
		return nil, err
	}
	return em.Marshal(docMap)
}

// String returns the document as a stringified JSON Object.
// Note: This representation should not be used for any cryptographic operations,
// such as signatures, or hashes as it does not guarantee canonical representation or ordering.
func (doc *Document) String() (string, error) {
	docMap, err := doc.toMap()
	if err != nil {
		return "", err
	}

	j, err := json.MarshalIndent(docMap, "", "\t")
	if err != nil {
		return "", err
	}

	return string(j), nil
}

// ToMap returns the document as a map[string]any
// object.
func (doc *Document) ToMap() (map[string]any, error) {
	return doc.toMapWithKey()
}

// Clean cleans the document by removing all dirty fields.
func (doc *Document) Clean() {
	for _, v := range doc.Fields() {
		val, _ := doc.GetValueWithField(v)
		if val.IsDirty() {
			if val.IsDelete() {
				doc.SetAs(v.Name(), nil, v.Type()) //nolint:errcheck
			}
			val.Clean()
		}
	}
}

// converts the document into a map[string]any
// including any sub documents
func (doc *Document) toMap() (map[string]any, error) {
	doc.mu.RLock()
	defer doc.mu.RUnlock()
	docMap := make(map[string]any)
	for k, v := range doc.fields {
		value, exists := doc.values[v]
		if !exists {
			return nil, NewErrFieldNotExist(v.Name())
		}

		if value.IsDocument() {
			subDoc := value.Value().(*Document)
			subDocMap, err := subDoc.toMap()
			if err != nil {
				return nil, err
			}
			docMap[k] = subDocMap
		}

		docMap[k] = value.Value()
	}

	return docMap, nil
}

func (doc *Document) toMapWithKey() (map[string]any, error) {
	doc.mu.RLock()
	defer doc.mu.RUnlock()
	docMap := make(map[string]any)
	for k, v := range doc.fields {
		value, exists := doc.values[v]
		if !exists {
			return nil, NewErrFieldNotExist(v.Name())
		}

		if value.IsDocument() {
			subDoc := value.Value().(*Document)
			subDocMap, err := subDoc.toMapWithKey()
			if err != nil {
				return nil, err
			}
			docMap[k] = subDocMap
		}

		docMap[k] = value.Value()
	}
	docMap["_key"] = doc.Key().String()

	return docMap, nil
}

// setDockeyFromBytes makes docKey out of the consumed bytes and then sets it as `doc.key`.
func (doc *Document) setDockeyFromBytes(bytes []byte) error {
	doc.mu.Lock()
	defer doc.mu.Unlock()

	pref := cid.Prefix{
		Version:  1,
		Codec:    cid.Raw,
		MhType:   mh.SHA2_256,
		MhLength: -1, // default length
	}

	c, err := pref.Sum(bytes)
	if err != nil {
		return err
	}

	doc.key = NewDocKeyV0(c)
	return nil
}

// generateAndSetDockey generates dockey and (re)sets `doc.key`.
func (doc *Document) generateAndSetDockey() error {
	newBytes, err := doc.Bytes()
	if err != nil {
		return err
	}

	return doc.setDockeyFromBytes(newBytes)
}

func (doc *Document) remapAliasFields(fieldDescriptions []FieldDescription) (bool, error) {
	doc.mu.Lock()
	defer doc.mu.Unlock()

	foundAlias := false
	for docField, docFieldValue := range doc.fields {
		for _, fieldDescription := range fieldDescriptions {
			maybeAliasField := docField + request.RelatedObjectID
			if fieldDescription.Name == maybeAliasField {
				foundAlias = true
				doc.fields[maybeAliasField] = docFieldValue
				delete(doc.fields, docField)
			}
		}
	}

	return foundAlias, nil
}

// RemapAliasFieldsAndDockey remaps the alias fields and fixes (overwrites) the dockey.
func (doc *Document) RemapAliasFieldsAndDockey(fieldDescriptions []FieldDescription) error {
	foundAlias, err := doc.remapAliasFields(fieldDescriptions)
	if err != nil {
		return err
	}

	if !foundAlias {
		return nil
	}

	// Update the dockey so dockey isn't based on an aliased name of a field.
	return doc.generateAndSetDockey()
}

// DocumentStatus represent the state of the document in the DAG store.
// It can either be `Active“ or `Deleted`.
type DocumentStatus uint8

const (
	// Active is the default state of a document.
	Active DocumentStatus = 1
	// Deleted represents a document that has been marked as deleted. This means that the document
	// can still be in the datastore but a normal request won't return it. The DAG store will still have all
	// the associated links.
	Deleted DocumentStatus = 2
)

var DocumentStatusToString = map[DocumentStatus]string{
	Active:  "Active",
	Deleted: "Deleted",
}

func (dStatus DocumentStatus) UInt8() uint8 {
	return uint8(dStatus)
}

func (dStatus DocumentStatus) IsDeleted() bool {
	return dStatus > 1
}

// loops through an object of the form map[string]any
// and fills in the Document with each field it finds in the object.
// Automatically handles sub objects and arrays.
// Does not allow anonymous fields, error is thrown in this case
// Eg. The JSON value [1,2,3,4] by itself is a valid JSON Object, but has no
// field name.
// func parseJSONObject(doc *Document, data map[string]any) error {
// 	for k, v := range data {
// 		switch v.(type) {

// 		// int (any number)
// 		case float64:
// 			// case int64:

// 			// Check if its actually a float or just an int
// 			val := v.(float64)
// 			if float64(int64(val)) == val { //int
// 				doc.setCBOR(crdt.LWW_REGISTER, k, int64(val))
// 			} else { //float
// 				panic("todo")
// 			}
// 			break

// 		// string
// 		case string:
// 			doc.setCBOR(crdt.LWW_REGISTER, k, v)
// 			break

// 		// array
// 		case []any:
// 			break

// 		// sub object, recurse down.
// 		// @TODO: Object Definitions
// 		// You can use an object as a way to override defaults
// 		// and types for JSON literals.
// 		// Eg.
// 		// Instead of { "Timestamp": 123 }
// 		//			- which is parsed as an int
// 		// Use { "Timestamp" : { "_Type": "uint64", "_Value": 123 } }
// 		//			- Which is parsed as an uint64
// 		case map[string]any:
// 			subDoc := newEmptyDoc()
// 			err := parseJSONObject(subDoc, v.(map[string]any))
// 			if err != nil {
// 				return err
// 			}

// 			doc.setObject(crdt.OBJECT, k, subDoc)
// 			break

// 		default:
// 			return errors.Wrap("Unhandled type in raw JSON: %v => %T", k, v)

// 		}
// 	}
// 	return nil
// }

// parses a document field path, can have sub elements if we have embedded objects.
// Returns the first path, the remaining split paths, and a bool indicating if there are sub paths
func parseFieldPath(path string) (string, string, bool) {
	splitKeys := strings.SplitN(path, "/", 2)
	return splitKeys[0], strings.Join(splitKeys[1:], ""), len(splitKeys) > 1
}

// Example Usage: Create/Insert new object
/*

obj := `{
	Hello: "World"
}`
objData := make(map[string]any)
err := json.Unmarshal(&objData, obj)

docA := document.NewFromJSON(objData)
err := db.Save(document)
		=> New batch transaction/store
		=> Loop through doc values
		=> 		instantiate MerkleCRDT objects
		=> 		Set/Publish new CRDT values


// One-to-one relationship example
obj := `{
	Hello: "world",
	Author: {
		Name: "Bob",
	}
}`

docA := document.NewFromJSON(obj)

// method 1
docA.Patch(...)
col.Save(docA)

// method 2
docA.Get("Author").Set("Name", "Eric")
col.Save(docA)

// method 3
docB := docA.GetObject("Author")
docB.Set("Name", "Eric")
authorCollection.Save(docB)

// method 4
docA.Set("Author.Name")

// method 5
doc := col.GetWithRelations("key")
// equivalent
doc := col.Get(key, db.WithRelationsOpt)

*/
