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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
)

const (
	schemaNamePathIndex int = 0
	schemaPathIndex     int = 1
	fieldsPathIndex     int = 2
	fieldIndexPathIndex int = 3
)

// addSchema takes the provided schema in SDL format, and applies it to the database,
// and creates the necessary collections, request types, etc.
func (db *db) addSchema(
	ctx context.Context,
	txn datastore.Txn,
	schemaString string,
) ([]client.CollectionDescription, error) {
	existingCollections, err := db.getAllCollections(ctx, txn)
	if err != nil {
		return nil, err
	}

	existingDefinitions := make([]client.CollectionDefinition, len(existingCollections))
	for i := range existingCollections {
		existingDefinitions[i] = existingCollections[i]
	}

	newDefinitions, err := db.parser.ParseSDL(ctx, schemaString)
	if err != nil {
		return nil, err
	}

	err = db.parser.SetSchema(ctx, txn, append(existingDefinitions, newDefinitions...))
	if err != nil {
		return nil, err
	}

	returnDescriptions := make([]client.CollectionDescription, len(newDefinitions))
	for i, definition := range newDefinitions {
		col, err := db.createCollection(ctx, txn, definition.Description(), definition.Schema())
		if err != nil {
			return nil, err
		}
		returnDescriptions[i] = col.Description()
	}

	return returnDescriptions, nil
}

func (db *db) loadSchema(ctx context.Context, txn datastore.Txn) error {
	collections, err := db.getAllCollections(ctx, txn)
	if err != nil {
		return err
	}

	definitions := make([]client.CollectionDefinition, len(collections))
	for i := range collections {
		definitions[i] = collections[i]
	}

	return db.parser.SetSchema(ctx, txn, definitions)
}

func (db *db) getCollectionDescriptions(
	ctx context.Context,
	txn datastore.Txn,
) ([]client.CollectionDescription, error) {
	collections, err := db.getAllCollections(ctx, txn)
	if err != nil {
		return nil, err
	}

	descriptions := make([]client.CollectionDescription, len(collections))
	for i, collection := range collections {
		descriptions[i] = collection.Description()
	}

	return descriptions, nil
}

// patchSchema takes the given JSON patch string and applies it to the set of CollectionDescriptions
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
func (db *db) patchSchema(ctx context.Context, txn datastore.Txn, patchString string, setAsDefaultVersion bool) error {
	patch, err := jsonpatch.DecodePatch([]byte(patchString))
	if err != nil {
		return err
	}

	collectionsByName, err := db.getCollectionsByName(ctx, txn)
	if err != nil {
		return err
	}

	// Here we swap out any string representations of enums for their integer values
	patch, err = substituteSchemaPatch(patch, collectionsByName)
	if err != nil {
		return err
	}

	existingDescriptionJson, err := json.Marshal(collectionsByName)
	if err != nil {
		return err
	}

	newDescriptionJson, err := patch.Apply(existingDescriptionJson)
	if err != nil {
		return err
	}

	var newDescriptionsByName map[string]client.CollectionDescription
	decoder := json.NewDecoder(strings.NewReader(string(newDescriptionJson)))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&newDescriptionsByName)
	if err != nil {
		return err
	}

	newCollections := []client.CollectionDefinition{}
	newSchemaByName := map[string]client.SchemaDescription{}
	for _, desc := range newDescriptionsByName {
		col, err := db.newCollection(desc)
		if err != nil {
			return err
		}

		newCollections = append(newCollections, col)
		newSchemaByName[col.schema.Name] = col.schema
	}

	for i, col := range newCollections {
		col, err := db.updateCollection(ctx, txn, collectionsByName, newSchemaByName, col.Description(), col.Schema(), setAsDefaultVersion)
		if err != nil {
			return err
		}

		newCollections[i] = col
	}

	return db.parser.SetSchema(ctx, txn, newCollections)
}

func (db *db) getCollectionsByName(
	ctx context.Context,
	txn datastore.Txn,
) (map[string]client.CollectionDescription, error) {
	collections, err := db.getAllCollections(ctx, txn)
	if err != nil {
		return nil, err
	}

	collectionsByName := map[string]client.CollectionDescription{}
	for _, collection := range collections {
		collectionsByName[collection.Name()] = collection.Description()
	}

	return collectionsByName, nil
}

// substituteSchemaPatch handles any substitution of values that may be required before
// the patch can be applied.
//
// For example Field [FieldKind] string representations will be replaced by the raw integer
// value.
func substituteSchemaPatch(
	patch jsonpatch.Patch,
	collectionsByName map[string]client.CollectionDescription,
) (jsonpatch.Patch, error) {
	fieldIndexesByCollection := make(map[string]map[string]int, len(collectionsByName))
	for colName, col := range collectionsByName {
		fieldIndexesByName := make(map[string]int, len(col.Schema.Fields))
		fieldIndexesByCollection[colName] = fieldIndexesByName
		for i, field := range col.Schema.Fields {
			fieldIndexesByName[field.Name] = i
		}
	}

	for _, patchOperation := range patch {
		path, err := patchOperation.Path()
		if err != nil {
			return nil, err
		}

		path = strings.TrimPrefix(path, "/")
		splitPath := strings.Split(path, "/")

		if value, hasValue := patchOperation["value"]; hasValue {
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

					desc := collectionsByName[splitPath[schemaNamePathIndex]]
					var index string
					if fieldIndexesByName, ok := fieldIndexesByCollection[desc.Name]; ok {
						if i, ok := fieldIndexesByName[fieldIndexer]; ok {
							index = fmt.Sprint(i)
						}
					}
					if index == "" {
						index = "-"
						// If this is a new field we need to track its location so that subsequent operations
						// within the patch may access it by field name.
						fieldIndexesByCollection[desc.Name][fieldIndexer] = len(fieldIndexesByCollection[desc.Name])
					}

					splitPath[fieldIndexPathIndex] = index
					path = strings.Join(splitPath, "/")
					opPath := json.RawMessage([]byte(fmt.Sprintf(`"/%s"`, path)))
					patchOperation["path"] = &opPath
				}
			}

			if isField {
				if kind, isString := field["Kind"].(string); isString {
					substitute, collectionName, err := getSubstituteFieldKind(kind, collectionsByName)
					if err != nil {
						return nil, err
					}

					field["Kind"] = substitute
					if collectionName != "" {
						if field["Schema"] != nil && field["Schema"] != collectionName {
							return nil, NewErrFieldKindDoesNotMatchFieldSchema(kind, field["Schema"].(string))
						}
						field["Schema"] = collectionName
					}

					newPatchValue = immutable.Some[any](field)
				}
			} else if isFieldKind(splitPath) {
				var kind any
				err = json.Unmarshal(*value, &kind)
				if err != nil {
					return nil, err
				}

				if kind, isString := kind.(string); isString {
					substitute, _, err := getSubstituteFieldKind(kind, collectionsByName)
					if err != nil {
						return nil, err
					}

					newPatchValue = immutable.Some[any](substitute)
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

// getSubstituteFieldKind checks and attempts to get the underlying integer value for the given string
// Field Kind value. It will return the value if one is found, else returns an [ErrFieldKindNotFound].
//
// If the value represents a foreign relation the collection name will also be returned.
func getSubstituteFieldKind(
	kind string,
	collectionsByName map[string]client.CollectionDescription,
) (client.FieldKind, string, error) {
	substitute, substituteFound := client.FieldKindStringToEnumMapping[kind]
	if substituteFound {
		return substitute, "", nil
	} else {
		var collectionName string
		var substitute client.FieldKind
		if len(kind) > 0 && kind[0] == '[' && kind[len(kind)-1] == ']' {
			collectionName = kind[1 : len(kind)-1]
			substitute = client.FieldKind_FOREIGN_OBJECT_ARRAY
		} else {
			collectionName = kind
			substitute = client.FieldKind_FOREIGN_OBJECT
		}

		if _, substituteFound := collectionsByName[collectionName]; substituteFound {
			return substitute, collectionName, nil
		}

		return 0, "", NewErrFieldKindNotFound(kind)
	}
}

// isFieldOrInner returns true if the given path points to a FieldDescription or a property within it.
func isFieldOrInner(path []string) bool {
	//nolint:goconst
	return len(path) >= 4 && path[fieldsPathIndex] == "Fields" && path[schemaPathIndex] == "Schema"
}

// isField returns true if the given path points to a FieldDescription.
func isField(path []string) bool {
	return len(path) == 4 && path[fieldsPathIndex] == "Fields" && path[schemaPathIndex] == "Schema"
}

// isField returns true if the given path points to a FieldDescription.Kind property.
func isFieldKind(path []string) bool {
	return len(path) == 5 &&
		path[fieldIndexPathIndex+1] == "Kind" &&
		path[fieldsPathIndex] == "Fields" &&
		path[schemaPathIndex] == "Schema"
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
