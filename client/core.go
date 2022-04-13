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
	ExecQuery(context.Context, string) *QueryResult
	ExecTransactionalQuery(ctx context.Context, query string, txn datastore.Txn) *QueryResult
	Close(context.Context)

	PrintDump(ctx context.Context)
}

type Collection interface {
	Description() CollectionDescription
	Name() string
	Schema() SchemaDescription
	ID() uint32
	SchemaID() string

	Indexes() []IndexDescription
	PrimaryIndex() IndexDescription
	Index(uint32) (IndexDescription, error)
	CreateIndex(IndexDescription) error

	Create(context.Context, *Document) error
	CreateMany(context.Context, []*Document) error
	Update(context.Context, *Document) error
	Save(context.Context, *Document) error
	Delete(context.Context, DocKey) (bool, error)
	Exists(context.Context, DocKey) (bool, error)

	UpdateWith(context.Context, interface{}, interface{}) error
	UpdateWithFilter(context.Context, interface{}, interface{}) (*UpdateResult, error)
	UpdateWithKey(context.Context, DocKey, interface{}) (*UpdateResult, error)
	UpdateWithKeys(context.Context, []DocKey, interface{}) (*UpdateResult, error)

	DeleteWith(context.Context, interface{}) error
	DeleteWithFilter(context.Context, interface{}) (*DeleteResult, error)
	DeleteWithKey(context.Context, DocKey) (*DeleteResult, error)
	DeleteWithKeys(context.Context, []DocKey) (*DeleteResult, error)

	Get(context.Context, DocKey) (*Document, error)

	WithTxn(datastore.Txn) Collection

	GetAllDocKeys(ctx context.Context) (<-chan DocKeysResult, error)
}

type DocKeysResult struct {
	Key DocKey
	Err error
}

type UpdateResult struct {
	Count   int64
	DocKeys []string
}

type DeleteResult struct {
	Count   int64
	DocKeys []string
}

type QueryResult struct {
	Errors []interface{} `json:"errors,omitempty"`
	Data   interface{}   `json:"data"`
}
