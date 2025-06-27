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

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/db/description"
	"github.com/sourcenetwork/defradb/internal/db/txnctx"
	"github.com/sourcenetwork/defradb/internal/keys"
)

func (db *DB) setMigration(ctx context.Context, cfg client.LensConfig) error {
	dstFound := true
	dstCol, err := description.GetCollectionByID(ctx, cfg.DestinationSchemaVersionID)
	if err != nil {
		if errors.Is(err, corekv.ErrNotFound) {
			dstFound = false
		} else {
			return err
		}
	}

	srcFound := true
	sourceCol, err := description.GetCollectionByID(ctx, cfg.SourceSchemaVersionID)
	if err != nil {
		if errors.Is(err, corekv.ErrNotFound) {
			srcFound = false
		} else {
			return err
		}
	}

	if !srcFound {
		desc := client.CollectionVersion{
			VersionID:      cfg.SourceSchemaVersionID,
			CollectionID:   client.OrphanCollectionID,
			IsMaterialized: true,
		}

		err = description.SaveCollection(ctx, desc)
		if err != nil {
			return err
		}

		sourceCol = desc
	}

	isDstCollectionFound := false
	if dstFound {
		if len(dstCol.Sources) == 0 {
			// If the destingation collection has no sources at all, it must have been added as an orphaned source
			// by another migration.  This can happen if the migrations are added in an unusual order, before
			// their schemas have been defined locally.
			dstCol.Sources = append(dstCol.Sources, &client.CollectionSource{
				SourceCollectionID: sourceCol.VersionID,
			})
		}

		for _, source := range dstCol.CollectionSources() {
			if source.SourceCollectionID == sourceCol.VersionID {
				isDstCollectionFound = true
				break
			}
		}
	}

	if !isDstCollectionFound {
		dstCol = client.CollectionVersion{
			Name:           sourceCol.Name,
			VersionID:      cfg.DestinationSchemaVersionID,
			IsMaterialized: true,
			CollectionID:   sourceCol.CollectionID,
			Sources: []any{
				&client.CollectionSource{
					SourceCollectionID: sourceCol.VersionID,
					// The transform will be set later, when updating all destination collections
					// whether they are newly created or not.
				},
			},
		}

		err = description.SaveCollection(ctx, dstCol)
		if err != nil {
			return err
		}

		if dstCol.CollectionID != "" { // todo- this makes no sense
			var schemaFound bool
			// If the root schema id is known, we need to add it to the index, even if the schema is not known locally
			schema, err := description.GetSchemaVersion(ctx, cfg.SourceSchemaVersionID)
			if err != nil {
				if !errors.Is(err, corekv.ErrNotFound) {
					return err
				}
			} else {
				schemaFound = true
			}

			if schemaFound {
				txn := txnctx.MustGet(ctx)
				schemaRootKey := keys.NewSchemaRootKey(schema.Root, cfg.DestinationSchemaVersionID)
				err = txn.Systemstore().Set(ctx, schemaRootKey.Bytes(), []byte{})
				if err != nil {
					return err
				}

				dstCol.CollectionID = schema.Root

				err = description.SaveCollection(ctx, dstCol)
				if err != nil {
					return err
				}
			}
		}
	}

	collectionSources := dstCol.CollectionSources()
	for _, source := range collectionSources {
		// WARNING: Here we assume that the collection source points at a collection of the source schema version.
		// This works currently, as collections only have a single source.  If/when this changes we need to make
		// sure we only update the correct source.

		source.Transform = immutable.Some(cfg.Lens)

		err = db.LensRegistry().SetMigration(ctx, dstCol.VersionID, cfg.Lens)
		if err != nil {
			return err
		}
	}

	err = description.SaveCollection(ctx, dstCol)
	if err != nil {
		return err
	}

	return nil
}
