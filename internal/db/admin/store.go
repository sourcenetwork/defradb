// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package admin

import (
	"context"

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/immutable"
)

func (adminDB *AdminDB) ExecRequest(
	ctx context.Context,
	request string,
	opts ...client.RequestOption,
) *client.RequestResult {
	return adminDB.db.ExecRequest(ctx, request, opts...)
}

func (adminDB *AdminDB) GetCollectionByName(ctx context.Context, name string) (client.Collection, error) {
	return adminDB.db.GetCollectionByName(ctx, name)
}

func (adminDB *AdminDB) GetCollections(
	ctx context.Context,
	options client.CollectionFetchOptions,
) ([]client.Collection, error) {
	return adminDB.db.GetCollections(ctx, options)
}

func (adminDB *AdminDB) GetSchemaByVersionID(ctx context.Context, versionID string) (client.SchemaDescription, error) {
	return adminDB.db.GetSchemaByVersionID(ctx, versionID)
}

func (adminDB *AdminDB) GetSchemas(
	ctx context.Context,
	options client.SchemaFetchOptions,
) ([]client.SchemaDescription, error) {
	return adminDB.db.GetSchemas(ctx, options)
}

func (adminDB *AdminDB) GetAllIndexes(
	ctx context.Context,
) (map[client.CollectionName][]client.IndexDescription, error) {
	return adminDB.db.GetAllIndexes(ctx)
}

func (adminDB *AdminDB) AddSchema(ctx context.Context, schemaString string) ([]client.CollectionVersion, error) {
	return adminDB.db.AddSchema(ctx, schemaString)
}

func (adminDB *AdminDB) PatchSchema(
	ctx context.Context,
	patchString string,
	migration immutable.Option[model.Lens],
	setAsDefaultVersion bool,
) error {
	return adminDB.db.PatchSchema(ctx, patchString, migration, setAsDefaultVersion)
}

func (adminDB *AdminDB) PatchCollection(
	ctx context.Context,
	patchString string,
) error {
	return adminDB.db.PatchCollection(ctx, patchString)
}

func (adminDB *AdminDB) SetActiveSchemaVersion(ctx context.Context, schemaVersionID string) error {
	return adminDB.db.SetActiveSchemaVersion(ctx, schemaVersionID)
}

func (adminDB *AdminDB) SetMigration(ctx context.Context, cfg client.LensConfig) error {
	return adminDB.db.SetMigration(ctx, cfg)
}

func (adminDB *AdminDB) LensRegistry() client.LensRegistry {
	return adminDB.db.LensRegistry()
}

func (adminDB *AdminDB) AddView(
	ctx context.Context,
	query string,
	sdl string,
	transform immutable.Option[model.Lens],
) ([]client.CollectionDefinition, error) {
	return adminDB.db.AddView(ctx, query, sdl, transform)
}

func (adminDB *AdminDB) RefreshViews(ctx context.Context, opts client.CollectionFetchOptions) error {
	return adminDB.db.RefreshViews(ctx, opts)
}

func (adminDB *AdminDB) BasicImport(ctx context.Context, filepath string) error {
	return adminDB.db.BasicImport(ctx, filepath)
}

func (adminDB *AdminDB) BasicExport(ctx context.Context, config *client.BackupConfig) error {
	return adminDB.db.BasicExport(ctx, config)
}
