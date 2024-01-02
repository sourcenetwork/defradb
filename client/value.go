// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import (
	"github.com/fxamacker/cbor/v2"
)

// Value is an interface that points to a concrete Value implementation.
// (TODO May collapse this down without an interface)
type Value interface {
	Value() any
	IsDocument() bool
	Type() CType
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

	Read() (any, error)
}

type simpleValue struct {
	t       CType
	value   any
	isDirty bool
	delete  bool
}

func newValue(t CType, val any) simpleValue {
	return simpleValue{
		t:       t,
		value:   val,
		isDirty: true,
	}
}

// func (val simpleValue) Set(val any)

func (val simpleValue) Value() any {
	return val.value
}

func (val simpleValue) Type() CType {
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

type cborValue struct {
	*simpleValue
}

// NewCBORValue creates a new CBOR value from a CRDT type and a value.
func NewCBORValue(t CType, val any) WriteableValue {
	return newCBORValue(t, val)
}

func newCBORValue(t CType, val any) WriteableValue {
	v := newValue(t, val)
	return cborValue{&v}
}

func (v cborValue) Bytes() ([]byte, error) {
	em, err := cbor.EncOptions{Time: cbor.TimeRFC3339}.EncMode()
	if err != nil {
		return nil, err
	}
	return em.Marshal(v.value)
}
