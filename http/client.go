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
	"sync/atomic"
)

type errorResponse struct {
	Error string `json:"error"`
}

type httpClient struct {
	client  *http.Client
	baseURL *url.URL
	txValue atomic.Uint64
}

func newHttpClient(baseURL *url.URL) *httpClient {
	client := httpClient{
		client:  http.DefaultClient,
		baseURL: baseURL,
	}
	client.txValue.Store(0)
	return &client
}

func (c *httpClient) withTxn(value uint64) *httpClient {
	client := httpClient{
		client:  c.client,
		baseURL: c.baseURL,
	}
	client.txValue.Store(value)
	return &client
}

func (c *httpClient) setDefaultHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(TX_HEADER_NAME, fmt.Sprintf("%d", c.txValue.Load()))
}

func (c *httpClient) request(req *http.Request) error {
	c.setDefaultHeaders(req)

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint:errcheck

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode == http.StatusOK {
		return nil
	}

	var errRes errorResponse
	if err := json.Unmarshal(data, &errRes); err != nil {
		return fmt.Errorf("%s", data)
	}
	return fmt.Errorf(errRes.Error)
}

func (c *httpClient) requestJson(req *http.Request, out any) error {
	c.setDefaultHeaders(req)

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint:errcheck

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode == http.StatusOK {
		return json.Unmarshal(data, out)
	}

	var errRes errorResponse
	if err := json.Unmarshal(data, &errRes); err != nil {
		return fmt.Errorf("%s", data)
	}
	return fmt.Errorf(errRes.Error)
}
