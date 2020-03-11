package document

import "github.com/sourcenetwork/defradb/merkle/crdt"

// Field is an interface to interact with Fields inside a document
type Field interface {
	Name() string
	Type() crdt.Type //TODO Abstract into a Field Type interface
}

type simpleField struct {
	name     string
	crdtType crdt.Type
}

func (field simpleField) Name() string {
	return field.name
}

func (field simpleField) Type() crdt.Type {
	return field.crdtType
}

func newField(key string, t crdt.Type) Field {
	return simpleField{
		name:     key,
		crdtType: t,
	}
}
