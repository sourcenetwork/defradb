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
	"maps"
	"strings"
	"unicode"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/ipfs/go-cid"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"slices"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/description"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

func (db *DB) addCollections(
	ctx context.Context,
	parseResults []core.Collection,
) ([]client.CollectionVersion, error) {
	returnDescriptions := make([]client.CollectionVersion, 0, len(parseResults))

	existingVersions, err := description.GetActiveCollections(ctx)
	if err != nil {
		return nil, err
	}

	err = finalizeRelations(parseResults, existingVersions)
	if err != nil {
		return nil, err
	}

	newCollections := make([]client.CollectionVersion, len(parseResults))
	for i, def := range parseResults {
		newCollections[i] = def.Definition
	}

	err = setCollectionIDs(ctx, newCollections, existingVersions)
	if err != nil {
		return nil, err
	}

	for i := range parseResults {
		// The secondary index code requires the useage of core.Collection which means we need to
		// map the CollectionVersion back on to the input param.
		parseResults[i].Definition = newCollections[i]
	}

	err = db.validateNewCollection(
		ctx,
		slices.Concat(newCollections, existingVersions),
		existingVersions,
	)
	if err != nil {
		return nil, err
	}

	for _, def := range parseResults {
		def.Definition.Indexes = make([]client.IndexDescription, 0, len(def.NewIndexes))
		for _, newIndex := range def.NewIndexes {
			desc, err := processNewIndexRequest(ctx, def.Definition, newIndex)
			if err != nil {
				return nil, err
			}
			def.Definition.Indexes = append(def.Definition.Indexes, desc)
		}

		err = description.SaveCollection(ctx, def.Definition)
		if err != nil {
			return nil, err
		}

		txnOpt := datastore.CtxTryGetClientTxnOption(ctx)

		col, err := db.newCollection(def.Definition, txnOpt)
		if err != nil {
			return nil, err
		}

		for _, index := range def.Definition.Indexes {
			if _, err := col.appendNewIndexAndIndexExistingDocs(ctx, index); err != nil {
				return nil, err
			}
		}

		result, err := description.GetCollectionByID(ctx, def.Definition.VersionID)
		if err != nil {
			return nil, err
		}

		returnDescriptions = append(returnDescriptions, result)
	}

	return returnDescriptions, nil
}

// patchCollection takes the given JSON patch string and applies it to the set of CollectionVersions
// present in the database.
//
// It will also update the GQL types used by the query system. It will error and not apply any of the
// requested, valid updates should the net result of the patch result in an invalid state.  The
// individual operations defined in the patch do not need to result in a valid state, only the net result
// of the full patch.
//
// New CollectionVersions added by modifying the global type definition (e.g. renaming, adding fields, etc)
// will automatically become the active version of the Collection, unless `IsActive` is set to false by the patch.
//
// Field [FieldKind] values may be provided in either their raw integer form, or as string as per
// [FieldKindStringToEnumMapping].
//
// CollectionVersions may be referenced by their VersionID, or their Name.  Referencing by name will patch
// the current active version, whereas referencing by VersionID will patch that specific version, whether it is
// currently active or not.
//
// A lens configuration may also be provided, and will become the migration to any new CollectionVersions added
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
	existingCols, err := description.GetCollections(ctx)
	if err != nil {
		return err
	}

	existingColsByName := map[string]client.CollectionVersion{}
	existingColsByID := map[string]client.CollectionVersion{}
	for _, col := range existingCols {
		if col.IsActive {
			existingColsByName[col.Name] = col
		}
		existingColsByID[col.VersionID] = col
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

	removedCollectionVersions := []client.CollectionVersion{}
existingVersionLoop:
	for versionID, version := range existingColsByID {
		if _, ok := newColsByID[versionID]; !ok {
			for _, newCol := range newColsByID {
				if newCol.VersionID == versionID {
					// If the missing version id is found in another location, we do not wish to delete the collection,
					// it has essentially been moved by the JSON patch for reasons known only to the user.
					continue existingVersionLoop
				}
			}
			removedCollectionVersions = append(removedCollectionVersions, version)
		}
	}

	for _, col := range newColsByID {
		// Automatically add any id fields for object fields added by the patch, if the patch did not explicitly
		// add one.
		for _, field := range col.Fields {
			if field.Kind.IsObject() && !field.Kind.IsArray() {
				idFieldName := request.ToFieldID(field.Name)
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

	oneToOneIndexRequests, err := getOneToOneIndexRequestsForPatch(newColsByID, existingColsByName)
	if err != nil {
		return err
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

	err = setCollectionIDs(ctx, newCollections, existingCols)
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

	// Track collections that were upgraded from placeholders and may need reindexing
	var placeholderReplacers []client.CollectionVersion

	for i := 0; i < len(newCollections); i++ {
		placeholder := newCollections[i]
		if placeholder.IsPlaceholder {
			isFound := false
			for j, col := range newCollections {
				if col.VersionID == placeholder.VersionID && !col.IsPlaceholder {
					newCollections[j].PreviousVersion = placeholder.PreviousVersion
					// Track this collection as it may have a migration that needs to be applied
					if col.IsActive {
						placeholderReplacers = append(placeholderReplacers, newCollections[j])
					}
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

	err = db.validateCollectionChanges(ctx, existingCols, newCollections)
	if err != nil {
		return err
	}

	err = db.deleteCollectionVersions(ctx, removedCollectionVersions)
	if err != nil {
		return err
	}

	for _, col := range newCollections {
		isDeleted := false
		for _, removedCol := range removedCollectionVersions {
			if col.VersionID == removedCol.VersionID {
				isDeleted = true
				break
			}
		}
		if isDeleted {
			// We need to make sure we dont save any collections that we have just deleted.
			// This check is needed due to the unfortunate way mutated collections have their
			// originals re-added to `newCollections` on line 260.
			//
			// This re-adding, and this check, are planned to be removed post v1 in issue:
			// https://github.com/sourcenetwork/defradb/issues/4197
			continue
		}

		existingCol, colExists := existingColsByID[col.VersionID]
		if colExists && col.Equal(existingCol) {
			continue
		}

		err := description.SaveCollection(ctx, col)
		if err != nil {
			return err
		}

		if col.IsActive {
			if indexReqs, hasReqs := oneToOneIndexRequests[col.Name]; hasReqs {
				txnOpt := datastore.CtxTryGetClientTxnOption(ctx)
				colObj, err := db.newCollection(col, txnOpt)
				if err != nil {
					return err
				}
				for _, indexReq := range indexReqs {
					if _, err := colObj.newIndex(ctx, indexReq); err != nil {
						return err
					}
				}
				col = colObj.Version()
			}
		}

		if colExists {
			if existingCol.IsMaterialized && !col.IsMaterialized {
				// If the collection is being de-materialized - delete any cached values.
				// Leaving them around will not break anything, but it would be a waste of
				// storage space.
				err := db.clearViewCache(ctx, col)
				if err != nil {
					return err
				}
			}
		} else if col.PreviousVersion.HasValue() && migration.HasValue() {
			_, err = db.setMigration(ctx, client.LensConfig{
				SourceCollectionVersionID:      col.PreviousVersion.Value().SourceCollectionID,
				DestinationCollectionVersionID: col.VersionID,
				Lens:                           migration.Value(),
			})
			if err != nil {
				return err
			}
		}
	}

	// Reindex any collections that were upgraded from placeholders with migrations
	for _, col := range placeholderReplacers {
		if col.PreviousVersion.HasValue() && col.PreviousVersion.Value().Transform.HasValue() {
			err = db.reindexNewActiveVersion(ctx, col)
			if err != nil {
				return err
			}
		}
	}

	return db.loadCollectionDefinitions(ctx)
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
	fieldIndexesByCollection := make(map[string]map[string]int, len(collectionsByName))
	for collectionName, collection := range collectionsByName {
		fieldIndexesByName := make(map[string]int, len(collection.Fields))
		fieldIndexesByCollection[collectionName] = fieldIndexesByName
		for i, field := range collection.Fields {
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

// isFieldOrInner returns true if the given path points to a CollectionFieldDescription or a property within it.
func isFieldOrInner(path []string) bool {
	//nolint:goconst
	return len(path) >= 3 && path[fieldsPathIndex] == "Fields"
}

// isField returns true if the given path points to a CollectionFieldDescription.
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

// SetActiveCollectionVersion activates all collection versions with the given collection version, and deactivates all
// those without it (if they share the same collection root).
//
// This will affect all operations interacting with the collection where a collection version is not explicitly
// provided.  This includes GQL queries and Collection operations.
//
// It will return an error if the provided collection version ID does not exist.
func (db *DB) setActiveCollectionVersion(
	ctx context.Context,
	versionID string,
) error {
	if versionID == "" {
		return ErrCollectionVersionIDEmpty
	}
	col, err := description.GetCollectionByID(ctx, versionID)
	if err != nil {
		return err
	}

	colsWithRoot, err := description.GetCollectionsByCollectionID(ctx, col.CollectionID)
	if err != nil {
		return err
	}

	// The optional collection is used to track if there was a switch to another version.
	newActiveCol := immutable.None[client.CollectionVersion]()

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

			newActiveCol = immutable.Some(col)

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

	if newActiveCol.HasValue() {
		shouldReindex, err := db.shouldReindexForVersionSwitch(ctx, newActiveCol.Value())
		if err != nil {
			return err
		}

		if shouldReindex {
			err = db.reindexNewActiveVersion(ctx, newActiveCol.Value())
			if err != nil {
				return err
			}
		}
	}

	// Load the collection definitions into the clients (e.g. GQL)
	return db.loadCollectionDefinitions(ctx)
}

// shouldReindexForVersionSwitch determines if reindexing is needed when switching
// to a new active version by examining the full version history DAG.
//
// This properly handles branching version histories by checking if any version
// reachable from the new active version has a migration.
func (db *DB) shouldReindexForVersionSwitch(
	ctx context.Context,
	newActiveCol client.CollectionVersion,
) (bool, error) {
	return description.HasMigrations(ctx, newActiveCol.CollectionID, newActiveCol.VersionID)
}

func (db *DB) deleteCollectionVersions(
	ctx context.Context,
	versions []client.CollectionVersion,
) error {
	versionsByVersionID := make(map[string]client.CollectionVersion, len(versions))
	for _, version := range versions {
		versionsByVersionID[version.VersionID] = version
	}

	// Order the versions to delete so that parents get deleted before their children.
	// This allows us to verify that a continuous history is always maintained.
	orderedVersions := make([]client.CollectionVersion, 0, len(versions))
	for len(orderedVersions) != len(versions) {
		for _, versionToAdd := range versionsByVersionID {
			hasParent := false
			for _, possibleParent := range versionsByVersionID {
				if possibleParent.PreviousVersion.HasValue() &&
					possibleParent.PreviousVersion.Value().SourceCollectionID == versionToAdd.VersionID {
					hasParent = true
					break
				}
			}

			if !hasParent {
				orderedVersions = append(orderedVersions, versionToAdd)
				delete(versionsByVersionID, versionToAdd.VersionID)
			}
		}
	}

	for _, version := range orderedVersions {
		err := db.deleteCollectionVersion(ctx, version)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) deleteCollectionVersion(
	ctx context.Context,
	version client.CollectionVersion,
) error {
	hasDocs, err := collectionHasDocuments(ctx, version)
	if err != nil {
		return err
	}
	if hasDocs {
		// If the collection contains any documents, we do not allow deletion of any version in the
		// collection - they must first delete the documents locally, and then delete the collection.
		//
		// This is thought to be much safer than allowing document deletion along with the collection.
		return NewErrCannotDeleteCollectionWithDocs(version.Name, version.VersionID)
	}

	err = validateCollectionDoesNotHaveHigherVersion(ctx, version)
	if err != nil {
		return err
	}

	err = description.DeleteCollection(ctx, db.lockSet, version)
	if err != nil {
		return err
	}

	err = deleteCollectionBlocks(ctx, version)
	if err != nil {
		return err
	}

	return nil
}

func collectionHasDocuments(
	ctx context.Context,
	version client.CollectionVersion,
) (bool, error) {
	if !version.IsMaterialized {
		// Assume that if the collection *was* materialized, and is no longer materialized, that the cached
		// state was properly disposed of (it should have been).
		return false, nil
	}

	txn := datastore.CtxMustGetTxn(ctx)

	shortID, err := id.GetShortCollectionID(ctx, version.CollectionID)
	if err != nil {
		return false, err
	}

	var prefixKey keys.Key
	if version.Query.HasValue() {
		prefixKey = keys.NewViewCacheColPrefix(shortID)
	} else {
		prefixKey = keys.PrimaryDataStoreKey{
			CollectionShortID: shortID,
		}
	}

	iter, err := txn.Datastore().Iterator(ctx, datastore.IterOptions{
		Prefix:   prefixKey.ToDS(),
		KeysOnly: true,
	})
	if err != nil {
		return false, errors.Join(err, iter.Close())
	}

	hasValue, err := iter.Next()
	if err != nil {
		return false, errors.Join(err, iter.Close())
	}

	return hasValue, iter.Close()
}

func validateCollectionDoesNotHaveHigherVersion(
	ctx context.Context,
	version client.CollectionVersion,
) error {
	allVersions, err := description.GetCollectionsByCollectionID(ctx, version.CollectionID)
	if err != nil {
		return err
	}

	for _, newVersion := range allVersions {
		if newVersion.PreviousVersion.HasValue() &&
			newVersion.PreviousVersion.Value().SourceCollectionID == version.VersionID {
			// We do not allow the deletion of versions that are not the head of their branch - this would
			// add a gap in the history, potentially causing problems that we do not wish to test for or
			// handle right now.
			return NewErrCannotDeleteOldVersion(version.VersionID, newVersion.VersionID)
		}
	}

	return nil
}

func deleteCollectionBlocks(
	ctx context.Context,
	version client.CollectionVersion,
) error {
	txn := datastore.CtxMustGetTxn(ctx)

	colCid, err := cid.Parse(version.VersionID)
	if err != nil {
		return err
	}

	err = txn.Blockstore().DeleteBlock(ctx, colCid)
	if err != nil {
		return err
	}

	for _, field := range version.Fields {
		if field.FieldID == "" {
			// Only fields with field IDs have backing blocks
			continue
		}

		fieldCid, err := cid.Parse(field.FieldID)
		if err != nil {
			return err
		}

		err = txn.Blockstore().DeleteBlock(ctx, fieldCid)
		if err != nil {
			return err
		}
	}

	return nil
}

// finalizeRelations determines which side of a relation is primary and sets IsPrimary=true
// on both the relation field and its corresponding _<fieldName>ID field.
//
// A relation field is marked as primary if:
// - The target collection has no corresponding field pointing back (one-sided relation), OR
// - The corresponding field in the target collection is an array (one-to-many relation)
//
// This function handles both within-batch relations (new collections referencing each other)
// and cross-batch relations (new collections referencing existing collections).
//
// Note on automatic IsPrimary assignment: When a new collection defines a relation to an
// existing collection that has no back-reference, the new collection's field MUST be primary.
// The existing collection cannot be modified to become primary, and a relation requires exactly
// one primary side to store the foreign key.
//
// For one-to-one relations, this function also ensures a unique index exists on the primary
// side's _<fieldName>ID field to enforce the 1-to-1 constraint efficiently.
func finalizeRelations(
	newCollections []core.Collection,
	existingCollections []client.CollectionVersion,
) error {
	existingByName := make(map[string]client.CollectionVersion)
	for _, col := range existingCollections {
		existingByName[col.Name] = col
	}

	newByName := make(map[string]client.CollectionVersion)
	for _, col := range newCollections {
		newByName[col.Definition.Name] = col.Definition
	}

	for i, newCol := range newCollections {
		if newCol.Definition.IsEmbeddedOnly {
			continue
		}

		for fieldIndex, field := range newCol.Definition.Fields {
			namedKind, ok := field.Kind.(*client.NamedKind)
			if !ok || namedKind.IsArray() {
				// We only need to process the primary side of a relation here.
				// If the field is not a relation or if it is an array, we can skip it.
				continue
			}

			var targetCol client.CollectionVersion
			var found bool

			if col, inBatch := newByName[namedKind.Name]; inBatch {
				targetCol = col
				found = true
			} else if col, exists := existingByName[namedKind.Name]; exists {
				targetCol = col
				found = true
			}

			if !found {
				// The target collection doesn't exist. Validation will catch this later.
				continue
			}

			isOneToOne := isOneToOneRelation(targetCol, field.RelationName.Value(), newCol.Definition.Name, field.Name)

			if !field.IsPrimary && !isOneToOne {
				newCollections[i].Definition.Fields[fieldIndex].IsPrimary = true

				idFieldName := request.ToFieldID(field.Name)
				for j, f := range newCollections[i].Definition.Fields {
					if f.Name == idFieldName {
						newCollections[i].Definition.Fields[j].IsPrimary = true
						break
					}
				}
			}

			if isOneToOne && field.IsPrimary {
				newIndex, err := ensureOneToOneUniqueIndex(
					newCollections[i].NewIndexes,
					nil,
					newCol.Definition.Name,
					field.Name,
				)
				if err != nil {
					return err
				}
				if newIndex != nil {
					newCollections[i].NewIndexes = append(newCollections[i].NewIndexes, *newIndex)
				}
			}
		}
	}

	return nil
}

// isOneToOneRelation checks if a relation field represents a one-to-one relationship.
// A relation is one-to-one if the target collection has a corresponding non-array field
// pointing back to the host collection.
func isOneToOneRelation(
	targetCol client.CollectionVersion,
	relationName string,
	hostColName string,
	fieldName string,
) bool {
	correspondingField, hasCorrespondingField := targetCol.GetFieldByRelation(
		relationName,
		hostColName,
		fieldName,
	)
	return hasCorrespondingField && !correspondingField.Kind.IsArray()
}

// findIndexWithFirstField checks if an index exists where the given field is the first field.
func findIndexWithFirstField(
	newIndexes []client.NewIndexRequest,
	existingIndexes []client.IndexDescription,
	fieldName string,
) (isUnique bool, found bool) {
	for _, index := range newIndexes {
		if len(index.Fields) > 0 && index.Fields[0].Name == fieldName {
			return index.Unique, true
		}
	}
	for _, index := range existingIndexes {
		if len(index.Fields) > 0 && index.Fields[0].Name == fieldName {
			return index.Unique, true
		}
	}
	return false, false
}

// ensureOneToOneUniqueIndex ensures a unique index exists for a one-to-one relation's _id field.
// If a user-defined index exists with the relation field as the first field, it validates that it's unique.
// If no user-defined index exists, it adds one automatically.
func ensureOneToOneUniqueIndex(
	newIndexes []client.NewIndexRequest,
	existingIndexes []client.IndexDescription,
	collectionName string,
	relationFieldName string,
) (newIndex *client.NewIndexRequest, err error) {
	idFieldName := request.ToFieldID(relationFieldName)

	// Check for user-defined index on either the _id field or the relation field name
	// (e.g., "_addressID" or "address" since @index on relation field uses field name)
	isUnique, hasIndex := findIndexWithFirstField(newIndexes, existingIndexes, idFieldName)
	if !hasIndex {
		isUnique, hasIndex = findIndexWithFirstField(newIndexes, existingIndexes, relationFieldName)
	}

	if hasIndex {
		if !isUnique {
			return nil, NewErrOneToOneRelationMustBeUnique(collectionName, relationFieldName)
		}
		return nil, nil
	}

	// No user-defined index exists, add one automatically
	return &client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: idFieldName}},
		Unique: true,
	}, nil
}

// getOneToOneIndexRequestsForPatch returns index add requests for one-to-one relations
// added via collection patch. This is needed because patches don't go through the
// standard collection creation flow that calls finalizeRelations.
// Returns a map of collectionName -> []NewIndexRequest for indexes that need to be added.
func getOneToOneIndexRequestsForPatch(
	newColsByID map[string]client.CollectionVersion,
	existingColsByName map[string]client.CollectionVersion,
) (map[string][]client.NewIndexRequest, error) {
	allColsByName := make(map[string]client.CollectionVersion)
	maps.Copy(allColsByName, existingColsByName)

	for _, col := range newColsByID {
		allColsByName[col.Name] = col
	}

	result := make(map[string][]client.NewIndexRequest)

	for _, col := range newColsByID {
		existingCol := existingColsByName[col.Name]
		existingFieldNames := make(map[string]struct{}, len(existingCol.Fields))
		for _, field := range existingCol.Fields {
			existingFieldNames[field.Name] = struct{}{}
		}

		for _, field := range col.Fields {
			if _, existed := existingFieldNames[field.Name]; existed {
				continue
			}

			namedKind, ok := field.Kind.(*client.NamedKind)
			if !ok || namedKind.IsArray() {
				continue
			}

			if !field.IsPrimary {
				continue
			}

			targetCol, found := allColsByName[namedKind.Name]
			if !found {
				continue
			}

			if !isOneToOneRelation(targetCol, field.RelationName.Value(), col.Name, field.Name) {
				continue
			}

			indexReq, err := ensureOneToOneUniqueIndex(nil, col.Indexes, col.Name, field.Name)
			if err != nil {
				return nil, err
			}

			if indexReq != nil {
				result[col.Name] = append(result[col.Name], *indexReq)
			}
		}
	}

	return result, nil
}
