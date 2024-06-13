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
)

// definitionState holds collection and schema descriptions in easily accessible
// sets.
//
// It is read only and will not and should not be mutated.
type definitionState struct {
	collections     []client.CollectionDescription
	collectionsByID map[uint32]client.CollectionDescription

	schemaByID   map[string]client.SchemaDescription
	schemaByName map[string]client.SchemaDescription

	definitionsByName map[string]client.CollectionDefinition
}

// newDefinitionState creates a new definitionState object given the provided
// descriptions.
func newDefinitionState(
	collections []client.CollectionDescription,
	schemasByID map[string]client.SchemaDescription,
) *definitionState {
	collectionsByID := map[uint32]client.CollectionDescription{}
	definitionsByName := map[string]client.CollectionDefinition{}
	schemaByName := map[string]client.SchemaDescription{}
	schemaVersionsAdded := map[string]struct{}{}

	for _, col := range collections {
		if len(col.Fields) == 0 {
			continue
		}

		schema := schemasByID[col.SchemaVersionID]
		definition := client.CollectionDefinition{
			Description: col,
			Schema:      schema,
		}

		definitionsByName[definition.GetName()] = definition
		schemaVersionsAdded[schema.VersionID] = struct{}{}
		collectionsByID[col.ID] = col
	}

	for _, schema := range schemasByID {
		schemaByName[schema.Name] = schema

		if _, ok := schemaVersionsAdded[schema.VersionID]; ok {
			continue
		}

		definitionsByName[schema.Name] = client.CollectionDefinition{
			Schema: schema,
		}
	}

	return &definitionState{
		collections:       collections,
		collectionsByID:   collectionsByID,
		schemaByID:        schemasByID,
		schemaByName:      schemaByName,
		definitionsByName: definitionsByName,
	}
}

// definitionValidator aliases the signature that all schema and collection
// validation functions should follow.
type definitionValidator = func(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error

// createOnlyValidators are executed on the creation of new descriptions only
// they will not be executed for updates to existing records.
var createOnlyValidators = []definitionValidator{}

// createOnlyValidators are executed on the update of existing descriptions only
// they will not be executed for new records.
var updateOnlyValidators = []definitionValidator{
	validateSourcesNotRedefined,
	validateIndexesNotModified,
	validateFieldsNotModified,
	validatePolicyNotModified,
	validateIDNotZero,
	validateIDUnique,
	validateIDExists,
	validateRootIDNotMutated,
	validateSchemaVersionIDNotMutated,
	validateCollectionNotRemoved,
	validateSingleVersionActive,
	validateSchemaFieldNotDeleted,
	validateFieldNotMutated,
	validateFieldNotMoved,
}

// globalValidators are run on create and update of records.
var globalValidators = []definitionValidator{
	validateCollectionNameUnique,
	validateRelationPointsToValidKind,
	validateSecondaryFieldsPairUp,
	validateSingleSidePrimary,
	validateCollectionDefinitionPolicyDesc,
	validateRelationalFieldIDType,
	validateSecondaryNotOnSchema,
	validateTypeSupported,
	validateTypeAndKindCompatible,
	validateFieldNotDuplicated,
}

var updateValidators = append(
	append([]definitionValidator{}, updateOnlyValidators...),
	globalValidators...,
)

var createValidators = append(
	append([]definitionValidator{}, createOnlyValidators...),
	globalValidators...,
)

func (db *db) validateCollectionChanges(
	ctx context.Context,
	oldCols []client.CollectionDescription,
	newColsByID map[uint32]client.CollectionDescription,
) error {
	newCols := make([]client.CollectionDescription, 0, len(newColsByID))
	for _, col := range newColsByID {
		newCols = append(newCols, col)
	}

	newState := newDefinitionState(newCols, map[string]client.SchemaDescription{})
	oldState := newDefinitionState(oldCols, map[string]client.SchemaDescription{})

	for _, validator := range updateValidators {
		err := validator(ctx, db, newState, oldState)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *db) validateNewCollection(
	ctx context.Context,
	newDefsByName map[string]client.CollectionDefinition,
) error {
	collections := []client.CollectionDescription{}
	schemasByID := map[string]client.SchemaDescription{}

	for _, def := range newDefsByName {
		if len(def.Description.Fields) != 0 {
			collections = append(collections, def.Description)
		}

		schemasByID[def.Schema.VersionID] = def.Schema
	}

	newState := newDefinitionState(collections, schemasByID)

	for _, validator := range createValidators {
		err := validator(ctx, db, newState, &definitionState{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *db) validateSchemaUpdate(
	ctx context.Context,
	newSchemaByName map[string]client.SchemaDescription,
	oldSchemaByName map[string]client.SchemaDescription,
) error {
	newSchemaByID := make(map[string]client.SchemaDescription, len(newSchemaByName))
	oldSchemaByID := make(map[string]client.SchemaDescription, len(oldSchemaByName))
	for _, schema := range newSchemaByName {
		newSchemaByID[schema.VersionID] = schema
	}
	for _, schema := range oldSchemaByName {
		oldSchemaByID[schema.VersionID] = schema
	}

	newState := newDefinitionState([]client.CollectionDescription{}, newSchemaByID)
	oldState := newDefinitionState([]client.CollectionDescription{}, oldSchemaByID)

	for _, validator := range updateValidators {
		err := validator(ctx, db, newState, oldState)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateRelationPointsToValidKind(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	for _, newCollection := range newState.collections {
		for _, field := range newCollection.Fields {
			if !field.Kind.HasValue() {
				continue
			}

			if !field.Kind.Value().IsObject() {
				continue
			}

			underlying := field.Kind.Value().Underlying()
			_, ok := newState.definitionsByName[underlying]
			if !ok {
				return NewErrFieldKindNotFound(field.Name, underlying)
			}
		}
	}

	for _, schema := range newState.schemaByName {
		for _, field := range schema.Fields {
			if !field.Kind.IsObject() {
				continue
			}

			underlying := field.Kind.Underlying()
			_, ok := newState.definitionsByName[underlying]
			if !ok {
				return NewErrFieldKindNotFound(field.Name, underlying)
			}
		}
	}

	return nil
}

func validateSecondaryFieldsPairUp(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	for _, newCollection := range newState.collections {
		schema, ok := newState.schemaByID[newCollection.SchemaVersionID]
		if !ok {
			continue
		}

		definition := client.CollectionDefinition{
			Description: newCollection,
			Schema:      schema,
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

			underlying := field.Kind.Value().Underlying()
			otherDef, ok := newState.definitionsByName[underlying]
			if !ok {
				continue
			}

			if len(otherDef.Description.Fields) == 0 {
				// Views/embedded objects do not require both sides of the relation to be defined.
				continue
			}

			otherField, ok := otherDef.Description.GetFieldByRelation(
				field.RelationName.Value(),
				definition.GetName(),
				field.Name,
			)
			if !ok {
				return NewErrRelationMissingField(underlying, field.RelationName.Value())
			}

			_, ok = otherDef.Schema.GetFieldByName(otherField.Name)
			if !ok {
				// This secondary is paired with another secondary, which is invalid
				return NewErrRelationMissingField(underlying, field.RelationName.Value())
			}
		}
	}

	return nil
}

func validateSingleSidePrimary(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	for _, newCollection := range newState.collections {
		schema, ok := newState.schemaByID[newCollection.SchemaVersionID]
		if !ok {
			continue
		}

		definition := client.CollectionDefinition{
			Description: newCollection,
			Schema:      schema,
		}

		for _, field := range definition.GetFields() {
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

			underlying := field.Kind.Underlying()
			otherDef, ok := newState.definitionsByName[underlying]
			if !ok {
				continue
			}

			otherField, ok := otherDef.Description.GetFieldByRelation(
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
				return ErrMultipleRelationPrimaries
			}
		}
	}

	return nil
}

func validateCollectionNameUnique(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	names := map[string]struct{}{}
	for _, col := range newState.collections {
		if !col.Name.HasValue() {
			continue
		}

		if _, ok := names[col.Name.Value()]; ok {
			return NewErrCollectionAlreadyExists(col.Name.Value())
		}
		names[col.Name.Value()] = struct{}{}
	}

	return nil
}

func validateSingleVersionActive(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	rootsWithActiveCol := map[uint32]struct{}{}
	for _, col := range newState.collections {
		if !col.Name.HasValue() {
			continue
		}

		if _, ok := rootsWithActiveCol[col.RootID]; ok {
			return NewErrMultipleActiveCollectionVersions(col.Name.Value(), col.RootID)
		}
		rootsWithActiveCol[col.RootID] = struct{}{}
	}

	return nil
}

// validateSourcesNotRedefined specifies the limitations on how the collection sources
// can be mutated.
//
// Currently new sources cannot be added, existing cannot be removed, and CollectionSources
// cannot be redirected to other collections.
func validateSourcesNotRedefined(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	for _, newCol := range newState.collections {
		oldCol, ok := oldState.collectionsByID[newCol.ID]
		if !ok {
			continue
		}

		newColSources := newCol.CollectionSources()
		oldColSources := oldCol.CollectionSources()

		if len(newColSources) != len(oldColSources) {
			return NewErrCollectionSourcesCannotBeAddedRemoved(newCol.ID)
		}

		for i := range newColSources {
			if newColSources[i].SourceCollectionID != oldColSources[i].SourceCollectionID {
				return NewErrCollectionSourceIDMutated(
					newCol.ID,
					newColSources[i].SourceCollectionID,
					oldColSources[i].SourceCollectionID,
				)
			}
		}

		newQuerySources := newCol.QuerySources()
		oldQuerySources := oldCol.QuerySources()

		if len(newQuerySources) != len(oldQuerySources) {
			return NewErrCollectionSourcesCannotBeAddedRemoved(newCol.ID)
		}
	}

	return nil
}

func validateIndexesNotModified(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	for _, newCol := range newState.collections {
		oldCol, ok := oldState.collectionsByID[newCol.ID]
		if !ok {
			continue
		}

		// DeepEqual is temporary, as this validation is temporary
		if !reflect.DeepEqual(oldCol.Indexes, newCol.Indexes) {
			return NewErrCollectionIndexesCannotBeMutated(newCol.ID)
		}
	}

	return nil
}

func validateFieldsNotModified(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	for _, newCol := range newState.collections {
		oldCol, ok := oldState.collectionsByID[newCol.ID]
		if !ok {
			continue
		}

		// DeepEqual is temporary, as this validation is temporary
		if !reflect.DeepEqual(oldCol.Fields, newCol.Fields) {
			return NewErrCollectionFieldsCannotBeMutated(newCol.ID)
		}
	}

	return nil
}

func validatePolicyNotModified(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	for _, newCol := range newState.collections {
		oldCol, ok := oldState.collectionsByID[newCol.ID]
		if !ok {
			continue
		}

		// DeepEqual is temporary, as this validation is temporary
		if !reflect.DeepEqual(oldCol.Policy, newCol.Policy) {
			return NewErrCollectionPolicyCannotBeMutated(newCol.ID)
		}
	}

	return nil
}

func validateIDNotZero(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	for _, newCol := range newState.collections {
		if newCol.ID == 0 {
			return ErrCollectionIDCannotBeZero
		}
	}

	return nil
}

func validateIDUnique(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	colIds := map[uint32]struct{}{}
	for _, newCol := range newState.collections {
		if _, ok := colIds[newCol.ID]; ok {
			return NewErrCollectionIDAlreadyExists(newCol.ID)
		}
		colIds[newCol.ID] = struct{}{}
	}

	return nil
}

func validateIDExists(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	for _, newCol := range newState.collections {
		if _, ok := oldState.collectionsByID[newCol.ID]; !ok {
			return NewErrAddCollectionIDWithPatch(newCol.ID)
		}
	}

	return nil
}

func validateRootIDNotMutated(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	for _, newCol := range newState.collections {
		oldCol, ok := oldState.collectionsByID[newCol.ID]
		if !ok {
			continue
		}

		if newCol.RootID != oldCol.RootID {
			return NewErrCollectionRootIDCannotBeMutated(newCol.ID)
		}
	}

	for _, newSchema := range newState.schemaByName {
		oldSchema := oldState.schemaByName[newSchema.Name]
		if newSchema.Root != oldSchema.Root {
			return NewErrSchemaRootDoesntMatch(
				newSchema.Name,
				oldSchema.Root,
				newSchema.Root,
			)
		}
	}

	return nil
}

func validateSchemaVersionIDNotMutated(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	for _, newCol := range newState.collections {
		oldCol, ok := oldState.collectionsByID[newCol.ID]
		if !ok {
			continue
		}

		if newCol.SchemaVersionID != oldCol.SchemaVersionID {
			return NewErrCollectionSchemaVersionIDCannotBeMutated(newCol.ID)
		}
	}

	for _, newSchema := range newState.schemaByName {
		oldSchema := oldState.schemaByName[newSchema.Name]
		if newSchema.VersionID != "" && newSchema.VersionID != oldSchema.VersionID {
			// If users specify this it will be overwritten, an error is preferred to quietly ignoring it.
			return ErrCannotSetVersionID
		}
	}

	return nil
}

func validateCollectionNotRemoved(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
oldLoop:
	for _, oldCol := range oldState.collections {
		for _, newCol := range newState.collectionsByID {
			// It is not enough to just match by the map index, in case the index does not pair
			// up with the ID (this can happen if a user moves the collection within the map)
			if newCol.ID == oldCol.ID {
				continue oldLoop
			}
		}

		return NewErrCollectionsCannotBeDeleted(oldCol.ID)
	}

	return nil
}

// validateCollectionDefinitionPolicyDesc validates that the policy definition is valid, beyond syntax.
//
// Ensures that the information within the policy definition makes sense,
// this function might also make relevant remote calls using the acp system.
func validateCollectionDefinitionPolicyDesc(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	for _, newCol := range newState.collections {
		if !newCol.Policy.HasValue() {
			// No policy validation needed, whether acp exists or not doesn't matter.
			continue
		}

		// If there is a policy specified, but the database does not have
		// acp enabled/available return an error, database must have an acp available
		// to enable access control (inorder to adhere to the policy specified).
		if !db.acp.HasValue() {
			return ErrCanNotHavePolicyWithoutACP
		}

		// If we have the policy specified on the collection, and acp is available/enabled,
		// then using the acp system we need to ensure the policy id specified
		// actually exists as a policy, and the resource name exists on that policy
		// and that the resource is a valid DPI.
		err := db.acp.Value().ValidateResourceExistsOnValidDPI(
			ctx,
			newCol.Policy.Value().ID,
			newCol.Policy.Value().ResourceName,
		)

		if err != nil {
			return err
		}
	}

	return nil
}

// validateUpdateSchema validates that the given schema description is a valid update.
//
// Will return true if the given description differs from the current persisted state of the
// schema. Will return an error if it fails validation.
func (db *db) validateUpdateSchema(
	existingDescriptionsByName map[string]client.SchemaDescription,
	proposedDescriptionsByName map[string]client.SchemaDescription,
	proposedDesc client.SchemaDescription,
) (bool, error) {
	if proposedDesc.Name == "" {
		return false, ErrSchemaNameEmpty
	}

	existingDesc, collectionExists := existingDescriptionsByName[proposedDesc.Name]
	if !collectionExists {
		return false, NewErrAddCollectionWithPatch(proposedDesc.Name)
	}

	hasChangedFields, err := validateUpdateSchemaFields(proposedDescriptionsByName, existingDesc, proposedDesc)
	if err != nil {
		return hasChangedFields, err
	}

	return hasChangedFields, err
}

func validateUpdateSchemaFields(
	descriptionsByName map[string]client.SchemaDescription,
	existingDesc client.SchemaDescription,
	proposedDesc client.SchemaDescription,
) (bool, error) {
	hasChanged := false
	existingFieldsByName := map[string]client.SchemaFieldDescription{}
	existingFieldIndexesByName := map[string]int{}
	for i, field := range existingDesc.Fields {
		existingFieldIndexesByName[field.Name] = i
		existingFieldsByName[field.Name] = field
	}

	newFieldNames := map[string]struct{}{}
	for _, proposedField := range proposedDesc.Fields {
		_, fieldAlreadyExists := existingFieldsByName[proposedField.Name]

		// If the field is new, then the collection has changed
		hasChanged = hasChanged || !fieldAlreadyExists

		newFieldNames[proposedField.Name] = struct{}{}
	}

	return hasChanged, nil
}

func validateSchemaFieldNotDeleted(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
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
				return NewErrCannotDeleteField(oldField.Name)
			}
		}
	}

	return nil
}

func validateTypeAndKindCompatible(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	for _, newSchema := range newState.schemaByName {
		for _, newField := range newSchema.Fields {
			if !newField.Typ.IsCompatibleWith(newField.Kind) {
				return client.NewErrCRDTKindMismatch(newField.Typ.String(), newField.Kind.String())
			}
		}
	}

	return nil
}

func validateTypeSupported(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	for _, newSchema := range newState.schemaByName {
		for _, newField := range newSchema.Fields {
			if !newField.Typ.IsSupportedFieldCType() {
				return client.NewErrInvalidCRDTType(newField.Name, newField.Typ.String())
			}
		}
	}

	return nil
}

func validateFieldNotMoved(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	for _, oldSchema := range oldState.schemaByName {
		oldFieldIndexesByName := map[string]int{}
		for i, field := range oldSchema.Fields {
			oldFieldIndexesByName[field.Name] = i
		}

		newSchema := newState.schemaByName[oldSchema.Name]

		for newIndex, newField := range newSchema.Fields {
			if existingIndex, exists := oldFieldIndexesByName[newField.Name]; exists && newIndex != existingIndex {
				return NewErrCannotMoveField(newField.Name, newIndex, existingIndex)
			}
		}
	}

	return nil
}

func validateFieldNotMutated(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	for _, oldSchema := range oldState.schemaByName {
		oldFieldsByName := map[string]client.SchemaFieldDescription{}
		for _, field := range oldSchema.Fields {
			oldFieldsByName[field.Name] = field
		}

		newSchema := newState.schemaByName[oldSchema.Name]

		for _, newField := range newSchema.Fields {
			oldField, exists := oldFieldsByName[newField.Name]
			if exists && oldField != newField {
				return NewErrCannotMutateField(newField.Name)
			}
		}
	}

	return nil
}

func validateFieldNotDuplicated(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	for _, schema := range newState.schemaByName {
		fieldNames := map[string]struct{}{}

		for _, field := range schema.Fields {
			if _, isDuplicate := fieldNames[field.Name]; isDuplicate {
				return NewErrDuplicateField(field.Name)
			}
			fieldNames[field.Name] = struct{}{}
		}
	}

	return nil
}

func validateSecondaryNotOnSchema(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	for _, newSchema := range newState.schemaByName {
		for _, newField := range newSchema.Fields {
			if newField.Kind.IsObjectArray() {
				return NewErrSecondaryFieldOnSchema(newField.Name)
			}
		}
	}

	return nil
}

func validateRelationalFieldIDType(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
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
						return NewErrRelationalFieldIDInvalidType(idField.Name, client.FieldKind_DocID, idField.Kind)
					}
				}
			}
		}
	}

	return nil
}
