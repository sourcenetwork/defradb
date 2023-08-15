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
	"strings"

	"github.com/sourcenetwork/defradb/client"
)

var _ client.Store = (*StoreClient)(nil)

type StoreClient struct {
	http *httpClient
}

func NewStoreClient(rawURL string) (*StoreClient, error) {
	baseURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	httpClient := &httpClient{
		client:  http.DefaultClient,
		baseURL: baseURL.JoinPath("/api/v0"),
	}
	return &StoreClient{httpClient}, nil
}

func (c *StoreClient) SetReplicator(ctx context.Context, rep client.Replicator) error {
	methodURL := c.http.baseURL.JoinPath("p2p", "replicators")

	body, err := json.Marshal(rep)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	return c.http.request(req)
}

func (c *StoreClient) DeleteReplicator(ctx context.Context, rep client.Replicator) error {
	methodURL := c.http.baseURL.JoinPath("p2p", "replicators")

	body, err := json.Marshal(rep)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	return c.http.request(req)
}

func (c *StoreClient) GetAllReplicators(ctx context.Context) ([]client.Replicator, error) {
	methodURL := c.http.baseURL.JoinPath("p2p", "replicators")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var reps []client.Replicator
	if err := c.http.requestJson(req, reps); err != nil {
		return nil, err
	}
	return reps, nil
}

func (c *StoreClient) AddP2PCollection(ctx context.Context, collectionID string) error {
	methodURL := c.http.baseURL.JoinPath("p2p", "collections", collectionID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), nil)
	if err != nil {
		return err
	}
	return c.http.request(req)
}

func (c *StoreClient) RemoveP2PCollection(ctx context.Context, collectionID string) error {
	methodURL := c.http.baseURL.JoinPath("p2p", "collections", collectionID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), nil)
	if err != nil {
		return err
	}
	return c.http.request(req)
}

func (c *StoreClient) GetAllP2PCollections(ctx context.Context) ([]string, error) {
	methodURL := c.http.baseURL.JoinPath("p2p", "collections")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var cols []string
	if err := c.http.requestJson(req, &cols); err != nil {
		return nil, err
	}
	return cols, nil
}

func (c *StoreClient) BasicImport(ctx context.Context, filepath string) error {
	methodURL := c.http.baseURL.JoinPath("backup", "import")

	body, err := json.Marshal(&client.BackupConfig{Filepath: filepath})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	return c.http.request(req)
}

func (c *StoreClient) BasicExport(ctx context.Context, config *client.BackupConfig) error {
	methodURL := c.http.baseURL.JoinPath("backup", "export")

	body, err := json.Marshal(config)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	return c.http.request(req)
}

func (c *StoreClient) AddSchema(ctx context.Context, schema string) ([]client.CollectionDescription, error) {
	methodURL := c.http.baseURL.JoinPath("schema")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), strings.NewReader(schema))
	if err != nil {
		return nil, err
	}
	var cols []client.CollectionDescription
	if err := c.http.requestJson(req, &cols); err != nil {
		return nil, err
	}
	return cols, nil
}

func (c *StoreClient) PatchSchema(ctx context.Context, patch string) error {
	methodURL := c.http.baseURL.JoinPath("schema")

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, methodURL.String(), strings.NewReader(patch))
	if err != nil {
		return err
	}
	return c.http.request(req)
}

func (c *StoreClient) SetMigration(ctx context.Context, config client.LensConfig) error {
	return c.LensRegistry().SetMigration(ctx, config)
}

func (c *StoreClient) LensRegistry() client.LensRegistry {
	return NewLensClient(c.http)
}

func (c *StoreClient) GetCollectionByName(ctx context.Context, name client.CollectionName) (client.Collection, error) {
	methodURL := c.http.baseURL.JoinPath("collections")
	methodURL.RawQuery = url.Values{"name": []string{name}}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var description client.CollectionDescription
	if err := c.http.requestJson(req, &description); err != nil {
		return nil, err
	}
	return NewCollectionClient(c.http, description), nil
}

func (c *StoreClient) GetCollectionBySchemaID(ctx context.Context, schemaId string) (client.Collection, error) {
	methodURL := c.http.baseURL.JoinPath("collections")
	methodURL.RawQuery = url.Values{"schema_id": []string{schemaId}}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var description client.CollectionDescription
	if err := c.http.requestJson(req, &description); err != nil {
		return nil, err
	}
	return NewCollectionClient(c.http, description), nil
}

func (c *StoreClient) GetCollectionByVersionID(ctx context.Context, versionId string) (client.Collection, error) {
	methodURL := c.http.baseURL.JoinPath("collections")
	methodURL.RawQuery = url.Values{"version_id": []string{versionId}}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var description client.CollectionDescription
	if err := c.http.requestJson(req, &description); err != nil {
		return nil, err
	}
	return NewCollectionClient(c.http, description), nil
}

func (c *StoreClient) GetAllCollections(ctx context.Context) ([]client.Collection, error) {
	methodURL := c.http.baseURL.JoinPath("collections")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var descriptions []client.CollectionDescription
	if err := c.http.requestJson(req, &descriptions); err != nil {
		return nil, err
	}
	collections := make([]client.Collection, len(descriptions))
	for i, d := range descriptions {
		collections[i] = NewCollectionClient(c.http, d)
	}
	return collections, nil
}

func (c *StoreClient) GetAllIndexes(ctx context.Context) (map[client.CollectionName][]client.IndexDescription, error) {
	methodURL := c.http.baseURL.JoinPath("indexes")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var indexes map[client.CollectionName][]client.IndexDescription
	if err := c.http.requestJson(req, &indexes); err != nil {
		return nil, err
	}
	return indexes, nil
}

func (c *StoreClient) ExecRequest(ctx context.Context, query string) (result *client.RequestResult) {
	methodURL := c.http.baseURL.JoinPath("graphql")
	result = &client.RequestResult{}

	body, err := json.Marshal(&GraphQLRequest{query})
	if err != nil {
		result.GQL.Errors = []error{err}
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		result.GQL.Errors = []error{err}
		return
	}
	var response GraphQLResponse
	if err = c.http.requestJson(req, &response); err != nil {
		result.GQL.Errors = []error{err}
		return
	}
	result.GQL.Data = response.Data
	for _, err := range response.Errors {
		result.GQL.Errors = append(result.GQL.Errors, fmt.Errorf(err))
	}
	return
}
