// Copyright 2022 Democratized Data Foundation
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
	"github.com/lens-vm/lens/host-go/config/model"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/db/description"
)

const (
	schemaNamePathIndex int = 0
	fieldsPathIndex     int = 1
	fieldIndexPathIndex int = 2
)

// addSchema takes the provided schema in SDL format, and applies it to the database,
// and creates the necessary collections, request types, etc.
func (db *db) addSchema(
	ctx context.Context,
	schemaString string,
) ([]client.CollectionDescription, error) {
	newDefinitions, err := db.parser.ParseSDL(ctx, schemaString)
	if err != nil {
		return nil, err
	}

	returnDefinitions, err := db.createCollections(ctx, newDefinitions)
	if err != nil {
		return nil, err
	}

	returnDescriptions := make([]client.CollectionDescription, len(returnDefinitions))
	for i, def := range returnDefinitions {
		returnDescriptions[i] = def.Description
	}

	err = db.loadSchema(ctx)
	if err != nil {
		return nil, err
	}

	return returnDescriptions, nil
}

func (db *db) loadSchema(ctx context.Context) error {
	txn := mustGetContextTxn(ctx)

	definitions, err := db.getAllActiveDefinitions(ctx)
	if err != nil {
		return err
	}

	return db.parser.SetSchema(ctx, txn, definitions)
}

// patchSchema takes the given JSON patch string and applies it to the set of SchemaDescriptions
// present in the database.
//
// It will also update the GQL types used by the query system. It will error and not apply any of the
// requested, valid updates should the net result of the patch result in an invalid state.  The
// individual operations defined in the patch do not need to result in a valid state, only the net result
// of the full patch.
//
// The collections (including the schema version ID) will only be updated if any changes have actually
// been made, if the net result of the patch matches the current persisted description then no changes
// will be applied.
func (db *db) patchSchema(
	ctx context.Context,
	patchString string,
	migration immutable.Option[model.Lens],
	setAsDefaultVersion bool,
) error {
	txn := mustGetContextTxn(ctx)

	patch, err := jsonpatch.DecodePatch([]byte(patchString))
	if err != nil {
		return err
	}

	schemas, err := description.GetSchemas(ctx, txn)
	if err != nil {
		return err
	}

	existingSchemaByName := map[string]client.SchemaDescription{}
	for _, schema := range schemas {
		existingSchemaByName[schema.Name] = schema
	}

	// Here we swap out any string representations of enums for their integer values
	patch, err = substituteSchemaPatch(patch, existingSchemaByName)
	if err != nil {
		return err
	}

	existingDescriptionJson, err := json.Marshal(existingSchemaByName)
	if err != nil {
		return err
	}

	newDescriptionJson, err := patch.Apply(existingDescriptionJson)
	if err != nil {
		return err
	}

	var newSchemaByName map[string]client.SchemaDescription
	decoder := json.NewDecoder(strings.NewReader(string(newDescriptionJson)))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&newSchemaByName)
	if err != nil {
		return err
	}

	for _, schema := range newSchemaByName {
		err := db.updateSchema(
			ctx,
			existingSchemaByName,
			newSchemaByName,
			schema,
			migration,
			setAsDefaultVersion,
		)
		if err != nil {
			return err
		}
	}

	return db.loadSchema(ctx)
}

// substituteSchemaPatch handles any substitution of values that may be required before
// the patch can be applied.
//
// For example Field [FieldKind] string representations will be replaced by the raw integer
// value.
func substituteSchemaPatch(
	patch jsonpatch.Patch,
	schemaByName map[string]client.SchemaDescription,
) (jsonpatch.Patch, error) {
	fieldIndexesBySchema := make(map[string]map[string]int, len(schemaByName))
	for schemaName, schema := range schemaByName {
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

		if value, hasValue := patchOperation["value"]; hasValue {
			path = strings.TrimPrefix(path, "/")
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

					desc := schemaByName[splitPath[schemaNamePathIndex]]
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
	}

	return patch, nil
}

func (db *db) getSchemaByVersionID(
	ctx context.Context,
	versionID string,
) (client.SchemaDescription, error) {
	schemas, err := db.getSchemas(ctx, client.SchemaFetchOptions{ID: immutable.Some(versionID)})
	if err != nil {
		return client.SchemaDescription{}, err
	}

	// schemas will always have length == 1 here
	return schemas[0], nil
}

func (db *db) getSchemas(
	ctx context.Context,
	options client.SchemaFetchOptions,
) ([]client.SchemaDescription, error) {
	txn := mustGetContextTxn(ctx)

	schemas := []client.SchemaDescription{}

	switch {
	case options.ID.HasValue():
		schema, err := description.GetSchemaVersion(ctx, txn, options.ID.Value())
		if err != nil {
			return nil, err
		}
		schemas = append(schemas, schema)

	case options.Root.HasValue():
		var err error
		schemas, err = description.GetSchemasByRoot(ctx, txn, options.Root.Value())
		if err != nil {
			return nil, err
		}
	case options.Name.HasValue():
		var err error
		schemas, err = description.GetSchemasByName(ctx, txn, options.Name.Value())
		if err != nil {
			return nil, err
		}
	default:
		return description.GetAllSchemas(ctx, txn)
	}

	result := []client.SchemaDescription{}
	for _, schema := range schemas {
		if options.Root.HasValue() && schema.Root != options.Root.Value() {
			continue
		}
		if options.Name.HasValue() && schema.Name != options.Name.Value() {
			continue
		}
		result = append(result, schema)
	}

	return result, nil
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

// updateSchema updates the persisted schema description matching the name of the given
// description, to the values in the given description.
//
// It will validate the given description using [validateUpdateSchema] before updating it.
//
// The schema (including the schema version ID) will only be updated if any changes have actually
// been made, if the given description matches the current persisted description then no changes will be
// applied.
func (db *db) updateSchema(
	ctx context.Context,
	existingSchemaByName map[string]client.SchemaDescription,
	proposedDescriptionsByName map[string]client.SchemaDescription,
	schema client.SchemaDescription,
	migration immutable.Option[model.Lens],
	setAsActiveVersion bool,
) error {
	previousSchema := existingSchemaByName[schema.Name]

	areEqual := areSchemasEqual(schema, previousSchema)
	if areEqual {
		return nil
	}

	err := db.validateSchemaUpdate(ctx, proposedDescriptionsByName, existingSchemaByName)
	if err != nil {
		return err
	}

	for _, field := range schema.Fields {
		if field.Kind.IsObject() && !field.Kind.IsArray() {
			idFieldName := field.Name + "_id"
			if _, ok := schema.GetFieldByName(idFieldName); !ok {
				schema.Fields = append(schema.Fields, client.SchemaFieldDescription{
					Name: idFieldName,
					Kind: client.FieldKind_DocID,
				})
			}
		}
	}

	previousFieldNames := make(map[string]struct{}, len(previousSchema.Fields))
	for _, field := range previousSchema.Fields {
		previousFieldNames[field.Name] = struct{}{}
	}

	for i, field := range schema.Fields {
		if _, existed := previousFieldNames[field.Name]; !existed && field.Typ == client.NONE_CRDT {
			// If no CRDT Type has been provided, default to LWW_REGISTER.
			field.Typ = client.LWW_REGISTER
			schema.Fields[i] = field
		}
	}

	txn := mustGetContextTxn(ctx)
	previousVersionID := schema.VersionID
	schema, err = description.CreateSchemaVersion(ctx, txn, schema)
	if err != nil {
		return err
	}

	// After creating the new schema version, we need to create new collection versions for
	// any collection using the previous version.  These will be inactive unless [setAsActiveVersion]
	// is true.

	cols, err := description.GetCollectionsBySchemaVersionID(ctx, txn, previousVersionID)
	if err != nil {
		return err
	}

	colSeq, err := db.getSequence(ctx, core.CollectionIDSequenceKey{})
	if err != nil {
		return err
	}

	for _, col := range cols {
		previousID := col.ID

		existingCols, err := description.GetCollectionsBySchemaVersionID(ctx, txn, schema.VersionID)
		if err != nil {
			return err
		}

		// The collection version may exist before the schema version was created locally.  This is
		// because migrations for the globally known schema version may have been registered locally
		// (typically to handle documents synced over P2P at higher versions) before the local schema
		// was updated.  We need to check for them now, and update them instead of creating new ones
		// if they exist.
		var isExistingCol bool
	existingColLoop:
		for _, existingCol := range existingCols {
			sources := existingCol.CollectionSources()
			for _, source := range sources {
				// Make sure that this collection is the parent of the current [col], and not part of
				// another collection set that happens to be using the same schema.
				if source.SourceCollectionID == previousID {
					if existingCol.RootID == client.OrphanRootID {
						existingCol.RootID = col.RootID
					}

					fieldSeq, err := db.getSequence(ctx, core.NewFieldIDSequenceKey(existingCol.RootID))
					if err != nil {
						return err
					}

					for _, globalField := range schema.Fields {
						var fieldID client.FieldID
						// We must check the source collection if the field already exists, and take its ID
						// from there, otherwise the field must be generated by the sequence.
						existingField, ok := col.GetFieldByName(globalField.Name)
						if ok {
							fieldID = existingField.ID
						} else {
							nextFieldID, err := fieldSeq.next(ctx)
							if err != nil {
								return err
							}
							fieldID = client.FieldID(nextFieldID)
						}

						existingCol.Fields = append(
							existingCol.Fields,
							client.CollectionFieldDescription{
								Name: globalField.Name,
								ID:   fieldID,
							},
						)
					}
					existingCol, err = description.SaveCollection(ctx, txn, existingCol)
					if err != nil {
						return err
					}
					isExistingCol = true
					break existingColLoop
				}
			}
		}

		if !isExistingCol {
			colID, err := colSeq.next(ctx)
			if err != nil {
				return err
			}

			fieldSeq, err := db.getSequence(ctx, core.NewFieldIDSequenceKey(col.RootID))
			if err != nil {
				return err
			}

			// Create any new collections without a name (inactive), if [setAsActiveVersion] is true
			// they will be activated later along with any existing collection versions.
			col.Name = immutable.None[string]()
			col.ID = uint32(colID)
			col.SchemaVersionID = schema.VersionID
			col.Sources = []any{
				&client.CollectionSource{
					SourceCollectionID: previousID,
					Transform:          migration,
				},
			}

			for _, globalField := range schema.Fields {
				_, exists := col.GetFieldByName(globalField.Name)
				if !exists {
					fieldID, err := fieldSeq.next(ctx)
					if err != nil {
						return err
					}

					col.Fields = append(
						col.Fields,
						client.CollectionFieldDescription{
							Name: globalField.Name,
							ID:   client.FieldID(fieldID),
						},
					)
				}
			}

			_, err = description.SaveCollection(ctx, txn, col)
			if err != nil {
				return err
			}

			if migration.HasValue() {
				err = db.LensRegistry().SetMigration(ctx, col.ID, migration.Value())
				if err != nil {
					return err
				}
			}
		}
	}

	if setAsActiveVersion {
		// activate collection versions using the new schema ID.  This call must be made after
		// all new collection versions have been saved.
		err = db.setActiveSchemaVersion(ctx, schema.VersionID)
		if err != nil {
			return err
		}
	}

	return nil
}

func areSchemasEqual(this client.SchemaDescription, that client.SchemaDescription) bool {
	if len(this.Fields) != len(that.Fields) {
		return false
	}

	for i, thisField := range this.Fields {
		if thisField != that.Fields[i] {
			return false
		}
	}

	return this.Name == that.Name &&
		this.Root == that.Root &&
		this.VersionID == that.VersionID
}
