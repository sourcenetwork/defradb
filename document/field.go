package document

import (
	ds "github.com/ipfs/go-datastore"
	"github.com/sourcenetwork/defradb/core"
)

// Field is an interface to interact with Fields inside a document
type Field interface {
	Key() ds.Key
	Name() string
	Type() core.CType //TODO Abstract into a Field Type interface
	SchemaType() string
}

type simpleField struct {
	name       string
	key        ds.Key
	crdtType   core.CType
	schemaType string
}

func (doc *Document) newField(t core.CType, name string, schemaType ...string) Field {
	f := simpleField{
		name:     name,
		key:      doc.Key().ChildString(name),
		crdtType: t,
	}
	if len(schemaType) > 0 {
		f.schemaType = schemaType[0]
	}
	return f
}

func (field simpleField) Name() string {
	return field.name
}

func (field simpleField) Type() core.CType {
	return field.crdtType
}

func (field simpleField) Key() ds.Key {
	return field.key
}

func (field simpleField) SchemaType() string {
	return field.schemaType
}
