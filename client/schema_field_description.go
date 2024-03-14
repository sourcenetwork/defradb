// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import "fmt"

// FieldKind describes the type of a field.
type FieldKind uint8

// SchemaFieldDescription describes a field on a Schema and its associated metadata.
type SchemaFieldDescription struct {
	// Name contains the name of this field.
	//
	// It is currently immutable.
	Name string

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

	// If true, this is the primary half of a relation, otherwise is false.
	IsPrimaryRelation bool
}

func (f FieldKind) String() string {
	switch f {
	case FieldKind_DocID:
		return "ID"
	case FieldKind_NILLABLE_BOOL:
		return "Boolean"
	case FieldKind_NILLABLE_BOOL_ARRAY:
		return "[Boolean]"
	case FieldKind_BOOL_ARRAY:
		return "[Boolean!]"
	case FieldKind_NILLABLE_INT:
		return "Int"
	case FieldKind_NILLABLE_INT_ARRAY:
		return "[Int]"
	case FieldKind_INT_ARRAY:
		return "[Int!]"
	case FieldKind_NILLABLE_DATETIME:
		return "DateTime"
	case FieldKind_NILLABLE_FLOAT:
		return "Float"
	case FieldKind_NILLABLE_FLOAT_ARRAY:
		return "[Float]"
	case FieldKind_FLOAT_ARRAY:
		return "[Float!]"
	case FieldKind_NILLABLE_STRING:
		return "String"
	case FieldKind_NILLABLE_STRING_ARRAY:
		return "[String]"
	case FieldKind_STRING_ARRAY:
		return "[String!]"
	case FieldKind_NILLABLE_BLOB:
		return "Blob"
	case FieldKind_NILLABLE_JSON:
		return "JSON"
	default:
		return fmt.Sprint(uint8(f))
	}
}

// IsObject returns true if this FieldKind is an object type.
func (f FieldKind) IsObject() bool {
	return f == FieldKind_FOREIGN_OBJECT ||
		f == FieldKind_FOREIGN_OBJECT_ARRAY
}

// IsObjectArray returns true if this FieldKind is an object array type.
func (f FieldKind) IsObjectArray() bool {
	return f == FieldKind_FOREIGN_OBJECT_ARRAY
}

// IsArray returns true if this FieldKind is an array type which includes inline arrays as well
// as relation arrays.
func (f FieldKind) IsArray() bool {
	return f == FieldKind_BOOL_ARRAY ||
		f == FieldKind_INT_ARRAY ||
		f == FieldKind_FLOAT_ARRAY ||
		f == FieldKind_STRING_ARRAY ||
		f == FieldKind_FOREIGN_OBJECT_ARRAY ||
		f == FieldKind_NILLABLE_BOOL_ARRAY ||
		f == FieldKind_NILLABLE_INT_ARRAY ||
		f == FieldKind_NILLABLE_FLOAT_ARRAY ||
		f == FieldKind_NILLABLE_STRING_ARRAY
}

// Note: These values are serialized and persisted in the database, avoid modifying existing values.
const (
	FieldKind_None              FieldKind = 0
	FieldKind_DocID             FieldKind = 1
	FieldKind_NILLABLE_BOOL     FieldKind = 2
	FieldKind_BOOL_ARRAY        FieldKind = 3
	FieldKind_NILLABLE_INT      FieldKind = 4
	FieldKind_INT_ARRAY         FieldKind = 5
	FieldKind_NILLABLE_FLOAT    FieldKind = 6
	FieldKind_FLOAT_ARRAY       FieldKind = 7
	_                           FieldKind = 8 // safe to repurpose (was never used)
	_                           FieldKind = 9 // safe to repurpose (previously old field)
	FieldKind_NILLABLE_DATETIME FieldKind = 10
	FieldKind_NILLABLE_STRING   FieldKind = 11
	FieldKind_STRING_ARRAY      FieldKind = 12
	FieldKind_NILLABLE_BLOB     FieldKind = 13
	FieldKind_NILLABLE_JSON     FieldKind = 14
	_                           FieldKind = 15 // safe to repurpose (was never used)

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
// [FieldKind] to be provided instead of their raw int values.  This usage may expand
// in the future.  They currently roughly correspond to the GQL field types, but this
// equality is not guaranteed.
var FieldKindStringToEnumMapping = map[string]FieldKind{
	"ID":         FieldKind_DocID,
	"Boolean":    FieldKind_NILLABLE_BOOL,
	"[Boolean]":  FieldKind_NILLABLE_BOOL_ARRAY,
	"[Boolean!]": FieldKind_BOOL_ARRAY,
	"Int":        FieldKind_NILLABLE_INT,
	"[Int]":      FieldKind_NILLABLE_INT_ARRAY,
	"[Int!]":     FieldKind_INT_ARRAY,
	"DateTime":   FieldKind_NILLABLE_DATETIME,
	"Float":      FieldKind_NILLABLE_FLOAT,
	"[Float]":    FieldKind_NILLABLE_FLOAT_ARRAY,
	"[Float!]":   FieldKind_FLOAT_ARRAY,
	"String":     FieldKind_NILLABLE_STRING,
	"[String]":   FieldKind_NILLABLE_STRING_ARRAY,
	"[String!]":  FieldKind_STRING_ARRAY,
	"Blob":       FieldKind_NILLABLE_BLOB,
	"JSON":       FieldKind_NILLABLE_JSON,
}

// IsRelation returns true if this field is a relation.
func (f SchemaFieldDescription) IsRelation() bool {
	return f.RelationName != ""
}
