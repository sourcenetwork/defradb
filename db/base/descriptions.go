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
}

func (index IndexDescription) IDString() string {
	return fmt.Sprint(index.ID)
}

type SchemaDescription struct {
	ID   uint32
	Name string
	Key  []byte
	// Schema schema.Schema
	FieldIDs []uint32
	Fields   []FieldDescription
}

//IsEmpty returns true if the SchemaDescription is empty and unitialized
func (sd SchemaDescription) IsEmpty() bool {
	if sd.ID == 0 &&
		len(sd.Key) == 0 &&
		len(sd.FieldIDs) == 0 &&
		len(sd.Fields) == 0 {
		return true
	}
	return false
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

const (
	Meta_Relation_ONE       uint8 = 0x01 << iota // 0b0001
	Meta_Relation_ONEMANY                        // 0b0010
	Meta_Relation_MANY_MANY                      // 0b1000
	_
	_
	_
	_
	Meta_Relation_Primary // 0b1000 0000 Primary reference entity on relation
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
	return (f.Kind == FieldKind_OBJECT) || (f.Kind == FieldKind_FOREIGN_OBJECT)
}
