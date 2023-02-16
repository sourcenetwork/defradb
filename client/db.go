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

	CreateCollection(context.Context, CollectionDescription) (Collection, error)
	CreateCollectionTxn(context.Context, datastore.Txn, CollectionDescription) (Collection, error)
	GetCollectionByName(context.Context, string) (Collection, error)
	GetCollectionByNameTxn(context.Context, datastore.Txn, string) (Collection, error)
	GetCollectionBySchemaID(context.Context, string) (Collection, error)
	GetCollectionBySchemaIDTxn(context.Context, datastore.Txn, string) (Collection, error)
	GetAllCollections(context.Context) ([]Collection, error)
	GetAllCollectionsTxn(context.Context, datastore.Txn) ([]Collection, error)

	Root() datastore.RootStore
	Blockstore() blockstore.Blockstore

	NewTxn(context.Context, bool) (datastore.Txn, error)
	NewConcurrentTxn(context.Context, bool) (datastore.Txn, error)
	ExecRequest(context.Context, string) *RequestResult
	ExecTransactionalRequest(context.Context, string, datastore.Txn) *RequestResult
	Close(context.Context)

	Events() events.Events

	MaxTxnRetries() int

	PrintDump(ctx context.Context) error

	// P2P holds the P2P related methods that must be implemented by the database.
	P2P
}

type GQLResult struct {
	Errors []any `json:"errors,omitempty"`
	Data   any   `json:"data"`
}

type RequestResult struct {
	GQL GQLResult
	Pub *events.Publisher[events.Update]
}
