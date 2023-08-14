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

	"github.com/sourcenetwork/immutable/enumerable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
)

var _ client.LensRegistry = (*LensClient)(nil)

// LensClient implements the client.LensRegistry interface over HTTP.
type LensClient struct {
	client  *http.Client
	baseURL *url.URL
}

func NewLensClient(s *Client) *LensClient {
	return &LensClient{
		client:  s.client,
		baseURL: s.baseURL,
	}
}

func (c *LensClient) WithTxn(datastore.Txn) client.LensRegistry {
	return c
}

func (c *LensClient) SetMigration(ctx context.Context, config client.LensConfig) error {
	methodURL := c.baseURL.JoinPath("lens", "migrate")

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

func (c *LensClient) ReloadLenses(ctx context.Context) error {
	methodURL := c.baseURL.JoinPath("lens", "reload")

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

func (c *LensClient) MigrateUp(
	ctx context.Context,
	src enumerable.Enumerable[map[string]any],
	schemaVersionID string,
) (enumerable.Enumerable[map[string]any], error) {
	methodURL := c.baseURL.JoinPath("lens", schemaVersionID, "up")

	body, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() //nolint:errcheck

	var result enumerable.Enumerable[map[string]any]
	if err := parseJsonResponse(res, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *LensClient) MigrateDown(
	ctx context.Context,
	src enumerable.Enumerable[map[string]any],
	schemaVersionID string,
) (enumerable.Enumerable[map[string]any], error) {
	methodURL := c.baseURL.JoinPath("lens", schemaVersionID, "down")

	body, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() //nolint:errcheck

	var result enumerable.Enumerable[map[string]any]
	if err := parseJsonResponse(res, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *LensClient) Config(ctx context.Context) ([]client.LensConfig, error) {
	methodURL := c.baseURL.JoinPath("lens")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() //nolint:errcheck

	var cfgs []client.LensConfig
	if err := parseJsonResponse(res, cfgs); err != nil {
		return nil, err
	}
	return cfgs, nil
}

func (c *LensClient) HasMigration(ctx context.Context, schemaVersionID string) (bool, error) {
	methodURL := c.baseURL.JoinPath("lens", schemaVersionID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return false, err
	}
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
