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
	ID  uint32
	Key []byte
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
)

type FieldDescription struct {
	Name string
	ID   uint32
	Kind FieldKind
	Typ  core.CType
}
