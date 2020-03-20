package document

import "github.com/sourcenetwork/defradb/merkle/crdt"

// Value is an interface that points to a concrete Value implementation
// May collapse this down without an interface
type Value interface {
	Value() interface{}
	IsDocument() bool
	Type() crdt.Type
	GetCRDT() crdt.MerkleCRDT
	SetCRDT(crdt.MerkleCRDT) error
}

type simpleValue struct {
	t     crdt.Type
	value interface{}
	crdt  crdt.MerkleCRDT
}

func newValue(t crdt.Type, val interface{}) *simpleValue {
	return &simpleValue{
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

func (val simpleValue) GetCRDT() crdt.MerkleCRDT {
	return val.crdt
}

func (val *simpleValue) SetCRDT(crdt crdt.MerkleCRDT) error {
	// if val.Type() != crdt.Type() {

	// } else {

	// }
	val.crdt = crdt
	return nil
}

// type merkleCRDTValue struct {
// 	crdt crdt.MerkleCRDT
// }

// func newMerkleCRDTValue(dt crdt.MerkleCRDT) *merkleCRDTValue {
// 	return &merkleCRDTValue{
// 		crdt: dt
// 	}
// }
