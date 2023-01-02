// Copyright 2022 Democratized Data Foundation
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

	blockstore "github.com/ipfs/go-ipfs-blockstore"

	"github.com/sourcenetwork/defradb/client/schema"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/events"
)

type DB interface {
	// This function gets reworked a bit, so that it becomes syntax sugar around `PatchSchemaS`.
	// I think it is really nice being able to define the schemas as we do now and do not want to
	// lose that (and force the use of the more complex patch for a simple schema create operation).
	// The signature should stay the same, but internally it can call `PatchSchemaS` or something
	// similar.
	AddSchema(context.Context, string) error

	// This function gets reworked a bit, so that it becomes syntax sugar around PatchSchema
	// it does not nessecarily need to keep CollectionDescription as an input param (it might
	// make sense to replace that with something more patch-like), we might also wish to rename
	// it.
	CreateCollection(context.Context, CollectionDescription) (Collection, error)

	// This applies the given patches, with transaction-like guarentees ensuring that if one
	// patch-item fails, they all fail and the database will not be left in a partially updated
	// state.  This includes JSONPatch `test` operations that can act as a kind of if-statement/
	// user-controlled sanity check: https://jsonpatch.com/#test.
	//
	// Having just the single func (plus the string version) should be easier on the users than
	// a function per operation.
	PatchSchema(context.Context, ...[]schema.Patch) ([]Collection, error)

	// Wraps `PatchSchema` allowing users to provide string based json patches. The implementation
	// of this function will most likely parse the string to `[]schema.Patch` and pass it on to
	// `PatchSchema`.
	PatchSchemaS(context.Context, string) ([]Collection, error)

	GetCollectionByName(context.Context, string) (Collection, error)
	GetCollectionBySchemaID(context.Context, string) (Collection, error)
	GetAllCollections(ctx context.Context) ([]Collection, error)

	Root() datastore.RootStore
	Blockstore() blockstore.Blockstore

	NewTxn(context.Context, bool) (datastore.Txn, error)
	ExecQuery(context.Context, string) *QueryResult
	ExecTransactionalQuery(ctx context.Context, query string, txn datastore.Txn) *QueryResult
	Close(context.Context)

	Events() events.Events

	PrintDump(ctx context.Context) error

	// SetReplicator adds a replicator to the persisted list or adds
	// schemas if the replicator already exists.
	SetReplicator(ctx context.Context, rep Replicator) error
	// DeleteReplicator deletes a replicator from the persisted list
	// or specific schemas if they are specified.
	DeleteReplicator(ctx context.Context, rep Replicator) error
	// GetAllReplicators returns the full list of replicators with their
	// subscribed schemas.
	GetAllReplicators(ctx context.Context) ([]Replicator, error)
}

type GQLResult struct {
	Errors []any `json:"errors,omitempty"`
	Data   any   `json:"data"`
}

type QueryResult struct {
	GQL GQLResult
	Pub *events.Publisher[events.Update]
}
