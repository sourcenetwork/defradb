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
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/db/description"
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
	txn datastore.Txn,
	schemaString string,
) ([]client.CollectionDescription, error) {
	newDefinitions, err := db.parser.ParseSDL(ctx, schemaString)
	if err != nil {
		return nil, err
	}

	returnDescriptions := make([]client.CollectionDescription, len(newDefinitions))
	for i, definition := range newDefinitions {
		// Only accept the schema if policy description is valid, otherwise reject the schema.
		err := db.validateCollectionDefinitionPolicyDesc(ctx, definition.Description.Policy)
		if err != nil {
			return nil, err
		}

		col, err := db.createCollection(ctx, txn, definition)
		if err != nil {
			return nil, err
		}
		returnDescriptions[i] = col.Description()
	}

	err = db.loadSchema(ctx, txn)
	if err != nil {
		return nil, err
	}

	return returnDescriptions, nil
}

func (db *db) loadSchema(ctx context.Context, txn datastore.Txn) error {
	definitions, err := db.getAllActiveDefinitions(ctx, txn)
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
	txn datastore.Txn,
	patchString string,
	migration immutable.Option[model.Lens],
	setAsDefaultVersion bool,
) error {
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
			txn,
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

	return db.loadSchema(ctx, txn)
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
	txn datastore.Txn,
	versionID string,
) (client.SchemaDescription, error) {
	schemas, err := db.getSchemas(ctx, txn, client.SchemaFetchOptions{ID: immutable.Some(versionID)})
	if err != nil {
		return client.SchemaDescription{}, err
	}

	// schemas will always have length == 1 here
	return schemas[0], nil
}

func (db *db) getSchemas(
	ctx context.Context,
	txn datastore.Txn,
	options client.SchemaFetchOptions,
) ([]client.SchemaDescription, error) {
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
