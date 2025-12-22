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
	"github.com/sourcenetwork/lens/host-go/config/model"
	"github.com/sourcenetwork/lens/host-go/store"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/description"
)

func (db *DB) getLensStore(ctx context.Context) store.Store {
	txn, ok := datastore.CtxTryGetTxn(ctx)
	if ok {
		return db.lensNode.Store.WithTxn(wrappedTxn{
			Txn:          txn,
			ReaderWriter: db.rootstore,
		})
	}

	return db.lensNode.Store
}

func (db *DB) addLens(ctx context.Context, lens model.Lens) (string, error) {
	cid, err := db.getLensStore(ctx).Add(ctx, lens)
	if err != nil {
		return "", err
	}
	return cid.String(), nil
}

func (db *DB) listLenses(ctx context.Context) (map[string]model.Lens, error) {
	lenses, err := db.getLensStore(ctx).List(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]model.Lens, len(lenses))
	for cid, lens := range lenses {
		result[cid.String()] = lens
	}
	return result, nil
}

func (db *DB) setMigration(ctx context.Context, cfg client.LensConfig) (string, error) {
	dstFound := true
	dstCol, err := description.GetCollectionByID(ctx, cfg.DestinationCollectionVersionID)
	if err != nil {
		if errors.Is(err, corekv.ErrNotFound) {
			dstFound = false
		} else {
			return "", err
		}
	}

	srcFound := true
	sourceCol, err := description.GetCollectionByID(ctx, cfg.SourceCollectionVersionID)
	if err != nil {
		if errors.Is(err, corekv.ErrNotFound) {
			srcFound = false
		} else {
			return "", err
		}
	}

	if !srcFound {
		sourceCol = client.CollectionVersion{
			VersionID:      cfg.SourceCollectionVersionID,
			CollectionID:   client.OrphanCollectionID,
			IsMaterialized: true,
			IsPlaceholder:  true,
		}

		err = description.SaveCollection(ctx, sourceCol)
		if err != nil {
			return "", err
		}
	}

	if !dstFound {
		dstCol = client.CollectionVersion{
			Name:           sourceCol.Name,
			VersionID:      cfg.DestinationCollectionVersionID,
			IsMaterialized: true,
			IsPlaceholder:  true,
			CollectionID:   sourceCol.CollectionID,
		}
	}

	if dstCol.PreviousVersion.HasValue() && dstCol.PreviousVersion.Value().SourceCollectionID != sourceCol.VersionID {
		return "", NewErrMigrationBetweenNonAdjacentVersions(cfg.SourceCollectionVersionID,
			cfg.DestinationCollectionVersionID)
	}

	id, err := db.getLensStore(ctx).Add(ctx, cfg.Lens)
	if err != nil {
		return "", err
	}

	dstCol.PreviousVersion = immutable.Some(client.CollectionSource{
		SourceCollectionID: sourceCol.VersionID,
		Transform:          immutable.Some(id.String()),
	})

	err = description.SaveCollection(ctx, dstCol)
	if err != nil {
		return "", err
	}

	shouldReindex, activeCol, err := db.shouldReindexAfterMigration(ctx, dstCol)
	if err != nil {
		return "", err
	}

	if shouldReindex {
		err = db.reindexNewActiveVersion(ctx, activeCol)
		if err != nil {
			return "", err
		}
	}

	return id.String(), nil
}

// shouldReindexAfterMigration determines if reindexing is needed after adding a migration.
// Reindexing is needed if:
// 1. The destination collection is currently active, OR
// 2. The destination collection is in the history chain of any currently active collection
// Returns: (shouldReindex bool, activeCollection, error)
func (db *DB) shouldReindexAfterMigration(
	ctx context.Context,
	dstCol client.CollectionVersion,
) (bool, client.CollectionVersion, error) {
	if dstCol.IsActive {
		return true, dstCol, nil
	}

	activeCol, err := description.GetActiveCollectionByCollectionID(ctx, dstCol.CollectionID)
	if err != nil {
		if errors.Is(err, corekv.ErrNotFound) {
			return false, client.CollectionVersion{}, nil
		}
		return false, client.CollectionVersion{}, err
	}

	history, err := description.GetTargetedCollectionHistory(
		ctx,
		activeCol.CollectionID,
		activeCol.VersionID,
	)
	if err != nil {
		return false, client.CollectionVersion{}, err
	}

	if history == nil {
		return false, activeCol, nil
	}

	_, found := history[dstCol.VersionID]
	return found, activeCol, nil
}
