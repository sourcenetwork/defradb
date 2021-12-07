// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package base

import (
	"fmt"

	"github.com/sourcenetwork/defradb/core"
)

const (
	ObjectMarker = byte(0xff) // @todo: Investigate object marker values
)

// CollectionDescription describes a Collection and
// all its associated metadata
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
	if !col.Schema.IsEmpty() {
		for _, field := range col.Schema.Fields {
			if field.Name == name {
				return field, true
			}
		}
	}
	return FieldDescription{}, false
}

// IndexDescription describes an Index on a Collection
// and its assocatied metadata.
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
	// and the DocKey of the target relation type. Morever, since
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
	// If we query more for Books, then Categories directly, then
	// we can set the Book type as the Primary type.
	Junction bool
	// RelationType is only used in the Index is a Junction Index.
	// It specifies what the other type is in the Many-to-Many
	// relationship.
	RelationType string
}

func (index IndexDescription) IDString() string {
	return fmt.Sprint(index.ID)
}

type SchemaDescription struct {
	ID   uint32
	Name string
	Key  []byte // DocKey for verioned source schema
	// Schema schema.Schema
	FieldIDs []uint32
	Fields   []FieldDescription
}

//IsEmpty returns true if the SchemaDescription is empty and unitialized
func (sd SchemaDescription) IsEmpty() bool {
	return len(sd.Fields) == 0
}

type FieldKind uint8

const (
	FieldKind_None FieldKind = iota
	FieldKind_DocKey
	FieldKind_BOOL
	FieldKind_INT
	FieldKind_FLOAT
	FieldKind_DECIMNAL
	FieldKind_DATE
	FieldKind_TIMESTAMP
	FieldKind_STRING
	FieldKind_BYTES
	FieldKind_OBJECT               // Embedded object within the type
	FieldKind_OBJECT_ARRAY         // Array of embedded objects
	FieldKind_FOREIGN_OBJECT       // Embedded object, but accessed via foreign keys
	FieldKind_FOREIGN_OBJECT_ARRAY // Array of embedded objects, accesed via foreign keys
)

// type RelationType uint8

const (
	Meta_Relation_ONE      uint8 = 0x01 << iota // 0b0000 0001
	Meta_Relation_MANY                          // 0b0000 0010
	Meta_Relation_ONEONE                        // 0b0000 0100
	Meta_Relation_ONEMANY                       // 0b0000 1000
	Meta_Relation_MANYMANY                      // 0b0001 0000
	_                                           // 0b0010 0000
	_                                           // 0b0100 0000
	Meta_Relation_Primary                       // 0b1000 0000 Primary reference entity on relation
)

type FieldID uint32

type FieldDescription struct {
	Name         string
	ID           FieldID
	Kind         FieldKind
	Schema       string // If the field is an OBJECT type, then it has a target schema
	RelationName string // The name of the relation index if the field is of type FORIEGN_OBJECT
	Typ          core.CType
	Meta         uint8
	// @todo: Add relation name for specifying target relation index
	// @body: If a type has two User sub objects, you need to specify the relation
	// name used. By default the relation name is "rootType_subType". However,
	// if you have two of the same sub types, then you need to specify to
	// avoid collision.
}

func (f FieldDescription) IsObject() bool {
	return (f.Kind == FieldKind_OBJECT) || (f.Kind == FieldKind_FOREIGN_OBJECT) ||
		(f.Kind == FieldKind_FOREIGN_OBJECT_ARRAY)
}

func IsSet(val, target uint8) bool {
	return val&target > 0
}
