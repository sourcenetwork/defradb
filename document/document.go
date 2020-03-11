package document

import (
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
type Document struct {
	key    key.DocKey
	fields map[string]Field
	values map[Field]Value
	// TODO: schemaInfo schema.Info
}

// Field is an interface to interact with Fields inside a document
type Field interface {
	Name() string
	Type() FieldType
}

// FieldType is an abstraction of managing the types associated with the Value of a Field
// This may or may not be schema enforced, but it always have a value.
type FieldType struct {
	Type crdt.Type
}

// Value is an interface that points to a concrete Value implementation
// May collapse this down without an interface
type Value interface {
	Value() interface{}
}
