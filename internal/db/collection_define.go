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
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"slices"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/db/description"
)

func (db *DB) createCollections(
	ctx context.Context,
	parseResults []core.Collection,
) ([]client.CollectionDefinition, error) {
	returnDescriptions := make([]client.CollectionDefinition, 0, len(parseResults))

	existingDefinitions, err := db.getAllActiveDefinitions(ctx)
	if err != nil {
		return nil, err
	}

	existingDefinitionsByID := make(map[string]client.CollectionDefinition, len(existingDefinitions))
	for _, col := range existingDefinitions {
		existingDefinitionsByID[col.Version.CollectionID] = col
	}

	newCollections := make([]client.CollectionVersion, len(parseResults))
	for i, def := range parseResults {
		newCollections[i] = def.Definition.Version
	}

	err = setCollectionIDs(ctx, newCollections, immutable.None[model.Lens]())
	if err != nil {
		return nil, err
	}

	for i := range parseResults {
		// The secondary index code requires the useage of core.Collection which means we need to
		// map the CollectionVersion back on to the input param.
		parseResults[i].Definition.Version = newCollections[i]
	}

	newDefinitions := make([]client.CollectionDefinition, len(parseResults))
	for i, def := range parseResults {
		newDefinitions[i] = def.Definition
		newDefinitions[i].Version = newCollections[i]
	}

	err = db.validateNewCollection(
		ctx,
		slices.Concat(newDefinitions, existingDefinitions),
		existingDefinitions,
	)
	if err != nil {
		return nil, err
	}

	for _, def := range parseResults {
		def.Definition.Version.Indexes = make([]client.IndexDescription, 0, len(def.CreateIndexes))
		for _, createIndex := range def.CreateIndexes {
			desc, err := processCreateIndexRequest(ctx, def.Definition, createIndex)
			if err != nil {
				return nil, err
			}
			def.Definition.Version.Indexes = append(def.Definition.Version.Indexes, desc)
		}

		err = description.SaveCollection(ctx, def.Definition.Version)
		if err != nil {
			return nil, err
		}

		col, err := db.newCollection(def.Definition.Version)
		if err != nil {
			return nil, err
		}

		for _, index := range def.Definition.Version.Indexes {
			if _, err := col.addNewIndex(ctx, index); err != nil {
				return nil, err
			}
		}

		result, err := db.getCollectionByID(ctx, def.Definition.Version.VersionID)
		if err != nil {
			return nil, err
		}

		returnDescriptions = append(returnDescriptions, result.Definition())
	}

	return returnDescriptions, nil
}

// PatchCollection takes the given JSON patch string and applies it to the set of CollectionVersions
// present in the database.
//
// It will also update the GQL types used by the query system. It will error and not apply any of the
// requested, valid updates should the net result of the patch result in an invalid state.  The
// individual operations defined in the patch do not need to result in a valid state, only the net result
// of the full patch.
//
// New CollectionVersions created by modifying the global type definition (e.g. renaming, adding fields, etc)
// will automatically become the active version of the Collection, unless `IsActive` is set to false by the patch.
//
// Field [FieldKind] values may be provided in either their raw integer form, or as string as per
// [FieldKindStringToEnumMapping].
//
// CollectionVersions may be referenced by their VersionID, or their Name.  Referencing by name will patch
// the current active version, whereas referencing by VersionID will patch that specific version, whether it is
// currently active or not.
//
// A lens configuration may also be provided, and will become the migration to any new CollectionVersions created
// by the patch.
func (db *DB) patchCollection(
	ctx context.Context,
	patchString string,
	migration immutable.Option[model.Lens],
) error {
	patch, err := jsonpatch.DecodePatch([]byte(patchString))
	if err != nil {
		return err
	}
	existingCols, err := db.getCollections(
		ctx,
		client.CollectionFetchOptions{IncludeInactive: immutable.Some(true)},
	)
	if err != nil {
		return err
	}

	existingColsByName := map[string]client.CollectionVersion{}
	existingColsByID := map[string]client.CollectionVersion{}
	existingDefinitions := make([]client.CollectionDefinition, 0, len(existingCols))
	for _, col := range existingCols {
		if col.Version().IsActive {
			existingColsByName[col.Version().Name] = col.Version()
		}
		existingColsByID[col.Version().VersionID] = col.Version()
		existingDefinitions = append(existingDefinitions, col.Definition())
	}

	// Here we swap out any string representations of enums for their integer values
	patch, err = substituteCollectionPatch(patch, existingColsByName)
	if err != nil {
		return err
	}

	existingDescriptionJson, err := json.Marshal(existingColsByID)
	if err != nil {
		return err
	}

	newDescriptionJson, err := patch.Apply(existingDescriptionJson)
	if err != nil {
		return err
	}

	var newColsByID map[string]client.CollectionVersion
	decoder := json.NewDecoder(strings.NewReader(string(newDescriptionJson)))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&newColsByID)
	if err != nil {
		return err
	}

	newDefinitions := make([]client.CollectionDefinition, len(existingCols))
	updatedColsByID := make(map[string]struct{})
	for i, col := range existingCols {
		newDefinitions[i].Version = newColsByID[col.Version().VersionID]
		updatedColsByID[col.Version().VersionID] = struct{}{}
	}

	for _, col := range newColsByID {
		// Automatically add any id fields for object fields added by the patch, if the patch did not explicitly
		// add one.
		for _, field := range col.Fields {
			if field.Kind.IsObject() && !field.Kind.IsArray() {
				idFieldName := field.Name + "_id"
				if _, ok := col.GetFieldByName(idFieldName); !ok {
					col.Fields = append(col.Fields, client.CollectionFieldDescription{
						Name:         idFieldName,
						Kind:         client.FieldKind_DocID,
						RelationName: field.RelationName,
						IsPrimary:    field.IsPrimary,
					})
				}
			}
		}
	}

	for key, col := range newColsByID {
		previousCol := existingColsByName[col.Name]

		previousFieldNames := make(map[string]struct{}, len(previousCol.Fields))
		for _, field := range previousCol.Fields {
			previousFieldNames[field.FieldID] = struct{}{}
		}

		for i, field := range col.Fields {
			if _, existed := previousFieldNames[field.FieldID]; !existed && field.Typ == client.NONE_CRDT {
				// If no CRDT Type has been provided to a new field, default to LWW_REGISTER.
				// If the field existed before it might have been explicitly cleared by the user, in which
				// case it is up to the validation logic to error or not.
				newColsByID[key].Fields[i].Typ = client.LWW_REGISTER
			}
		}
	}

	newCollections := make([]client.CollectionVersion, 0, len(newColsByID))
	for _, col := range newColsByID {
		newCollections = append(newCollections, col)
	}

	err = setCollectionIDs(ctx, newCollections, migration)
	if err != nil {
		return err
	}

	for _, existingCol := range existingColsByName {
		isMissing := true
		for _, newCol := range newCollections {
			if newCol.VersionID == existingCol.VersionID {
				isMissing = false
				break
			}
		}

		// If an existing collection is not present in the new collection set,
		// it must have mutated into a new collection version.
		// The original still needs to exist and must be validated against.
		// It may also be mutated later in this function.
		if isMissing {
			for _, newCol := range newCollections {
				if newCol.CollectionID == existingCol.CollectionID && newCol.IsActive {
					existingCol.IsActive = false
					break
				}
			}
			newCollections = append(newCollections, existingCol)
		}
	}

	for i := 0; i < len(newCollections); i++ {
		placeholder := newCollections[i]
		if placeholder.IsPlaceholder {
			isFound := false
			for j, col := range newCollections {
				if col.VersionID == placeholder.VersionID && !col.IsPlaceholder {
					newCollections[j].Sources = placeholder.Sources
					isFound = true
					break
				}
			}

			if isFound {
				// Remove the original placeholder from the collection set, its sources
				// have been copied to the actual definition (with the same VersionID)
				newCollections = append(newCollections[:i], newCollections[i+1:]...)
				i--
			}
		}
	}

	newDefinitions = make([]client.CollectionDefinition, 0, len(newCollections))
	for _, col := range newCollections {
		newDefinitions = append(newDefinitions, client.CollectionDefinition{Version: col})
	}

	err = db.validateCollectionChanges(ctx, existingDefinitions, newDefinitions)
	if err != nil {
		return err
	}

	for _, col := range newCollections {
		existingCol, ok := existingColsByID[col.VersionID]
		if ok && col.Equal(existingCol) {
			continue
		}

		err := description.SaveCollection(ctx, col)
		if err != nil {
			return err
		}

		if ok {
			if existingCol.IsMaterialized && !col.IsMaterialized {
				// If the collection is being de-materialized - delete any cached values.
				// Leaving them around will not break anything, but it would be a waste of
				// storage space.
				err := db.clearViewCache(ctx, client.CollectionDefinition{
					Version: col,
				})
				if err != nil {
					return err
				}
			}

			// Clear any existing migrations in the registry, using this semi-hacky way
			// to avoid adding more functions to a public interface that we wish to remove.

			for _, src := range existingCol.CollectionSources() {
				if src.Transform.HasValue() {
					err = db.LensRegistry().SetMigration(ctx, existingCol.VersionID, model.Lens{})
					if err != nil {
						return err
					}
				}
			}
			for _, src := range existingCol.QuerySources() {
				if src.Transform.HasValue() {
					err = db.LensRegistry().SetMigration(ctx, existingCol.VersionID, model.Lens{})
					if err != nil {
						return err
					}
				}
			}
		}

		for _, src := range col.CollectionSources() {
			if src.Transform.HasValue() {
				err = db.LensRegistry().SetMigration(ctx, col.VersionID, src.Transform.Value())
				if err != nil {
					return err
				}
			}
		}

		for _, src := range col.QuerySources() {
			if src.Transform.HasValue() {
				err = db.LensRegistry().SetMigration(ctx, col.VersionID, src.Transform.Value())
				if err != nil {
					return err
				}
			}
		}
	}

	return db.loadSchema(ctx)
}

const (
	collectionNamePathIndex int = 0
	fieldsPathIndex         int = 1
	fieldIndexPathIndex     int = 2
)

// substituteCollectionPatch handles any substitution of values that may be required before
// the patch can be applied.
//
// For example Field [FieldKind] string representations will be replaced by the raw integer
// value.
func substituteCollectionPatch(
	patch jsonpatch.Patch,
	collectionsByName map[string]client.CollectionVersion,
) (jsonpatch.Patch, error) {
	fieldIndexesBySchema := make(map[string]map[string]int, len(collectionsByName))
	for schemaName, schema := range collectionsByName {
		fieldIndexesByName := make(map[string]int, len(schema.Fields))
		fieldIndexesBySchema[schemaName] = fieldIndexesByName
		for i, field := range schema.Fields {
			fieldIndexesByName[field.Name] = i
		}
	}

	for _, patchOperation := range patch {
		path, err := patchOperation.Path()
		if err != nil {
			return nil, err
		}
		path = strings.TrimPrefix(path, "/")

		if value, hasValue := patchOperation["value"]; hasValue {
			splitPath := strings.Split(path, "/")

			var newPatchValue immutable.Option[any]
			var field map[string]any
			isField := isField(splitPath)

			if isField {
				// We unmarshal the full field-value into a map to ensure that all user
				// specified properties are maintained.
				err = json.Unmarshal(*value, &field)
				if err != nil {
					return nil, err
				}
			}

			if isFieldOrInner(splitPath) {
				fieldIndexer := splitPath[fieldIndexPathIndex]

				if containsLetter(fieldIndexer) {
					if isField {
						if nameValue, hasName := field["Name"]; hasName {
							if name, isString := nameValue.(string); isString && name != fieldIndexer {
								return nil, NewErrIndexDoesNotMatchName(fieldIndexer, name)
							}
						} else {
							field["Name"] = fieldIndexer
						}
						newPatchValue = immutable.Some[any](field)
					}

					desc := collectionsByName[splitPath[collectionNamePathIndex]]
					var index string
					if fieldIndexesByName, ok := fieldIndexesBySchema[desc.Name]; ok {
						if i, ok := fieldIndexesByName[fieldIndexer]; ok {
							index = fmt.Sprint(i)
						}
					}
					if index == "" {
						index = "-"
						// If this is a new field we need to track its location so that subsequent operations
						// within the patch may access it by field name.
						fieldIndexesBySchema[desc.Name][fieldIndexer] = len(fieldIndexesBySchema[desc.Name])
					}

					splitPath[fieldIndexPathIndex] = index
					path = strings.Join(splitPath, "/")
					opPath := json.RawMessage([]byte(fmt.Sprintf(`"/%s"`, path)))
					patchOperation["path"] = &opPath
				}
			}

			if newPatchValue.HasValue() {
				substitute, err := json.Marshal(newPatchValue.Value())
				if err != nil {
					return nil, err
				}

				substitutedValue := json.RawMessage(substitute)
				patchOperation["value"] = &substitutedValue
			}
		}

		splitPath := strings.Split(path, "/")
		if len(splitPath) > 0 {
			// If the path contains a collection name, substitute it for the version id
			if col, ok := collectionsByName[splitPath[0]]; ok {
				splitPath[0] = col.VersionID
				path = strings.Join(splitPath, "/")
				opPath := json.RawMessage([]byte(fmt.Sprintf(`"/%s"`, path)))
				patchOperation["path"] = &opPath
			}
		}

		fromPath, ok := patchOperation["from"]
		if ok {
			var from string
			err := json.Unmarshal(*fromPath, &from)
			if err != nil {
				return nil, err
			}
			from = strings.TrimPrefix(from, "/")

			splitPath := strings.Split(from, "/")
			if len(splitPath) > 0 {
				// If 'from' exists, and contains a collection name, substitute it for the version id
				if col, ok := collectionsByName[splitPath[0]]; ok {
					splitPath[0] = col.VersionID
					from = strings.Join(splitPath, "/")
					opPath := json.RawMessage([]byte(fmt.Sprintf(`"/%s"`, from)))
					patchOperation["from"] = &opPath
				}
			}
		}
	}

	return patch, nil
}

// isFieldOrInner returns true if the given path points to a SchemaFieldDescription or a property within it.
func isFieldOrInner(path []string) bool {
	//nolint:goconst
	return len(path) >= 3 && path[fieldsPathIndex] == "Fields"
}

// isField returns true if the given path points to a SchemaFieldDescription.
func isField(path []string) bool {
	return len(path) == 3 && path[fieldsPathIndex] == "Fields"
}

// containsLetter returns true if the string contains a single unicode character.
func containsLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

// SetActiveCollectionVersion activates all collection versions with the given schema version, and deactivates all
// those without it (if they share the same schema root).
//
// This will affect all operations interacting with the schema where a schema version is not explicitly
// provided.  This includes GQL queries and Collection operations.
//
// It will return an error if the provided schema version ID does not exist.
func (db *DB) setActiveCollectionVersion(
	ctx context.Context,
	versionID string,
) error {
	if versionID == "" {
		return ErrSchemaVersionIDEmpty
	}
	col, err := description.GetCollectionByID(ctx, versionID)
	if err != nil {
		return err
	}

	colsWithRoot, err := description.GetCollectionsByCollectionID(ctx, col.CollectionID)
	if err != nil {
		return err
	}

	for _, col := range colsWithRoot {
		if col.VersionID == versionID {
			if col.IsActive {
				continue
			}

			col.IsActive = true
			err = description.SaveCollection(ctx, col)
			if err != nil {
				return err
			}

			continue
		}

		if !col.IsActive {
			continue
		}

		col.IsActive = false
		err = description.SaveCollection(ctx, col)
		if err != nil {
			return err
		}
	}

	// Load the schema into the clients (e.g. GQL)
	return db.loadSchema(ctx)
}
