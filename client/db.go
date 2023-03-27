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

	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/events"
)

type DB interface {
	Store

	NewTxn(context.Context, bool) (datastore.Txn, error)
	NewConcurrentTxn(context.Context, bool) (datastore.Txn, error)
	// WithTxn returns a new [client.Store] that respects the given transaction.
	WithTxn(datastore.Txn) Store

	Root() datastore.RootStore
	Blockstore() blockstore.Blockstore

	Close(context.Context)

	Events() events.Events

	MaxTxnRetries() int

	PrintDump(ctx context.Context) error
}

type Store interface {
	// P2P holds the P2P related methods that must be implemented by the database.
	P2P

	AddSchema(context.Context, string) error

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
	//
	// Field [FieldKind] values may be provided in either their raw integer form, or as string as per
	// [FieldKindStringToEnumMapping].
	PatchSchema(context.Context, string) error

	CreateCollection(context.Context, CollectionDescription) (Collection, error)

	// UpdateCollection updates the persisted collection description matching the name of the given
	// description, to the values in the given description.
	//
	// It will validate the given description using [ValidateUpdateCollection] before updating it.
	//
	// The collection (including the schema version ID) will only be updated if any changes have actually
	// been made, if the given description matches the current persisted description then no changes will be
	// applied.
	UpdateCollection(context.Context, CollectionDescription) (Collection, error)

	// ValidateUpdateCollection validates that the given collection description is a valid update.
	//
	// Will return true if the given desctiption differs from the current persisted state of the
	// collection. Will return an error if it fails validation.
	ValidateUpdateCollection(context.Context, CollectionDescription) (bool, error)

	GetCollectionByName(context.Context, string) (Collection, error)
	GetCollectionBySchemaID(context.Context, string) (Collection, error)
	GetAllCollections(context.Context) ([]Collection, error)

	ExecRequest(context.Context, string) *RequestResult
}

type GQLResult struct {
	Errors []any `json:"errors,omitempty"`
	Data   any   `json:"data"`
}

type RequestResult struct {
	GQL GQLResult
	Pub *events.Publisher[events.Update]
}
