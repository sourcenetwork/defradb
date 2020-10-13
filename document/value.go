package document

import (
	"encoding/binary"

	"github.com/fxamacker/cbor/v2"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/merkle/crdt"
)

// Value is an interface that points to a concrete Value implementation
// May collapse this down without an interface
type Value interface {
	Value() interface{}
	IsDocument() bool
	Type() core.CType
	IsDirty() bool
	Clean()
	IsDelete() bool //todo: Update IsDelete naming
	Delete()
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

type ReadableValue interface {
	Value

	Read() (interface{}, error)
}

type simpleValue struct {
	t       core.CType
	value   interface{}
	crdt    crdt.MerkleCRDT
	isDirty bool
	delete  bool
}

func newValue(t core.CType, val interface{}) simpleValue {
	return simpleValue{
		t:       t,
		value:   val,
		isDirty: true,
	}
}

// func (val simpleValue) Set(val interface{})

func (val simpleValue) Value() interface{} {
	return val.value
}

func (val simpleValue) Type() core.CType {
	return val.t
}

func (val simpleValue) IsDocument() bool {
	_, ok := val.value.(*Document)
	return ok
}

// IsDirty returns if the value is marked as dirty (unsaved/changed)
func (val simpleValue) IsDirty() bool {
	return val.isDirty
}

func (val *simpleValue) Clean() {
	val.isDirty = false
	val.delete = false
}

func (val *simpleValue) Delete() {
	val.delete = true
	val.isDirty = true
}

func (val simpleValue) IsDelete() bool {
	return val.delete
}

// // MakeDirty sets the value as
// func (val *simpleValue) MakeDirty() {
// 	val.isDirty = true
// }

// StringValue is a String wrapper for a simple Value
type StringValue struct {
	*simpleValue
}

// NewStringValue creates a new typed String Value
func NewStringValue(t core.CType, val string) WriteableValue {
	v := newValue(t, val)
	return StringValue{&v}
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
	*simpleValue
}

// NewInt64Value creates a new typed int64 value
func NewInt64Value(t core.CType, val int64) WriteableValue {
	v := newValue(t, val)
	return Int64Value{&v}
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
	*simpleValue
}

func newCBORValue(t core.CType, val interface{}) WriteableValue {
	v := newValue(t, val)
	return cborValue{&v}
}

func (v cborValue) Bytes() ([]byte, error) {
	return cbor.Marshal(v.value)
}

// func ReadCBORValue()

// func (val simpleValue) GetCRDT() crdt.MerkleCRDT {
// 	return val.crdt
// }

// func (val *simpleValue) SetCRDT(crdt crdt.MerkleCRDT) error {
// 	// if val.Type() != core.CType() {

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
