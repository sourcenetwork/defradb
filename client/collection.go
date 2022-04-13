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
)

type Collection interface {
	Description() CollectionDescription
	Name() string
	Schema() SchemaDescription
	ID() uint32
	SchemaID() string

	Indexes() []IndexDescription
	PrimaryIndex() IndexDescription
	Index(uint32) (IndexDescription, error)

	Create(context.Context, *Document) error
	CreateMany(context.Context, []*Document) error
	Update(context.Context, *Document) error
	Save(context.Context, *Document) error
	Delete(context.Context, DocKey) (bool, error)
	Exists(context.Context, DocKey) (bool, error)

	UpdateWith(ctx context.Context, target interface{}, updater interface{}) (*UpdateResult, error)
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
