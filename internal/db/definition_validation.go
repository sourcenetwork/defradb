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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
)

var patchCollectionValidators = []func(
	map[uint32]client.CollectionDescription,
	map[uint32]client.CollectionDescription,
) error{
	validateCollectionNameUnique,
	validateSingleVersionActive,
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
}

var newCollectionValidators = []func(
	client.CollectionDefinition,
	map[string]client.CollectionDefinition,
) error{
	validateSecondaryFieldsPairUp,
	validateRelationPointsToValidKind,
	validateSingleSidePrimary,
}

func (db *db) validateCollectionChanges(
	oldColsByID map[uint32]client.CollectionDescription,
	newColsByID map[uint32]client.CollectionDescription,
) error {
	for _, validators := range patchCollectionValidators {
		err := validators(oldColsByID, newColsByID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *db) validateNewCollection(
	def client.CollectionDefinition,
	defsByName map[string]client.CollectionDefinition,
) error {
	for _, validators := range newCollectionValidators {
		err := validators(def, defsByName)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateRelationPointsToValidKind(
	def client.CollectionDefinition,
	defsByName map[string]client.CollectionDefinition,
) error {
	for _, field := range def.Description.Fields {
		if !field.Kind.HasValue() {
			continue
		}

		if !field.Kind.Value().IsObject() {
			continue
		}

		underlying := field.Kind.Value().Underlying()
		_, ok := defsByName[underlying]
		if !ok {
			return NewErrFieldKindNotFound(field.Name, underlying)
		}
	}

	return nil
}

func validateSecondaryFieldsPairUp(
	def client.CollectionDefinition,
	defsByName map[string]client.CollectionDefinition,
) error {
	for _, field := range def.Description.Fields {
		if !field.Kind.HasValue() {
			continue
		}

		if !field.Kind.Value().IsObject() {
			continue
		}

		if !field.RelationName.HasValue() {
			continue
		}

		_, hasSchemaField := def.Schema.GetFieldByName(field.Name)
		if hasSchemaField {
			continue
		}

		underlying := field.Kind.Value().Underlying()
		otherDef, ok := defsByName[underlying]
		if !ok {
			continue
		}

		if len(otherDef.Description.Fields) == 0 {
			// Views/embedded objects do not require both sides of the relation to be defined.
			continue
		}

		otherField, ok := otherDef.Description.GetFieldByRelation(
			field.RelationName.Value(),
			def.GetName(),
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

	return nil
}

func validateSingleSidePrimary(
	def client.CollectionDefinition,
	defsByName map[string]client.CollectionDefinition,
) error {
	for _, field := range def.Description.Fields {
		if !field.Kind.HasValue() {
			continue
		}

		if !field.Kind.Value().IsObject() {
			continue
		}

		if !field.RelationName.HasValue() {
			continue
		}

		_, hasSchemaField := def.Schema.GetFieldByName(field.Name)
		if !hasSchemaField {
			// This is a secondary field and thus passes this rule
			continue
		}

		underlying := field.Kind.Value().Underlying()
		otherDef, ok := defsByName[underlying]
		if !ok {
			continue
		}

		otherField, ok := otherDef.Description.GetFieldByRelation(
			field.RelationName.Value(),
			def.GetName(),
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

	return nil
}

func validateCollectionNameUnique(
	oldColsByID map[uint32]client.CollectionDescription,
	newColsByID map[uint32]client.CollectionDescription,
) error {
	names := map[string]struct{}{}
	for _, col := range newColsByID {
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
	oldColsByID map[uint32]client.CollectionDescription,
	newColsByID map[uint32]client.CollectionDescription,
) error {
	rootsWithActiveCol := map[uint32]struct{}{}
	for _, col := range newColsByID {
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
	oldColsByID map[uint32]client.CollectionDescription,
	newColsByID map[uint32]client.CollectionDescription,
) error {
	for _, newCol := range newColsByID {
		oldCol, ok := oldColsByID[newCol.ID]
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
	oldColsByID map[uint32]client.CollectionDescription,
	newColsByID map[uint32]client.CollectionDescription,
) error {
	for _, newCol := range newColsByID {
		oldCol, ok := oldColsByID[newCol.ID]
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
	oldColsByID map[uint32]client.CollectionDescription,
	newColsByID map[uint32]client.CollectionDescription,
) error {
	for _, newCol := range newColsByID {
		oldCol, ok := oldColsByID[newCol.ID]
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
	oldColsByID map[uint32]client.CollectionDescription,
	newColsByID map[uint32]client.CollectionDescription,
) error {
	for _, newCol := range newColsByID {
		oldCol, ok := oldColsByID[newCol.ID]
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
	oldColsByID map[uint32]client.CollectionDescription,
	newColsByID map[uint32]client.CollectionDescription,
) error {
	for _, newCol := range newColsByID {
		if newCol.ID == 0 {
			return ErrCollectionIDCannotBeZero
		}
	}

	return nil
}

func validateIDUnique(
	oldColsByID map[uint32]client.CollectionDescription,
	newColsByID map[uint32]client.CollectionDescription,
) error {
	colIds := map[uint32]struct{}{}
	for _, newCol := range newColsByID {
		if _, ok := colIds[newCol.ID]; ok {
			return NewErrCollectionIDAlreadyExists(newCol.ID)
		}
		colIds[newCol.ID] = struct{}{}
	}

	return nil
}

func validateIDExists(
	oldColsByID map[uint32]client.CollectionDescription,
	newColsByID map[uint32]client.CollectionDescription,
) error {
	for _, newCol := range newColsByID {
		if _, ok := oldColsByID[newCol.ID]; !ok {
			return NewErrAddCollectionIDWithPatch(newCol.ID)
		}
	}

	return nil
}

func validateRootIDNotMutated(
	oldColsByID map[uint32]client.CollectionDescription,
	newColsByID map[uint32]client.CollectionDescription,
) error {
	for _, newCol := range newColsByID {
		oldCol, ok := oldColsByID[newCol.ID]
		if !ok {
			continue
		}

		if newCol.RootID != oldCol.RootID {
			return NewErrCollectionRootIDCannotBeMutated(newCol.ID)
		}
	}

	return nil
}

func validateSchemaVersionIDNotMutated(
	oldColsByID map[uint32]client.CollectionDescription,
	newColsByID map[uint32]client.CollectionDescription,
) error {
	for _, newCol := range newColsByID {
		oldCol, ok := oldColsByID[newCol.ID]
		if !ok {
			continue
		}

		if newCol.SchemaVersionID != oldCol.SchemaVersionID {
			return NewErrCollectionSchemaVersionIDCannotBeMutated(newCol.ID)
		}
	}

	return nil
}

func validateCollectionNotRemoved(
	oldColsByID map[uint32]client.CollectionDescription,
	newColsByID map[uint32]client.CollectionDescription,
) error {
oldLoop:
	for _, oldCol := range oldColsByID {
		for _, newCol := range newColsByID {
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
func (db *db) validateCollectionDefinitionPolicyDesc(
	ctx context.Context,
	policyDesc immutable.Option[client.PolicyDescription],
) error {
	if !policyDesc.HasValue() {
		// No policy validation needed, whether acp exists or not doesn't matter.
		return nil
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
	return db.acp.Value().ValidateResourceExistsOnValidDPI(
		ctx,
		policyDesc.Value().ID,
		policyDesc.Value().ResourceName,
	)
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

	if proposedDesc.Root != existingDesc.Root {
		return false, NewErrSchemaRootDoesntMatch(
			proposedDesc.Name,
			existingDesc.Root,
			proposedDesc.Root,
		)
	}

	if proposedDesc.Name != existingDesc.Name {
		// There is actually little reason to not support this atm besides controlling the surface area
		// of the new feature.  Changing this should not break anything, but it should be tested first.
		return false, NewErrCannotModifySchemaName(existingDesc.Name, proposedDesc.Name)
	}

	if proposedDesc.VersionID != "" && proposedDesc.VersionID != existingDesc.VersionID {
		// If users specify this it will be overwritten, an error is preferred to quietly ignoring it.
		return false, ErrCannotSetVersionID
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
	for proposedIndex, proposedField := range proposedDesc.Fields {
		existingField, fieldAlreadyExists := existingFieldsByName[proposedField.Name]

		// If the field is new, then the collection has changed
		hasChanged = hasChanged || !fieldAlreadyExists

		if !fieldAlreadyExists && proposedField.Kind.IsObject() {
			_, relatedDescFound := descriptionsByName[proposedField.Kind.Underlying()]

			if !relatedDescFound {
				return false, NewErrFieldKindNotFound(proposedField.Name, proposedField.Kind.Underlying())
			}

			if proposedField.Kind.IsObject() && !proposedField.Kind.IsArray() {
				idFieldName := proposedField.Name + request.RelatedObjectID
				idField, idFieldFound := proposedDesc.GetFieldByName(idFieldName)
				if idFieldFound {
					if idField.Kind != client.FieldKind_DocID {
						return false, NewErrRelationalFieldIDInvalidType(idField.Name, client.FieldKind_DocID, idField.Kind)
					}
				}
			}
		}

		if proposedField.Kind.IsObjectArray() {
			return false, NewErrSecondaryFieldOnSchema(proposedField.Name)
		}

		if _, isDuplicate := newFieldNames[proposedField.Name]; isDuplicate {
			return false, NewErrDuplicateField(proposedField.Name)
		}

		if fieldAlreadyExists && proposedField != existingField {
			return false, NewErrCannotMutateField(proposedField.Name)
		}

		if existingIndex := existingFieldIndexesByName[proposedField.Name]; fieldAlreadyExists &&
			proposedIndex != existingIndex {
			return false, NewErrCannotMoveField(proposedField.Name, proposedIndex, existingIndex)
		}

		if !proposedField.Typ.IsSupportedFieldCType() {
			return false, client.NewErrInvalidCRDTType(proposedField.Name, proposedField.Typ.String())
		}

		if !proposedField.Typ.IsCompatibleWith(proposedField.Kind) {
			return false, client.NewErrCRDTKindMismatch(proposedField.Typ.String(), proposedField.Kind.String())
		}

		newFieldNames[proposedField.Name] = struct{}{}
	}

	for _, field := range existingDesc.Fields {
		if _, stillExists := newFieldNames[field.Name]; !stillExists {
			return false, NewErrCannotDeleteField(field.Name)
		}
	}
	return hasChanged, nil
}
