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
	"strings"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/description"
	"github.com/sourcenetwork/defradb/internal/db/id"
)

func (db *DB) createCollections(
	ctx context.Context,
	newDefinitions []client.CollectionDefinition,
) ([]client.CollectionDefinition, error) {
	returnDescriptions := make([]client.CollectionDefinition, 0, len(newDefinitions))

	existingDefinitions, err := db.getAllActiveDefinitions(ctx)
	if err != nil {
		return nil, err
	}

	newSchemas := make([]client.SchemaDescription, len(newDefinitions))
	for i, def := range newDefinitions {
		newSchemas[i] = def.Schema
	}

	err = setSchemaIDs(newSchemas)
	if err != nil {
		return nil, err
	}

	for i := range newDefinitions {
		newDefinitions[i].Version.ID = newSchemas[i].VersionID
		newDefinitions[i].Version.CollectionID = newSchemas[i].Root
		newDefinitions[i].Schema = newSchemas[i]
	}

	txn := mustGetContextTxn(ctx)
	err = id.SetFieldIDs(ctx, txn, newDefinitions)
	if err != nil {
		return nil, err
	}

	err = db.validateNewCollection(
		ctx,
		append(
			append(
				[]client.CollectionDefinition{},
				newDefinitions...,
			),
			existingDefinitions...,
		),
		existingDefinitions,
	)
	if err != nil {
		return nil, err
	}

	for _, def := range newDefinitions {
		_, err := description.CreateSchemaVersion(ctx, txn, def.Schema)
		if err != nil {
			return nil, err
		}

		if len(def.Version.Fields) == 0 {
			// This is a schema-only definition, we should not create a collection for it
			returnDescriptions = append(returnDescriptions, def)
			continue
		}

		desc, err := description.SaveCollection(ctx, txn, def.Version)
		if err != nil {
			return nil, err
		}

		col := db.newCollection(desc, def.Schema)

		for _, index := range desc.Indexes {
			descWithoutID := client.IndexDescriptionCreateRequest{
				Name:   index.Name,
				Fields: index.Fields,
				Unique: index.Unique,
			}
			if _, err := col.createIndex(ctx, descWithoutID); err != nil {
				return nil, err
			}
		}

		result, err := db.getCollectionByID(ctx, desc.ID)
		if err != nil {
			return nil, err
		}

		returnDescriptions = append(returnDescriptions, result.Definition())
	}

	return returnDescriptions, nil
}

func (db *DB) patchCollection(
	ctx context.Context,
	patchString string,
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

	existingColsByID := map[string]client.CollectionVersion{}
	existingDefinitions := make([]client.CollectionDefinition, len(existingCols))
	for _, col := range existingCols {
		existingColsByID[col.Version().ID] = col.Version()
		existingDefinitions = append(existingDefinitions, col.Definition())
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
		newDefinitions[i].Schema = col.Schema()
		newDefinitions[i].Version = newColsByID[col.Version().ID]
		updatedColsByID[col.Version().ID] = struct{}{}
	}
	// append new cols
	for id, col := range newColsByID {
		if _, ok := updatedColsByID[id]; ok {
			continue
		}
		newDefinitions = append(newDefinitions, client.CollectionDefinition{Version: col})
	}

	err = db.validateCollectionChanges(ctx, existingDefinitions, newDefinitions)
	if err != nil {
		return err
	}

	txn := mustGetContextTxn(ctx)
	for _, col := range newColsByID {
		_, err := description.SaveCollection(ctx, txn, col)
		if err != nil {
			return err
		}

		existingCol, ok := existingColsByID[col.ID]
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
					err = db.LensRegistry().SetMigration(ctx, existingCol.ID, model.Lens{})
					if err != nil {
						return err
					}
				}
			}
			for _, src := range existingCol.QuerySources() {
				if src.Transform.HasValue() {
					err = db.LensRegistry().SetMigration(ctx, existingCol.ID, model.Lens{})
					if err != nil {
						return err
					}
				}
			}
		}

		for _, src := range col.CollectionSources() {
			if src.Transform.HasValue() {
				err = db.LensRegistry().SetMigration(ctx, col.ID, src.Transform.Value())
				if err != nil {
					return err
				}
			}
		}

		for _, src := range col.QuerySources() {
			if src.Transform.HasValue() {
				err = db.LensRegistry().SetMigration(ctx, col.ID, src.Transform.Value())
				if err != nil {
					return err
				}
			}
		}
	}

	return db.loadSchema(ctx)
}

// SetActiveSchemaVersion activates all collection versions with the given schema version, and deactivates all
// those without it (if they share the same schema root).
//
// This will affect all operations interacting with the schema where a schema version is not explicitly
// provided.  This includes GQL queries and Collection operations.
//
// It will return an error if the provided schema version ID does not exist.
func (db *DB) setActiveSchemaVersion(
	ctx context.Context,
	schemaVersionID string,
) error {
	if schemaVersionID == "" {
		return ErrSchemaVersionIDEmpty
	}
	txn := mustGetContextTxn(ctx)
	col, err := description.GetCollectionByID(ctx, txn, schemaVersionID)
	if err != nil {
		return err
	}

	schema, err := description.GetSchemaVersion(ctx, txn, schemaVersionID)
	if err != nil {
		return err
	}

	colsWithRoot, err := description.GetCollectionsBySchemaRoot(ctx, txn, schema.Root)
	if err != nil {
		return err
	}

	colsBySourceID := map[string][]client.CollectionVersion{}
	colsByID := make(map[string]client.CollectionVersion, len(colsWithRoot))
	for _, col := range colsWithRoot {
		colsByID[col.ID] = col

		sources := col.CollectionSources()
		if len(sources) > 0 {
			// For now, we assume that each collection can only have a single source.  This will likely need
			// to change later.
			slice := colsBySourceID[sources[0].SourceCollectionID]
			slice = append(slice, col)
			colsBySourceID[sources[0].SourceCollectionID] = slice
		}
	}

	if col.IsActive {
		// The collection is already active, so we can skip it and continue
		return db.loadSchema(ctx)
	}

	sources := col.CollectionSources()

	var activeCol client.CollectionVersion
	var rootCol client.CollectionVersion
	var isActiveFound bool
	if len(sources) > 0 {
		// For now, we assume that each collection can only have a single source.  This will likely need
		// to change later.
		activeCol, rootCol, isActiveFound = db.getActiveCollectionDown(ctx, colsByID, sources[0].SourceCollectionID)
	}
	if !isActiveFound {
		// We need to look both down and up for the active version - the most recent is not necessarily the active one.
		activeCol, isActiveFound = db.getActiveCollectionUp(ctx, colsBySourceID, rootCol.ID)
	}

	col.IsActive = true
	_, err = description.SaveCollection(ctx, txn, col)
	if err != nil {
		return err
	}

	if isActiveFound {
		activeCol.IsActive = false
		_, err = description.SaveCollection(ctx, txn, activeCol)
		if err != nil {
			return err
		}
	}

	// Load the schema into the clients (e.g. GQL)
	return db.loadSchema(ctx)
}

func (db *DB) getActiveCollectionDown(
	ctx context.Context,
	colsByID map[string]client.CollectionVersion,
	id string,
) (client.CollectionVersion, client.CollectionVersion, bool) {
	col, ok := colsByID[id]
	if !ok {
		return client.CollectionVersion{}, client.CollectionVersion{}, false
	}

	if col.IsActive {
		return col, client.CollectionVersion{}, true
	}

	sources := col.CollectionSources()
	if len(sources) == 0 {
		// If a collection has zero sources it is likely the initial collection version, or
		// this collection set is currently orphaned (can happen when setting migrations that
		// do not yet link all the way back to a non-orphaned set)
		return client.CollectionVersion{}, col, false
	}

	// For now, we assume that each collection can only have a single source.  This will likely need
	// to change later.
	return db.getActiveCollectionDown(ctx, colsByID, sources[0].SourceCollectionID)
}

func (db *DB) getActiveCollectionUp(
	ctx context.Context,
	colsBySourceID map[string][]client.CollectionVersion,
	id string,
) (client.CollectionVersion, bool) {
	cols, ok := colsBySourceID[id]
	if !ok {
		// We have reached the top of the set, and have not found an active collection
		return client.CollectionVersion{}, false
	}

	for _, col := range cols {
		if col.IsActive {
			return col, true
		}
		activeCol, isFound := db.getActiveCollectionUp(ctx, colsBySourceID, col.ID)
		if isFound {
			return activeCol, isFound
		}
	}

	return client.CollectionVersion{}, false
}
