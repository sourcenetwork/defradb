package document

import (
	"github.com/sourcenetwork/defradb/merkle/crdt"

	ds "github.com/ipfs/go-datastore"
)

// Field is an interface to interact with Fields inside a document
type Field interface {
	Key() ds.Key
	Name() string
	Type() crdt.Type //TODO Abstract into a Field Type interface
}

type simpleField struct {
	name     string
	key      ds.Key
	crdtType crdt.Type
}

func (doc *Document) newField(t crdt.Type, name string) Field {
	return simpleField{
		name:     name,
		key:      doc.Key().ChildString(name),
		crdtType: t,
	}
}

func (field simpleField) Name() string {
	return field.name
}

func (field simpleField) Type() crdt.Type {
	return field.crdtType
}

func (field simpleField) Key() ds.Key {
	return field.key
}
