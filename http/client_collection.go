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
	"strings"

	sse "github.com/vito/go-sse/sse"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/datastore"
)

var _ client.Collection = (*Collection)(nil)

// Collection implements the client.Collection interface over HTTP.
type Collection struct {
	http *httpClient
	desc client.CollectionDescription
}

func (c *Collection) Description() client.CollectionDescription {
	return c.desc
}

func (c *Collection) Name() string {
	return c.desc.Name
}

func (c *Collection) Schema() client.SchemaDescription {
	return c.desc.Schema
}

func (c *Collection) ID() uint32 {
	return c.desc.ID
}

func (c *Collection) SchemaID() string {
	return c.desc.Schema.SchemaID
}

func (c *Collection) Create(ctx context.Context, doc *client.Document) error {
	methodURL := c.http.baseURL.JoinPath("collections", c.desc.Name)

	body, err := doc.String()
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), strings.NewReader(body))
	if err != nil {
		return err
	}
	_, err = c.http.request(req)
	if err != nil {
		return err
	}
	doc.Clean()
	return nil
}

func (c *Collection) CreateMany(ctx context.Context, docs []*client.Document) error {
	methodURL := c.http.baseURL.JoinPath("collections", c.desc.Name)

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
	_, err = c.http.request(req)
	if err != nil {
		return err
	}
	for _, doc := range docs {
		doc.Clean()
	}
	return nil
}

func (c *Collection) Update(ctx context.Context, doc *client.Document) error {
	methodURL := c.http.baseURL.JoinPath("collections", c.desc.Name, doc.Key().String())

	docMap, err := doc.ToMap()
	if err != nil {
		return err
	}
	for field, value := range doc.Values() {
		if !value.IsDirty() {
			delete(docMap, field.Name())
		}
	}
	body, err := json.Marshal(docMap)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	_, err = c.http.request(req)
	if err != nil {
		return err
	}
	doc.Clean()
	return nil
}

func (c *Collection) Save(ctx context.Context, doc *client.Document) error {
	methodURL := c.http.baseURL.JoinPath("collections", c.desc.Name, doc.Key().String())

	docMap, err := doc.ToMap()
	if err != nil {
		return err
	}
	for field, value := range doc.Values() {
		if !value.IsDirty() {
			delete(docMap, field.Name())
		}
	}
	body, err := json.Marshal(docMap)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	_, err = c.http.request(req)
	if err != nil {
		return err
	}
	doc.Clean()
	return nil
}

func (c *Collection) Delete(ctx context.Context, docKey client.DocKey) (bool, error) {
	methodURL := c.http.baseURL.JoinPath("collections", c.desc.Name, docKey.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), nil)
	if err != nil {
		return false, err
	}
	_, err = c.http.request(req)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Collection) Exists(ctx context.Context, docKey client.DocKey) (bool, error) {
	_, err := c.Get(ctx, docKey, false)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Collection) UpdateWith(ctx context.Context, target any, updater string) (*client.UpdateResult, error) {
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

func (c *Collection) updateWith(
	ctx context.Context,
	request CollectionUpdateRequest,
) (*client.UpdateResult, error) {
	methodURL := c.http.baseURL.JoinPath("collections", c.desc.Name)

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	var result client.UpdateResult
	if err := c.http.requestJson(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Collection) UpdateWithFilter(
	ctx context.Context,
	filter any,
	updater string,
) (*client.UpdateResult, error) {
	return c.updateWith(ctx, CollectionUpdateRequest{
		Filter:  filter,
		Updater: updater,
	})
}

func (c *Collection) UpdateWithKey(
	ctx context.Context,
	key client.DocKey,
	updater string,
) (*client.UpdateResult, error) {
	return c.updateWith(ctx, CollectionUpdateRequest{
		Key:     key.String(),
		Updater: updater,
	})
}

func (c *Collection) UpdateWithKeys(
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

func (c *Collection) DeleteWith(ctx context.Context, target any) (*client.DeleteResult, error) {
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

func (c *Collection) deleteWith(
	ctx context.Context,
	request CollectionDeleteRequest,
) (*client.DeleteResult, error) {
	methodURL := c.http.baseURL.JoinPath("collections", c.desc.Name)

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	var result client.DeleteResult
	if err := c.http.requestJson(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Collection) DeleteWithFilter(ctx context.Context, filter any) (*client.DeleteResult, error) {
	return c.deleteWith(ctx, CollectionDeleteRequest{
		Filter: filter,
	})
}

func (c *Collection) DeleteWithKey(ctx context.Context, docKey client.DocKey) (*client.DeleteResult, error) {
	return c.deleteWith(ctx, CollectionDeleteRequest{
		Key: docKey.String(),
	})
}

func (c *Collection) DeleteWithKeys(ctx context.Context, docKeys []client.DocKey) (*client.DeleteResult, error) {
	var keys []string
	for _, key := range docKeys {
		keys = append(keys, key.String())
	}
	return c.deleteWith(ctx, CollectionDeleteRequest{
		Keys: keys,
	})
}

func (c *Collection) Get(ctx context.Context, key client.DocKey, showDeleted bool) (*client.Document, error) {
	methodURL := c.http.baseURL.JoinPath("collections", c.desc.Name, key.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var docMap map[string]any
	if err := c.http.requestJson(req, &docMap); err != nil {
		return nil, err
	}
	return client.NewDocFromMap(docMap)
}

func (c *Collection) WithTxn(tx datastore.Txn) client.Collection {
	return &Collection{
		http: c.http.withTxn(tx.ID()),
		desc: c.desc,
	}
}

func (c *Collection) GetAllDocKeys(ctx context.Context) (<-chan client.DocKeysResult, error) {
	methodURL := c.http.baseURL.JoinPath("collections", c.desc.Name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	c.http.setDefaultHeaders(req)

	res, err := c.http.client.Do(req)
	if err != nil {
		return nil, err
	}
	docKeyCh := make(chan client.DocKeysResult)

	go func() {
		eventReader := sse.NewReadCloser(res.Body)
		// ignore close errors because the headers
		// and body are already written and we
		// cannot do anything useful with the error
		defer eventReader.Close() //nolint:errcheck
		defer close(docKeyCh)

		for {
			evt, err := eventReader.Next()
			if err != nil {
				return
			}
			var res DocKeyResult
			if err := json.Unmarshal(evt.Data, &res); err != nil {
				return
			}
			key, err := client.NewDocKeyFromString(res.Key)
			if err != nil {
				return
			}
			docKey := client.DocKeysResult{
				Key: key,
			}
			if res.Error != "" {
				docKey.Err = fmt.Errorf(res.Error)
			}
			docKeyCh <- docKey
		}
	}()

	return docKeyCh, nil
}

func (c *Collection) CreateIndex(
	ctx context.Context,
	indexDesc client.IndexDescription,
) (client.IndexDescription, error) {
	methodURL := c.http.baseURL.JoinPath("collections", c.desc.Name, "indexes")

	body, err := json.Marshal(&indexDesc)
	if err != nil {
		return client.IndexDescription{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return client.IndexDescription{}, err
	}
	var index client.IndexDescription
	if err := c.http.requestJson(req, &index); err != nil {
		return client.IndexDescription{}, err
	}
	return index, nil
}

func (c *Collection) DropIndex(ctx context.Context, indexName string) error {
	methodURL := c.http.baseURL.JoinPath("collections", c.desc.Name, "indexes", indexName)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), nil)
	if err != nil {
		return err
	}
	_, err = c.http.request(req)
	return err
}

func (c *Collection) GetIndexes(ctx context.Context) ([]client.IndexDescription, error) {
	methodURL := c.http.baseURL.JoinPath("collections", c.desc.Name, "indexes")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var indexes []client.IndexDescription
	if err := c.http.requestJson(req, &indexes); err != nil {
		return nil, err
	}
	return c.desc.Indexes, nil
}
