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
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/datastore"
)

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

	// DefaultValue contains the default value for this field.
	DefaultValue any
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
		DefaultValue:      local.DefaultValue,
	}
}

// NewLocalFieldDefinition returns a new [FieldDefinition] from the given local [CollectionFieldDescription].
func NewLocalFieldDefinition(local CollectionFieldDescription) FieldDefinition {
	return FieldDefinition{
		Name:         local.Name,
		ID:           local.ID,
		Kind:         local.Kind.Value(),
		RelationName: local.RelationName.Value(),
		DefaultValue: local.DefaultValue,
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

// GetSecondaryRelationField returns the secondary side field definition of this field
// from the relationship on the given collection definition and a bool indicating
// if the secondary side of the relation was found.
func (f FieldDefinition) GetSecondaryRelationField(c CollectionDefinition) (FieldDefinition, bool) {
	if f.RelationName == "" || f.Kind != FieldKind_DocID {
		return FieldDefinition{}, false
	}
	secondary, valid := c.GetFieldByName(strings.TrimSuffix(f.Name, request.RelatedObjectID))
	return secondary, valid && !secondary.IsPrimaryRelation
}

// DefinitionCache is an object providing easy access to cached collection definitions.
type DefinitionCache struct {
	// The full set of [CollectionDefinition]s within this cache
	Definitions []CollectionDefinition

	// The cached Definitions mapped by the Root of their [SchemaDescription]
	DefinitionsBySchemaRoot map[string]CollectionDefinition

	// The cached Definitions mapped by the Root of their [CollectionDescription]
	DefinitionsByCollectionRoot map[uint32]CollectionDefinition
}

// NewDefinitionCache creates a new [DefinitionCache] populated with the given [CollectionDefinition]s.
func NewDefinitionCache(definitions []CollectionDefinition) DefinitionCache {
	definitionsBySchemaRoot := make(map[string]CollectionDefinition, len(definitions))
	definitionsByCollectionRoot := make(map[uint32]CollectionDefinition, len(definitions))

	for _, def := range definitions {
		definitionsBySchemaRoot[def.Schema.Root] = def
		definitionsByCollectionRoot[def.Description.RootID] = def
	}

	return DefinitionCache{
		Definitions:                 definitions,
		DefinitionsBySchemaRoot:     definitionsBySchemaRoot,
		DefinitionsByCollectionRoot: definitionsByCollectionRoot,
	}
}

// GetDefinition returns the definition that the given [FieldKind] points to, if it is found in the
// given [DefinitionCache].
//
// If the related definition is not found, default and false will be returned.
func GetDefinition(
	cache DefinitionCache,
	host CollectionDefinition,
	kind FieldKind,
) (CollectionDefinition, bool) {
	switch typedKind := kind.(type) {
	case *NamedKind:
		for _, def := range cache.Definitions {
			if def.GetName() == typedKind.Name {
				return def, true
			}
		}

		return CollectionDefinition{}, false

	case *SchemaKind:
		def, ok := cache.DefinitionsBySchemaRoot[typedKind.Root]
		return def, ok

	case *CollectionKind:
		def, ok := cache.DefinitionsByCollectionRoot[typedKind.Root]
		return def, ok

	case *SelfKind:
		if host.Description.RootID != 0 {
			return host, true
		}

		if typedKind.RelativeID == "" {
			return host, true
		}

		hostIDBase := strings.Split(host.Schema.Root, "-")[0]
		targetID := fmt.Sprintf("%s-%s", hostIDBase, typedKind.RelativeID)

		def, ok := cache.DefinitionsBySchemaRoot[targetID]
		return def, ok

	default:
		// no-op
	}

	return CollectionDefinition{}, false
}

// GetDefinitionFromStore returns the definition that the given [FieldKind] points to, if it is found
// in the given store.
//
// If the related definition is not found, or an error occurs, default and false will be returned.
func GetDefinitionFromStore(
	ctx context.Context,
	store Store,
	host CollectionDefinition,
	kind FieldKind,
) (CollectionDefinition, bool, error) {
	switch typedKind := kind.(type) {
	case *NamedKind:
		col, err := store.GetCollectionByName(ctx, typedKind.Name)
		if errors.Is(err, datastore.ErrNotFound) {
			schemas, err := store.GetSchemas(ctx, SchemaFetchOptions{
				Name: immutable.Some(typedKind.Name),
			})
			if len(schemas) == 0 || err != nil {
				return CollectionDefinition{}, false, err
			}

			return CollectionDefinition{
				// todo - returning the first is a temporary simplification until
				// https://github.com/sourcenetwork/defradb/issues/2934
				Schema: schemas[0],
			}, true, nil
		} else if err != nil {
			return CollectionDefinition{}, false, err
		}

		return col.Definition(), true, nil

	case *SchemaKind:
		schemas, err := store.GetSchemas(ctx, SchemaFetchOptions{
			Root: immutable.Some(typedKind.Root),
		})
		if len(schemas) == 0 || err != nil {
			return CollectionDefinition{}, false, err
		}

		return CollectionDefinition{
			// todo - returning the first is a temporary simplification until
			// https://github.com/sourcenetwork/defradb/issues/2934
			Schema: schemas[0],
		}, true, nil

	case *CollectionKind:
		cols, err := store.GetCollections(ctx, CollectionFetchOptions{
			Root: immutable.Some(typedKind.Root),
		})

		if len(cols) == 0 || err != nil {
			return CollectionDefinition{}, false, err
		}

		return cols[0].Definition(), true, nil

	case *SelfKind:
		if host.Description.RootID != 0 {
			return host, true, nil
		}

		if typedKind.RelativeID == "" {
			return host, true, nil
		}

		hostIDBase := strings.Split(host.Schema.Root, "-")[0]
		targetID := fmt.Sprintf("%s-%s", hostIDBase, typedKind.RelativeID)

		cols, err := store.GetCollections(ctx, CollectionFetchOptions{
			SchemaRoot: immutable.Some(targetID),
		})
		if len(cols) == 0 || err != nil {
			return CollectionDefinition{}, false, err
		}
		def := cols[0].Definition()
		def.Description = CollectionDescription{}

		return def, true, nil

	default:
		// no-op
	}

	return CollectionDefinition{}, false, nil
}
