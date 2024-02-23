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
type FieldDefinition struct {
	// Name contains the name of this field.
	Name string

	// ID contains the local, internal ID of this field.
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
		Schema:            global.Schema,
		RelationName:      global.RelationName,
		Typ:               global.Typ,
		IsPrimaryRelation: global.IsPrimaryRelation,
	}
}

// IsObject returns true if this field is an object type.
func (f FieldDefinition) IsObject() bool {
	return (f.Kind == FieldKind_FOREIGN_OBJECT) ||
		(f.Kind == FieldKind_FOREIGN_OBJECT_ARRAY)
}

// IsObjectArray returns true if this field is an object array type.
func (f FieldDefinition) IsObjectArray() bool {
	return (f.Kind == FieldKind_FOREIGN_OBJECT_ARRAY)
}

// IsRelation returns true if this field is a relation.
func (f FieldDefinition) IsRelation() bool {
	return f.RelationName != ""
}

// IsArray returns true if this field is an array type which includes inline arrays as well
// as relation arrays.
func (f FieldDefinition) IsArray() bool {
	return f.Kind == FieldKind_BOOL_ARRAY ||
		f.Kind == FieldKind_INT_ARRAY ||
		f.Kind == FieldKind_FLOAT_ARRAY ||
		f.Kind == FieldKind_STRING_ARRAY ||
		f.Kind == FieldKind_FOREIGN_OBJECT_ARRAY ||
		f.Kind == FieldKind_NILLABLE_BOOL_ARRAY ||
		f.Kind == FieldKind_NILLABLE_INT_ARRAY ||
		f.Kind == FieldKind_NILLABLE_FLOAT_ARRAY ||
		f.Kind == FieldKind_NILLABLE_STRING_ARRAY
}
