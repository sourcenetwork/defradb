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

	"github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
)

// ExecRequest executes a request against the database.
func (db *DB) ExecRequest(ctx context.Context, request string, opts ...*options.ExecRequestOptions) *client.RequestResult {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		res := &client.RequestResult{}
		res.GQL.Errors = append(res.GQL.Errors, err)
		return res
	}
	defer txn.Discard()

	gqlOpts := &client.GQLOptions{}
	if len(opts) > 0 && opts[0] != nil {
		opt := opts[0]
		if opt.OperationName.HasValue() {
			gqlOpts.OperationName = opt.OperationName.Value()
		}
		gqlOpts.Variables = opt.Variables
	}

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
	opts ...*options.GetCollectionByNameOptions,
) (client.Collection, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	var ident immutable.Option[identity.Identity]
	if len(opts) > 0 && opts[0] != nil {
		ident = opts[0].Identity
	}

	if err := db.checkNodeAccess(ctx, ident, acpTypes.NodeCollectionGetPerm); err != nil {
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
	opts ...*options.GetCollectionsOptions,
) ([]client.Collection, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	var opt *options.GetCollectionsOptions
	if len(opts) > 0 && opts[0] != nil {
		opt = opts[0]
	}

	var ident immutable.Option[identity.Identity]
	if opt != nil {
		ident = opt.Identity
	}

	if err := db.checkNodeAccess(ctx, ident, acpTypes.NodeCollectionGetPerm); err != nil {
		return nil, err
	}

	ctx, txn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard()

	return db.getCollections(ctx, opt)
}

// GetAllIndexes gets all the indexes in the database.
func (db *DB) GetAllIndexes(
	ctx context.Context,
	opts ...*options.GetAllIndexesOptions,
) (map[client.CollectionName][]client.IndexDescription, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	var ident immutable.Option[identity.Identity]
	if len(opts) > 0 && opts[0] != nil {
		ident = opts[0].Identity
	}

	if err := db.checkNodeAccess(ctx, ident, acpTypes.NodeIndexListPerm); err != nil {
		return nil, err
	}

	ctx, txn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard()

	return db.getAllIndexDescriptions(ctx)
}

// ListAllEncryptedIndexes gets all the encrypted indexes in the database.
func (db *DB) ListAllEncryptedIndexes(
	ctx context.Context,
) (map[client.CollectionName][]client.EncryptedIndexDescription, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	ctx, txn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard()

	return db.listAllEncryptedIndexDescriptions(ctx)
}

// AddSchema takes the provided GQL schema in SDL format, and applies it to the database,
// creating the necessary collections, request types, etc.
//
// All schema types provided must not exist prior to calling this, and they may not reference existing
// types previously defined.
func (db *DB) AddSchema(
	ctx context.Context,
	schemaString string,
	opts ...*options.AddSchemaOptions,
) ([]client.CollectionVersion, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	var ident immutable.Option[identity.Identity]
	if len(opts) > 0 && opts[0] != nil {
		ident = opts[0].Identity
	}

	if err := db.checkNodeAccess(ctx, ident, acpTypes.NodeCollectionPatchPerm); err != nil {
		return nil, err
	}

	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return nil, err
	}
	defer txn.Discard()

	cols, err := db.addSchema(ctx, schemaString)
	if err != nil {
		return nil, err
	}

	if err := txn.Commit(); err != nil {
		return nil, err
	}
	return cols, nil
}

// PatchSchema takes the given JSON patch string and applies it to the set of SchemaDescriptions
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

func (db *DB) PatchCollection(
	ctx context.Context,
	patchString string,
	migration immutable.Option[model.Lens],
	opts ...*options.PatchCollectionOptions,
) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	var ident immutable.Option[identity.Identity]
	if len(opts) > 0 && opts[0] != nil {
		ident = opts[0].Identity
	}

	if err := db.checkNodeAccess(ctx, ident, acpTypes.NodeCollectionPatchPerm); err != nil {
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
	schemaVersionID string,
	opts ...*options.SetActiveCollectionVersionOptions,
) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	var ident immutable.Option[identity.Identity]
	if len(opts) > 0 && opts[0] != nil {
		ident = opts[0].Identity
	}

	if err := db.checkNodeAccess(ctx, ident, acpTypes.NodeCollectionPatchPerm); err != nil {
		return err
	}

	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	err = db.setActiveCollectionVersion(ctx, schemaVersionID)
	if err != nil {
		return err
	}

	return txn.Commit()
}

func (db *DB) SetMigration(ctx context.Context, cfg client.LensConfig) (string, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

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

func (db *DB) AddLens(ctx context.Context, lens model.Lens) (string, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

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

func (db *DB) ListLenses(ctx context.Context) (map[string]model.Lens, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return nil, err
	}
	defer txn.Discard()

	return db.listLenses(ctx)
}

func (db *DB) AddView(
	ctx context.Context,
	query string,
	sdl string,
	transformCID immutable.Option[string],
) ([]client.CollectionVersion, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return nil, err
	}
	defer txn.Discard()

	defs, err := db.addView(ctx, query, sdl, transformCID)
	if err != nil {
		return nil, err
	}

	err = txn.Commit()
	if err != nil {
		return nil, err
	}

	return defs, nil
}

func (db *DB) RefreshViews(ctx context.Context, opts ...*options.RefreshViewsOptions) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	var getCollOpts *options.GetCollectionsOptions
	if len(opts) > 0 && opts[0] != nil {
		getCollOpts = opts[0].ToGetCollectionsOptions()
	} else {
		getCollOpts = options.GetCollections()
	}

	err = db.refreshViews(ctx, getCollOpts)
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
func (db *DB) BasicExport(ctx context.Context, config *client.BackupConfig) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	ctx, txn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return err
	}
	defer txn.Discard()

	err = db.basicExport(ctx, config)
	if err != nil {
		return err
	}

	return txn.Commit()
}
