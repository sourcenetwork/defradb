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

// CollectionDefinition contains the metadata defining what a Collection is.
//
// The definition types ([CollectionDefinition], [FieldDefinition]) are read-only types returned
// from various functions as a convienient means to access the computated convergence of schema
// and collection descriptions.
type CollectionDefinition struct {
	// Description returns the CollectionDescription of this Collection.
	Description CollectionDescription `json:"description"`
	// Schema returns the SchemaDescription used to define this Collection.
	Schema SchemaDescription `json:"schema"`
}

// GetFieldByName returns the field for the given field name. If such a field is found it
// will return it and true, if it is not found it will return false.
func (def CollectionDefinition) GetFieldByName(fieldName string) (FieldDefinition, bool) {
	collectionField, ok := def.Description.GetFieldByName(fieldName)
	if ok {
		schemaField, ok := def.Schema.GetFieldByName(fieldName)
		if ok {
			return NewFieldDefinition(
				collectionField,
				schemaField,
			), true
		}
	}
	return FieldDefinition{}, false
}

// GetFields returns the combined local and global field elements on this [CollectionDefinition]
// as a single set.
func (def CollectionDefinition) GetFields() []FieldDefinition {
	fields := []FieldDefinition{}
	for _, localField := range def.Description.Fields {
		globalField, ok := def.Schema.GetFieldByName(localField.Name)
		if ok {
			fields = append(
				fields,
				NewFieldDefinition(localField, globalField),
			)
		}
	}
	return fields
}

// FieldDefinition describes the combined local and global set of properties that constitutes
// a field on a collection.
//
// It draws it's information from the [CollectionFieldDescription] on the [CollectionDescription],
// and the [SchemaFieldDescription] on the [SchemaDescription].
//
// It is to [CollectionFieldDescription] and [SchemaFieldDescription] what [CollectionDefinition]
// is to [CollectionDescription] and [SchemaDescription].
//
// The definition types ([CollectionDefinition], [FieldDefinition]) are read-only types returned
// from various functions as a convienient means to access the computated convergence of schema
// and collection descriptions.
type FieldDefinition struct {
	// Name contains the name of this field.
	Name string

	// ID contains the local, internal ID of this field.
	ID FieldID

	// The data type that this field holds.
	//
	// Must contain a valid value. It is currently immutable.
	Kind FieldKind

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

// NewFieldDefinition returns a new [FieldDefinition], combining the given local and global elements
// into a single object.
func NewFieldDefinition(local CollectionFieldDescription, global SchemaFieldDescription) FieldDefinition {
	return FieldDefinition{
		Name:              global.Name,
		ID:                local.ID,
		Kind:              global.Kind,
		RelationName:      global.RelationName,
		Typ:               global.Typ,
		IsPrimaryRelation: global.IsPrimaryRelation,
	}
}

// IsRelation returns true if this field is a relation.
func (f FieldDefinition) IsRelation() bool {
	return f.RelationName != ""
}
