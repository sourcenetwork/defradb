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
)

// ExecRequest executes a request against the database.
func (db *DB) ExecRequest(ctx context.Context, request string, opts ...client.RequestOption) *client.RequestResult {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		res := &client.RequestResult{}
		res.GQL.Errors = append(res.GQL.Errors, err)
		return res
	}
	defer txn.Discard()

	options := &client.GQLOptions{}
	for _, o := range opts {
		o(options)
	}

	res := db.execRequest(ctx, request, options)
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
func (db *DB) GetCollectionByName(ctx context.Context, name string) (client.Collection, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	if err := db.checkNodeAccess(ctx, acpTypes.NodeCollectionGetPerm); err != nil {
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
	options client.CollectionFetchOptions,
) ([]client.Collection, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	if err := db.checkNodeAccess(ctx, acpTypes.NodeCollectionGetPerm); err != nil {
		return nil, err
	}

	ctx, txn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard()

	return db.getCollections(ctx, options)
}

// GetAllIndexes gets all the indexes in the database.
func (db *DB) GetAllIndexes(
	ctx context.Context,
) (map[client.CollectionName][]client.IndexDescription, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	if err := db.checkNodeAccess(ctx, acpTypes.NodeIndexListPerm); err != nil {
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
func (db *DB) AddSchema(ctx context.Context, schemaString string) ([]client.CollectionVersion, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	if err := db.checkNodeAccess(ctx, acpTypes.NodeCollectionPatchPerm); err != nil {
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
) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	if err := db.checkNodeAccess(ctx, acpTypes.NodeCollectionPatchPerm); err != nil {
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

func (db *DB) SetActiveCollectionVersion(ctx context.Context, schemaVersionID string) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	if err := db.checkNodeAccess(ctx, acpTypes.NodeCollectionPatchPerm); err != nil {
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

func (db *DB) AddView(
	ctx context.Context,
	query string,
	sdl string,
	transform immutable.Option[model.Lens],
) ([]client.CollectionVersion, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return nil, err
	}
	defer txn.Discard()

	defs, err := db.addView(ctx, query, sdl, transform)
	if err != nil {
		return nil, err
	}

	err = txn.Commit()
	if err != nil {
		return nil, err
	}

	return defs, nil
}

func (db *DB) RefreshViews(ctx context.Context, opts client.CollectionFetchOptions) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	err = db.refreshViews(ctx, opts)
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
