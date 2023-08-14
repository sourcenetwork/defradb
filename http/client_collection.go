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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/datastore"
)

var _ client.Collection = (*CollectionClient)(nil)

// CollectionClient implements the client.Collection interface over HTTP.
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

func (c *CollectionClient) Create(ctx context.Context, doc *client.Document) error {
	methodURL := c.baseURL.JoinPath("collections", c.description.Name)

	docMap, err := doc.ToMap()
	if err != nil {
		return err
	}
	body, err := json.Marshal(docMap)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint:errcheck

	return parseResponse(res)
}

func (c *CollectionClient) CreateMany(ctx context.Context, docs []*client.Document) error {
	methodURL := c.baseURL.JoinPath("collections", c.description.Name)

	var docMapList []map[string]any
	for _, doc := range docs {
		docMap, err := doc.ToMap()
		if err != nil {
			return err
		}
		docMapList = append(docMapList, docMap)
	}
	body, err := json.Marshal(docMapList)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint:errcheck

	return parseResponse(res)
}

func (c *CollectionClient) Update(ctx context.Context, doc *client.Document) error {
	methodURL := c.baseURL.JoinPath("collections", c.description.Name, doc.Key().String())

	docMap, err := doc.ToMap()
	if err != nil {
		return err
	}
	body, err := json.Marshal(docMap)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint:errcheck

	return parseResponse(res)
}

func (c *CollectionClient) Save(ctx context.Context, doc *client.Document) error {
	methodURL := c.baseURL.JoinPath("collections", c.description.Name, doc.Key().String())

	docMap, err := doc.ToMap()
	if err != nil {
		return err
	}
	body, err := json.Marshal(docMap)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint:errcheck

	return parseResponse(res)
}

func (c *CollectionClient) Delete(ctx context.Context, docKey client.DocKey) (bool, error) {
	methodURL := c.baseURL.JoinPath("collections", c.description.Name, docKey.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), nil)
	if err != nil {
		return false, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close() //nolint:errcheck

	err = parseResponse(res)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *CollectionClient) Exists(ctx context.Context, docKey client.DocKey) (bool, error) {
	_, err := c.Get(ctx, docKey, false)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *CollectionClient) UpdateWith(ctx context.Context, target any, updater string) (*client.UpdateResult, error) {
	switch t := target.(type) {
	case string, map[string]any, *request.Filter:
		return c.UpdateWithFilter(ctx, t, updater)
	case client.DocKey:
		return c.UpdateWithKey(ctx, t, updater)
	case []client.DocKey:
		return c.UpdateWithKeys(ctx, t, updater)
	default:
		return nil, client.ErrInvalidUpdateTarget
	}
}

func (c *CollectionClient) updateWith(
	ctx context.Context,
	request CollectionUpdateRequest,
) (*client.UpdateResult, error) {
	methodURL := c.baseURL.JoinPath("collections", c.description.Name)

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() //nolint:errcheck

	var result client.UpdateResult
	if err := parseJsonResponse(res, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *CollectionClient) UpdateWithFilter(
	ctx context.Context,
	filter any,
	updater string,
) (*client.UpdateResult, error) {
	return c.updateWith(ctx, CollectionUpdateRequest{
		Filter:  filter,
		Updater: updater,
	})
}

func (c *CollectionClient) UpdateWithKey(
	ctx context.Context,
	key client.DocKey,
	updater string,
) (*client.UpdateResult, error) {
	return c.updateWith(ctx, CollectionUpdateRequest{
		Key:     key.String(),
		Updater: updater,
	})
}

func (c *CollectionClient) UpdateWithKeys(
	ctx context.Context,
	docKeys []client.DocKey,
	updater string,
) (*client.UpdateResult, error) {
	var keys []string
	for _, key := range docKeys {
		keys = append(keys, key.String())
	}
	return c.updateWith(ctx, CollectionUpdateRequest{
		Keys:    keys,
		Updater: updater,
	})
}

func (c *CollectionClient) DeleteWith(ctx context.Context, target any) (*client.DeleteResult, error) {
	switch t := target.(type) {
	case string, map[string]any, *request.Filter:
		return c.DeleteWithFilter(ctx, t)
	case client.DocKey:
		return c.DeleteWithKey(ctx, t)
	case []client.DocKey:
		return c.DeleteWithKeys(ctx, t)
	default:
		return nil, client.ErrInvalidDeleteTarget
	}
}

func (c *CollectionClient) deleteWith(
	ctx context.Context,
	request CollectionDeleteRequest,
) (*client.DeleteResult, error) {
	methodURL := c.baseURL.JoinPath("collections", c.description.Name)

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() //nolint:errcheck

	var result client.DeleteResult
	if err := parseJsonResponse(res, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *CollectionClient) DeleteWithFilter(ctx context.Context, filter any) (*client.DeleteResult, error) {
	return c.deleteWith(ctx, CollectionDeleteRequest{
		Filter: filter,
	})
}

func (c *CollectionClient) DeleteWithKey(ctx context.Context, docKey client.DocKey) (*client.DeleteResult, error) {
	return c.deleteWith(ctx, CollectionDeleteRequest{
		Key: docKey.String(),
	})
}

func (c *CollectionClient) DeleteWithKeys(ctx context.Context, docKeys []client.DocKey) (*client.DeleteResult, error) {
	var keys []string
	for _, key := range docKeys {
		keys = append(keys, key.String())
	}
	return c.deleteWith(ctx, CollectionDeleteRequest{
		Keys: keys,
	})
}

func (c *CollectionClient) Get(ctx context.Context, key client.DocKey, showDeleted bool) (*client.Document, error) {
	methodURL := c.baseURL.JoinPath("collections", c.description.Name, key.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() //nolint:errcheck

	var docMap map[string]any
	if err := parseJsonResponse(res, docMap); err != nil {
		return nil, err
	}
	return client.NewDocFromMap(docMap)
}

func (c *CollectionClient) WithTxn(datastore.Txn) client.Collection {
	return c
}

func (c *CollectionClient) GetAllDocKeys(ctx context.Context) (<-chan client.DocKeysResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *CollectionClient) CreateIndex(
	ctx context.Context,
	indexDesc client.IndexDescription,
) (client.IndexDescription, error) {
	methodURL := c.baseURL.JoinPath("collections", c.description.Name, "indexes")

	body, err := json.Marshal(&indexDesc)
	if err != nil {
		return client.IndexDescription{}, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return client.IndexDescription{}, nil
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return client.IndexDescription{}, nil
	}
	defer res.Body.Close() //nolint:errcheck

	var index client.IndexDescription
	if err := parseJsonResponse(res, &index); err != nil {
		return client.IndexDescription{}, nil
	}
	return index, nil
}

func (c *CollectionClient) DropIndex(ctx context.Context, indexName string) error {
	methodURL := c.baseURL.JoinPath("collections", c.description.Name, "indexes", indexName)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint:errcheck

	return parseResponse(res)
}

func (c *CollectionClient) GetIndexes(ctx context.Context) ([]client.IndexDescription, error) {
	methodURL := c.baseURL.JoinPath("collections", c.description.Name, "indexes")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() //nolint:errcheck

	var indexes []client.IndexDescription
	if err := parseJsonResponse(res, indexes); err != nil {
		return nil, err
	}
	return c.description.Indexes, nil
}
