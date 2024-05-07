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

	ds "github.com/ipfs/go-datastore"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/db/description"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/core"
)

func (db *db) setMigration(ctx context.Context, cfg client.LensConfig) error {
	txn := mustGetContextTxn(ctx)

	dstCols, err := description.GetCollectionsBySchemaVersionID(ctx, txn, cfg.DestinationSchemaVersionID)
	if err != nil {
		return err
	}

	sourceCols, err := description.GetCollectionsBySchemaVersionID(ctx, txn, cfg.SourceSchemaVersionID)
	if err != nil {
		return err
	}

	colSeq, err := db.getSequence(ctx, core.CollectionIDSequenceKey{})
	if err != nil {
		return err
	}

	if len(sourceCols) == 0 {
		// If no collections are found with the given [SourceSchemaVersionID], this migration must be from
		// a collection/schema version that does not yet exist locally.  We must now create it.
		colID, err := colSeq.next(ctx)
		if err != nil {
			return err
		}

		desc := client.CollectionDescription{
			ID:              uint32(colID),
			RootID:          client.OrphanRootID,
			SchemaVersionID: cfg.SourceSchemaVersionID,
		}

		col, err := description.SaveCollection(ctx, txn, desc)
		if err != nil {
			return err
		}

		sourceCols = append(sourceCols, col)
	}

	for _, sourceCol := range sourceCols {
		isDstCollectionFound := false
	dstColsLoop:
		for i, dstCol := range dstCols {
			if len(dstCol.Sources) == 0 {
				// If the destingation collection has no sources at all, it must have been added as an orphaned source
				// by another migration.  This can happen if the migrations are added in an unusual order, before
				// their schemas have been defined locally.
				dstCol.Sources = append(dstCol.Sources, &client.CollectionSource{
					SourceCollectionID: sourceCol.ID,
				})
				dstCols[i] = dstCol
			}

			for _, source := range dstCol.CollectionSources() {
				if source.SourceCollectionID == sourceCol.ID {
					isDstCollectionFound = true
					break dstColsLoop
				}
			}
		}

		if !isDstCollectionFound {
			// If the destination collection was not found, we must create it.  This can happen when setting a migration
			// to a schema version that does not yet exist locally.
			colID, err := colSeq.next(ctx)
			if err != nil {
				return err
			}

			desc := client.CollectionDescription{
				ID:              uint32(colID),
				RootID:          sourceCol.RootID,
				SchemaVersionID: cfg.DestinationSchemaVersionID,
				Sources: []any{
					&client.CollectionSource{
						SourceCollectionID: sourceCol.ID,
						// The transform will be set later, when updating all destination collections
						// whether they are newly created or not.
					},
				},
			}

			col, err := description.SaveCollection(ctx, txn, desc)
			if err != nil {
				return err
			}

			if desc.RootID != client.OrphanRootID {
				var schemaFound bool
				// If the root schema id is known, we need to add it to the index, even if the schema is not known locally
				schema, err := description.GetSchemaVersion(ctx, txn, cfg.SourceSchemaVersionID)
				if err != nil {
					if !errors.Is(err, ds.ErrNotFound) {
						return err
					}
				} else {
					schemaFound = true
				}

				if schemaFound {
					schemaRootKey := core.NewSchemaRootKey(schema.Root, cfg.DestinationSchemaVersionID)
					err = txn.Systemstore().Put(ctx, schemaRootKey.ToDS(), []byte{})
					if err != nil {
						return err
					}
				}
			}

			dstCols = append(dstCols, col)
		}
	}

	for _, col := range dstCols {
		collectionSources := col.CollectionSources()

		for _, source := range collectionSources {
			// WARNING: Here we assume that the collection source points at a collection of the source schema version.
			// This works currently, as collections only have a single source.  If/when this changes we need to make
			// sure we only update the correct source.

			source.Transform = immutable.Some(cfg.Lens)

			err = db.LensRegistry().SetMigration(ctx, col.ID, cfg.Lens)
			if err != nil {
				return err
			}
		}

		_, err = description.SaveCollection(ctx, txn, col)
		if err != nil {
			return err
		}
	}

	return nil
}
