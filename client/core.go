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

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/document"
	"github.com/sourcenetwork/defradb/document/key"
	"github.com/sourcenetwork/defradb/query/graphql/schema"

	blocks "github.com/ipfs/go-block-format"
	cid "github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
)

type DB interface {
	// Collections
	CreateCollection(context.Context, base.CollectionDescription) (Collection, error)
	GetCollectionByName(context.Context, string) (Collection, error)
	GetCollectionBySchemaID(context.Context, string) (Collection, error)
	ExecQuery(context.Context, string) *QueryResult
	SchemaManager() *schema.SchemaManager
	AddSchema(context.Context, string) error
	PrintDump(ctx context.Context)
	GetBlock(ctx context.Context, c cid.Cid) (blocks.Block, error)
	Root() ds.Batching
	// Rootstore() core.DSReaderWriter
	// Headstore() core.DSReaderWriter
	// Datastore() core.DSReaderWriter
	DAGstore() core.DAGStore

	NewTxn(context.Context, bool) (core.Txn, error)
	GetAllCollections(ctx context.Context) ([]Collection, error)
}

type Sequence interface{}

type Collection interface {
	Description() base.CollectionDescription
	Name() string
	Schema() base.SchemaDescription
	ID() uint32
	SchemaID() string

	Indexes() []base.IndexDescription
	PrimaryIndex() base.IndexDescription
	Index(uint32) (base.IndexDescription, error)
	CreateIndex(base.IndexDescription) error

	Create(context.Context, *document.Document) error
	CreateMany(context.Context, []*document.Document) error
	Update(context.Context, *document.Document) error
	Save(context.Context, *document.Document) error
	Delete(context.Context, key.DocKey) (bool, error)
	Exists(context.Context, key.DocKey) (bool, error)

	UpdateWith(context.Context, interface{}, interface{}, ...UpdateOpt) error
	UpdateWithFilter(context.Context, interface{}, interface{}, ...UpdateOpt) (*UpdateResult, error)
	UpdateWithKey(context.Context, key.DocKey, interface{}, ...UpdateOpt) (*UpdateResult, error)
	UpdateWithKeys(context.Context, []key.DocKey, interface{}, ...UpdateOpt) (*UpdateResult, error)

	DeleteWith(context.Context, interface{}, ...DeleteOpt) error
	DeleteWithFilter(context.Context, interface{}, ...DeleteOpt) (*DeleteResult, error)
	DeleteWithKey(context.Context, key.DocKey, ...DeleteOpt) (*DeleteResult, error)
	DeleteWithKeys(context.Context, []key.DocKey, ...DeleteOpt) (*DeleteResult, error)

	Get(context.Context, key.DocKey) (*document.Document, error)

	WithTxn(core.Txn) Collection

	GetAllDocKeys(ctx context.Context) (<-chan DocKeysResult, error)
}

type DocKeysResult struct {
	Key key.DocKey
	Err error
}

type UpdateOpt struct{}
type CreateOpt struct{}
type DeleteOpt struct{}

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
