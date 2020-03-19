package document

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"

	"github.com/sourcenetwork/defradb/document/key"
	"github.com/sourcenetwork/defradb/merkle/crdt"
)

// This is the main implementation stating point for accessing the internal Document API
// Which provides API access to the various operations available for Documents
// IE. CRUD.
//
// Documents in this case refer to the core database type of DefraDB which is a
// "NoSQL Document Datastore"
//
// This section is not concerned with the outer query layer used to interact with the
// Document API, but instead is soley consered with carrying out the internal API
// operations. IE. CRUD.
//
// Note: These actions on the outside are deceivingly simple, but require a number
// of complex interactions with the underlying KV Datastore, as well as the
// Merkle CRDT semantics.

var (
	ErrFieldNotExist  = errors.New("The given field does not exist")
	ErrFieldNotObject = errors.New("Trying to access field on a non object type")
)

// Document is a generalized struct referring to a stored document in the database.
//
// It *can* have a reference to a enforced schema, which is enforced at the time
// of an operation.
//
// Documents are similar to JSON Objects stored in MongoDB, which are collections
// of Fields and Values.
// Fields are Key names that point to values
// Values are literal or complex objects such as strings, integers, or sub documents (objects)
//
// Note: Documents represent the serialized state of the underlying MerkleCRDTs
type Document struct {
	key    key.DocKey
	fields map[string]Field
	values map[Field]Value
	// @TODO: schemaInfo schema.Info
}

func newEmptyDoc() *Document {
	return &Document{
		fields: make(map[string]Field),
		values: make(map[Field]Value),
	}
}

// NewFromJSON creates a new instance of a Document from a raw JSON object byte array
func NewFromJSON(obj []byte) (*Document, error) {
	pref := cid.Prefix{
		Version:  1,
		Codec:    cid.Raw,
		MhType:   mh.SHA2_256,
		MhLength: -1, // default length
	}

	// And then feed it some data
	c, err := pref.Sum(obj)
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	err = json.Unmarshal(obj, &data)
	if err != nil {
		return nil, err
	}

	doc := &Document{
		key:    key.NewDocKeyV0(c),
		fields: make(map[string]Field),
		values: make(map[Field]Value),
	}

	err = parseJSONObject(doc, data)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

// Key returns the generated DocKey for this document
func (doc *Document) Key() key.DocKey {
	return doc.key
}

// Get returns the value for a given field
// Since Documents are objects with potentially sub objects
// a supplied field string can be of the form "A/B/C"
// Where field A is an object containing a object B which has
// a field C
// If no matching field exists then return an empty interface
// and an error.
func (doc *Document) Get(field string) (interface{}, error) {
	splitKeys := strings.SplitN(field, "/", 2)
	hasChildKeys := len(splitKeys) > 1
	f, exists := doc.fields[splitKeys[0]]
	if !exists {
		return nil, ErrFieldNotExist
	}

	val := doc.values[f]
	if !hasChildKeys {
		return val.Value(), nil
	} else if hasChildKeys && !val.IsDocument() {
		return nil, ErrFieldNotObject
	} else {
		return val.Value().(*Document).Get(splitKeys[1])
	}
}

// loops through a parsed JSON object of the form map[string]interface{}
// and fills in the Document with each field it finds in the JSON object.
// Automatically handles sub objects and arrays.
// Does not allow anonymous fields, error is thrown in this case
// Eg. The JSON value [1,2,3,4] by itself is a valid JSON Object, but has no
// field name.
func parseJSONObject(doc *Document, data map[string]interface{}) error {
	for k, v := range data {
		switch v.(type) {

		// int (any number)
		case float64:
			// case int64:

			// Check if its actually a float or just an int
			val := v.(float64)
			// var val interface{}
			if float64(int64(val)) == val {
				v = int64(val)
			}

			field := doc.newField(crdt.LWW_REGISTER, k)
			doc.fields[k] = field
			doc.values[field] = newValue(crdt.LWW_REGISTER, v)
			break

		// string
		case string:
			field := doc.newField(crdt.LWW_REGISTER, k)
			doc.fields[k] = field
			doc.values[field] = newValue(crdt.LWW_REGISTER, v)
			break

		// array
		case []interface{}:
			break

		// sub object, recurse down.
		// @TODO: Object Definitions
		// You can use an object as a way to override defults
		// and types for JSON literals.
		// Eg.
		// Instead of { "Timestamp": 123 }
		//			- which is parsed as an int
		// Use { "Timestamp" : { "_Type": "uint64", "_Value": 123 } }
		//			- Which is parsed as an uint64
		case map[string]interface{}:
			subDoc := newEmptyDoc()
			err := parseJSONObject(subDoc, v.(map[string]interface{}))
			if err != nil {
				return err
			}

			field := subDoc.newField(crdt.OBJECT, k)
			doc.fields[k] = field
			doc.values[field] = newValue(crdt.OBJECT, subDoc)
			break

		default:
			return fmt.Errorf("Unhandled type in raw JSON: %v => %T", k, v)

		}
	}
	return nil
}

// Exmaple Usage: Create/Insert new object
/*

obj := `{
	Hello: "World"
}`
objData := make(map[string]interface{})
err := json.Unmarshal(&objData, obj)

docA := document.NewFromJSON(objData)
err := db.Save(document)
		=> New batch transaction/store
		=> Loop through doc values
		=> 		instanciate MerkleCRDT objects
		=> 		Set/Publish new CRDT values


*/
