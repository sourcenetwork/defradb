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
type CollectionDescription struct {
	// Name contains the name of the collection.
	//
	// It is conceptually local to node hosting the DefraDB instance, but currently there
	// is no means to update the local value so that it differs from the (global) schema name.
	Name string

	// ID is the local identifier of this collection.
	//
	// It is immutable.
	ID uint32

	// Schema contains the data type information that this Collection uses.
	Schema SchemaDescription
}

// IDString returns the collection ID as a string.
func (col CollectionDescription) IDString() string {
	return fmt.Sprint(col.ID)
}

// GetField returns the field of the given name.
func (col CollectionDescription) GetField(name string) (FieldDescription, bool) {
	if !col.Schema.IsEmpty() {
		for _, field := range col.Schema.Fields {
			if field.Name == name {
				return field, true
			}
		}
	}
	return FieldDescription{}, false
}

// GetRelation returns the field that supports the relation of the given name.
func (col CollectionDescription) GetRelation(name string) (FieldDescription, bool) {
	if !col.Schema.IsEmpty() {
		for _, field := range col.Schema.Fields {
			if field.RelationName == name {
				return field, true
			}
		}
	}
	return FieldDescription{}, false
}

// IndexDescription describes an Index on a Collection and its associated metadata.
type IndexDescription struct {
	Name     string
	ID       uint32
	Primary  bool
	Unique   bool
	FieldIDs []uint32

	// Junction is a special field, it indicates if this Index is
	// being used as a junction table for a Many-to-Many relation.
	// A Junction index needs to index the DocKey from two different
	// collections, so the usual method of storing the indexed fields
	// in the FieldIDs property won't work, since thats scoped to the
	// local schema.
	//
	// The Junction stores the DocKey of the type its assigned to,
	// and the DocKey of the target relation type. Moreover, since
	// we use a Composite Key Index system, the ordering of the keys
	// affects how we can use in the index. The initial Junction
	// Index for a type, needs to be assigned to the  "Primary"
	// type in the Many-to-Many relation. This is usually the type
	// that expects more reads from.
	//
	// Eg:
	// A Book type can have many Categories,
	// and Categories can belong to many Books.
	//
	// If we request Books more, then Categories directly, then
	// we can set the Book type as the Primary type.
	Junction bool

	// RelationType is only used in the Index is a Junction Index.
	// It specifies what the other type is in the Many-to-Many
	// relationship.
	RelationType string
}

// IDString returns the index ID as a string.
func (index IndexDescription) IDString() string {
	return fmt.Sprint(index.ID)
}

// SchemaDescription describes a Schema and its associated metadata.
type SchemaDescription struct {
	// SchemaID is the version agnostic identifier for this schema.
	//
	// It remains constant throughout the lifetime of this schema.
	SchemaID string

	// VersionID is the version-specific identifier for this schema.
	//
	// It is generated on mutation of this schema and can be used to uniquely
	// identify a schema at a specific version.
	VersionID string
	Name      string
	Fields    []FieldDescription
}

// IsEmpty returns true if the SchemaDescription is empty and uninitialized
func (sd SchemaDescription) IsEmpty() bool {
	return len(sd.Fields) == 0
}

// GetFieldKey returns the field ID for the given field name.
func (sd SchemaDescription) GetFieldKey(fieldName string) uint32 {
	for _, field := range sd.Fields {
		if field.Name == fieldName {
			return uint32(field.ID)
		}
	}
	return uint32(0)
}

// FieldKind describes the type of a field.
type FieldKind uint8

// Note: These values are serialized and persisted in the database, avoid modifying existing values.
const (
	FieldKind_None         FieldKind = 0
	FieldKind_DocKey       FieldKind = 1
	FieldKind_BOOL         FieldKind = 2
	FieldKind_BOOL_ARRAY   FieldKind = 3
	FieldKind_INT          FieldKind = 4
	FieldKind_INT_ARRAY    FieldKind = 5
	FieldKind_FLOAT        FieldKind = 6
	FieldKind_FLOAT_ARRAY  FieldKind = 7
	_                      FieldKind = 8 // safe to repurpose (was never used)
	_                      FieldKind = 9 // safe to repurpose (previoulsy old field)
	FieldKind_DATETIME     FieldKind = 10
	FieldKind_STRING       FieldKind = 11
	FieldKind_STRING_ARRAY FieldKind = 12
	_                      FieldKind = 13 // safe to repurpose (was never used)
	_                      FieldKind = 14 // safe to repurpose (was never used)
	_                      FieldKind = 15 // safe to repurpose (was never used)

	// Embedded object, but accessed via foreign keys
	FieldKind_FOREIGN_OBJECT FieldKind = 16

	// Array of embedded objects, accessed via foreign keys
	FieldKind_FOREIGN_OBJECT_ARRAY FieldKind = 17

	FieldKind_NILLABLE_BOOL_ARRAY   FieldKind = 18
	FieldKind_NILLABLE_INT_ARRAY    FieldKind = 19
	FieldKind_NILLABLE_FLOAT_ARRAY  FieldKind = 20
	FieldKind_NILLABLE_STRING_ARRAY FieldKind = 21
)

// FieldKindStringToEnumMapping maps string representations of [FieldKind] values to
// their enum values.
//
// It is currently used to by [db.PatchSchema] to allow string representations of
// [FieldKind] to be provided instead of their raw int values.  This useage may expand
// in the future.  They currently roughly correspond to the GQL field types, but this
// equality is not guarenteed.
var FieldKindStringToEnumMapping = map[string]FieldKind{
	"ID":         FieldKind_DocKey,
	"Boolean":    FieldKind_BOOL,
	"[Boolean]":  FieldKind_NILLABLE_BOOL_ARRAY,
	"[Boolean!]": FieldKind_BOOL_ARRAY,
	"Integer":    FieldKind_INT,
	"[Integer]":  FieldKind_NILLABLE_INT_ARRAY,
	"[Integer!]": FieldKind_INT_ARRAY,
	"DateTime":   FieldKind_DATETIME,
	"Float":      FieldKind_FLOAT,
	"[Float]":    FieldKind_NILLABLE_FLOAT_ARRAY,
	"[Float!]":   FieldKind_FLOAT_ARRAY,
	"String":     FieldKind_STRING,
	"[String]":   FieldKind_NILLABLE_STRING_ARRAY,
	"[String!]":  FieldKind_STRING_ARRAY,
}

// RelationType describes the type of relation between two types.
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

// FieldID is a unique identifier for a field in a schema.
type FieldID uint32

func (f FieldID) String() string {
	return fmt.Sprint(uint32(f))
}

// FieldDescription describes a field on a Schema and its associated metadata.
type FieldDescription struct {
	Name string
	ID   FieldID

	// The data type that this field holds.
	//
	// Must contain a valid value.
	Kind         FieldKind
	Schema       string // If the field is an OBJECT type, then it has a target schema
	RelationName string // The name of the relation index if the field is of type FOREIGN_OBJECT

	// The CRDT Type of this field. If no type has been provided it will default to [LWW_REGISTER].
	Typ          CType
	RelationType RelationType
	// @todo: Add relation name for specifying target relation index
	// @body: If a type has two User sub objects, you need to specify the relation
	// name used. By default the relation name is "rootType_subType". However,
	// if you have two of the same sub types, then you need to specify to
	// avoid collision.
}

// IsObject returns true if this field is an object type.
func (f FieldDescription) IsObject() bool {
	return (f.Kind == FieldKind_FOREIGN_OBJECT) ||
		(f.Kind == FieldKind_FOREIGN_OBJECT_ARRAY)
}

// IsObjectArray returns true if this field is an object array type.
func (f FieldDescription) IsObjectArray() bool {
	return (f.Kind == FieldKind_FOREIGN_OBJECT_ARRAY)
}

// IsPrimaryRelation returns true if this field is a relation, and is the primary side.
func (f FieldDescription) IsPrimaryRelation() bool {
	return f.RelationType > 0 && f.RelationType&Relation_Type_Primary != 0
}

// IsSet returns true if the target relation type is set.
func (m RelationType) IsSet(target RelationType) bool {
	return m&target > 0
}
