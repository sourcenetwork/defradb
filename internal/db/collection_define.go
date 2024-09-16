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
)

func (db *db) createCollections(
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
		newDefinitions[i].Description.SchemaVersionID = newSchemas[i].VersionID
		newDefinitions[i].Schema = newSchemas[i]
	}

	err = db.setCollectionIDs(ctx, newDefinitions)
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

	txn := mustGetContextTxn(ctx)

	for _, def := range newDefinitions {
		_, err := description.CreateSchemaVersion(ctx, txn, def.Schema)
		if err != nil {
			return nil, err
		}

		if len(def.Description.Fields) == 0 {
			// This is a schema-only definition, we should not create a collection for it
			returnDescriptions = append(returnDescriptions, def)
			continue
		}

		desc, err := description.SaveCollection(ctx, txn, def.Description)
		if err != nil {
			return nil, err
		}

		col := db.newCollection(desc, def.Schema)

		for _, index := range desc.Indexes {
			if _, err := col.createIndex(ctx, index); err != nil {
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

func (db *db) patchCollection(
	ctx context.Context,
	patchString string,
) error {
	patch, err := jsonpatch.DecodePatch([]byte(patchString))
	if err != nil {
		return err
	}
	txn := mustGetContextTxn(ctx)
	cols, err := description.GetCollections(ctx, txn)
	if err != nil {
		return err
	}

	existingColsByID := map[uint32]client.CollectionDescription{}
	for _, col := range cols {
		existingColsByID[col.ID] = col
	}

	existingDescriptionJson, err := json.Marshal(existingColsByID)
	if err != nil {
		return err
	}

	newDescriptionJson, err := patch.Apply(existingDescriptionJson)
	if err != nil {
		return err
	}

	var newColsByID map[uint32]client.CollectionDescription
	decoder := json.NewDecoder(strings.NewReader(string(newDescriptionJson)))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&newColsByID)
	if err != nil {
		return err
	}

	err = db.validateCollectionChanges(ctx, cols, newColsByID)
	if err != nil {
		return err
	}

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
					Description: col,
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
func (db *db) setActiveSchemaVersion(
	ctx context.Context,
	schemaVersionID string,
) error {
	if schemaVersionID == "" {
		return ErrSchemaVersionIDEmpty
	}
	txn := mustGetContextTxn(ctx)
	cols, err := description.GetCollectionsBySchemaVersionID(ctx, txn, schemaVersionID)
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

	colsBySourceID := map[uint32][]client.CollectionDescription{}
	colsByID := make(map[uint32]client.CollectionDescription, len(colsWithRoot))
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

	for _, col := range cols {
		if col.Name.HasValue() {
			// The collection is already active, so we can skip it and continue
			continue
		}
		sources := col.CollectionSources()

		var activeCol client.CollectionDescription
		var rootCol client.CollectionDescription
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

		var newName string
		if isActiveFound {
			newName = activeCol.Name.Value()
		} else {
			// If there are no active versions in the collection set, take the name of the schema to be the name of the
			// collection.
			newName = schema.Name
		}
		col.Name = immutable.Some(newName)

		_, err = description.SaveCollection(ctx, txn, col)
		if err != nil {
			return err
		}

		if isActiveFound {
			// Deactivate the currently active collection by setting its name to none.
			activeCol.Name = immutable.None[string]()
			_, err = description.SaveCollection(ctx, txn, activeCol)
			if err != nil {
				return err
			}
		}
	}

	// Load the schema into the clients (e.g. GQL)
	return db.loadSchema(ctx)
}

func (db *db) getActiveCollectionDown(
	ctx context.Context,
	colsByID map[uint32]client.CollectionDescription,
	id uint32,
) (client.CollectionDescription, client.CollectionDescription, bool) {
	col, ok := colsByID[id]
	if !ok {
		return client.CollectionDescription{}, client.CollectionDescription{}, false
	}

	if col.Name.HasValue() {
		return col, client.CollectionDescription{}, true
	}

	sources := col.CollectionSources()
	if len(sources) == 0 {
		// If a collection has zero sources it is likely the initial collection version, or
		// this collection set is currently orphaned (can happen when setting migrations that
		// do not yet link all the way back to a non-orphaned set)
		return client.CollectionDescription{}, col, false
	}

	// For now, we assume that each collection can only have a single source.  This will likely need
	// to change later.
	return db.getActiveCollectionDown(ctx, colsByID, sources[0].SourceCollectionID)
}

func (db *db) getActiveCollectionUp(
	ctx context.Context,
	colsBySourceID map[uint32][]client.CollectionDescription,
	id uint32,
) (client.CollectionDescription, bool) {
	cols, ok := colsBySourceID[id]
	if !ok {
		// We have reached the top of the set, and have not found an active collection
		return client.CollectionDescription{}, false
	}

	for _, col := range cols {
		if col.Name.HasValue() {
			return col, true
		}
		activeCol, isFound := db.getActiveCollectionUp(ctx, colsBySourceID, col.ID)
		if isFound {
			return activeCol, isFound
		}
	}

	return client.CollectionDescription{}, false
}
