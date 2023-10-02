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
	// It is conceptually local to the node hosting the DefraDB instance, but currently there
	// is no means to update the local value so that it differs from the (global) schema name.
	Name string

	// ID is the local identifier of this collection.
	//
	// It is immutable.
	ID uint32

	// Schema contains the data type information that this Collection uses.
	Schema SchemaDescription

	// Indexes contains the secondary indexes that this Collection has.
	Indexes []IndexDescription
}

// IDString returns the collection ID as a string.
func (col CollectionDescription) IDString() string {
	return fmt.Sprint(col.ID)
}

// GetFieldByID searches for a field with the given ID. If such a field is found it
// will return it and true, if it is not found it will return false.
func (col CollectionDescription) GetFieldByID(id FieldID, schema *SchemaDescription) (FieldDescription, bool) {
	for _, field := range schema.Fields {
		if field.ID == id {
			return field, true
		}
	}
	return FieldDescription{}, false
}

// GetFieldByName returns the field for the given field name. If such a field is found it
// will return it and true, if it is not found it will return false.
func (col CollectionDescription) GetFieldByName(fieldName string, schema *SchemaDescription) (FieldDescription, bool) {
	for _, field := range schema.Fields {
		if field.Name == fieldName {
			return field, true
		}
	}
	return FieldDescription{}, false
}

// GetFieldByRelation returns the field that supports the relation of the given name.
func (col CollectionDescription) GetFieldByRelation(
	relationName string,
	otherCollectionName string,
	otherFieldName string,
	schema *SchemaDescription,
) (FieldDescription, bool) {
	for _, field := range schema.Fields {
		if field.RelationName == relationName && !(col.Name == otherCollectionName && otherFieldName == field.Name) {
			return field, true
		}
	}
	return FieldDescription{}, false
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

	// Name is the name of this Schema.
	//
	// It is currently used to define the Collection Name, and as such these two properties
	// will currently share the same name.
	//
	// It is immutable.
	Name string

	// Fields contains the fields within this Schema.
	//
	// Currently new fields may be added after initial declaration, but they cannot be removed.
	Fields []FieldDescription
}

// GetField returns the field of the given name.
func (sd SchemaDescription) GetField(name string) (FieldDescription, bool) {
	for _, field := range sd.Fields {
		if field.Name == name {
			return field, true
		}
	}
	return FieldDescription{}, false
}

// FieldKind describes the type of a field.
type FieldKind uint8

func (f FieldKind) String() string {
	switch f {
	case FieldKind_DocKey:
		return "ID"
	case FieldKind_BOOL:
		return "Boolean"
	case FieldKind_NILLABLE_BOOL_ARRAY:
		return "[Boolean]"
	case FieldKind_BOOL_ARRAY:
		return "[Boolean!]"
	case FieldKind_INT:
		return "Int"
	case FieldKind_NILLABLE_INT_ARRAY:
		return "[Int]"
	case FieldKind_INT_ARRAY:
		return "[Int!]"
	case FieldKind_DATETIME:
		return "DateTime"
	case FieldKind_FLOAT:
		return "Float"
	case FieldKind_NILLABLE_FLOAT_ARRAY:
		return "[Float]"
	case FieldKind_FLOAT_ARRAY:
		return "[Float!]"
	case FieldKind_STRING:
		return "String"
	case FieldKind_NILLABLE_STRING_ARRAY:
		return "[String]"
	case FieldKind_STRING_ARRAY:
		return "[String!]"
	default:
		return fmt.Sprint(uint8(f))
	}
}

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
	"Int":        FieldKind_INT,
	"[Int]":      FieldKind_NILLABLE_INT_ARRAY,
	"[Int!]":     FieldKind_INT_ARRAY,
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
	// Name contains the name of this field.
	//
	// It is currently immutable.
	Name string

	// ID contains the internal ID of this field.
	//
	// Whilst this ID will typically match the field's index within the Schema's Fields
	// slice, there is no guarentee that they will be the same.
	//
	// It is immutable.
	ID FieldID

	// The data type that this field holds.
	//
	// Must contain a valid value. It is currently immutable.
	Kind FieldKind

	// Schema contains the schema name of the type this field contains if this field is
	// a relation field.  Otherwise this will be empty.
	Schema string

	// RelationName the name of the relationship that this field represents if this field is
	// a relation field.  Otherwise this will be empty.
	RelationName string

	// The CRDT Type of this field. If no type has been provided it will default to [LWW_REGISTER].
	//
	// It is currently immutable.
	Typ CType

	// RelationType contains the relationship type if this field is a relation field. Otherwise this
	// will be empty.
	RelationType RelationType
}

// IsInternal returns true if this field is internally generated.
func (f FieldDescription) IsInternal() bool {
	return (f.Name == "_key") || f.RelationType&Relation_Type_INTERNAL_ID != 0
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
