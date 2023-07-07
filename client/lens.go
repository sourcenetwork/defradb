// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import (
	"context"

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/immutable/enumerable"

	"github.com/sourcenetwork/defradb/datastore"
)

// LensConfig represents the configuration of a Lens migration in Defra.
type LensConfig struct {
	// SourceSchemaVersionID is the ID of the schema version from which to migrate
	// from.
	//
	// The source and destination versions must be next to each other in the history.
	SourceSchemaVersionID string

	// DestinationSchemaVersionID is the ID of the schema version from which to migrate
	// to.
	//
	// The source and destination versions must be next to each other in the history.
	DestinationSchemaVersionID string

	// The configuration of the Lens module.
	//
	// For now, the wasm module must remain at the location specified as long as the
	// migration is active.
	model.Lens
}

// LensRegistry exposes several useful thread-safe migration related functions which may
// be used to manage migrations.
type LensRegistry interface {
	// SetMigration sets the migration for the given source-destination schema version IDs. Is equivilent to
	// calling `Store.SetMigration(ctx, cfg)`.
	//
	// There may only be one migration per schema version id.  If another migration was registered it will be
	// overwritten by this migration.
	//
	// Neither of the schema version IDs specified in the configuration need to exist at the time of calling.
	// This is to allow the migration of documents of schema versions unknown to the local node recieved by the
	// P2P system.
	//
	// Migrations will only run if there is a complete path from the document schema version to the latest local
	// schema version.
	SetMigration(context.Context, datastore.Txn, LensConfig) error

	// ReloadLenses clears any cached migrations, loads their configurations from the database and re-initializes
	// them.  It is run on database start if the database already existed.
	ReloadLenses(ctx context.Context, txn datastore.Txn) error

	// MigrateUp returns an enumerable that feeds the given source through the Lens migration for the given
	// schema version id if one is found, if there is no matching migration the given source will be returned.
	MigrateUp(enumerable.Enumerable[map[string]any], string) (enumerable.Enumerable[map[string]any], error)

	// MigrateDown returns an enumerable that feeds the given source through the Lens migration for the schema
	// version that precedes the given schema version id in reverse, if one is found, if there is no matching
	// migration the given source will be returned.
	//
	// This downgrades any documents in the source enumerable if/when enumerated.
	MigrateDown(enumerable.Enumerable[map[string]any], string) (enumerable.Enumerable[map[string]any], error)

	// Config returns a slice of the configurations of the currently loaded migrations.
	//
	// Modifying the slice does not affect the loaded configurations.
	Config() []LensConfig
}
