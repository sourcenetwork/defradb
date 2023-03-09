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
	PatchSchema(context.Context, string) error

	Root() datastore.RootStore
	Blockstore() blockstore.Blockstore

	NewTxn(context.Context, bool) (datastore.Txn, error)
	NewConcurrentTxn(context.Context, bool) (datastore.Txn, error)

	// WithTxn returns a new Store instanciated with a new transaction
	// WithTxn(context.Context, bool) (TxnStore, error)

	// or... Instead of (implicitly) creating its own transaction, can
	// be given an explicit transaction. Which would simplify this API
	// further since we wouldn't need a dedicated `TxnStore` and instead
	// this function only returns a `Store` as you can see.
	WithTxn(context.Context, datastore.Txn) Store

	Events() events.Events

	MaxTxnRetries() int
	PrintDump(ctx context.Context) error
	Close(context.Context)

	// Store interface for actual Collection (and peer) based CRUD operations
	Store
}

// type TxnStore interface {
// 	Store
// 	datastore.Txn
// }

type Store interface {
	Read
	PeerRead
	Write
	PeerWrite

	// This could be on client.Write interface alternatively
	// but since its a query it has the capacity to be either
	// read or write or both, so I have put it on the client.Store
	// interface for now.
	ExecRequest(context.Context, string) *RequestResult
}

type Write interface {
	CreateCollection(context.Context, CollectionDescription) (Collection, error)

	// UpdateCollectionTxn updates the persisted collection description matching the name of the given
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
}

type Read interface {
	GetCollectionByName(context.Context, string) (Collection, error)
	GetCollectionBySchemaID(context.Context, string) (Collection, error)
	GetAllCollections(context.Context) ([]Collection, error)
}

type GQLResult struct {
	Errors []any `json:"errors,omitempty"`
	Data   any   `json:"data"`
}

type RequestResult struct {
	GQL GQLResult
	Pub *events.Publisher[events.Update]
}
