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
	"fmt"
	"strings"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client/request"
)

// CollectionDefinition contains the metadata defining what a Collection is.
//
// The definition types ([CollectionDefinition], [FieldDefinition]) are read-only types returned
// from various functions as a convienient means to access the computated convergence of schema
// and collection versions.
type CollectionDefinition struct {
	// Version returns the CollectionVersion of this Collection.
	Version CollectionVersion `json:"version"`
}

// GetFieldByName returns the field for the given field name. If such a field is found it
// will return it and true, if it is not found it will return false.
func (def CollectionDefinition) GetFieldByName(fieldName string) (FieldDefinition, bool) {
	collectionField, existsOnCollection := def.Version.GetFieldByName(fieldName)

	if existsOnCollection {
		// If the field exists only on the collection, it is a local only field, for example the
		// secondary side of a relation.
		return NewLocalFieldDefinition(
			collectionField,
		), true
	}

	return FieldDefinition{}, false
}

// GetFields returns the combined local and global field elements on this [CollectionDefinition]
// as a single set.
func (def CollectionDefinition) GetFields() []FieldDefinition {
	fields := []FieldDefinition{}

	for _, localField := range def.Version.Fields {
		// This must be a local only field, for example the secondary side of a relation.
		fields = append(
			fields,
			NewLocalFieldDefinition(localField),
		)
	}

	return fields
}

// GetName gets the name of this definition.
//
// If the collection version has a name (e.g. it is an active collection) it will return that,
// otherwise it will return the schema name.
func (def CollectionDefinition) GetName() string {
	return def.Version.Name
}

// FieldDefinition describes the combined local and global set of properties that constitutes
// a field on a collection.
//
// It draws it's information from the [CollectionFieldDescription] on the [CollectionVersion],
// and the [SchemaFieldDescription] on the [SchemaDescription].
//
// It is to [CollectionFieldDescription] and [SchemaFieldDescription] what [CollectionDefinition]
// is to [CollectionVersion] and [SchemaDescription].
//
// The definition types ([CollectionDefinition], [FieldDefinition]) are read-only types returned
// from various functions as a convienient means to access the computated convergence of schema
// and collection versions.
type FieldDefinition struct {
	// The immutable ID of this field.
	//
	// Only global fields persisted in the DAG will have a value - virtual fields such as secondary
	// relation fields will not have a FieldID.
	FieldID string

	// Name contains the name of this field.
	Name string

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

	// Size is a constraint that can be applied to fields that are arrays.
	//
	// Mutations on fields with a size constraint will fail if the size of the array
	// does not match the constraint.
	Size int
}

// NewLocalFieldDefinition returns a new [FieldDefinition] from the given local [CollectionFieldDescription].
func NewLocalFieldDefinition(local CollectionFieldDescription) FieldDefinition {
	return FieldDefinition{
		FieldID:           local.FieldID,
		Name:              local.Name,
		Kind:              local.Kind,
		Typ:               local.Typ,
		RelationName:      local.RelationName.Value(),
		DefaultValue:      local.DefaultValue,
		IsPrimaryRelation: local.IsPrimary,
		Size:              local.Size,
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
}

// NewDefinitionCache creates a new [DefinitionCache] populated with the given [CollectionDefinition]s.
func NewDefinitionCache(definitions []CollectionDefinition) DefinitionCache {
	definitionsBySchemaRoot := make(map[string]CollectionDefinition, len(definitions))

	for _, def := range definitions {
		definitionsBySchemaRoot[def.Version.CollectionID] = def
	}

	return DefinitionCache{
		Definitions:             definitions,
		DefinitionsBySchemaRoot: definitionsBySchemaRoot,
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

	case *CollectionKind:
		def, ok := cache.DefinitionsBySchemaRoot[typedKind.CollectionID]
		return def, ok

	case *SelfKind:
		if typedKind.RelativeID == "" {
			return host, true
		}

		for _, col := range cache.Definitions {
			if col.Version.CollectionID == host.Version.CollectionID {
				continue
			}

			if col.Version.CollectionSet.Value().CollectionSetID != host.Version.CollectionSet.Value().CollectionSetID {
				continue
			}

			if fmt.Sprint(col.Version.CollectionSet.Value().RelativeID) == typedKind.RelativeID {
				return col, true
			}
		}

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
	store TxnStore,
	host CollectionDefinition,
	kind FieldKind,
) (CollectionDefinition, bool, error) {
	switch typedKind := kind.(type) {
	case *NamedKind:
		col, err := store.GetCollectionByName(ctx, typedKind.Name)
		if err != nil {
			return CollectionDefinition{}, false, err
		}

		return col.Definition(), true, nil

	case *CollectionKind:
		cols, err := store.GetCollections(ctx, CollectionFetchOptions{
			CollectionID: immutable.Some(typedKind.CollectionID),
		})

		if len(cols) == 0 {
			return CollectionDefinition{}, false, ErrNotFound
		}

		if err != nil {
			return CollectionDefinition{}, false, err
		}

		return cols[0].Definition(), true, nil

	case *SelfKind:
		if typedKind.RelativeID == "" {
			return host, true, nil
		}

		cols, err := store.GetCollections(ctx, CollectionFetchOptions{
			CollectionSetID: immutable.Some(host.Version.CollectionSet.Value().CollectionSetID),
		})
		if err != nil {
			return CollectionDefinition{}, false, err
		}

		for _, col := range cols {
			if col.Version().CollectionID == host.Version.CollectionID {
				continue
			}

			if col.Version().CollectionSet.Value().CollectionSetID != host.Version.CollectionSet.Value().CollectionSetID {
				continue
			}

			if fmt.Sprint(col.Version().CollectionSet.Value().RelativeID) == typedKind.RelativeID {
				return col.Definition(), true, nil
			}
		}

	default:
		// no-op
	}

	return CollectionDefinition{}, false, nil
}
