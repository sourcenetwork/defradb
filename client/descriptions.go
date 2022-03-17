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
	"fmt"
)

// CollectionDescription describes a Collection and all its associated metadata.
// It a single instance of its schema, containing localized additions tailored to
// this instance's needs.
type CollectionDescription struct {
	Name    string
	ID      uint32
	Schema  SchemaDescription
	Indexes []IndexDescription
}

// IDString returns the collection ID as a string
func (col CollectionDescription) IDString() string {
	return fmt.Sprint(col.ID)
}

func (col CollectionDescription) GetField(name string) (FieldDescription, bool) {
	if !col.IsEmpty() {
		for _, field := range col.Schema.Fields {
			if field.Name == name {
				return field, true
			}
		}
	}
	return FieldDescription{}, false
}

func (c CollectionDescription) GetPrimaryIndex() IndexDescription {
	return c.Indexes[0]
}

//IsEmpty returns true if the CollectionDescription is empty and uninitialized
func (c CollectionDescription) IsEmpty() bool {
	return c.Schema.IsEmpty()
}

func (c CollectionDescription) GetFieldKey(fieldName string) uint32 {
	return c.Schema.GetFieldKey(fieldName)
}

// IndexDescription describes an Index on a Collection
// and its associated metadata.
type IndexDescription struct {
	ID       uint32
	FieldIDs []uint32
}

func (index IndexDescription) IDString() string {
	return fmt.Sprint(index.ID)
}

// SchemaDescription describes the core data structure of a given type
// that can be shared between multiple collections, potentially on multiple machines.
type SchemaDescription struct {
	Fields []FieldDescription
}

//IsEmpty returns true if the SchemaDescription is empty and uninitialized
func (sd SchemaDescription) IsEmpty() bool {
	return len(sd.Fields) == 0
}

func (sd SchemaDescription) GetFieldKey(fieldName string) uint32 {
	for _, field := range sd.Fields {
		if field.Name == fieldName {
			return uint32(field.ID)
		}
	}
	return uint32(0)
}

type FieldKind uint8

// Note: These values are serialized and persisted in the database, avoid modifying existing values
const (
	FieldKind_None                 FieldKind = 0
	FieldKind_DocKey               FieldKind = 1
	FieldKind_BOOL                 FieldKind = 2
	FieldKind_BOOL_ARRAY           FieldKind = 3
	FieldKind_INT                  FieldKind = 4
	FieldKind_INT_ARRAY            FieldKind = 5
	FieldKind_FLOAT                FieldKind = 6
	FieldKind_FLOAT_ARRAY          FieldKind = 7
	FieldKind_DECIMAL              FieldKind = 8
	FieldKind_DATE                 FieldKind = 9
	FieldKind_TIMESTAMP            FieldKind = 10
	FieldKind_STRING               FieldKind = 11
	FieldKind_STRING_ARRAY         FieldKind = 12
	FieldKind_BYTES                FieldKind = 13
	FieldKind_OBJECT               FieldKind = 14 // Embedded object within the type
	FieldKind_OBJECT_ARRAY         FieldKind = 15 // Array of embedded objects
	FieldKind_FOREIGN_OBJECT       FieldKind = 16 // Embedded object, but accessed via foreign keys
	FieldKind_FOREIGN_OBJECT_ARRAY FieldKind = 17 // Array of embedded objects, accessed via foreign keys
)

type RelationType uint8

// Note: These values are serialized and persisted in the database, avoid modifying existing values
const (
	Relation_Type_ONE         RelationType = 1   // 0b0000 0001
	Relation_Type_MANY        RelationType = 2   // 0b0000 0010
	Relation_Type_ONEONE      RelationType = 4   // 0b0000 0100
	Relation_Type_ONEMANY     RelationType = 8   // 0b0000 1000
	Relation_Type_MANYMANY    RelationType = 16  // 0b0001 0000
	_                         RelationType = 32  // 0b0010 0000
	Relation_Type_INTERNAL_ID RelationType = 64  // 0b0100 0000
	Relation_Type_Primary     RelationType = 128 // 0b1000 0000 Primary reference entity on relation
)

type FieldID uint32

func (f FieldID) String() string {
	return fmt.Sprint(uint32(f))
}

type FieldDescription struct {
	Name         string
	ID           FieldID
	Kind         FieldKind
	Schema       string // If the field is an OBJECT type, then it has a target schema
	RelationName string // The name of the relation index if the field is of type FOREIGN_OBJECT
	Typ          CType
	RelationType RelationType
}

func (f FieldDescription) IsObject() bool {
	return (f.Kind == FieldKind_OBJECT) || (f.Kind == FieldKind_FOREIGN_OBJECT) ||
		(f.Kind == FieldKind_FOREIGN_OBJECT_ARRAY)
}

func (f FieldDescription) IsObjectArray() bool {
	return (f.Kind == FieldKind_FOREIGN_OBJECT_ARRAY)
}

func (m RelationType) IsSet(target RelationType) bool {
	return m&target > 0
}
