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
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcenetwork/defradb/client"
)

var _ client.Store = (*Client)(nil)

// Client implements the client.Store interface over HTTP.
type Client struct {
	client  *http.Client
	baseURL *url.URL
}

func NewClient(rawURL string) (*Client, error) {
	baseURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return &Client{
		client:  http.DefaultClient,
		baseURL: baseURL.JoinPath("/api/v0"),
	}, nil
}

func (c *Client) SetReplicator(ctx context.Context, rep client.Replicator) error {
	methodURL := c.baseURL.JoinPath("p2p", "replicators")

	body, err := json.Marshal(rep)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint:errcheck

	return parseResponse(res)
}

func (c *Client) DeleteReplicator(ctx context.Context, rep client.Replicator) error {
	methodURL := c.baseURL.JoinPath("p2p", "replicators")

	body, err := json.Marshal(rep)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint:errcheck

	return parseResponse(res)
}

func (c *Client) GetAllReplicators(ctx context.Context) ([]client.Replicator, error) {
	methodURL := c.baseURL.JoinPath("p2p", "replicators")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() //nolint:errcheck

	var reps []client.Replicator
	if err := parseJsonResponse(res, &reps); err != nil {
		return nil, err
	}
	return reps, nil
}

func (c *Client) AddP2PCollection(ctx context.Context, collectionID string) error {
	methodURL := c.baseURL.JoinPath("p2p", "collections", collectionID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), nil)
	if err != nil {
		return err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint:errcheck

	return parseResponse(res)
}

func (c *Client) RemoveP2PCollection(ctx context.Context, collectionID string) error {
	methodURL := c.baseURL.JoinPath("p2p", "collections", collectionID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), nil)
	if err != nil {
		return err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint:errcheck

	return parseResponse(res)
}

func (c *Client) GetAllP2PCollections(ctx context.Context) ([]string, error) {
	methodURL := c.baseURL.JoinPath("p2p", "collections")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() //nolint:errcheck

	var cols []string
	if err := parseJsonResponse(res, &cols); err != nil {
		return nil, err
	}
	return cols, nil
}

func (c *Client) BasicImport(ctx context.Context, filepath string) error {
	methodURL := c.baseURL.JoinPath("backup", "import")

	body, err := json.Marshal(&client.BackupConfig{Filepath: filepath})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint:errcheck

	return parseResponse(res)
}

func (c *Client) BasicExport(ctx context.Context, config *client.BackupConfig) error {
	methodURL := c.baseURL.JoinPath("backup", "export")

	body, err := json.Marshal(config)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint:errcheck

	return parseResponse(res)
}

func (c *Client) AddSchema(ctx context.Context, schema string) ([]client.CollectionDescription, error) {
	methodURL := c.baseURL.JoinPath("schema")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), strings.NewReader(schema))
	if err != nil {
		return nil, err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() //nolint:errcheck

	var cols []client.CollectionDescription
	if err := parseJsonResponse(res, &cols); err != nil {
		return nil, err
	}
	return cols, nil
}

func (c *Client) PatchSchema(ctx context.Context, patch string) error {
	methodURL := c.baseURL.JoinPath("schema")

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, methodURL.String(), strings.NewReader(patch))
	if err != nil {
		return err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint:errcheck

	return parseResponse(res)
}

func (c *Client) SetMigration(ctx context.Context, config client.LensConfig) error {
	return c.LensRegistry().SetMigration(ctx, config)
}

func (c *Client) LensRegistry() client.LensRegistry {
	return NewLensClient(c)
}

func (c *Client) GetCollectionByName(ctx context.Context, name client.CollectionName) (client.Collection, error) {
	methodURL := c.baseURL.JoinPath("collections")
	methodURL.RawQuery = url.Values{"name": []string{name}}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() //nolint:errcheck

	var description client.CollectionDescription
	if err := parseJsonResponse(res, &description); err != nil {
		return nil, err
	}
	return NewCollectionClient(c, description), nil
}

func (c *Client) GetCollectionBySchemaID(ctx context.Context, schemaId string) (client.Collection, error) {
	methodURL := c.baseURL.JoinPath("collections")
	methodURL.RawQuery = url.Values{"schema_id": []string{schemaId}}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() //nolint:errcheck

	var description client.CollectionDescription
	if err := parseJsonResponse(res, &description); err != nil {
		return nil, err
	}
	return NewCollectionClient(c, description), nil
}

func (c *Client) GetCollectionByVersionID(ctx context.Context, versionId string) (client.Collection, error) {
	methodURL := c.baseURL.JoinPath("collections")
	methodURL.RawQuery = url.Values{"version_id": []string{versionId}}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() //nolint:errcheck

	var description client.CollectionDescription
	if err := parseJsonResponse(res, &description); err != nil {
		return nil, err
	}
	return NewCollectionClient(c, description), nil
}

func (c *Client) GetAllCollections(ctx context.Context) ([]client.Collection, error) {
	methodURL := c.baseURL.JoinPath("collections")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() //nolint:errcheck

	var descriptions []client.CollectionDescription
	if err := parseJsonResponse(res, &descriptions); err != nil {
		return nil, err
	}
	collections := make([]client.Collection, len(descriptions))
	for i, d := range descriptions {
		collections[i] = NewCollectionClient(c, d)
	}
	return collections, nil
}

func (c *Client) GetAllIndexes(ctx context.Context) (map[client.CollectionName][]client.IndexDescription, error) {
	methodURL := c.baseURL.JoinPath("indexes")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() //nolint:errcheck

	var indexes map[client.CollectionName][]client.IndexDescription
	if err := parseJsonResponse(res, &indexes); err != nil {
		return nil, err
	}
	return indexes, nil
}

func (c *Client) ExecRequest(context.Context, string) *client.RequestResult {
	return nil
}
