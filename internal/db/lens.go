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
		sourceCol = client.CollectionVersion{
			VersionID:      cfg.SourceSchemaVersionID,
			CollectionID:   client.OrphanCollectionID,
			IsMaterialized: true,
			IsPlaceholder:  true,
		}

		err = description.SaveCollection(ctx, sourceCol)
		if err != nil {
			return err
		}
	}

	if !dstFound {
		dstCol = client.CollectionVersion{
			Name:           sourceCol.Name,
			VersionID:      cfg.DestinationSchemaVersionID,
			IsMaterialized: true,
			IsPlaceholder:  true,
			CollectionID:   sourceCol.CollectionID,
		}
	}

	if dstCol.PreviousVersion.HasValue() && dstCol.PreviousVersion.Value().SourceCollectionID != sourceCol.VersionID {
		return NewErrMigrationBetweenNonAdjacentVersions(cfg.SourceSchemaVersionID, cfg.DestinationSchemaVersionID)
	}

	dstCol.PreviousVersion = immutable.Some(client.CollectionSource{
		SourceCollectionID: sourceCol.VersionID,
		Transform:          immutable.Some(cfg.Lens),
	})

	err = description.SaveCollection(ctx, dstCol)
	if err != nil {
		return err
	}

	return db.LensRegistry().SetMigration(ctx, dstCol.VersionID, cfg.Lens)
}
