package document

import (
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/document/key"
)

type ValueType interface {
	// returns what kind of field it is
	// Kind() string
}

// Scalar Value Type
type Scalar struct {
	value interface{}
}

// List Value Type
type List struct {
	vals []ValueType
}

type SimpleDocument struct {
	schema *base.SchemaDescription

	data map[string]ValueType
}

func NewSimpleFromJSON(schema *base.SchemaDescription, data []byte) (*SimpleDocument, error) {
	return nil, nil
}

func NewSimpleFromMap(schema *base.SchemaDescription, data map[string]interface{}) (*SimpleDocument, error) {
	return nil, nil
}

func NewSimpleWithKey(schema *base.SchemaDescription, key key.DocKey) *SimpleDocument {
	return nil
}

func SimpleFromJSON(schema *base.SchemaDescription, data []byte) (*SimpleDocument, error) {
	return nil, nil
}

func SimpleFromMap(schema *base.SchemaDescription, data map[string]interface{}) (*SimpleDocument, error) {
	return nil, nil
}

func (doc *SimpleDocument) Get(field string) interface{} {
	return nil
}

/* API

doc := userCollection.GetByID(db.WithRealtions)
doc := userCollection.GetByFilter()
userCollection.UpdateB(filter | doc | docs | docID | docIDs, patch)

doc.Patch(...)
doc.GetRelations
doc.GetObject("Author")
doc.GetList()

userCollection.Save(doc) // checks for dirty fields, checks for patch/merge, apply

*/
