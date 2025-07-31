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

// ScalarKind represents singular scalar field kinds, such as `Int`.
type ScalarKind uint8

// ScalarArrayKind represents arrays of simple scalar field kinds, such as `[Int]`.
type ScalarArrayKind uint8

// CollectionKind represents a relationship with a collection.
type CollectionKind struct {
	// If true, this side of the relationship points to many related records.
	Array bool

	// The CollectionID of the related collection.
	CollectionID string
}

// NamedKind represents a temporary declaration of a relationship to another
// [CollectionDefinition].
//
// This is used only to temporarily describe a relationship, this kind will
// never be persisted in the store and instead will be converted to one of
// [CollectionKind], [CollectionKind] or [SelfKind] first.
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
// defined using [SelfKind] instead of [CollectionKind].
//
// This is because schema IDs are content IDs and cannot be generated for a
// single element within a circular dependency tree.
type SelfKind struct {
	// todo: RelativeID should be made into an Option[int], see the following
	// ticket for more info: https://github.com/sourcenetwork/defradb/issues/3880

	// The relative ID to the related type.  If this points at its host this
	// will be empty.
	RelativeID string

	// If true, this side of the relationship points to many related records.
	Array bool
}

var _ FieldKind = ScalarKind(0)
var _ FieldKind = ScalarArrayKind(0)
var _ FieldKind = (*CollectionKind)(nil)
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
	case FieldKind_NILLABLE_FLOAT64:
		return "Float64"
	case FieldKind_NILLABLE_FLOAT32:
		return "Float32"
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
	case FieldKind_NILLABLE_FLOAT64_ARRAY:
		return "[Float64]"
	case FieldKind_FLOAT64_ARRAY:
		return "[Float64!]"
	case FieldKind_NILLABLE_FLOAT32_ARRAY:
		return "[Float32]"
	case FieldKind_FLOAT32_ARRAY:
		return "[Float32!]"
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

func (k ScalarArrayKind) SubKind() ScalarKind {
	switch k {
	case FieldKind_NILLABLE_BOOL_ARRAY:
		return FieldKind_NILLABLE_BOOL
	case FieldKind_BOOL_ARRAY:
		return FieldKind_NILLABLE_BOOL
	case FieldKind_NILLABLE_INT_ARRAY:
		return FieldKind_NILLABLE_INT
	case FieldKind_INT_ARRAY:
		return FieldKind_NILLABLE_INT
	case FieldKind_NILLABLE_FLOAT64_ARRAY:
		return FieldKind_NILLABLE_FLOAT64
	case FieldKind_FLOAT64_ARRAY:
		return FieldKind_NILLABLE_FLOAT64
	case FieldKind_NILLABLE_FLOAT32_ARRAY:
		return FieldKind_NILLABLE_FLOAT32
	case FieldKind_FLOAT32_ARRAY:
		return FieldKind_NILLABLE_FLOAT32
	case FieldKind_NILLABLE_STRING_ARRAY:
		return FieldKind_NILLABLE_STRING
	case FieldKind_STRING_ARRAY:
		return FieldKind_NILLABLE_STRING
	default:
		return FieldKind_None
	}
}

func NewCollectionKind(root string, isArray bool) *CollectionKind {
	return &CollectionKind{
		CollectionID: root,
		Array:        isArray,
	}
}

func (k *CollectionKind) String() string {
	if k.Array {
		return fmt.Sprintf("[%v]", k.CollectionID)
	}
	return k.CollectionID
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
	FieldKind_None                   ScalarKind      = 0
	FieldKind_DocID                  ScalarKind      = 1
	FieldKind_NILLABLE_BOOL          ScalarKind      = 2
	FieldKind_BOOL_ARRAY             ScalarArrayKind = 3
	FieldKind_NILLABLE_INT           ScalarKind      = 4
	FieldKind_INT_ARRAY              ScalarArrayKind = 5
	FieldKind_NILLABLE_FLOAT64       ScalarKind      = 6
	FieldKind_FLOAT64_ARRAY          ScalarArrayKind = 7
	FieldKind_NILLABLE_FLOAT32       ScalarKind      = 8
	FieldKind_FLOAT32_ARRAY          ScalarArrayKind = 9
	FieldKind_NILLABLE_DATETIME      ScalarKind      = 10
	FieldKind_NILLABLE_STRING        ScalarKind      = 11
	FieldKind_STRING_ARRAY           ScalarArrayKind = 12
	FieldKind_NILLABLE_BLOB          ScalarKind      = 13
	FieldKind_NILLABLE_JSON          ScalarKind      = 14
	_                                ScalarKind      = 15 // safe to repurpose (was never used)
	_                                ScalarKind      = 16 // Deprecated 2024-03-15, was FieldKind_FOREIGN_OBJECT
	_                                ScalarKind      = 17 // Deprecated 2024-03-15, was FieldKind_FOREIGN_OBJECT_ARRAY
	FieldKind_NILLABLE_BOOL_ARRAY    ScalarArrayKind = 18
	FieldKind_NILLABLE_INT_ARRAY     ScalarArrayKind = 19
	FieldKind_NILLABLE_FLOAT64_ARRAY ScalarArrayKind = 20
	FieldKind_NILLABLE_STRING_ARRAY  ScalarArrayKind = 21
	FieldKind_NILLABLE_FLOAT32_ARRAY ScalarArrayKind = 22
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
	"Float":              FieldKind_NILLABLE_FLOAT64,
	"[Float]":            FieldKind_NILLABLE_FLOAT64_ARRAY,
	"[Float!]":           FieldKind_FLOAT64_ARRAY,
	"Float64":            FieldKind_NILLABLE_FLOAT64,
	"[Float64]":          FieldKind_NILLABLE_FLOAT64_ARRAY,
	"[Float64!]":         FieldKind_FLOAT64_ARRAY,
	"Float32":            FieldKind_NILLABLE_FLOAT32,
	"[Float32]":          FieldKind_NILLABLE_FLOAT32_ARRAY,
	"[Float32!]":         FieldKind_FLOAT32_ARRAY,
	"String":             FieldKind_NILLABLE_STRING,
	"[String]":           FieldKind_NILLABLE_STRING_ARRAY,
	"[String!]":          FieldKind_STRING_ARRAY,
	"Blob":               FieldKind_NILLABLE_BLOB,
	"JSON":               FieldKind_NILLABLE_JSON,
	request.SelfTypeName: NewSelfKind("", false),
	fmt.Sprintf("[%s]", request.SelfTypeName): NewSelfKind("", true),
}

// objectKind is a private type used to facilitate the unmarshalling
// of json to a [FieldKind].
type objectKind struct {
	Array        bool
	CollectionID string
	RelativeID   string
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

		if objKind.CollectionID == "" {
			return NewSelfKind(objKind.RelativeID, objKind.Array), nil
		}

		return NewCollectionKind(objKind.CollectionID, objKind.Array), nil
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
		case uint8(FieldKind_BOOL_ARRAY), uint8(FieldKind_INT_ARRAY), uint8(FieldKind_FLOAT64_ARRAY),
			uint8(FieldKind_STRING_ARRAY), uint8(FieldKind_NILLABLE_BOOL_ARRAY), uint8(FieldKind_NILLABLE_INT_ARRAY),
			uint8(FieldKind_NILLABLE_FLOAT64_ARRAY), uint8(FieldKind_NILLABLE_STRING_ARRAY),
			uint8(FieldKind_FLOAT32_ARRAY), uint8(FieldKind_NILLABLE_FLOAT32_ARRAY):
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

// IsVectorEmbeddingCompatible returns true if the FieldKind is an array that contains
// a supported scalar type for vector embeddings.
func IsVectorEmbeddingCompatible(kind FieldKind) bool {
	if arrKind, ok := kind.(ScalarArrayKind); ok {
		switch arrKind {
		case FieldKind_FLOAT32_ARRAY:
			return true
		default:
			return false
		}
	}
	return false
}
