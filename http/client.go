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
	"io"
	"net/http"
	"net/url"
	"strings"

	blockstore "github.com/ipfs/boxo/blockstore"
	sse "github.com/vito/go-sse/sse"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/events"
)

var _ client.DB = (*Client)(nil)

// Client implements the client.DB interface over HTTP.
type Client struct {
	http *httpClient
}

func NewClient(rawURL string) (*Client, error) {
	httpClient, err := newHttpClient(rawURL)
	if err != nil {
		return nil, err
	}
	return &Client{httpClient}, nil
}

func (c *Client) NewTxn(ctx context.Context, readOnly bool) (datastore.Txn, error) {
	query := url.Values{}
	if readOnly {
		query.Add("read_only", "true")
	}

	methodURL := c.http.baseURL.JoinPath("tx")
	methodURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var txRes CreateTxResponse
	if err := c.http.requestJson(req, &txRes); err != nil {
		return nil, err
	}
	return &Transaction{txRes.ID, c.http}, nil
}

func (c *Client) NewConcurrentTxn(ctx context.Context, readOnly bool) (datastore.Txn, error) {
	query := url.Values{}
	if readOnly {
		query.Add("read_only", "true")
	}

	methodURL := c.http.baseURL.JoinPath("tx", "concurrent")
	methodURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var txRes CreateTxResponse
	if err := c.http.requestJson(req, &txRes); err != nil {
		return nil, err
	}
	return &Transaction{txRes.ID, c.http}, nil
}

func (c *Client) WithTxn(tx datastore.Txn) client.Store {
	client := c.http.withTxn(tx.ID())
	return &Client{client}
}

func (c *Client) BasicImport(ctx context.Context, filepath string) error {
	methodURL := c.http.baseURL.JoinPath("backup", "import")

	body, err := json.Marshal(&client.BackupConfig{Filepath: filepath})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	_, err = c.http.request(req)
	return err
}

func (c *Client) BasicExport(ctx context.Context, config *client.BackupConfig) error {
	methodURL := c.http.baseURL.JoinPath("backup", "export")

	body, err := json.Marshal(config)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	_, err = c.http.request(req)
	return err
}

func (c *Client) AddSchema(ctx context.Context, schema string) ([]client.CollectionDescription, error) {
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

type patchSchemaRequest struct {
	Patch               string
	SetAsDefaultVersion bool
}

func (c *Client) PatchSchema(ctx context.Context, patch string, setAsDefaultVersion bool) error {
	methodURL := c.http.baseURL.JoinPath("schema")

	body, err := json.Marshal(patchSchemaRequest{patch, setAsDefaultVersion})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	_, err = c.http.request(req)
	return err
}

func (c *Client) SetDefaultSchemaVersion(ctx context.Context, schemaVersionID string) error {
	methodURL := c.http.baseURL.JoinPath("schema", "default")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), strings.NewReader(schemaVersionID))
	if err != nil {
		return err
	}
	_, err = c.http.request(req)
	return err
}

func (c *Client) SetMigration(ctx context.Context, config client.LensConfig) error {
	return c.LensRegistry().SetMigration(ctx, config)
}

func (c *Client) LensRegistry() client.LensRegistry {
	return &LensRegistry{c.http}
}

func (c *Client) GetCollectionByName(ctx context.Context, name client.CollectionName) (client.Collection, error) {
	methodURL := c.http.baseURL.JoinPath("collections")
	methodURL.RawQuery = url.Values{"name": []string{name}}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var definition client.CollectionDefinition
	if err := c.http.requestJson(req, &definition); err != nil {
		return nil, err
	}
	return &Collection{c.http, definition}, nil
}

func (c *Client) GetCollectionsBySchemaRoot(ctx context.Context, schemaRoot string) ([]client.Collection, error) {
	methodURL := c.http.baseURL.JoinPath("collections")
	methodURL.RawQuery = url.Values{"schema_root": []string{schemaRoot}}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var descriptions []client.CollectionDefinition
	if err := c.http.requestJson(req, &descriptions); err != nil {
		return nil, err
	}
	collections := make([]client.Collection, len(descriptions))
	for i, d := range descriptions {
		collections[i] = &Collection{c.http, d}
	}
	return collections, nil
}

func (c *Client) GetCollectionsByVersionID(ctx context.Context, versionId string) ([]client.Collection, error) {
	methodURL := c.http.baseURL.JoinPath("collections")
	methodURL.RawQuery = url.Values{"version_id": []string{versionId}}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var descriptions []client.CollectionDefinition
	if err := c.http.requestJson(req, &descriptions); err != nil {
		return nil, err
	}
	collections := make([]client.Collection, len(descriptions))
	for i, d := range descriptions {
		collections[i] = &Collection{c.http, d}
	}
	return collections, nil
}

func (c *Client) GetAllCollections(ctx context.Context) ([]client.Collection, error) {
	methodURL := c.http.baseURL.JoinPath("collections")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var descriptions []client.CollectionDefinition
	if err := c.http.requestJson(req, &descriptions); err != nil {
		return nil, err
	}
	collections := make([]client.Collection, len(descriptions))
	for i, d := range descriptions {
		collections[i] = &Collection{c.http, d}
	}
	return collections, nil
}

func (c *Client) GetSchemaByVersionID(ctx context.Context, versionID string) (client.SchemaDescription, error) {
	methodURL := c.http.baseURL.JoinPath("schema")
	methodURL.RawQuery = url.Values{"version_id": []string{versionID}}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return client.SchemaDescription{}, err
	}
	var schema client.SchemaDescription
	if err := c.http.requestJson(req, &schema); err != nil {
		return client.SchemaDescription{}, err
	}
	return schema, nil
}

func (c *Client) GetAllSchema(ctx context.Context) ([]client.SchemaDescription, error) {
	methodURL := c.http.baseURL.JoinPath("schema")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var schema []client.SchemaDescription
	if err := c.http.requestJson(req, &schema); err != nil {
		return nil, err
	}
	return schema, nil
}

func (c *Client) GetAllIndexes(ctx context.Context) (map[client.CollectionName][]client.IndexDescription, error) {
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

func (c *Client) ExecRequest(ctx context.Context, query string) *client.RequestResult {
	methodURL := c.http.baseURL.JoinPath("graphql")
	result := &client.RequestResult{}

	body, err := json.Marshal(&GraphQLRequest{query})
	if err != nil {
		result.GQL.Errors = []error{err}
		return result
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		result.GQL.Errors = []error{err}
		return result
	}
	c.http.setDefaultHeaders(req)

	res, err := c.http.client.Do(req)
	if err != nil {
		result.GQL.Errors = []error{err}
		return result
	}
	if res.Header.Get("Content-Type") == "text/event-stream" {
		result.Pub = c.execRequestSubscription(ctx, res.Body)
		return result
	}
	// ignore close errors because they have
	// no perceivable effect on the end user
	// and cannot be reconciled easily
	defer res.Body.Close() //nolint:errcheck

	data, err := io.ReadAll(res.Body)
	if err != nil {
		result.GQL.Errors = []error{err}
		return result
	}
	var response GraphQLResponse
	if err = json.Unmarshal(data, &response); err != nil {
		result.GQL.Errors = []error{err}
		return result
	}
	result.GQL.Data = response.Data
	result.GQL.Errors = response.Errors
	return result
}

func (c *Client) execRequestSubscription(ctx context.Context, r io.ReadCloser) *events.Publisher[events.Update] {
	pubCh := events.New[events.Update](0, 0)
	pub, err := events.NewPublisher[events.Update](pubCh, 0)
	if err != nil {
		return nil
	}

	go func() {
		eventReader := sse.NewReadCloser(r)
		// ignore close errors because the status
		// and body of the request are already
		// checked and it cannot be handled properly
		defer eventReader.Close() //nolint:errcheck

		for {
			evt, err := eventReader.Next()
			if err != nil {
				return
			}
			var response GraphQLResponse
			if err := json.Unmarshal(evt.Data, &response); err != nil {
				return
			}
			pub.Publish(client.GQLResult{
				Errors: response.Errors,
				Data:   response.Data,
			})
		}
	}()

	return pub
}

func (c *Client) PrintDump(ctx context.Context) error {
	methodURL := c.http.baseURL.JoinPath("debug", "dump")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return err
	}
	_, err = c.http.request(req)
	return err
}

func (c *Client) Close() {
	// do nothing
}

func (c *Client) Root() datastore.RootStore {
	panic("client side database")
}

func (c *Client) Blockstore() blockstore.Blockstore {
	panic("client side database")
}

func (c *Client) Peerstore() datastore.DSBatching {
	panic("client side database")
}

func (c *Client) Events() events.Events {
	panic("client side database")
}

func (c *Client) MaxTxnRetries() int {
	panic("client side database")
}
