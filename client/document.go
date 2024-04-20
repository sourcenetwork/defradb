// Copyright 2023 Democratized Data Foundation
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
	"errors"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"
	"github.com/sourcenetwork/immutable"
	"github.com/valyala/fastjson"

	"github.com/sourcenetwork/defradb/client/request"
	ccid "github.com/sourcenetwork/defradb/core/cid"
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
	id     DocID
	fields map[string]Field
	values map[Field]*FieldValue
	head   cid.Cid
	mu     sync.RWMutex
	// marks if document has unsaved changes
	isDirty bool

	collectionDefinition CollectionDefinition
}

func newEmptyDoc(collectionDefinition CollectionDefinition) *Document {
	return &Document{
		fields:               make(map[string]Field),
		values:               make(map[Field]*FieldValue),
		collectionDefinition: collectionDefinition,
	}
}

// NewDocWithID creates a new Document with a specified key.
func NewDocWithID(docID DocID, collectionDefinition CollectionDefinition) *Document {
	doc := newEmptyDoc(collectionDefinition)
	doc.id = docID
	return doc
}

// NewDocFromMap creates a new Document from a data map.
func NewDocFromMap(data map[string]any, collectionDefinition CollectionDefinition) (*Document, error) {
	var err error
	doc := newEmptyDoc(collectionDefinition)

	// check if document contains special _docID field
	k, hasDocID := data[request.DocIDFieldName]
	if hasDocID {
		delete(data, request.DocIDFieldName) // remove the DocID so it isn't parsed further
		kstr, ok := k.(string)
		if !ok {
			return nil, NewErrUnexpectedType[string]("data["+request.DocIDFieldName+"]", k)
		}
		if doc.id, err = NewDocIDFromString(kstr); err != nil {
			return nil, err
		}
	}

	err = doc.setAndParseObjectType(data)
	if err != nil {
		return nil, err
	}

	// if no DocID was specified, then we assume it doesn't exist and we generate, and set it.
	if !hasDocID {
		err = doc.generateAndSetDocID()
		if err != nil {
			return nil, err
		}
	}

	return doc, nil
}

var jsonArrayPattern = regexp.MustCompile(`^\s*\[.*\]\s*$`)

// IsJSONArray returns true if the given byte array is a JSON Array.
func IsJSONArray(obj []byte) bool {
	return jsonArrayPattern.Match(obj)
}

// NewFromJSON creates a new instance of a Document from a raw JSON object byte array.
func NewDocFromJSON(obj []byte, collectionDefinition CollectionDefinition) (*Document, error) {
	doc := newEmptyDoc(collectionDefinition)
	err := doc.SetWithJSON(obj)
	if err != nil {
		return nil, err
	}
	err = doc.generateAndSetDocID()
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// ManyFromJSON creates a new slice of Documents from a raw JSON array byte array.
// It will return an error if the given byte array is not a valid JSON array.
func NewDocsFromJSON(obj []byte, collectionDefinition CollectionDefinition) ([]*Document, error) {
	v, err := fastjson.ParseBytes(obj)
	if err != nil {
		return nil, err
	}
	a, err := v.Array()
	if err != nil {
		return nil, err
	}

	docs := make([]*Document, len(a))
	for i, v := range a {
		o, err := v.Object()
		if err != nil {
			return nil, err
		}
		doc := newEmptyDoc(collectionDefinition)
		err = doc.setWithFastJSONObject(o)
		if err != nil {
			return nil, err
		}
		err = doc.generateAndSetDocID()
		if err != nil {
			return nil, err
		}
		docs[i] = doc
	}

	return docs, nil
}

// validateFieldSchema takes a given value as an interface,
// and ensures it matches the supplied field description.
// It will do any minor parsing, like dates, and return
// the typed value again as an interface.
func validateFieldSchema(val any, field FieldDefinition) (NormalValue, error) {
	if field.Kind.IsNillable() {
		if val == nil {
			return NewNormalNil(field.Kind)
		}
		if v, ok := val.(*fastjson.Value); ok && v.Type() == fastjson.TypeNull {
			return NewNormalNil(field.Kind)
		}
	}

	if field.Kind.IsObjectArray() {
		return nil, NewErrFieldNotExist(field.Name)
	}

	if field.Kind.IsObject() {
		v, err := getString(val)
		if err != nil {
			return nil, err
		}
		return NewNormalString(v), nil
	}

	switch field.Kind {
	case FieldKind_DocID, FieldKind_NILLABLE_STRING, FieldKind_NILLABLE_BLOB:
		v, err := getString(val)
		if err != nil {
			return nil, err
		}
		return NewNormalString(v), nil

	case FieldKind_STRING_ARRAY:
		v, err := getArray(val, getString)
		if err != nil {
			return nil, err
		}
		return NewNormalStringArray(v), nil

	case FieldKind_NILLABLE_STRING_ARRAY:
		v, err := getNillableArray(val, getString)
		if err != nil {
			return nil, err
		}
		return NewNormalNillableStringArray(v), nil

	case FieldKind_NILLABLE_BOOL:
		v, err := getBool(val)
		if err != nil {
			return nil, err
		}
		return NewNormalBool(v), nil

	case FieldKind_BOOL_ARRAY:
		v, err := getArray(val, getBool)
		if err != nil {
			return nil, err
		}
		return NewNormalBoolArray(v), nil

	case FieldKind_NILLABLE_BOOL_ARRAY:
		v, err := getNillableArray(val, getBool)
		if err != nil {
			return nil, err
		}
		return NewNormalNillableBoolArray(v), nil

	case FieldKind_NILLABLE_FLOAT:
		v, err := getFloat64(val)
		if err != nil {
			return nil, err
		}
		return NewNormalFloat(v), nil

	case FieldKind_FLOAT_ARRAY:
		v, err := getArray(val, getFloat64)
		if err != nil {
			return nil, err
		}
		return NewNormalFloatArray(v), nil

	case FieldKind_NILLABLE_FLOAT_ARRAY:
		v, err := getNillableArray(val, getFloat64)
		if err != nil {
			return nil, err
		}
		return NewNormalNillableFloatArray(v), nil

	case FieldKind_NILLABLE_DATETIME:
		v, err := getDateTime(val)
		if err != nil {
			return nil, err
		}
		return NewNormalTime(v), nil

	case FieldKind_NILLABLE_INT:
		v, err := getInt64(val)
		if err != nil {
			return nil, err
		}
		return NewNormalInt(v), nil

	case FieldKind_INT_ARRAY:
		v, err := getArray(val, getInt64)
		if err != nil {
			return nil, err
		}
		return NewNormalIntArray(v), nil

	case FieldKind_NILLABLE_INT_ARRAY:
		v, err := getNillableArray(val, getInt64)
		if err != nil {
			return nil, err
		}
		return NewNormalNillableIntArray(v), nil

	case FieldKind_NILLABLE_JSON:
		v, err := getJSON(val)
		if err != nil {
			return nil, err
		}
		return NewNormalString(v), nil
	}

	return nil, NewErrUnhandledType("FieldKind", field.Kind)
}

func getString(v any) (string, error) {
	switch val := v.(type) {
	case *fastjson.Value:
		b, err := val.StringBytes()
		return string(b), err
	default:
		return val.(string), nil
	}
}

func getBool(v any) (bool, error) {
	switch val := v.(type) {
	case *fastjson.Value:
		return val.Bool()
	default:
		return val.(bool), nil
	}
}

func getFloat64(v any) (float64, error) {
	switch val := v.(type) {
	case *fastjson.Value:
		return val.Float64()
	case int:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case float64:
		return val, nil
	default:
		return 0, NewErrUnexpectedType[float64]("field", v)
	}
}

func getInt64(v any) (int64, error) {
	switch val := v.(type) {
	case *fastjson.Value:
		return val.Int64()
	case int:
		return int64(val), nil
	case int32:
		return int64(val), nil
	case int64:
		return val, nil
	case float64:
		return int64(val), nil
	default:
		return 0, NewErrUnexpectedType[int64]("field", v)
	}
}

func getDateTime(v any) (time.Time, error) {
	var s string
	switch val := v.(type) {
	case *fastjson.Value:
		b, err := val.StringBytes()
		if err != nil {
			return time.Time{}, err
		}
		s = string(b)
	case time.Time:
		return val, nil
	default:
		s = val.(string)
	}
	return time.Parse(time.RFC3339, s)
}

func getJSON(v any) (string, error) {
	s, err := getString(v)
	if err != nil {
		return "", err
	}
	val, err := fastjson.Parse(s)
	if err != nil {
		return "", NewErrInvalidJSONPaylaod(s)
	}
	return val.String(), nil
}

func getArray[T any](
	v any,
	typeGetter func(any) (T, error),
) ([]T, error) {
	switch val := v.(type) {
	case *fastjson.Value:
		if val.Type() == fastjson.TypeNull {
			return nil, nil
		}

		valArray, err := val.Array()
		if err != nil {
			return nil, err
		}

		arr := make([]T, len(valArray))
		for i, arrItem := range valArray {
			if arrItem.Type() == fastjson.TypeNull {
				continue
			}
			arr[i], err = typeGetter(arrItem)
			if err != nil {
				return nil, err
			}
		}

		return arr, nil
	case []any:
		arr := make([]T, len(val))
		for i, arrItem := range val {
			var err error
			arr[i], err = typeGetter(arrItem)
			if err != nil {
				return nil, err
			}
		}

		return arr, nil
	case []T:
		return val, nil
	default:
		return []T{}, nil
	}
}

func getNillableArray[T any](
	v any,
	typeGetter func(any) (T, error),
) ([]immutable.Option[T], error) {
	switch val := v.(type) {
	case *fastjson.Value:
		if val.Type() == fastjson.TypeNull {
			return nil, nil
		}

		valArray, err := val.Array()
		if err != nil {
			return nil, err
		}

		arr := make([]immutable.Option[T], len(valArray))
		for i, arrItem := range valArray {
			if arrItem.Type() == fastjson.TypeNull {
				arr[i] = immutable.None[T]()
				continue
			}
			v, err := typeGetter(arrItem)
			if err != nil {
				return nil, err
			}
			arr[i] = immutable.Some(v)
		}

		return arr, nil
	case []any:
		arr := make([]immutable.Option[T], len(val))
		for i, arrItem := range val {
			if arrItem == nil {
				arr[i] = immutable.None[T]()
				continue
			}
			v, err := typeGetter(arrItem)
			if err != nil {
				return nil, err
			}
			arr[i] = immutable.Some(v)
		}

		return arr, nil
	case []immutable.Option[T]:
		return val, nil
	default:
		return []immutable.Option[T]{}, nil
	}
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

// ID returns the generated DocID for this document.
func (doc *Document) ID() DocID {
	// Reading without a read-lock as we assume the DocID is immutable
	return doc.id
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
func (doc *Document) GetValue(field string) (*FieldValue, error) {
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

// TryGetValue returns the value for a given field, if it exists.
// If the field does not exist then return nil and an error.
func (doc *Document) TryGetValue(field string) (*FieldValue, error) {
	val, err := doc.GetValue(field)
	if err != nil && errors.Is(err, ErrFieldNotExist) {
		return nil, nil
	}
	return val, err
}

// GetValueWithField gets the Value type from a given Field type
func (doc *Document) GetValueWithField(f Field) (*FieldValue, error) {
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
func (doc *Document) SetWithJSON(obj []byte) error {
	v, err := fastjson.ParseBytes(obj)
	if err != nil {
		return err
	}
	o, err := v.Object()
	if err != nil {
		return err
	}

	return doc.setWithFastJSONObject(o)
}

func (doc *Document) setWithFastJSONObject(obj *fastjson.Object) error {
	var visitErr error
	obj.Visit(func(k []byte, v *fastjson.Value) {
		fieldName := string(k)
		err := doc.Set(fieldName, v)
		if err != nil {
			visitErr = err
			return
		}
	})
	return visitErr
}

// Set the value of a field.
func (doc *Document) Set(field string, value any) error {
	fd, exists := doc.collectionDefinition.GetFieldByName(field)
	if !exists {
		return NewErrFieldNotExist(field)
	}
	if fd.Kind.IsObject() && !fd.Kind.IsObjectArray() {
		if !strings.HasSuffix(field, request.RelatedObjectID) {
			field = field + request.RelatedObjectID
		}
		fd, exists = doc.collectionDefinition.GetFieldByName(field)
		if !exists {
			return NewErrFieldNotExist(field)
		}
	}
	val, err := validateFieldSchema(value, fd)
	if err != nil {
		return err
	}
	return doc.setCBOR(fd.Typ, field, val)
}

func (doc *Document) set(t CType, field string, value *FieldValue) error {
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

func (doc *Document) setCBOR(t CType, field string, val NormalValue) error {
	value := NewFieldValue(t, val)
	return doc.set(t, field, value)
}

func (doc *Document) setAndParseObjectType(value map[string]any) error {
	for k, v := range value {
		err := doc.Set(k, v)
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
func (doc *Document) Values() map[Field]*FieldValue {
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

// ToMap returns the document as a map[string]any object.
func (doc *Document) ToMap() (map[string]any, error) {
	return doc.toMapWithKey()
}

// ToJSONPatch returns a json patch that can be used to update
// a document by calling SetWithJSON.
func (doc *Document) ToJSONPatch() ([]byte, error) {
	docMap, err := doc.toMap()
	if err != nil {
		return nil, err
	}

	for field, value := range doc.Values() {
		if !value.IsDirty() {
			delete(docMap, field.Name())
		}
	}

	return json.Marshal(docMap)
}

// Clean cleans the document by removing all dirty fields.
func (doc *Document) Clean() {
	for _, v := range doc.Fields() {
		val, _ := doc.GetValueWithField(v)
		if val.IsDirty() {
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
	docMap[request.DocIDFieldName] = doc.ID().String()

	return docMap, nil
}

// GenerateDocID generates the DocID corresponding to the document.
func (doc *Document) GenerateDocID() (DocID, error) {
	bytes, err := doc.Bytes()
	if err != nil {
		return DocID{}, err
	}

	cid, err := ccid.NewSHA256CidV1(bytes)
	if err != nil {
		return DocID{}, err
	}

	return NewDocIDV0(cid), nil
}

// setDocID sets the `doc.id` (should NOT be public).
func (doc *Document) setDocID(docID DocID) {
	doc.mu.Lock()
	defer doc.mu.Unlock()

	doc.id = docID
}

// generateAndSetDocID generates the DocID and then (re)sets `doc.id`.
func (doc *Document) generateAndSetDocID() error {
	docID, err := doc.GenerateDocID()
	if err != nil {
		return err
	}

	doc.setDocID(docID)
	return nil
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
