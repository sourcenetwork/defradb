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

	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/immutable"

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/internal/utils"
)

// ExecRequest executes a request against the database.
func (db *DB) ExecRequest(
	ctx context.Context,
	request string, opts ...options.Enumerable[options.ExecRequestOptions],
) *client.RequestResult {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)
	if opt.Identity.HasValue() {
		ctx = identity.WithContext(ctx, opt.Identity)
	}

	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		res := &client.RequestResult{}
		res.GQL.Errors = append(res.GQL.Errors, err)
		return res
	}

	defer txn.Discard()

	gqlOpts := &client.GQLOptions{}
	if opt.OperationName.HasValue() {
		gqlOpts.OperationName = opt.OperationName.Value()
	}
	gqlOpts.Variables = opt.Variables

	res := db.execRequest(ctx, request, gqlOpts)
	if len(res.GQL.Errors) > 0 {
		return res
	}

	if err := txn.Commit(); err != nil {
		res.GQL.Errors = append(res.GQL.Errors, err)
		return res
	}

	return res
}

// GetCollectionByName returns an existing collection within the database.
func (db *DB) GetCollectionByName(
	ctx context.Context,
	name string,
	opts ...options.Enumerable[options.GetCollectionByNameOptions],
) (client.Collection, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeGetCollectionPerm); err != nil {
		return nil, err
	}

	ctx, txn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return nil, err
	}

	defer txn.Discard()

	return db.getCollectionByName(ctx, name)
}

// GetCollections gets all the currently defined collections.
func (db *DB) GetCollections(
	ctx context.Context,
	opts ...options.Enumerable[options.GetCollectionsOptions],
) ([]client.Collection, error) {
	_, hadTxn := datastore.CtxTryGetTxn(ctx)

	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeGetCollectionPerm); err != nil {
		return nil, err
	}

	var err error
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return nil, err
	}

	defer txn.Discard()

	return db.getCollections(ctx, opt, !hadTxn)
}

// ListIndexes gets all the indexes in the database.
func (db *DB) ListIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListIndexesOptions],
) (map[client.CollectionName][]client.IndexDescription, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeListIndexPerm); err != nil {
		return nil, err
	}

	ctx, txn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return nil, err
	}

	defer txn.Discard()

	return db.listIndexDescriptions(ctx)
}

// ListAllEncryptedIndexes gets all the encrypted indexes in the database.
func (db *DB) ListAllEncryptedIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListAllEncryptedIndexesOptions],
) (map[client.CollectionName][]client.EncryptedIndexDescription, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)
	ident := opt.GetIdentity()
	if err := db.checkNodeAccess(ctx, ident, acpTypes.NodeListAllEncryptedIndexPerm); err != nil {
		return nil, err
	}

	ctx, txn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return nil, err
	}

	defer txn.Discard()

	return db.listAllEncryptedIndexDescriptions(ctx)
}

// AddCollection takes the provided GQL SDL and applies it to the database,
// creating the necessary collections, request types, etc.
//
// All collection types provided must not exist prior to calling this, and they may not
// reference existing types previously defined.
func (db *DB) AddCollection(
	ctx context.Context,
	sdl string,
	opts ...options.Enumerable[options.AddCollectionOptions],
) ([]client.CollectionVersion, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodePatchCollectionPerm); err != nil {
		return nil, err
	}

	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return nil, err
	}

	defer txn.Discard()

	cols, err := db.addCollection(ctx, sdl)
	if err != nil {
		return nil, err
	}

	if err := txn.Commit(); err != nil {
		return nil, err
	}
	return cols, nil
}

// PatchCollection takes the given JSON patch string and applies it to the set of CollectionVersions
// present in the database.
//
// It will also update the GQL types used by the query system. It will error and not apply any of the
// requested, valid updates should the net result of the patch result in an invalid state.  The
// individual operations defined in the patch do not need to result in a valid state, only the net result
// of the full patch.
//
// The collections (including the collection version ID) will only be updated if any changes have actually
// been made, if the net result of the patch matches the current persisted description then no changes
// will be applied.

func (db *DB) PatchCollection(
	ctx context.Context,
	patchString string,
	migration immutable.Option[model.Lens],
	opts ...options.Enumerable[options.PatchCollectionOptions],
) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodePatchCollectionPerm); err != nil {
		return err
	}

	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}

	defer txn.Discard()

	err = db.patchCollection(ctx, patchString, migration)
	if err != nil {
		return err
	}

	return txn.Commit()
}

func (db *DB) SetActiveCollectionVersion(
	ctx context.Context,
	collectionVersionID string,
	opts ...options.Enumerable[options.SetActiveCollectionVersionOptions],
) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodePatchCollectionPerm); err != nil {
		return err
	}

	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}

	defer txn.Discard()

	err = db.setActiveCollectionVersion(ctx, collectionVersionID)
	if err != nil {
		return err
	}

	return txn.Commit()
}

func (db *DB) SetMigration(
	ctx context.Context,
	cfg client.LensConfig,
	opts ...options.Enumerable[options.SetMigrationOptions],
) (string, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)
	ident := opt.GetIdentity()

	if err := db.checkNodeAccess(ctx, ident, acpTypes.NodeSetMigrationPerm); err != nil {
		return "", err
	}

	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return "", err
	}
	defer txn.Discard()

	lensID, err := db.setMigration(ctx, cfg)
	if err != nil {
		return "", err
	}

	err = txn.Commit()
	if err != nil {
		return "", err
	}

	return lensID, nil
}

func (db *DB) AddLens(
	ctx context.Context,
	lens model.Lens,
	opts ...options.Enumerable[options.AddLensOptions],
) (string, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)
	ident := opt.GetIdentity()

	if err := db.checkNodeAccess(ctx, ident, acpTypes.NodeAddLensPerm); err != nil {
		return "", err
	}

	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return "", err
	}

	defer txn.Discard()

	lensID, err := db.addLens(ctx, lens)
	if err != nil {
		return "", err
	}

	err = txn.Commit()
	if err != nil {
		return "", err
	}

	return lensID, nil
}

func (db *DB) ListLenses(
	ctx context.Context,
	opts ...options.Enumerable[options.ListLensesOptions],
) (map[string]model.Lens, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)
	ident := opt.GetIdentity()

	if err := db.checkNodeAccess(ctx, ident, acpTypes.NodeListLensPerm); err != nil {
		return nil, err
	}

	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return nil, err
	}

	defer txn.Discard()

	lenses, err := db.listLenses(ctx)
	if err != nil {
		return nil, err
	}

	return lenses, nil
}

func (db *DB) AddView(
	ctx context.Context,
	query string,
	sdl string,
	opts ...options.Enumerable[options.AddViewOptions],
) ([]client.CollectionVersion, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.GetIdentity(), acpTypes.NodeAddViewPerm); err != nil {
		return nil, err
	}

	ctx = identity.WithContext(ctx, opt.GetIdentity())

	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return nil, err
	}
	defer txn.Discard()

	defs, err := db.addView(ctx, query, sdl, opt.TransformCID)
	if err != nil {
		return nil, err
	}

	err = txn.Commit()
	if err != nil {
		return nil, err
	}

	return defs, nil
}

func (db *DB) RefreshViews(ctx context.Context, opts ...options.Enumerable[options.RefreshViewsOptions]) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.GetIdentity(), acpTypes.NodeRefreshViewPerm); err != nil {
		return err
	}

	ctx = identity.WithContext(ctx, opt.GetIdentity())

	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}

	defer txn.Discard()

	err = db.refreshViews(ctx, opt)
	if err != nil {
		return err
	}

	err = txn.Commit()
	if err != nil {
		return err
	}

	return nil
}

// BasicImport imports a json dataset.
// filepath must be accessible to the node.
func (db *DB) BasicImport(ctx context.Context, filepath string) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	err = db.basicImport(ctx, filepath)
	if err != nil {
		return err
	}

	return txn.Commit()
}

// BasicExport exports the current data or subset of data to file in json format.
func (db *DB) BasicExport(
	ctx context.Context,
	filepath string,
	opts ...options.Enumerable[options.BasicExportOptions],
) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	ctx, txn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return err
	}
	defer txn.Discard()

	config := &client.BackupConfig{
		Filepath:    filepath,
		Format:      opt.Format,
		Pretty:      opt.Pretty,
		Collections: opt.Collections,
	}

	err = db.basicExport(ctx, config)
	if err != nil {
		return err
	}

	return txn.Commit()
}
