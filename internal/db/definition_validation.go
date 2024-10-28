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
	definitionCache   client.DefinitionCache
}

// newDefinitionStateFromCols creates a new definitionState object given the provided
// collection descriptions.
func newDefinitionStateFromCols(
	collections []client.CollectionDescription,
) *definitionState {
	collectionsByID := map[uint32]client.CollectionDescription{}
	definitionsByName := map[string]client.CollectionDefinition{}
	definitions := []client.CollectionDefinition{}
	schemaByName := map[string]client.SchemaDescription{}

	for _, col := range collections {
		if len(col.Fields) == 0 {
			continue
		}

		definition := client.CollectionDefinition{
			Description: col,
		}

		definitionsByName[definition.GetName()] = definition
		definitions = append(definitions, definition)
		collectionsByID[col.ID] = col
	}

	return &definitionState{
		collections:       collections,
		collectionsByID:   collectionsByID,
		schemaByID:        map[string]client.SchemaDescription{},
		schemaByName:      schemaByName,
		definitionsByName: definitionsByName,
		definitionCache:   client.NewDefinitionCache(definitions),
	}
}

// newDefinitionState creates a new definitionState object given the provided
// definitions.
func newDefinitionState(
	definitions []client.CollectionDefinition,
) *definitionState {
	collectionsByID := map[uint32]client.CollectionDescription{}
	schemasByID := map[string]client.SchemaDescription{}
	definitionsByName := map[string]client.CollectionDefinition{}
	collections := []client.CollectionDescription{}
	schemaByName := map[string]client.SchemaDescription{}

	for _, def := range definitions {
		definitionsByName[def.GetName()] = def
		schemasByID[def.Schema.VersionID] = def.Schema
		schemaByName[def.Schema.Name] = def.Schema

		if len(def.Description.Fields) != 0 {
			collectionsByID[def.Description.ID] = def.Description
			collections = append(collections, def.Description)
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
	db *db,
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
	validateIDNotZero,
	validateIDUnique,
	validateRootIDNotMutated,
	validateSingleVersionActive,
	validateSchemaNotAdded,
	validateSchemaFieldNotDeleted,
	validateFieldNotMutated,
	validateFieldNotMoved,
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
	validateRelationalFieldIDType,
	validateSecondaryNotOnSchema,
	validateTypeSupported,
	validateTypeAndKindCompatible,
	validateFieldNotDuplicated,
	validateSelfReferences,
	validateCollectionMaterialized,
	validateMaterializedHasNoPolicy,
	validateCollectionFieldDefaultValue,
}

var createValidators = append(
	append([]definitionValidator{}, createOnlyValidators...),
	globalValidators...,
)

func (db *db) validateSchemaUpdate(
	ctx context.Context,
	oldDefinitions []client.CollectionDefinition,
	newDefinitions []client.CollectionDefinition,
) error {
	newState := newDefinitionState(newDefinitions)
	oldState := newDefinitionState(oldDefinitions)

	for _, validator := range schemaUpdateValidators {
		err := validator(ctx, db, newState, oldState)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *db) validateCollectionChanges(
	ctx context.Context,
	oldCols []client.CollectionDescription,
	newColsByID map[uint32]client.CollectionDescription,
) error {
	newCols := make([]client.CollectionDescription, 0, len(newColsByID))
	for _, col := range newColsByID {
		newCols = append(newCols, col)
	}

	newState := newDefinitionStateFromCols(newCols)
	oldState := newDefinitionStateFromCols(oldCols)

	for _, validator := range collectionUpdateValidators {
		err := validator(ctx, db, newState, oldState)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *db) validateNewCollection(
	ctx context.Context,
	newDefinitions []client.CollectionDefinition,
	oldDefinitions []client.CollectionDefinition,
) error {
	newState := newDefinitionState(newDefinitions)
	oldState := newDefinitionState(oldDefinitions)

	for _, validator := range createValidators {
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

			definition := newState.definitionsByName[newCollection.Name.Value()]
			_, ok := client.GetDefinition(newState.definitionCache, definition, field.Kind.Value())
			if !ok {
				return NewErrFieldKindNotFound(field.Name, field.Kind.Value().String())
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
				return NewErrFieldKindNotFound(field.Name, field.Kind.String())
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

			otherDef, ok := client.GetDefinition(newState.definitionCache, definition, field.Kind.Value())
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
				return NewErrRelationMissingField(otherDef.GetName(), field.RelationName.Value())
			}

			_, ok = otherDef.Schema.GetFieldByName(otherField.Name)
			if !ok {
				// This secondary is paired with another secondary, which is invalid
				return NewErrRelationMissingField(otherDef.GetName(), field.RelationName.Value())
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

			otherDef, ok := client.GetDefinition(newState.definitionCache, definition, field.Kind)
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
		oldSchema, ok := oldState.schemaByName[newSchema.Name]
		if !ok {
			continue
		}

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

			// DeepEqual is temporary, as this validation is temporary
			if exists && !reflect.DeepEqual(oldField, newField) {
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

func validateSelfReferences(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
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
				return NewErrSelfReferenceWithoutSelf(field.Name)
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

			definition := newState.definitionsByName[col.Name.Value()]
			otherDef, ok := client.GetDefinition(newState.definitionCache, definition, field.Kind.Value())
			if !ok {
				continue
			}

			if otherDef.Description.RootID == col.RootID {
				return NewErrSelfReferenceWithoutSelf(field.Name)
			}
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
			if newField.Kind.IsObject() && newField.Kind.IsArray() {
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

func validateSchemaNotAdded(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	for _, newSchema := range newState.schemaByName {
		if newSchema.Name == "" {
			// continue, and allow a more appropriate rule to return a nicer error
			// for the user
			continue
		}

		if _, exists := oldState.schemaByName[newSchema.Name]; !exists {
			return NewErrAddSchemaWithPatch(newSchema.Name)
		}
	}

	return nil
}

func validateSchemaNameNotEmpty(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	for _, schema := range newState.schemaByName {
		if schema.Name == "" {
			return ErrSchemaNameEmpty
		}
	}

	return nil
}

// validateCollectionMaterialized verifies that a non-view collection is materialized.
//
// Long term we wish to support this, however for now we block it off.
func validateCollectionMaterialized(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	for _, col := range newState.collections {
		if len(col.QuerySources()) == 0 && !col.IsMaterialized {
			return NewErrColNotMaterialized(col.Name.Value())
		}
	}

	return nil
}

// validateMaterializedHasNoPolicy verifies that a materialized view has no ACP policy.
//
// Long term we wish to support this, however for now we block it off.
func validateMaterializedHasNoPolicy(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	for _, col := range newState.collections {
		if col.IsMaterialized && len(col.QuerySources()) != 0 && col.Policy.HasValue() {
			return NewErrMaterializedViewAndACPNotSupported(col.Name.Value())
		}
	}

	return nil
}

func validateCollectionFieldDefaultValue(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	for name, col := range newState.definitionsByName {
		// default values are set when a doc is first created
		_, err := client.NewDocFromMap(map[string]any{}, col)
		if err != nil {
			return NewErrDefaultFieldValueInvalid(name, err)
		}
	}

	return nil
}

// validateCollectionIsBranchableNotMutated is a temporary restriction that prevents users from toggling
// whether or not a collection is branchable.
func validateCollectionIsBranchableNotMutated(
	ctx context.Context,
	db *db,
	newState *definitionState,
	oldState *definitionState,
) error {
	for _, newCol := range newState.collections {
		oldCol := oldState.collectionsByID[newCol.ID]

		if newCol.IsBranchable != oldCol.IsBranchable {
			return NewErrColMutatingIsBranchable(newCol.Name.Value())
		}
	}

	return nil
}
