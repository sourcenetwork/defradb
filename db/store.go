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

var _ client.Store = (*store)(nil)

type store struct {
	*db
}

// ExecRequest executes a request against the database.
func (s *store) ExecRequest(
	ctx context.Context,
	identity immutable.Option[string],
	request string,
) *client.RequestResult {
	txn, err := getContextTxn(ctx, s, false)
	if err != nil {
		res := &client.RequestResult{}
		res.GQL.Errors = []error{err}
		return res
	}
	defer txn.Discard(ctx)

	res := s.db.execRequest(ctx, identity, request, txn)
	if len(res.GQL.Errors) > 0 {
		return res
	}

	if err := txn.Commit(ctx); err != nil {
		res.GQL.Errors = []error{err}
		return res
	}

	return res
}

// GetCollectionByName returns an existing collection within the database.
func (s *store) GetCollectionByName(ctx context.Context, name string) (client.Collection, error) {
	txn, err := getContextTxn(ctx, s, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	return s.db.getCollectionByName(ctx, txn, name)
}

// GetCollections gets all the currently defined collections.
func (s *store) GetCollections(
	ctx context.Context,
	options client.CollectionFetchOptions,
) ([]client.Collection, error) {
	txn, err := getContextTxn(ctx, s, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	return s.db.getCollections(ctx, txn, options)
}

// GetSchemaByVersionID returns the schema description for the schema version of the
// ID provided.
//
// Will return an error if it is not found.
func (s *store) GetSchemaByVersionID(ctx context.Context, versionID string) (client.SchemaDescription, error) {
	txn, err := getContextTxn(ctx, s, true)
	if err != nil {
		return client.SchemaDescription{}, err
	}
	defer txn.Discard(ctx)

	return s.db.getSchemaByVersionID(ctx, txn, versionID)
}

// GetSchemas returns all schema versions that currently exist within
// this [Store].
func (s *store) GetSchemas(
	ctx context.Context,
	options client.SchemaFetchOptions,
) ([]client.SchemaDescription, error) {
	txn, err := getContextTxn(ctx, s, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	return s.db.getSchemas(ctx, txn, options)
}

// GetAllIndexes gets all the indexes in the database.
func (s *store) GetAllIndexes(
	ctx context.Context,
) (map[client.CollectionName][]client.IndexDescription, error) {
	txn, err := getContextTxn(ctx, s, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	return s.db.getAllIndexDescriptions(ctx, txn)
}

// AddSchema takes the provided GQL schema in SDL format, and applies it to the database,
// creating the necessary collections, request types, etc.
//
// All schema types provided must not exist prior to calling this, and they may not reference existing
// types previously defined.
func (s *store) AddSchema(ctx context.Context, schemaString string) ([]client.CollectionDescription, error) {
	txn, err := getContextTxn(ctx, s, false)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	cols, err := s.db.addSchema(ctx, txn, schemaString)
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
func (s *store) PatchSchema(
	ctx context.Context,
	patchString string,
	migration immutable.Option[model.Lens],
	setAsDefaultVersion bool,
) error {
	txn, err := getContextTxn(ctx, s, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	err = s.db.patchSchema(ctx, txn, patchString, migration, setAsDefaultVersion)
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

func (s *store) PatchCollection(
	ctx context.Context,
	patchString string,
) error {
	txn, err := getContextTxn(ctx, s, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	err = s.db.patchCollection(ctx, txn, patchString)
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

func (s *store) SetActiveSchemaVersion(ctx context.Context, schemaVersionID string) error {
	txn, err := getContextTxn(ctx, s, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	err = s.db.setActiveSchemaVersion(ctx, txn, schemaVersionID)
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

func (s *store) SetMigration(ctx context.Context, cfg client.LensConfig) error {
	txn, err := getContextTxn(ctx, s, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	err = s.db.setMigration(ctx, txn, cfg)
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

func (s *store) AddView(
	ctx context.Context,
	query string,
	sdl string,
	transform immutable.Option[model.Lens],
) ([]client.CollectionDefinition, error) {
	txn, err := getContextTxn(ctx, s, false)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	defs, err := s.db.addView(ctx, txn, query, sdl, transform)
	if err != nil {
		return nil, err
	}

	err = txn.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return defs, nil
}

// BasicImport imports a json dataset.
// filepath must be accessible to the node.
func (s *store) BasicImport(ctx context.Context, filepath string) error {
	txn, err := getContextTxn(ctx, s, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	err = s.db.basicImport(ctx, txn, filepath)
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

// BasicExport exports the current data or subset of data to file in json format.
func (s *store) BasicExport(ctx context.Context, config *client.BackupConfig) error {
	txn, err := getContextTxn(ctx, s, true)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	err = s.db.basicExport(ctx, txn, config)
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

func (s *store) LensRegistry() client.LensRegistry {
	return s.db.lensRegistry
}
