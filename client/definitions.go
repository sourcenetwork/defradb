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
	collectionField, existsOnCollection := def.Description.GetFieldByName(fieldName)
	schemaField, existsOnSchema := def.Schema.GetFieldByName(fieldName)

	if existsOnCollection && existsOnSchema {
		return NewFieldDefinition(
			collectionField,
			schemaField,
		), true
	} else if existsOnCollection && !existsOnSchema {
		// If the field exists only on the collection, it is a local only field, for example the
		// secondary side of a relation.
		return NewLocalFieldDefinition(
			collectionField,
		), true
	} else if !existsOnCollection && existsOnSchema {
		// If the field only exist on the schema it is likely that this is a schema-only object
		// definition, for example for an embedded object.
		return NewSchemaOnlyFieldDefinition(
			schemaField,
		), true
	}

	return FieldDefinition{}, false
}

// GetFields returns the combined local and global field elements on this [CollectionDefinition]
// as a single set.
func (def CollectionDefinition) GetFields() []FieldDefinition {
	fields := []FieldDefinition{}
	localFieldNames := map[string]struct{}{}

	for _, localField := range def.Description.Fields {
		globalField, ok := def.Schema.GetFieldByName(localField.Name)
		if ok {
			fields = append(
				fields,
				NewFieldDefinition(localField, globalField),
			)
		} else {
			// This must be a local only field, for example the secondary side of a relation.
			fields = append(
				fields,
				NewLocalFieldDefinition(localField),
			)
		}
		localFieldNames[localField.Name] = struct{}{}
	}

	for _, schemaField := range def.Schema.Fields {
		if _, ok := localFieldNames[schemaField.Name]; ok {
			continue
		}
		// This must be a global only field, for example on an embedded object.
		fields = append(
			fields,
			NewSchemaOnlyFieldDefinition(schemaField),
		)
	}

	return fields
}

// GetName gets the name of this definition.
//
// If the collection description has a name (e.g. it is an active collection) it will return that,
// otherwise it will return the schema name.
func (def CollectionDefinition) GetName() string {
	if def.Description.Name.HasValue() {
		return def.Description.Name.Value()
	}
	return def.Schema.Name
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
	var kind FieldKind
	if local.Kind.HasValue() {
		kind = local.Kind.Value()
	} else {
		kind = global.Kind
	}

	return FieldDefinition{
		Name:              global.Name,
		ID:                local.ID,
		Kind:              kind,
		RelationName:      local.RelationName.Value(),
		Typ:               global.Typ,
		IsPrimaryRelation: kind.IsObject() && !kind.IsArray(),
	}
}

// NewLocalFieldDefinition returns a new [FieldDefinition] from the given local [CollectionFieldDescription].
func NewLocalFieldDefinition(local CollectionFieldDescription) FieldDefinition {
	return FieldDefinition{
		Name:         local.Name,
		ID:           local.ID,
		Kind:         local.Kind.Value(),
		RelationName: local.RelationName.Value(),
	}
}

// NewSchemaOnlyFieldDefinition returns a new [FieldDefinition] from the given global [SchemaFieldDescription].
func NewSchemaOnlyFieldDefinition(global SchemaFieldDescription) FieldDefinition {
	return FieldDefinition{
		Name: global.Name,
		Kind: global.Kind,
		Typ:  global.Typ,
	}
}

// IsRelation returns true if this field is a relation.
func (f FieldDefinition) IsRelation() bool {
	return f.RelationName != ""
}
