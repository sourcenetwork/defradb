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

	"github.com/ipfs/go-cid"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

// definitionState holds collection and schema descriptions in easily accessible
// sets.
//
// It is read only and will not and should not be mutated.
type definitionState struct {
	collections []client.CollectionVersion

	collectionsByID         map[string]client.CollectionVersion
	activeCollectionsByName map[string]client.CollectionVersion

	definitionCache client.DefinitionCache
}

// newDefinitionState creates a new definitionState object given the provided
// definitions.
func newDefinitionState(
	definitions []client.CollectionDefinition,
) *definitionState {
	collectionsByID := map[string]client.CollectionVersion{}
	collections := []client.CollectionVersion{}
	activeCollectionsByName := map[string]client.CollectionVersion{}

	for _, def := range definitions {
		collections = append(collections, def.Version)

		if def.Version.IsActive {
			activeCollectionsByName[def.Version.Name] = def.Version
		}

		if def.Version.VersionID != "" {
			collectionsByID[def.Version.VersionID] = def.Version
		}
	}

	return &definitionState{
		collections:             collections,
		collectionsByID:         collectionsByID,
		activeCollectionsByName: activeCollectionsByName,
		definitionCache:         client.NewDefinitionCache(definitions),
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
	validatePolicyNotModified,
	validateIDNotEmpty,
	validateIDUnique,
	validateSingleVersionActive,
	validateCollectionIDNotMutated,
	validateFieldNotMutated,
	validateFieldNotMoved,
	validateCollectionNameNotMutated,
}

var collectionUpdateValidators = append(
	append(
		append(
			[]definitionValidator{},
			updateOnlyValidators...,
		),
		validateCollectionNotAdded,
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
	validateRelationNameSet,
	validateCollectionDefinitionPolicyDesc,
	validateCollectionNameNotEmpty,
	validateRelationalFieldIDType,
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
	validateVersionID,
	validateCollectionID,
	validateCollectionSourceFromSameCollection,
}

var createValidators = append(
	append([]definitionValidator{}, createOnlyValidators...),
	globalValidators...,
)

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
	for _, col := range newState.collections {
		for _, field := range col.Fields {
			if !field.Kind.IsObject() {
				continue
			}

			_, ok := client.GetDefinition(newState.definitionCache, client.CollectionDefinition{Version: col}, field.Kind)
			if !ok {
				errs = append(errs, NewErrFieldKindNotFound(field.Name, field.Kind.String()))
			}
		}
	}

	return errors.Join(errs...)
}

func validateRelationNameSet(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, newCollection := range newState.collections {
		for _, field := range newCollection.Fields {
			if !field.Kind.IsObject() {
				continue
			}

			if !field.RelationName.HasValue() {
				errs = append(errs, NewErrRelationNameEmpty(field.Name))
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
		if len(newCollection.QuerySources()) > 0 {
			// Views do not require both sides of the relation to be defined.
			continue
		}

		definition := client.CollectionDefinition{
			Version: newCollection,
		}

		for _, field := range newCollection.Fields {
			if !field.Kind.IsObject() {
				continue
			}

			if !field.RelationName.HasValue() {
				continue
			}

			if field.IsPrimary {
				continue
			}

			otherDef, ok := client.GetDefinition(newState.definitionCache, definition, field.Kind)
			if !ok {
				continue
			}

			if otherDef.Version.IsEmbeddedOnly {
				// Embedded objects do not require both sides of the relation to be defined.
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

			if !otherField.IsPrimary {
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
		definition := client.CollectionDefinition{
			Version: newCollection,
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
			if otherField.IsPrimary {
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
		if oldCol.IsPlaceholder {
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

		if oldCol.IsPlaceholder {
			continue
		}

		if len(oldCol.Indexes) != len(newCol.Indexes) {
			// DeepEqual distinguishes between an empty set and a nil set, and the value is
			// inconsistent for this property, so we have to check the length and elements
			// manually instead of using DeepEqual.
			errs = append(errs, NewErrCollectionIndexesCannotBeMutated(newCol.VersionID))
		}

		for i := range oldCol.Indexes {
			// DeepEqual is temporary, as this validation is temporary
			if !reflect.DeepEqual(oldCol.Indexes[i], newCol.Indexes[i]) {
				errs = append(errs, NewErrCollectionIndexesCannotBeMutated(newCol.VersionID))
			}
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

func validateCollectionNotAdded(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, newCol := range newState.collections {
		if newCol.IsPlaceholder {
			continue
		}

		existed := false
		for _, oldCol := range oldState.collections {
			if oldCol.CollectionID == newCol.CollectionID {
				existed = true
				break
			}
		}

		if !existed {
			errs = append(errs, NewErrAddCollectionWithPatch(newCol.Name))
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

		if oldCol.IsPlaceholder {
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

	return errors.Join(errs...)
}

func validateCollectionNotRemoved(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, oldCol := range oldState.collections {
		if oldCol.IsPlaceholder {
			continue
		}

		if _, ok := newState.collectionsByID[oldCol.VersionID]; !ok {
			errs = append(errs, NewErrCollectionsCannotBeDeleted(oldCol.VersionID))
		}
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

func validateTypeAndKindCompatible(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, col := range newState.collections {
		for _, newField := range col.Fields {
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
	for _, colDef := range newState.activeCollectionsByName {
		for _, embedding := range colDef.VectorEmbeddings {
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
	for _, colDef := range newState.activeCollectionsByName {
		for _, embedding := range colDef.VectorEmbeddings {
			if len(embedding.Fields) == 0 {
				errs = append(errs, client.ErrEmptyFieldsForEmbedding)
			}
			for _, fieldName := range embedding.Fields {
				// Check that no fields used for embedding generation refers to self of another embedding field.
				for _, embedding := range colDef.VectorEmbeddings {
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
	for _, colDef := range newState.activeCollectionsByName {
		for _, embedding := range colDef.VectorEmbeddings {
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
	for _, col := range newState.collections {
		for _, newField := range col.Fields {
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
	for _, oldCol := range oldState.collections {
		oldFieldIndexesByName := map[string]int{}
		for i, field := range oldCol.Fields {
			oldFieldIndexesByName[field.Name] = i
		}

		newCol := newState.activeCollectionsByName[oldCol.Name]

		for newIndex, newField := range newCol.Fields {
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
	for _, oldCol := range oldState.activeCollectionsByName {
		oldFieldsByID := map[string]client.CollectionFieldDescription{}
		for _, field := range oldCol.Fields {
			oldFieldsByID[field.FieldID] = field
		}

		newCol := newState.activeCollectionsByName[oldCol.Name]

		for _, newField := range newCol.Fields {
			if newField.FieldID == "" {
				continue
			}

			oldField, exists := oldFieldsByID[newField.FieldID]

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
	for _, col := range newState.collections {
		fieldNames := map[string]struct{}{}

		for _, field := range col.Fields {
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
	for _, col := range newState.collections {
		for _, field := range col.Fields {
			if _, ok := field.Kind.(*client.SelfKind); ok {
				continue
			}

			otherDef, ok := client.GetDefinition(
				newState.definitionCache,
				client.CollectionDefinition{Version: col},
				field.Kind,
			)
			if !ok {
				continue
			}

			if otherDef.Version.CollectionID == col.CollectionID {
				errs = append(errs, NewErrSelfReferenceWithoutSelf(field.Name))
			}
		}
	}

	for _, col := range newState.collections {
		for _, field := range col.Fields {
			if _, ok := field.Kind.(*client.SelfKind); ok {
				continue
			}

			activeCol := newState.activeCollectionsByName[col.Name]
			otherDef, ok := client.GetDefinition(
				newState.definitionCache,
				client.CollectionDefinition{Version: activeCol},
				field.Kind,
			)
			if !ok {
				continue
			}

			if otherDef.Version.CollectionID == newState.collectionsByID[col.VersionID].VersionID {
				errs = append(errs, NewErrSelfReferenceWithoutSelf(field.Name))
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
	for _, col := range newState.collections {
		fieldsByName := map[string]client.CollectionFieldDescription{}

		for _, field := range col.Fields {
			fieldsByName[field.Name] = field
		}

		for _, field := range col.Fields {
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

func validateCollectionNameNotEmpty(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, col := range newState.collections {
		if col.CollectionID == client.OrphanCollectionID || col.IsPlaceholder {
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

func validateCollectionSourceFromSameCollection(
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

		for _, source := range col.CollectionSources() {
			for _, otherCol := range newState.collections {
				if otherCol.VersionID == source.SourceCollectionID &&
					otherCol.CollectionID != col.CollectionID {
					errs = append(errs, NewErrCollectionSourceWrongCollection(col.CollectionID, otherCol.CollectionID))
				}
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
	for name, col := range newState.activeCollectionsByName {
		// default values are set when a doc is first created
		_, err := client.NewDocFromMap(map[string]any{}, client.CollectionDefinition{Version: col})
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

func validateVersionID(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, col := range newState.collections {
		txn := datastore.CtxMustGetTxn(ctx)

		key, err := cid.Parse(col.VersionID)
		if err != nil {
			errs = append(errs, NewErrInvalidCID("VersionID", col.VersionID, err))
			continue
		}

		if col.IsPlaceholder {
			continue
		}

		exists, err := txn.Blockstore().Has(ctx, key)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		if !exists {
			errs = append(errs, NewErrUnknownCID("VersionID", col.VersionID))
		}
	}

	return errors.Join(errs...)
}

func validateCollectionID(
	ctx context.Context,
	db *DB,
	newState *definitionState,
	oldState *definitionState,
) error {
	var errs []error
	for _, col := range newState.collections {
		txn := datastore.CtxMustGetTxn(ctx)

		if col.IsPlaceholder {
			continue
		}

		key, err := cid.Parse(col.CollectionID)
		if err != nil {
			errs = append(errs, NewErrInvalidCID("CollectionID", col.CollectionID, err))
			continue
		}

		exists, err := txn.Blockstore().Has(ctx, key)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		if !exists {
			errs = append(errs, NewErrUnknownCID("CollectionID", col.CollectionID))
		}
	}

	return errors.Join(errs...)
}
