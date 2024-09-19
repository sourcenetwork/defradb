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

	"github.com/lens-vm/lens/host-go/config/model"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
)

// ExecRequest executes a request against the database.
func (db *db) ExecRequest(ctx context.Context, request string, opts ...client.RequestOption) *client.RequestResult {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		res := &client.RequestResult{}
		res.GQL.AddErrors(err)
		return res
	}
	defer txn.Discard(ctx)

	options := &client.GQLOptions{}
	for _, o := range opts {
		o(options)
	}

	res := db.execRequest(ctx, request, options)
	if len(res.GQL.Errors) > 0 {
		return res
	}

	if err := txn.Commit(ctx); err != nil {
		res.GQL.AddErrors(err)
		return res
	}

	return res
}

// GetCollectionByName returns an existing collection within the database.
func (db *db) GetCollectionByName(ctx context.Context, name string) (client.Collection, error) {
	ctx, txn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	return db.getCollectionByName(ctx, name)
}

// GetCollections gets all the currently defined collections.
func (db *db) GetCollections(
	ctx context.Context,
	options client.CollectionFetchOptions,
) ([]client.Collection, error) {
	ctx, txn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	return db.getCollections(ctx, options)
}

// GetSchemaByVersionID returns the schema description for the schema version of the
// ID provided.
//
// Will return an error if it is not found.
func (db *db) GetSchemaByVersionID(ctx context.Context, versionID string) (client.SchemaDescription, error) {
	ctx, txn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return client.SchemaDescription{}, err
	}
	defer txn.Discard(ctx)

	return db.getSchemaByVersionID(ctx, versionID)
}

// GetSchemas returns all schema versions that currently exist within
// this [Store].
func (db *db) GetSchemas(
	ctx context.Context,
	options client.SchemaFetchOptions,
) ([]client.SchemaDescription, error) {
	ctx, txn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	return db.getSchemas(ctx, options)
}

// GetAllIndexes gets all the indexes in the database.
func (db *db) GetAllIndexes(
	ctx context.Context,
) (map[client.CollectionName][]client.IndexDescription, error) {
	ctx, txn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	return db.getAllIndexDescriptions(ctx)
}

// AddSchema takes the provided GQL schema in SDL format, and applies it to the database,
// creating the necessary collections, request types, etc.
//
// All schema types provided must not exist prior to calling this, and they may not reference existing
// types previously defined.
func (db *db) AddSchema(ctx context.Context, schemaString string) ([]client.CollectionDescription, error) {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	cols, err := db.addSchema(ctx, schemaString)
	if err != nil {
		return nil, err
	}

	if err := txn.Commit(ctx); err != nil {
		return nil, err
	}
	return cols, nil
}

// PatchSchema takes the given JSON patch string and applies it to the set of CollectionDescriptions
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
func (db *db) PatchSchema(
	ctx context.Context,
	patchString string,
	migration immutable.Option[model.Lens],
	setAsDefaultVersion bool,
) error {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	err = db.patchSchema(ctx, patchString, migration, setAsDefaultVersion)
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

func (db *db) PatchCollection(
	ctx context.Context,
	patchString string,
) error {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	err = db.patchCollection(ctx, patchString)
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

func (db *db) SetActiveSchemaVersion(ctx context.Context, schemaVersionID string) error {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	err = db.setActiveSchemaVersion(ctx, schemaVersionID)
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

func (db *db) SetMigration(ctx context.Context, cfg client.LensConfig) error {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	err = db.setMigration(ctx, cfg)
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

func (db *db) AddView(
	ctx context.Context,
	query string,
	sdl string,
	transform immutable.Option[model.Lens],
) ([]client.CollectionDefinition, error) {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	defs, err := db.addView(ctx, query, sdl, transform)
	if err != nil {
		return nil, err
	}

	err = txn.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return defs, nil
}

func (db *db) RefreshViews(ctx context.Context, opts client.CollectionFetchOptions) error {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	err = db.refreshViews(ctx, opts)
	if err != nil {
		return err
	}

	err = txn.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

// BasicImport imports a json dataset.
// filepath must be accessible to the node.
func (db *db) BasicImport(ctx context.Context, filepath string) error {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	err = db.basicImport(ctx, filepath)
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

// BasicExport exports the current data or subset of data to file in json format.
func (db *db) BasicExport(ctx context.Context, config *client.BackupConfig) error {
	ctx, txn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	err = db.basicExport(ctx, config)
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}
