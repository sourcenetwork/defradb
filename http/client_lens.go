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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/immutable/enumerable"
)

var _ client.LensRegistry = (*LensClient)(nil)

type LensClient struct {
	client  *http.Client
	baseURL *url.URL
}

func NewLensClient(s *StoreClient) *LensClient {
	return &LensClient{
		client:  s.client,
		baseURL: s.baseURL,
	}
}

func (c *LensClient) WithTxn(datastore.Txn) client.LensRegistry {
	return nil
}

func (c *LensClient) SetMigration(ctx context.Context, config client.LensConfig) error {
	url := c.baseURL.JoinPath("lens", "migration").String()

	body, err := json.Marshal(config)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return parseResponse(res)
}

func (c *LensClient) ReloadLenses(context.Context) error {
	return nil
}

func (c *LensClient) MigrateUp(context.Context, enumerable.Enumerable[map[string]any], string) (enumerable.Enumerable[map[string]any], error) {
	return nil, nil
}

func (c *LensClient) MigrateDown(context.Context, enumerable.Enumerable[map[string]any], string) (enumerable.Enumerable[map[string]any], error) {
	return nil, nil
}

func (c *LensClient) Config(context.Context) ([]client.LensConfig, error) {
	return nil, nil
}

func (c *LensClient) HasMigration(context.Context, string) (bool, error) {
	return false, nil
}
