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

import (
	"encoding/json"
	"strconv"
	"strings"
)

// FieldKind describes the type of a field.
type FieldKind interface {
	// String returns the string representation of this FieldKind.
	String() string

	// Underlying returns the underlying Kind as a string.
	//
	// If this is an array, it will return the element kind, else it will return the same as
	// [String()].
	Underlying() string

	// IsNillable returns true if this kind supports nil values.
	IsNillable() bool

	// IsObject returns true if this FieldKind is an object type, or an array of object types.
	IsObject() bool

	// IsObjectArray returns true if this FieldKind is an object array type.
	IsObjectArray() bool

	// IsArray returns true if this FieldKind is an array type which includes inline arrays as well
	// as relation arrays.
	IsArray() bool
}

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

	// The CRDT Type of this field. If no type has been provided it will default to [LWW_REGISTER].
	//
	// It is currently immutable.
	Typ CType
}

// ScalarKind represents singular scalar field kinds, such as `Int`.
type ScalarKind uint8

// ScalarArrayKind represents arrays of simple scalar field kinds, such as `[Int]`.
type ScalarArrayKind uint8

// ObjectKind represents singular objects (foreign and embedded), such as `User`.
type ObjectKind string

// ObjectKind represents arrays of objects (foreign and embedded), such as `[User]`.
type ObjectArrayKind string

var _ FieldKind = ScalarKind(0)
var _ FieldKind = ScalarArrayKind(0)
var _ FieldKind = ObjectKind("")
var _ FieldKind = ObjectArrayKind("")

func (k ScalarKind) String() string {
	switch k {
	case FieldKind_DocID:
		return "ID"
	case FieldKind_NILLABLE_BOOL:
		return "Boolean"
	case FieldKind_NILLABLE_INT:
		return "Int"
	case FieldKind_NILLABLE_DATETIME:
		return "DateTime"
	case FieldKind_NILLABLE_FLOAT:
		return "Float"
	case FieldKind_NILLABLE_STRING:
		return "String"
	case FieldKind_NILLABLE_BLOB:
		return "Blob"
	case FieldKind_NILLABLE_JSON:
		return "JSON"
	default:
		return strconv.Itoa(int(k))
	}
}

func (k ScalarKind) Underlying() string {
	return k.String()
}

func (k ScalarKind) IsNillable() bool {
	return k != FieldKind_DocID
}

func (k ScalarKind) IsObject() bool {
	return false
}

func (k ScalarKind) IsObjectArray() bool {
	return false
}

func (k ScalarKind) IsArray() bool {
	return false
}

func (k ScalarArrayKind) String() string {
	switch k {
	case FieldKind_NILLABLE_BOOL_ARRAY:
		return "[Boolean]"
	case FieldKind_BOOL_ARRAY:
		return "[Boolean!]"
	case FieldKind_NILLABLE_INT_ARRAY:
		return "[Int]"
	case FieldKind_INT_ARRAY:
		return "[Int!]"
	case FieldKind_NILLABLE_FLOAT_ARRAY:
		return "[Float]"
	case FieldKind_FLOAT_ARRAY:
		return "[Float!]"
	case FieldKind_NILLABLE_STRING_ARRAY:
		return "[String]"
	case FieldKind_STRING_ARRAY:
		return "[String!]"
	default:
		return strconv.Itoa(int(k))
	}
}

func (k ScalarArrayKind) Underlying() string {
	return strings.Trim(k.String(), "[]")
}

func (k ScalarArrayKind) IsNillable() bool {
	return k == FieldKind_NILLABLE_BOOL_ARRAY ||
		k == FieldKind_NILLABLE_INT_ARRAY ||
		k == FieldKind_NILLABLE_FLOAT_ARRAY ||
		k == FieldKind_NILLABLE_STRING_ARRAY
}

func (k ScalarArrayKind) IsObject() bool {
	return false
}

func (k ScalarArrayKind) IsObjectArray() bool {
	return false
}

func (k ScalarArrayKind) IsArray() bool {
	return true
}

func (k ObjectKind) String() string {
	return string(k)
}

func (k ObjectKind) Underlying() string {
	return k.String()
}

func (k ObjectKind) IsNillable() bool {
	return true
}

func (k ObjectKind) IsObject() bool {
	return true
}

func (k ObjectKind) IsObjectArray() bool {
	return false
}

func (k ObjectKind) IsArray() bool {
	return false
}

func (k ObjectArrayKind) String() string {
	return "[" + string(k) + "]"
}

func (k ObjectArrayKind) Underlying() string {
	return strings.Trim(k.String(), "[]")
}

func (k ObjectArrayKind) IsNillable() bool {
	return true
}

func (k ObjectArrayKind) IsObject() bool {
	return true
}

func (k ObjectArrayKind) IsObjectArray() bool {
	return true
}

func (k ObjectArrayKind) IsArray() bool {
	return true
}

func (k ObjectArrayKind) MarshalJSON() ([]byte, error) {
	return []byte(`"` + k.String() + `"`), nil
}

// Note: These values are serialized and persisted in the database, avoid modifying existing values.
const (
	FieldKind_None                  ScalarKind      = 0
	FieldKind_DocID                 ScalarKind      = 1
	FieldKind_NILLABLE_BOOL         ScalarKind      = 2
	FieldKind_BOOL_ARRAY            ScalarArrayKind = 3
	FieldKind_NILLABLE_INT          ScalarKind      = 4
	FieldKind_INT_ARRAY             ScalarArrayKind = 5
	FieldKind_NILLABLE_FLOAT        ScalarKind      = 6
	FieldKind_FLOAT_ARRAY           ScalarArrayKind = 7
	_                               ScalarKind      = 8 // safe to repurpose (was never used)
	_                               ScalarKind      = 9 // safe to repurpose (previously old field)
	FieldKind_NILLABLE_DATETIME     ScalarKind      = 10
	FieldKind_NILLABLE_STRING       ScalarKind      = 11
	FieldKind_STRING_ARRAY          ScalarArrayKind = 12
	FieldKind_NILLABLE_BLOB         ScalarKind      = 13
	FieldKind_NILLABLE_JSON         ScalarKind      = 14
	_                               ScalarKind      = 15 // safe to repurpose (was never used)
	_                               ScalarKind      = 16 // Deprecated 2024-03-15, was FieldKind_FOREIGN_OBJECT
	_                               ScalarKind      = 17 // Deprecated 2024-03-15, was FieldKind_FOREIGN_OBJECT_ARRAY
	FieldKind_NILLABLE_BOOL_ARRAY   ScalarArrayKind = 18
	FieldKind_NILLABLE_INT_ARRAY    ScalarArrayKind = 19
	FieldKind_NILLABLE_FLOAT_ARRAY  ScalarArrayKind = 20
	FieldKind_NILLABLE_STRING_ARRAY ScalarArrayKind = 21
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
	return f.Kind.IsObject()
}

// schemaFieldDescription is a private type used to facilitate the unmarshalling
// of json to a [SchemaFieldDescription].
type schemaFieldDescription struct {
	Name string
	Typ  CType

	// Properties below this line are unmarshalled using custom logic in [UnmarshalJSON]
	Kind json.RawMessage
}

func (f *SchemaFieldDescription) UnmarshalJSON(bytes []byte) error {
	var descMap schemaFieldDescription
	err := json.Unmarshal(bytes, &descMap)
	if err != nil {
		return err
	}

	f.Name = descMap.Name
	f.Typ = descMap.Typ
	f.Kind, err = parseFieldKind(descMap.Kind)
	if err != nil {
		return err
	}

	return nil
}

func parseFieldKind(bytes json.RawMessage) (FieldKind, error) {
	if len(bytes) == 0 {
		return FieldKind_None, nil
	}

	if bytes[0] != '"' {
		// If the Kind is not represented by a string, assume try to parse it to an int, as
		// that is the only other type we support.
		var intKind uint8
		err := json.Unmarshal(bytes, &intKind)
		if err != nil {
			return nil, err
		}
		switch intKind {
		case uint8(FieldKind_BOOL_ARRAY), uint8(FieldKind_INT_ARRAY), uint8(FieldKind_FLOAT_ARRAY),
			uint8(FieldKind_STRING_ARRAY), uint8(FieldKind_NILLABLE_BOOL_ARRAY), uint8(FieldKind_NILLABLE_INT_ARRAY),
			uint8(FieldKind_NILLABLE_FLOAT_ARRAY), uint8(FieldKind_NILLABLE_STRING_ARRAY):
			return ScalarArrayKind(intKind), nil
		default:
			return ScalarKind(intKind), nil
		}
	}

	var strKind string
	err := json.Unmarshal(bytes, &strKind)
	if err != nil {
		return nil, err
	}

	kind, ok := FieldKindStringToEnumMapping[strKind]
	if ok {
		return kind, nil
	}

	// If we don't find the string representation of this type in the
	// scalar mapping, assume it is an object - if it is not, validation
	// will catch this later.  If it is unknown we have no way of telling
	// as to whether the user thought it was a scalar or an object anyway.
	if strKind[0] == '[' {
		return ObjectArrayKind(strings.Trim(strKind, "[]")), nil
	}
	return ObjectKind(strKind), nil
}
