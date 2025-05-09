// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"
	"reflect"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/errors"
)

// definitionState holds collection and schema descriptions in easily accessible
// sets.
//
// It is read only and will not and should not be mutated.
type definitionState struct {
	collections     []client.CollectionVersion
	collectionsByID map[string]client.CollectionVersion

	schemaByID   map[string]client.SchemaDescription
	schemaByName map[string]client.SchemaDescription

	definitionsByName map[string]client.CollectionDefinition
	definitionCache   client.DefinitionCache
}

// newDefinitionState creates a new definitionState object given the provided
// definitions.
func newDefinitionState(
	definitions []client.CollectionDefinition,
) *definitionState {
	collectionsByID := map[string]client.CollectionVersion{}
	schemasByID := map[string]client.SchemaDescription{}
	definitionsByName := map[string]client.CollectionDefinition{}
	collections := []client.CollectionVersion{}
	schemaByName := map[string]client.SchemaDescription{}

	for _, def := range definitions {
		definitionsByName[def.GetName()] = def
		schemasByID[def.Schema.VersionID] = def.Schema
		schemaByName[def.Schema.Name] = def.Schema

		if def.Version.VersionID != "" {
			collectionsByID[def.Version.VersionID] = def.Version
			collections = append(collections, def.Version)
		}
	}

	return &definitionState{
		collections:       collections,
		collectionsByID:   collectionsByID,
		schemaByID:        schemasByID,
		schemaByName:      schemaByName,
		definitionsByName: definitionsByName,
		definitionCache:   client.NewDefinitionCache(definitions),
	}
}

// definitionValidator aliases the signature that all schema and collection
// validation functions should follow.
type definitionValidator = func(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error

// createOnlyValidators are executed on the creation of new descriptions only
// they will not be executed for updates to existing records.
var createOnlyValidators = []definitionValidator{}

// updateOnlyValidators are executed on the update of existing descriptions only
// they will not be executed for new records.
var updateOnlyValidators = []definitionValidator{
	validateSourcesNotRedefined,
	validateIndexesNotModified,
	validateFieldsNotModified,
	validatePolicyNotModified,
	validateIDNotEmpty,
	validateIDUnique,
	validateSingleVersionActive,
	validateCollectionIDNotMutated,
	validateSchemaNotAdded,
	validateSchemaFieldNotDeleted,
	validateFieldNotMutated,
	validateFieldNotMoved,
	validateCollectionNameNotMutated,
}

var schemaUpdateValidators = append(
	append(
		[]definitionValidator{},
		updateOnlyValidators...,
	),
	globalValidators...,
)

var collectionUpdateValidators = append(
	append(
		append(
			[]definitionValidator{},
			updateOnlyValidators...,
		),
		validateIDExists,
		validateSchemaVersionIDNotMutated,
		validateCollectionNotRemoved,
		validateCollectionIsBranchableNotMutated,
	),
	globalValidators...,
)

// globalValidators are run on create and update of records.
var globalValidators = []definitionValidator{
	validateCollectionNameUnique,
	validateRelationPointsToValidKind,
	validateSecondaryFieldsPairUp,
	validateSingleSidePrimary,
	validateCollectionDefinitionPolicyDesc,
	validateSchemaNameNotEmpty,
	validateCollectionNameNotEmpty,
	validateRelationalFieldIDType,
	validateSecondaryNotOnSchema,
	validateTypeSupported,
	validateTypeAndKindCompatible,
	validateFieldNotDuplicated,
	validateSelfReferences,
	validateCollectionMaterialized,
	validateMaterializedHasNoPolicy,
	validateCollectionFieldDefaultValue,
	validateEmbeddingAndKindCompatible,
	validateEmbeddingFieldsForGeneration,
	validateEmbeddingProviderAndModel,
}

var createValidators = append(
	append([]definitionValidator{}, createOnlyValidators...),
	globalValidators...,
)

func (db *DB) validateSchemaUpdate(
	ctx context.Context,
	oldDefinitions []client.CollectionDefinition,
	newDefinitions []client.CollectionDefinition,
) error {
	var errs []error
	newState := newDefinitionState(newDefinitions)
	oldState := newDefinitionState(oldDefinitions)

	for _, validator := range schemaUpdateValidators {
		if err := validator(ctx, db, newState, oldState); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (db *DB) validateCollectionChanges(
	ctx context.Context,
	oldDefinitions []client.CollectionDefinition,
	newDefinitions []client.CollectionDefinition,
) error {
	newState := newDefinitionState(newDefinitions)
	oldState := newDefinitionState(oldDefinitions)
	var errs []error
	for _, validator := range collectionUpdateValidators {
		err := validator(ctx, db, newState, oldState)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (db *DB) validateNewCollection(
	ctx context.Context,
	newDefinitions []client.CollectionDefinition,
	oldDefinitions []client.CollectionDefinition,
) error {
	newState := newDefinitionState(newDefinitions)
	oldState := newDefinitionState(oldDefinitions)
	var errs []error
	for _, validator := range createValidators {
		err := validator(ctx, db, newState, oldState)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func validateRelationPointsToValidKind(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, newCollection := range newState.collections {
		for _, field := range newCollection.Fields {
			if !field.Kind.HasValue() {
				continue
			}

			if !field.Kind.Value().IsObject() {
				continue
			}

			definition := newState.definitionsByName[newCollection.Name]
			_, ok := client.GetDefinition(newState.definitionCache, definition, field.Kind.Value())
			if !ok {
				errs = append(errs, NewErrFieldKindNotFound(field.Name, field.Kind.Value().String()))
			}
		}
	}

	for _, schema := range newState.schemaByName {
		for _, field := range schema.Fields {
			if !field.Kind.IsObject() {
				continue
			}

			_, ok := client.GetDefinition(newState.definitionCache, client.CollectionDefinition{Schema: schema}, field.Kind)
			if !ok {
				errs = append(errs, NewErrFieldKindNotFound(field.Name, field.Kind.String()))
			}
		}
	}

	return errors.Join(errs...)
}

func validateSecondaryFieldsPairUp(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, newCollection := range newState.collections {
		schema, ok := newState.schemaByID[newCollection.VersionID]
		if !ok {
			continue
		}

		definition := client.CollectionDefinition{
			Version: newCollection,
			Schema:  schema,
		}

		for _, field := range newCollection.Fields {
			if !field.Kind.HasValue() {
				continue
			}

			if !field.Kind.Value().IsObject() {
				continue
			}

			if !field.RelationName.HasValue() {
				continue
			}

			_, hasSchemaField := schema.GetFieldByName(field.Name)
			if hasSchemaField {
				continue
			}

			otherDef, ok := client.GetDefinition(newState.definitionCache, definition, field.Kind.Value())
			if !ok {
				continue
			}

			if otherDef.Version.IsEmbeddedOnly {
				// Views/embedded objects do not require both sides of the relation to be defined.
				continue
			}

			otherField, ok := otherDef.Version.GetFieldByRelation(
				field.RelationName.Value(),
				definition.GetName(),
				field.Name,
			)
			if !ok {
				errs = append(errs, NewErrRelationMissingField(otherDef.GetName(), field.RelationName.Value()))
			}

			_, ok = otherDef.Schema.GetFieldByName(otherField.Name)
			if !ok {
				// This secondary is paired with another secondary, which is invalid
				errs = append(errs, NewErrRelationMissingField(otherDef.GetName(), field.RelationName.Value()))
			}
		}
	}

	return errors.Join(errs...)
}

func validateSingleSidePrimary(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, newCollection := range newState.collections {
		schema, ok := newState.schemaByID[newCollection.VersionID]
		if !ok {
			continue
		}
		definition := client.CollectionDefinition{
			Version: newCollection,
			Schema:  schema,
		}
		for _, field := range definition.GetFields() {
			if field.Kind == nil {
				continue
			}
			if !field.Kind.IsObject() {
				continue
			}
			if field.RelationName == "" {
				continue
			}
			if !field.IsPrimaryRelation {
				// This is a secondary field and thus passes this rule
				continue
			}
			otherDef, ok := client.GetDefinition(newState.definitionCache, definition, field.Kind)
			if !ok {
				continue
			}
			otherField, ok := otherDef.Version.GetFieldByRelation(
				field.RelationName,
				definition.GetName(),
				field.Name,
			)
			if !ok {
				// This must be a one-sided relation, in which case it passes this rule
				continue
			}
			_, ok = otherDef.Schema.GetFieldByName(otherField.Name)
			if ok {
				// This primary is paired with another primary, which is invalid
				errs = append(errs, ErrMultipleRelationPrimaries)
			}
		}
	}

	return errors.Join(errs...)
}

func validateCollectionNameUnique(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	names := map[string]struct{}{}
	for _, col := range newState.collections {
		if !col.IsActive || col.Name == "" {
			continue
		}

		if _, ok := names[col.Name]; ok {
			errs = append(errs, NewErrCollectionAlreadyExists(col.Name))
		}
		names[col.Name] = struct{}{}
	}

	return errors.Join(errs...)
}

func validateSingleVersionActive(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	colsWithActiveCol := map[string]struct{}{}
	for _, def := range newState.collections {
		if !def.IsActive {
			continue
		}

		if _, isDuplicate := colsWithActiveCol[def.CollectionID]; isDuplicate {
			errs = append(
				errs,
				NewErrMultipleActiveCollectionVersions(
					def.Name,
					def.CollectionID,
				),
			)
		}
		colsWithActiveCol[def.CollectionID] = struct{}{}
	}

	return errors.Join(errs...)
}

// validateSourcesNotRedefined specifies the limitations on how the collection sources
// can be mutated.
//
// Currently new sources cannot be added, existing cannot be removed, and CollectionSources
// cannot be redirected to other collections.
func validateSourcesNotRedefined(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, newCol := range newState.collections {
		oldCol, ok := oldState.collectionsByID[newCol.VersionID]
		if !ok {
			continue
		}

		newColSources := newCol.CollectionSources()
		oldColSources := oldCol.CollectionSources()

		if len(newColSources) != len(oldColSources) {
			errs = append(errs, NewErrCollectionSourcesCannotBeAddedRemoved(newCol.VersionID))
		}

		for i := range newColSources {
			if i >= len(oldColSources) {
				continue // Avoid out-of-bounds panic
			}
			if newColSources[i].SourceCollectionID != oldColSources[i].SourceCollectionID {
				errs = append(errs, NewErrCollectionSourceIDMutated(
					newCol.VersionID,
					newColSources[i].SourceCollectionID,
					oldColSources[i].SourceCollectionID,
				))
			}
		}

		newQuerySources := newCol.QuerySources()
		oldQuerySources := oldCol.QuerySources()

		if len(newQuerySources) != len(oldQuerySources) {
			errs = append(errs, NewErrCollectionSourcesCannotBeAddedRemoved(newCol.VersionID))
		}
	}

	return errors.Join(errs...)
}

func validateIndexesNotModified(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, newCol := range newState.collections {
		oldCol, ok := oldState.collectionsByID[newCol.VersionID]
		if !ok {
			continue
		}

		// DeepEqual is temporary, as this validation is temporary
		if !reflect.DeepEqual(oldCol.Indexes, newCol.Indexes) {
			errs = append(errs, NewErrCollectionIndexesCannotBeMutated(newCol.VersionID))
		}
	}

	return errors.Join(errs...)
}

func validateFieldsNotModified(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, newCol := range newState.collections {
		oldCol, ok := oldState.collectionsByID[newCol.VersionID]
		if !ok {
			continue
		}

		// DeepEqual is temporary, as this validation is temporary
		if !reflect.DeepEqual(oldCol.Fields, newCol.Fields) {
			errs = append(errs, NewErrCollectionFieldsCannotBeMutated(newCol.VersionID))
		}
	}

	return errors.Join(errs...)
}

func validatePolicyNotModified(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, newCol := range newState.collections {
		oldCol, ok := oldState.collectionsByID[newCol.VersionID]
		if !ok {
			continue
		}

		// DeepEqual is temporary, as this validation is temporary
		if !reflect.DeepEqual(oldCol.Policy, newCol.Policy) {
			errs = append(errs, NewErrCollectionPolicyCannotBeMutated(newCol.VersionID))
		}
	}

	return errors.Join(errs...)
}

func validateIDNotEmpty(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, newCol := range newState.collections {
		if newCol.VersionID == "" {
			errs = append(errs, ErrCollectionIDCannotBeEmpty)
		}
	}

	return errors.Join(errs...)
}

func validateIDUnique(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	colIds := map[string]struct{}{}
	for _, newCol := range newState.collections {
		if _, ok := colIds[newCol.VersionID]; ok {
			errs = append(errs, NewErrCollectionIDAlreadyExists(newCol.VersionID))
		}
		colIds[newCol.VersionID] = struct{}{}
	}

	return errors.Join(errs...)
}

func validateIDExists(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, newCol := range newState.collections {
		if _, ok := oldState.collectionsByID[newCol.VersionID]; !ok {
			errs = append(errs, NewErrAddCollectionIDWithPatch(newCol.VersionID))
		}
	}

	return errors.Join(errs...)
}

func validateCollectionIDNotMutated(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, newCol := range newState.collections {
		oldCol, ok := oldState.collectionsByID[newCol.VersionID]
		if !ok {
			continue
		}

		if newCol.CollectionID != oldCol.CollectionID {
			errs = append(errs, NewErrCollectionIDCannotBeMutated(newCol.VersionID))
		}
	}

	return errors.Join(errs...)
}

func validateSchemaVersionIDNotMutated(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, newCol := range newState.collections {
		oldCol, ok := oldState.collectionsByID[newCol.VersionID]
		if !ok {
			continue
		}

		if newCol.VersionID != oldCol.VersionID {
			errs = append(errs, NewErrCollectionSchemaVersionIDCannotBeMutated(newCol.VersionID))
		}
	}

	for _, newSchema := range newState.schemaByName {
		oldSchema := oldState.schemaByName[newSchema.Name]
		if newSchema.VersionID != "" && newSchema.VersionID != oldSchema.VersionID {
			// If users specify this it will be overwritten, an error is preferred to quietly ignoring it.
			errs = append(errs, ErrCannotSetVersionID)
		}
	}

	return errors.Join(errs...)
}

func validateCollectionNotRemoved(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
oldLoop:
	for _, oldCol := range oldState.collections {
		for _, newCol := range newState.collectionsByID {
			// It is not enough to just match by the map index, in case the index does not pair
			// up with the ID (this can happen if a user moves the collection within the map)
			if newCol.VersionID == oldCol.VersionID {
				continue oldLoop
			}
		}

		errs = append(errs, NewErrCollectionsCannotBeDeleted(oldCol.VersionID))
	}

	return errors.Join(errs...)
}

// validateCollectionDefinitionPolicyDesc validates that the policy definition is valid, beyond syntax.
//
// Ensures that the information within the policy definition makes sense,
// this function might also make relevant remote calls using the acp system.
func validateCollectionDefinitionPolicyDesc(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, newCol := range newState.collections {
		if !newCol.Policy.HasValue() {
			// No policy validation needed, whether acp exists or not doesn't matter.
			continue
		}

		// If there is a policy specified, but the database does not have
		// acp enabled/available return an error, database must have an acp available
		// to enable access control (inorder to adhere to the policy specified).
		if !db.documentACP.HasValue() {
			errs = append(errs, ErrCanNotHavePolicyWithoutACP)
		}

		// If we have the policy specified on the collection, and document acp is
		// available/enabled, then using the document acp system we need to ensure
		// the policy id specified actually exists as a policy, and the resource name
		// exists on that policy and that the resource is a valid document interface resource.
		err := db.documentACP.Value().ValidateResourceInterface(
			ctx,
			newCol.Policy.Value().ID,
			newCol.Policy.Value().ResourceName,
		)

		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func validateSchemaFieldNotDeleted(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, newSchema := range newState.schemaByName {
		oldSchema := oldState.schemaByName[newSchema.Name]

		for _, oldField := range oldSchema.Fields {
			stillExists := false
			for _, newField := range newSchema.Fields {
				if newField.Name == oldField.Name {
					stillExists = true
					break
				}
			}

			if !stillExists {
				errs = append(errs, NewErrCannotDeleteField(oldField.Name))
			}
		}
	}

	return errors.Join(errs...)
}

func validateTypeAndKindCompatible(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, newSchema := range newState.schemaByName {
		for _, newField := range newSchema.Fields {
			if !newField.Typ.IsCompatibleWith(newField.Kind) {
				errs = append(errs, client.NewErrCRDTKindMismatch(newField.Typ.String(), newField.Kind.String()))
			}
		}
	}

	return errors.Join(errs...)
}

func validateEmbeddingAndKindCompatible(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, colDef := range newState.definitionsByName {
		for _, embedding := range colDef.Version.VectorEmbeddings {
			if embedding.FieldName == "" {
				errs = append(errs, client.ErrEmptyFieldNameForEmbedding)
				continue
			}

			field, fieldExists := colDef.GetFieldByName(embedding.FieldName)
			if !fieldExists {
				errs = append(errs, client.NewErrVectorFieldDoesNotExist(embedding.FieldName))
				continue
			}

			if field.Kind == nil {
				errs = append(errs, client.NewErrVectorFieldDoesNotExist(embedding.FieldName))
				continue
			}

			if !client.IsVectorEmbeddingCompatible(field.Kind) {
				errs = append(errs, client.NewErrInvalidTypeForEmbedding(field.Kind))
			}
		}
	}
	return errors.Join(errs...)
}

func validateEmbeddingFieldsForGeneration(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, colDef := range newState.definitionsByName {
		for _, embedding := range colDef.Version.VectorEmbeddings {
			if len(embedding.Fields) == 0 {
				errs = append(errs, client.ErrEmptyFieldsForEmbedding)
			}
			for _, fieldName := range embedding.Fields {
				// Check that no fields used for embedding generation refers to self of another embedding field.
				for _, embedding := range colDef.Version.VectorEmbeddings {
					if embedding.FieldName == fieldName {
						errs = append(errs, client.NewErrEmbeddingFieldEmbedding(fieldName))
					}
				}
				// Check that the field exists.
				field, fieldExists := colDef.GetFieldByName(fieldName)
				if !fieldExists {
					errs = append(errs, client.NewErrFieldForEmbeddingGenerationDoesNotExist(fieldName))
				}

				if field.Kind == nil {
					continue
				}

				// Check that the field is of a supperted kind.
				if !client.IsSupportedVectorEmbeddingSourceKind(field.Kind) {
					errs = append(errs, client.NewErrInvalidTypeForEmbeddingGeneration(field.Kind))
				}
			}
		}
	}
	return errors.Join(errs...)
}

func validateEmbeddingProviderAndModel(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, colDef := range newState.definitionsByName {
		for _, embedding := range colDef.Version.VectorEmbeddings {
			if embedding.Provider == "" {
				errs = append(errs, client.ErrEmptyProviderForEmbedding)
			}
			if _, supported := supportedEmbeddingProviders[embedding.Provider]; !supported {
				errs = append(errs, client.NewErrUnknownEmbeddingProvider(embedding.Provider))
			}
			if embedding.Model == "" {
				errs = append(errs, client.ErrEmptyModelForEmbedding)
			}
		}
	}
	return errors.Join(errs...)
}

func validateTypeSupported(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, newSchema := range newState.schemaByName {
		for _, newField := range newSchema.Fields {
			if !newField.Typ.IsSupportedFieldCType() {
				errs = append(errs, client.NewErrInvalidCRDTType(newField.Name, newField.Typ.String()))
			}
		}
	}
	return errors.Join(errs...)
}

func validateFieldNotMoved(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, oldSchema := range oldState.schemaByName {
		oldFieldIndexesByName := map[string]int{}
		for i, field := range oldSchema.Fields {
			oldFieldIndexesByName[field.Name] = i
		}

		newSchema := newState.schemaByName[oldSchema.Name]

		for newIndex, newField := range newSchema.Fields {
			if existingIndex, exists := oldFieldIndexesByName[newField.Name]; exists && newIndex != existingIndex {
				errs = append(errs, NewErrCannotMoveField(newField.Name, newIndex, existingIndex))
			}
		}
	}

	return errors.Join(errs...)
}

func validateFieldNotMutated(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, oldSchema := range oldState.schemaByName {
		oldFieldsByName := map[string]client.SchemaFieldDescription{}
		for _, field := range oldSchema.Fields {
			oldFieldsByName[field.Name] = field
		}

		newSchema := newState.schemaByName[oldSchema.Name]

		for _, newField := range newSchema.Fields {
			oldField, exists := oldFieldsByName[newField.Name]

			// DeepEqual is temporary, as this validation is temporary
			if exists && !reflect.DeepEqual(oldField, newField) {
				errs = append(errs, NewErrCannotMutateField(newField.Name))
			}
		}
	}

	return errors.Join(errs...)
}

func validateFieldNotDuplicated(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, schema := range newState.schemaByName {
		fieldNames := map[string]struct{}{}

		for _, field := range schema.Fields {
			if _, isDuplicate := fieldNames[field.Name]; isDuplicate {
				errs = append(errs, NewErrDuplicateField(field.Name))
			}
			fieldNames[field.Name] = struct{}{}
		}
	}

	return errors.Join(errs...)
}

func validateSelfReferences(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, schema := range newState.schemaByName {
		for _, field := range schema.Fields {
			if _, ok := field.Kind.(*client.SelfKind); ok {
				continue
			}

			otherDef, ok := client.GetDefinition(
				newState.definitionCache,
				client.CollectionDefinition{Schema: schema},
				field.Kind,
			)
			if !ok {
				continue
			}

			if otherDef.Schema.Root == schema.Root {
				errs = append(errs, NewErrSelfReferenceWithoutSelf(field.Name))
			}
		}
	}

	for _, col := range newState.collections {
		for _, field := range col.Fields {
			if !field.Kind.HasValue() {
				continue
			}

			if _, ok := field.Kind.Value().(*client.SelfKind); ok {
				continue
			}

			definition := newState.definitionsByName[col.Name]
			otherDef, ok := client.GetDefinition(newState.definitionCache, definition, field.Kind.Value())
			if !ok {
				continue
			}

			if otherDef.Schema.Root == newState.schemaByID[col.VersionID].Root {
				errs = append(errs, NewErrSelfReferenceWithoutSelf(field.Name))
			}
		}
	}

	return errors.Join(errs...)
}

func validateSecondaryNotOnSchema(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, newSchema := range newState.schemaByName {
		for _, newField := range newSchema.Fields {
			if newField.Kind.IsObject() && newField.Kind.IsArray() {
				errs = append(errs, NewErrSecondaryFieldOnSchema(newField.Name))
			}
		}
	}

	return errors.Join(errs...)
}

func validateRelationalFieldIDType(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, schema := range newState.schemaByName {
		fieldsByName := map[string]client.SchemaFieldDescription{}

		for _, field := range schema.Fields {
			fieldsByName[field.Name] = field
		}

		for _, field := range schema.Fields {
			if field.Kind.IsObject() && !field.Kind.IsArray() {
				idFieldName := field.Name + request.RelatedObjectID
				idField, idFieldFound := fieldsByName[idFieldName]
				if idFieldFound {
					if idField.Kind != client.FieldKind_DocID {
						errs = append(errs, NewErrRelationalFieldIDInvalidType(idField.Name, client.FieldKind_DocID, idField.Kind))
					}
				}
			}
		}
	}

	return errors.Join(errs...)
}

func validateSchemaNotAdded(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, newSchema := range newState.schemaByName {
		if newSchema.Name == "" {
			// continue, and allow a more appropriate rule to return a nicer error
			// for the user
			continue
		}
		if _, exists := oldState.schemaByName[newSchema.Name]; !exists {
			errs = append(errs, NewErrAddSchemaWithPatch(newSchema.Name))
		}
	}

	return errors.Join(errs...)
}

func validateSchemaNameNotEmpty(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, schema := range newState.schemaByName {
		if schema.Name == "" {
			errs = append(errs, ErrSchemaNameEmpty)
		}
	}

	return errors.Join(errs...)
}

func validateCollectionNameNotEmpty(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, col := range newState.collections {
		if col.CollectionID == client.OrphanCollectionID {
			// CollectionVersions can exist before they are are linked to a Collection, as
			// users can register migrations for unknown version ids, in which case the name
			// will be empty.
			continue
		}

		if col.Name == "" {
			errs = append(errs, ErrCollectionNameEmpty)
		}
	}

	return errors.Join(errs...)
}

func validateCollectionNameNotMutated(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, col := range newState.collections {
		if col.Name == "" {
			continue
		}

		for _, oldCol := range oldState.collections {
			if oldCol.CollectionID == col.CollectionID &&
				oldCol.Name != col.Name {
				errs = append(errs, NewErrCollectionNameMutated(col.Name, oldCol.Name))
			}
		}
	}

	return errors.Join(errs...)
}

// validateCollectionMaterialized verifies that a non-view collection is materialized.
//
// Long term we wish to support this, however for now we block it off.
func validateCollectionMaterialized(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, col := range newState.collections {
		if len(col.QuerySources()) == 0 && !col.IsMaterialized {
			errs = append(errs, NewErrColNotMaterialized(col.Name))
		}
	}

	return errors.Join(errs...)
}

// validateMaterializedHasNoPolicy verifies that a materialized view has no ACP policy.
//
// Long term we wish to support this, however for now we block it off.
func validateMaterializedHasNoPolicy(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, col := range newState.collections {
		if col.IsMaterialized && len(col.QuerySources()) != 0 && col.Policy.HasValue() {
			errs = append(errs, NewErrMaterializedViewAndACPNotSupported(col.Name))
		}
	}

	return errors.Join(errs...)
}

func validateCollectionFieldDefaultValue(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for name, col := range newState.definitionsByName {
		// default values are set when a doc is first created
		_, err := client.NewDocFromMap(map[string]any{}, col)
		if err != nil {
			errs = append(errs, NewErrDefaultFieldValueInvalid(name, err))
		}
	}

	return errors.Join(errs...)
}

// validateCollectionIsBranchableNotMutated is a temporary restriction that prevents users from toggling
// whether or not a collection is branchable.
// https://github.com/sourcenetwork/defradb/issues/3219
func validateCollectionIsBranchableNotMutated(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, newCol := range newState.collections {
		oldCol := oldState.collectionsByID[newCol.VersionID]

		if newCol.IsBranchable != oldCol.IsBranchable {
			errs = append(errs, NewErrColMutatingIsBranchable(newCol.Name))
		}
	}

	return errors.Join(errs...)
}
