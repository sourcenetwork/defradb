package document

import "github.com/sourcenetwork/defradb/merkle/crdt"

// Value is an interface that points to a concrete Value implementation
// May collapse this down without an interface
type Value interface {
	Value() interface{}
	Type() crdt.Type
	IsDocument() bool
}

type simpleValue struct {
	t     crdt.Type
	value interface{}
}

func newValue(t crdt.Type, val interface{}) simpleValue {
	return simpleValue{
		t:     t,
		value: val,
	}
}

// func (val simpleValue) Set(val interface{})

func (val simpleValue) Value() interface{} {
	return val.value
}

func (val simpleValue) Type() crdt.Type {
	return val.t
}

func (val simpleValue) IsDocument() bool {
	_, ok := val.value.(*Document)
	return ok
}

// type merkleCRDTValue struct {
// 	crdt crdt.MerkleCRDT
// }

// func newMerkleCRDTValue(dt crdt.MerkleCRDT) *merkleCRDTValue {
// 	return &merkleCRDTValue{
// 		crdt: dt
// 	}
// }
