// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package http

import (
	"context"
	"net/http"
	"net/url"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
)

var _ client.Collection = (*CollectionClient)(nil)

// LensClient implements the client.Collection interface over HTTP.
type CollectionClient struct {
	client      *http.Client
	baseURL     *url.URL
	description client.CollectionDescription
}

func NewCollectionClient(s *Client, description client.CollectionDescription) *CollectionClient {
	return &CollectionClient{
		client:      s.client,
		baseURL:     s.baseURL,
		description: description,
	}
}

func (c *CollectionClient) Description() client.CollectionDescription {
	return c.description
}

func (c *CollectionClient) Name() string {
	return c.description.Name
}

func (c *CollectionClient) Schema() client.SchemaDescription {
	return c.description.Schema
}

func (c *CollectionClient) ID() uint32 {
	return c.description.ID
}

func (c *CollectionClient) SchemaID() string {
	return c.description.Schema.SchemaID
}

func (c *CollectionClient) Create(context.Context, *client.Document) error {
	return nil
}

func (c *CollectionClient) CreateMany(context.Context, []*client.Document) error {
	return nil
}

func (c *CollectionClient) Update(context.Context, *client.Document) error {
	return nil
}

func (c *CollectionClient) Save(context.Context, *client.Document) error {
	return nil
}

func (c *CollectionClient) Delete(context.Context, client.DocKey) (bool, error) {
	return false, nil
}

func (c *CollectionClient) Exists(context.Context, client.DocKey) (bool, error) {
	return false, nil
}

func (c *CollectionClient) UpdateWith(ctx context.Context, target any, updater string) (*client.UpdateResult, error) {
	return nil, nil
}

func (c *CollectionClient) UpdateWithFilter(ctx context.Context, filter any, updater string) (*client.UpdateResult, error) {
	return nil, nil
}

func (c *CollectionClient) UpdateWithKey(ctx context.Context, key client.DocKey, updater string) (*client.UpdateResult, error) {
	return nil, nil
}

func (c *CollectionClient) UpdateWithKeys(context.Context, []client.DocKey, string) (*client.UpdateResult, error) {
	return nil, nil
}

func (c *CollectionClient) DeleteWith(ctx context.Context, target any) (*client.DeleteResult, error) {
	return nil, nil
}

func (c *CollectionClient) DeleteWithFilter(ctx context.Context, filter any) (*client.DeleteResult, error) {
	return nil, nil
}

func (c *CollectionClient) DeleteWithKey(context.Context, client.DocKey) (*client.DeleteResult, error) {
	return nil, nil
}

func (c *CollectionClient) DeleteWithKeys(context.Context, []client.DocKey) (*client.DeleteResult, error) {
	return nil, nil
}

func (c *CollectionClient) Get(ctx context.Context, key client.DocKey, showDeleted bool) (*client.Document, error) {
	return nil, nil
}

func (c *CollectionClient) WithTxn(datastore.Txn) client.Collection {
	return nil
}

func (c *CollectionClient) GetAllDocKeys(ctx context.Context) (<-chan client.DocKeysResult, error) {
	return nil, nil
}

func (c *CollectionClient) CreateIndex(context.Context, client.IndexDescription) (client.IndexDescription, error) {
	return client.IndexDescription{}, nil
}

func (c *CollectionClient) DropIndex(ctx context.Context, indexName string) error {
	return nil
}

func (c *CollectionClient) GetIndexes(ctx context.Context) ([]client.IndexDescription, error) {
	return c.description.Indexes, nil
}
