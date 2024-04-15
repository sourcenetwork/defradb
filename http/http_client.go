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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcenetwork/defradb/db"
)

type httpClient struct {
	client  *http.Client
	baseURL *url.URL
}

func newHttpClient(rawURL string) (*httpClient, error) {
	if !strings.HasPrefix(rawURL, "http") {
		rawURL = "http://" + rawURL
	}
	baseURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	client := httpClient{
		client:  http.DefaultClient,
		baseURL: baseURL.JoinPath("/api/v0"),
	}
	return &client, nil
}

func (c *httpClient) setDefaultHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	txn, ok := db.TryGetContextTxn(req.Context())
	if ok {
		req.Header.Set(txHeaderName, fmt.Sprintf("%d", txn.ID()))
	}
	id, ok := db.TryGetContextIdentity(req.Context())
	if ok && id.HasValue() {
		req.Header.Add(authHeaderName, authSchemaPrefix+id.Value())
	}
}

func (c *httpClient) request(req *http.Request) ([]byte, error) {
	c.setDefaultHeaders(req)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	// ignore close errors because they have
	// no perceivable effect on the end user
	// and cannot be reconciled easily
	defer res.Body.Close() //nolint:errcheck

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	// request was successful
	if res.StatusCode == http.StatusOK {
		return data, nil
	}
	// attempt to parse json error
	var errRes errorResponse
	if err := json.Unmarshal(data, &errRes); err != nil {
		return nil, fmt.Errorf("%s", data)
	}
	return nil, errRes.Error
}

func (c *httpClient) requestJson(req *http.Request, out any) error {
	data, err := c.request(req)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}
