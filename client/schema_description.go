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

// SchemaDescription describes a Schema and its associated metadata.
type SchemaDescription struct {
	// Root is the version agnostic identifier for this schema.
	//
	// It remains constant throughout the lifetime of this schema.
	Root string

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
	Fields []SchemaFieldDescription
}

// GetFieldByName returns the field for the given field name. If such a field is found it
// will return it and true, if it is not found it will return false.
func (s SchemaDescription) GetFieldByName(fieldName string) (SchemaFieldDescription, bool) {
	for _, field := range s.Fields {
		if field.Name == fieldName {
			return field, true
		}
	}
	return SchemaFieldDescription{}, false
}
