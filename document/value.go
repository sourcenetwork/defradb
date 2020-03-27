package document

import (
	"encoding/binary"

	"github.com/fxamacker/cbor/v2"

	"github.com/sourcenetwork/defradb/merkle/crdt"
)

// Value is an interface that points to a concrete Value implementation
// May collapse this down without an interface
type Value interface {
	Value() interface{}
	IsDocument() bool
	Type() crdt.Type
}

// WriteableValue defines a simple interface with a Bytes() method
// which is used to indicate if a Value is writeable type versus
// a composite type like a Sub-Document.
// Writeable types include simple Strings/Ints/Floats/Binary
// that can be loaded into a CRDT Register, Set, Counter, etc.
type WriteableValue interface {
	Value

	Bytes() ([]byte, error)
}

type simpleValue struct {
	t     crdt.Type
	value interface{}
	crdt  crdt.MerkleCRDT
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

// StringValue is a String wrapper for a simple Value
type StringValue struct {
	simpleValue
}

// NewStringValue creates a new typed String Value
func NewStringValue(t crdt.Type, val string) WriteableValue {
	return StringValue{newValue(t, val)}
}

// Bytes implements WriteableValue and encodes a string into a byte array
func (s StringValue) Bytes() ([]byte, error) {
	str, ok := s.value.(string)
	if !ok {
		return []byte(nil), ErrValueTypeMismatch
	}
	return []byte(str), nil
}

// Int64Value is a String wrapper for a simple Value
type Int64Value struct {
	simpleValue
}

// NewInt64Value creates a new typed int64 value
func NewInt64Value(t crdt.Type, val int64) WriteableValue {
	return Int64Value{newValue(t, val)}
}

// Bytes implements WriteableValue and encodes an int64 into a byte array
func (s Int64Value) Bytes() ([]byte, error) {
	i, ok := s.value.(int64)
	if !ok {
		return []byte(nil), ErrValueTypeMismatch
	}
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, i)
	b := buf[:n]
	return b, nil
}

type cborValue struct {
	simpleValue
}

func newCBORValue(t crdt.Type, val interface{}) WriteableValue {
	return cborValue{newValue(t, val)}
}

func (v cborValue) Bytes() ([]byte, error) {
	return cbor.Marshal(v.value)
}

// func (val simpleValue) GetCRDT() crdt.MerkleCRDT {
// 	return val.crdt
// }

// func (val *simpleValue) SetCRDT(crdt crdt.MerkleCRDT) error {
// 	// if val.Type() != crdt.Type() {

// 	// } else {

// 	// }
// 	val.crdt = crdt
// 	return nil
// }

// type merkleCRDTValue struct {
// 	crdt crdt.MerkleCRDT
// }

// func newMerkleCRDTValue(dt crdt.MerkleCRDT) *merkleCRDTValue {
// 	return &merkleCRDTValue{
// 		crdt: dt
// 	}
// }
