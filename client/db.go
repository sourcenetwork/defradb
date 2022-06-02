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

	"github.com/sourcenetwork/defradb/datastore"

	ds "github.com/ipfs/go-datastore"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
)

type DB interface {
	AddSchema(context.Context, string) error

	CreateCollection(context.Context, CollectionDescription) (Collection, error)
	GetCollectionByName(context.Context, string) (Collection, error)
	GetCollectionBySchemaID(context.Context, string) (Collection, error)
	GetAllCollections(ctx context.Context) ([]Collection, error)
	GetRelationshipIdField(fieldName, targetType, thisType string) (string, error)

	Root() ds.Batching
	Blockstore() blockstore.Blockstore

	NewTxn(context.Context, bool) (datastore.Txn, error)
	ExecuteRequest(context.Context, string) *RequestResult
	ExecuteTransactionalRequest(ctx context.Context, request string, txn datastore.Txn) *RequestResult
	Close(context.Context)

	PrintDump(ctx context.Context)
}

type RequestResult struct {
	Errors []interface{} `json:"errors,omitempty"`
	Data   interface{}   `json:"data"`
}
