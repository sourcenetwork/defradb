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
	"fmt"
	"strconv"

	"github.com/sourcenetwork/defradb/client/request"
)

// FieldKind describes the type of a field.
type FieldKind interface {
	// String returns the string representation of this FieldKind.
	String() string

	// IsNillable returns true if this kind supports nil values.
	IsNillable() bool

	// IsObject returns true if this FieldKind is an object type, or an array of object types.
	IsObject() bool

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

	// DefaultValue contains the default value for this field.
	DefaultValue any
}

// ScalarKind represents singular scalar field kinds, such as `Int`.
type ScalarKind uint8

// ScalarArrayKind represents arrays of simple scalar field kinds, such as `[Int]`.
type ScalarArrayKind uint8

// CollectionKind represents a relationship with a [CollectionDescription].
type CollectionKind struct {
	// If true, this side of the relationship points to many related records.
	Array bool

	// The root ID of the related [CollectionDescription].
	Root uint32
}

// SchemaKind represents a relationship with a [SchemaDescription].
type SchemaKind struct {
	// If true, this side of the relationship points to many related records.
	Array bool

	// The root ID of the related [SchemaDescription].
	Root string
}

// NamedKind represents a temporary declaration of a relationship to another
// [CollectionDefinition].
//
// This is used only to temporarily describe a relationship, this kind will
// never be persisted in the store and instead will be converted to one of
// [CollectionKind], [SchemaKind] or [SelfKind] first.
type NamedKind struct {
	// The current name of the related [CollectionDefinition].
	Name string

	// If true, this side of the relationship points to many related records.
	Array bool
}

// SelfKind represents a relationship with the host.
//
// This includes any other schema that formed a circular dependency with the
// host at the point at which they were created.
//
// For example: the relations in User=>Dog=>User form a circle, and would be
// defined using [SelfKind] instead of [SchemaKind].
//
// This is because schema IDs are content IDs and cannot be generated for a
// single element within a circular dependency tree.
type SelfKind struct {
	// The relative ID to the related type.  If this points at its host this
	// will be empty.
	RelativeID string

	// If true, this side of the relationship points to many related records.
	Array bool
}

var _ FieldKind = ScalarKind(0)
var _ FieldKind = ScalarArrayKind(0)
var _ FieldKind = (*CollectionKind)(nil)
var _ FieldKind = (*SchemaKind)(nil)
var _ FieldKind = (*SelfKind)(nil)
var _ FieldKind = (*NamedKind)(nil)

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

func (k ScalarKind) IsNillable() bool {
	return true
}

func (k ScalarKind) IsObject() bool {
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

func (k ScalarArrayKind) IsNillable() bool {
	return true
}

func (k ScalarArrayKind) IsObject() bool {
	return false
}

func (k ScalarArrayKind) IsArray() bool {
	return true
}

func NewCollectionKind(root uint32, isArray bool) *CollectionKind {
	return &CollectionKind{
		Root:  root,
		Array: isArray,
	}
}

func (k *CollectionKind) String() string {
	if k.Array {
		return fmt.Sprintf("[%v]", k.Root)
	}
	return strconv.FormatInt(int64(k.Root), 10)
}

func (k *CollectionKind) IsNillable() bool {
	return true
}

func (k *CollectionKind) IsObject() bool {
	return true
}

func (k *CollectionKind) IsArray() bool {
	return k.Array
}

func NewSchemaKind(root string, isArray bool) *SchemaKind {
	return &SchemaKind{
		Root:  root,
		Array: isArray,
	}
}

func (k *SchemaKind) String() string {
	if k.Array {
		return fmt.Sprintf("[%v]", k.Root)
	}
	return k.Root
}

func (k *SchemaKind) IsNillable() bool {
	return true
}

func (k *SchemaKind) IsObject() bool {
	return true
}

func (k *SchemaKind) IsArray() bool {
	return k.Array
}

func NewSelfKind(relativeID string, isArray bool) *SelfKind {
	return &SelfKind{
		RelativeID: relativeID,
		Array:      isArray,
	}
}

func (k *SelfKind) String() string {
	var relativeName string
	if k.RelativeID != "" {
		relativeName = fmt.Sprintf("%s-%s", request.SelfTypeName, k.RelativeID)
	} else {
		relativeName = request.SelfTypeName
	}

	if k.Array {
		return fmt.Sprintf("[%s]", relativeName)
	}
	return relativeName
}

func (k *SelfKind) IsNillable() bool {
	return true
}

func (k *SelfKind) IsObject() bool {
	return true
}

func (k *SelfKind) IsArray() bool {
	return k.Array
}

func NewNamedKind(name string, isArray bool) *NamedKind {
	return &NamedKind{
		Name:  name,
		Array: isArray,
	}
}

func (k *NamedKind) String() string {
	if k.Array {
		return fmt.Sprintf("[%v]", k.Name)
	}
	return k.Name
}

func (k *NamedKind) IsNillable() bool {
	return true
}

func (k *NamedKind) IsObject() bool {
	return true
}

func (k *NamedKind) IsArray() bool {
	return k.Array
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
	"ID":                 FieldKind_DocID,
	"Boolean":            FieldKind_NILLABLE_BOOL,
	"[Boolean]":          FieldKind_NILLABLE_BOOL_ARRAY,
	"[Boolean!]":         FieldKind_BOOL_ARRAY,
	"Int":                FieldKind_NILLABLE_INT,
	"[Int]":              FieldKind_NILLABLE_INT_ARRAY,
	"[Int!]":             FieldKind_INT_ARRAY,
	"DateTime":           FieldKind_NILLABLE_DATETIME,
	"Float":              FieldKind_NILLABLE_FLOAT,
	"[Float]":            FieldKind_NILLABLE_FLOAT_ARRAY,
	"[Float!]":           FieldKind_FLOAT_ARRAY,
	"String":             FieldKind_NILLABLE_STRING,
	"[String]":           FieldKind_NILLABLE_STRING_ARRAY,
	"[String!]":          FieldKind_STRING_ARRAY,
	"Blob":               FieldKind_NILLABLE_BLOB,
	"JSON":               FieldKind_NILLABLE_JSON,
	request.SelfTypeName: NewSelfKind("", false),
	fmt.Sprintf("[%s]", request.SelfTypeName): NewSelfKind("", true),
}

// IsRelation returns true if this field is a relation.
func (f SchemaFieldDescription) IsRelation() bool {
	return f.Kind.IsObject()
}

// schemaFieldDescription is a private type used to facilitate the unmarshalling
// of json to a [SchemaFieldDescription].
type schemaFieldDescription struct {
	Name         string
	Typ          CType
	DefaultValue any

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
	f.DefaultValue = descMap.DefaultValue
	f.Kind, err = parseFieldKind(descMap.Kind)
	if err != nil {
		return err
	}

	return nil
}

// objectKind is a private type used to facilitate the unmarshalling
// of json to a [FieldKind].
type objectKind struct {
	Array      bool
	Root       any
	RelativeID string
}

func parseFieldKind(bytes json.RawMessage) (FieldKind, error) {
	if len(bytes) == 0 {
		return FieldKind_None, nil
	}

	if bytes[0] == '{' {
		var objKind objectKind
		err := json.Unmarshal(bytes, &objKind)
		if err != nil {
			return nil, err
		}

		if objKind.Root == nil {
			return NewSelfKind(objKind.RelativeID, objKind.Array), nil
		}

		switch root := objKind.Root.(type) {
		case float64:
			return NewCollectionKind(uint32(root), objKind.Array), nil
		case string:
			return NewSchemaKind(root, objKind.Array), nil
		default:
			return nil, NewErrFailedToParseKind(bytes)
		}
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

	isArray := strKind[0] == '['
	if isArray {
		// Strip the brackets
		strKind = strKind[1 : len(strKind)-1]
	}

	// This is used by patch schema/collection, where new fields added
	// by users will be initially added as [NamedKind]s.
	return NewNamedKind(strKind, isArray), nil
}
