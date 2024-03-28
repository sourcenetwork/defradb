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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcenetwork/immutable"
	sse "github.com/vito/go-sse/sse"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/datastore"
)

var _ client.Collection = (*Collection)(nil)

// Collection implements the client.Collection interface over HTTP.
type Collection struct {
	http *httpClient
	def  client.CollectionDefinition
}

func (c *Collection) Description() client.CollectionDescription {
	return c.def.Description
}

func (c *Collection) Name() immutable.Option[string] {
	return c.Description().Name
}

func (c *Collection) Schema() client.SchemaDescription {
	return c.def.Schema
}

func (c *Collection) ID() uint32 {
	return c.Description().ID
}

func (c *Collection) SchemaRoot() string {
	return c.Schema().Root
}

func (c *Collection) Definition() client.CollectionDefinition {
	return c.def
}

func (c *Collection) Create(
	ctx context.Context,
	identity immutable.Option[string],
	doc *client.Document,
) error {
	if !c.Description().Name.HasValue() {
		return client.ErrOperationNotPermittedOnNamelessCols
	}

	methodURL := c.http.baseURL.JoinPath("collections", c.Description().Name.Value())

	body, err := doc.String()
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), strings.NewReader(body))
	if err != nil {
		return err
	}

	addIdentityToAuthHeader(req, identity)

	_, err = c.http.request(req)
	if err != nil {
		return err
	}
	doc.Clean()
	return nil
}

func (c *Collection) CreateMany(
	ctx context.Context,
	identity immutable.Option[string],
	docs []*client.Document,
) error {
	if !c.Description().Name.HasValue() {
		return client.ErrOperationNotPermittedOnNamelessCols
	}
	methodURL := c.http.baseURL.JoinPath("collections", c.Description().Name.Value())

	var docMapList []json.RawMessage
	for _, doc := range docs {
		docMap, err := doc.ToJSONPatch()
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

	addIdentityToAuthHeader(req, identity)

	_, err = c.http.request(req)
	if err != nil {
		return err
	}

	for _, doc := range docs {
		doc.Clean()
	}
	return nil
}

func (c *Collection) Update(
	ctx context.Context,
	identity immutable.Option[string],
	doc *client.Document,
) error {
	if !c.Description().Name.HasValue() {
		return client.ErrOperationNotPermittedOnNamelessCols
	}

	methodURL := c.http.baseURL.JoinPath("collections", c.Description().Name.Value(), doc.ID().String())

	body, err := doc.ToJSONPatch()
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	addIdentityToAuthHeader(req, identity)

	_, err = c.http.request(req)
	if err != nil {
		return err
	}
	doc.Clean()
	return nil
}

func (c *Collection) Save(
	ctx context.Context,
	identity immutable.Option[string],
	doc *client.Document,
) error {
	_, err := c.Get(ctx, identity, doc.ID(), true)
	if err == nil {
		return c.Update(ctx, identity, doc)
	}
	if errors.Is(err, client.ErrDocumentNotFound) {
		return c.Create(ctx, identity, doc)
	}
	return err
}

func (c *Collection) Delete(
	ctx context.Context,
	identity immutable.Option[string],
	docID client.DocID,
) (bool, error) {
	if !c.Description().Name.HasValue() {
		return false, client.ErrOperationNotPermittedOnNamelessCols
	}

	methodURL := c.http.baseURL.JoinPath("collections", c.Description().Name.Value(), docID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), nil)
	if err != nil {
		return false, err
	}

	addIdentityToAuthHeader(req, identity)

	_, err = c.http.request(req)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Collection) Exists(
	ctx context.Context,
	identity immutable.Option[string],
	docID client.DocID,
) (bool, error) {
	_, err := c.Get(ctx, identity, docID, false)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Collection) UpdateWith(
	ctx context.Context,
	identity immutable.Option[string],
	target any,
	updater string,
) (*client.UpdateResult, error) {
	switch t := target.(type) {
	case string, map[string]any, *request.Filter:
		return c.UpdateWithFilter(ctx, identity, t, updater)
	case client.DocID:
		return c.UpdateWithDocID(ctx, identity, t, updater)
	case []client.DocID:
		return c.UpdateWithDocIDs(ctx, identity, t, updater)
	default:
		return nil, client.ErrInvalidUpdateTarget
	}
}

func (c *Collection) updateWith(
	ctx context.Context,
	identity immutable.Option[string],
	request CollectionUpdateRequest,
) (*client.UpdateResult, error) {
	if !c.Description().Name.HasValue() {
		return nil, client.ErrOperationNotPermittedOnNamelessCols
	}

	methodURL := c.http.baseURL.JoinPath("collections", c.Description().Name.Value())

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	addIdentityToAuthHeader(req, identity)

	var result client.UpdateResult
	if err := c.http.requestJson(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Collection) UpdateWithFilter(
	ctx context.Context,
	identity immutable.Option[string],
	filter any,
	updater string,
) (*client.UpdateResult, error) {
	return c.updateWith(
		ctx,
		identity,
		CollectionUpdateRequest{
			Filter:  filter,
			Updater: updater,
		},
	)
}

func (c *Collection) UpdateWithDocID(
	ctx context.Context,
	identity immutable.Option[string],
	docID client.DocID,
	updater string,
) (*client.UpdateResult, error) {
	return c.updateWith(
		ctx,
		identity,
		CollectionUpdateRequest{
			DocID:   docID.String(),
			Updater: updater,
		},
	)
}

func (c *Collection) UpdateWithDocIDs(
	ctx context.Context,
	identity immutable.Option[string],
	docIDs []client.DocID,
	updater string,
) (*client.UpdateResult, error) {
	var strDocIDs []string
	for _, docID := range docIDs {
		strDocIDs = append(strDocIDs, docID.String())
	}
	return c.updateWith(
		ctx,
		identity,
		CollectionUpdateRequest{
			DocIDs:  strDocIDs,
			Updater: updater,
		},
	)
}

func (c *Collection) DeleteWith(
	ctx context.Context,
	identity immutable.Option[string],
	target any,
) (*client.DeleteResult, error) {
	switch t := target.(type) {
	case string, map[string]any, *request.Filter:
		return c.DeleteWithFilter(ctx, identity, t)
	case client.DocID:
		return c.DeleteWithDocID(ctx, identity, t)
	case []client.DocID:
		return c.DeleteWithDocIDs(ctx, identity, t)
	default:
		return nil, client.ErrInvalidDeleteTarget
	}
}

func (c *Collection) deleteWith(
	ctx context.Context,
	identity immutable.Option[string],
	request CollectionDeleteRequest,
) (*client.DeleteResult, error) {
	if !c.Description().Name.HasValue() {
		return nil, client.ErrOperationNotPermittedOnNamelessCols
	}

	methodURL := c.http.baseURL.JoinPath("collections", c.Description().Name.Value())

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	addIdentityToAuthHeader(req, identity)

	var result client.DeleteResult
	if err := c.http.requestJson(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Collection) DeleteWithFilter(
	ctx context.Context,
	identity immutable.Option[string],
	filter any,
) (*client.DeleteResult, error) {
	return c.deleteWith(
		ctx,
		identity,
		CollectionDeleteRequest{
			Filter: filter,
		},
	)
}

func (c *Collection) DeleteWithDocID(
	ctx context.Context,
	identity immutable.Option[string],
	docID client.DocID,
) (*client.DeleteResult, error) {
	return c.deleteWith(
		ctx,
		identity,
		CollectionDeleteRequest{
			DocID: docID.String(),
		},
	)
}

func (c *Collection) DeleteWithDocIDs(
	ctx context.Context,
	identity immutable.Option[string],
	docIDs []client.DocID,
) (*client.DeleteResult, error) {
	var strDocIDs []string
	for _, docID := range docIDs {
		strDocIDs = append(strDocIDs, docID.String())
	}
	return c.deleteWith(
		ctx,
		identity,
		CollectionDeleteRequest{
			DocIDs: strDocIDs,
		},
	)
}

func (c *Collection) Get(
	ctx context.Context,
	identity immutable.Option[string],
	docID client.DocID,
	showDeleted bool,
) (*client.Document, error) {
	if !c.Description().Name.HasValue() {
		return nil, client.ErrOperationNotPermittedOnNamelessCols
	}

	query := url.Values{}
	if showDeleted {
		query.Add("show_deleted", "true")
	}

	methodURL := c.http.baseURL.JoinPath("collections", c.Description().Name.Value(), docID.String())
	methodURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}

	addIdentityToAuthHeader(req, identity)

	data, err := c.http.request(req)
	if err != nil {
		return nil, err
	}
	doc := client.NewDocWithID(docID, c.def.Schema)
	err = doc.SetWithJSON(data)
	if err != nil {
		return nil, err
	}
	doc.Clean()
	return doc, nil
}

func (c *Collection) WithTxn(tx datastore.Txn) client.Collection {
	return &Collection{
		http: c.http.withTxn(tx.ID()),
		def:  c.def,
	}
}

func (c *Collection) GetAllDocIDs(ctx context.Context) (<-chan client.DocIDResult, error) {
	if !c.Description().Name.HasValue() {
		return nil, client.ErrOperationNotPermittedOnNamelessCols
	}

	methodURL := c.http.baseURL.JoinPath("collections", c.Description().Name.Value())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	c.http.setDefaultHeaders(req)

	res, err := c.http.client.Do(req)
	if err != nil {
		return nil, err
	}
	docIDCh := make(chan client.DocIDResult)

	go func() {
		eventReader := sse.NewReadCloser(res.Body)
		// ignore close errors because the status
		// and body of the request are already
		// checked and it cannot be handled properly
		defer eventReader.Close() //nolint:errcheck
		defer close(docIDCh)

		for {
			evt, err := eventReader.Next()
			if err != nil {
				return
			}
			var res DocIDResult
			if err := json.Unmarshal(evt.Data, &res); err != nil {
				return
			}
			docID, err := client.NewDocIDFromString(res.DocID)
			if err != nil {
				return
			}
			docIDResult := client.DocIDResult{
				ID: docID,
			}
			if res.Error != "" {
				docIDResult.Err = fmt.Errorf(res.Error)
			}
			docIDCh <- docIDResult
		}
	}()

	return docIDCh, nil
}

func (c *Collection) CreateIndex(
	ctx context.Context,
	indexDesc client.IndexDescription,
) (client.IndexDescription, error) {
	if !c.Description().Name.HasValue() {
		return client.IndexDescription{}, client.ErrOperationNotPermittedOnNamelessCols
	}

	methodURL := c.http.baseURL.JoinPath("collections", c.Description().Name.Value(), "indexes")

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
	if !c.Description().Name.HasValue() {
		return client.ErrOperationNotPermittedOnNamelessCols
	}

	methodURL := c.http.baseURL.JoinPath("collections", c.Description().Name.Value(), "indexes", indexName)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), nil)
	if err != nil {
		return err
	}
	_, err = c.http.request(req)
	return err
}

func (c *Collection) GetIndexes(ctx context.Context) ([]client.IndexDescription, error) {
	if !c.Description().Name.HasValue() {
		return nil, client.ErrOperationNotPermittedOnNamelessCols
	}

	methodURL := c.http.baseURL.JoinPath("collections", c.Description().Name.Value(), "indexes")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var indexes []client.IndexDescription
	if err := c.http.requestJson(req, &indexes); err != nil {
		return nil, err
	}
	return indexes, nil
}

func (c *Collection) CreateDocIndex(context.Context, *client.Document) error {
	return ErrMethodIsNotImplemented
}

func (c *Collection) UpdateDocIndex(ctx context.Context, oldDoc, newDoc *client.Document) error {
	return ErrMethodIsNotImplemented
}

func (c *Collection) DeleteDocIndex(context.Context, *client.Document) error {
	return ErrMethodIsNotImplemented
}
