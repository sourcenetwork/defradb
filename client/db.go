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

	ds "github.com/ipfs/go-datastore"
	blockstore "github.com/ipfs/go-ipfs-blockstore"

	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/events"
)

type DB interface {
	AddSchema(context.Context, string) error

	CreateCollection(context.Context, CollectionDescription) (Collection, error)
	GetCollectionByName(context.Context, string) (Collection, error)
	GetCollectionBySchemaID(context.Context, string) (Collection, error)
	GetAllCollections(ctx context.Context) ([]Collection, error)

	Root() ds.Batching
	Blockstore() blockstore.Blockstore

	NewTxn(context.Context, bool) (datastore.Txn, error)
	ExecQuery(context.Context, string) *QueryResult
	ExecTransactionalQuery(ctx context.Context, query string, txn datastore.Txn) *QueryResult
	Close(context.Context)

	Events() Events

	PrintDump(ctx context.Context) error
}

type GQLResult struct {
	Errors []any `json:"errors,omitempty"`
	Data   any   `json:"data"`
}

type QueryResult struct {
	GQL    GQLResult
	Stream *events.Publisher[GQLResult]
}
