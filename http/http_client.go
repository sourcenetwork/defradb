// Copyright 2025 Democratized Data Foundation
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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
)

type httpClient struct {
	client  *http.Client
	baseURL *url.URL
	apiURL  *url.URL
}

func newHttpClient(rawURL string) (*httpClient, error) {
	baseURL, err := parseBaseURL(rawURL)
	if err != nil {
		return nil, err
	}
	return &httpClient{
		client:  http.DefaultClient,
		baseURL: baseURL,
		apiURL:  baseURL.JoinPath("/api/" + Version),
	}, nil
}

// newInsecureHttpClient returns an httpClient that skips TLS certificate
// verification. Only use for loopback health checks against a server whose
// cert is not trusted by the system CA pool (e.g. self-signed certs).
func newInsecureHttpClient(rawURL string) (*httpClient, error) {
	baseURL, err := parseBaseURL(rawURL)
	if err != nil {
		return nil, err
	}
	return &httpClient{
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
			},
		},
		baseURL: baseURL,
		apiURL:  baseURL.JoinPath("/api/" + Version),
	}, nil
}

// parseBaseURL normalizes a raw API address into a URL, defaulting to the
// http scheme when none is provided. This is the same address the client
// dials, so its host is what the server sees in the request Host header.
func parseBaseURL(rawURL string) (*url.URL, error) {
	// Detect a scheme by the "://" separator rather than a "http" prefix, so
	// that scheme-less hosts that merely begin with "http" (e.g. httpbin:9181)
	// still get a scheme prepended and parse with a non-empty host.
	if !strings.Contains(rawURL, "://") {
		rawURL = "http://" + rawURL
	}
	return url.Parse(rawURL)
}

// AuthAudienceForURL returns the audience an auth token must carry to be
// accepted by a server reached at rawURL. It matches the server-side check
// in AuthMiddleware, which validates the token audience against the
// lower-cased request Host header.
func AuthAudienceForURL(rawURL string) (string, error) {
	baseURL, err := parseBaseURL(rawURL)
	if err != nil {
		return "", err
	}
	if baseURL.Host == "" {
		return "", NewErrNoHostInURL(rawURL)
	}
	return strings.ToLower(baseURL.Host), nil
}

func (c *httpClient) setDefaultHeaders(req *http.Request) error {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	txn, ok := datastore.CtxTryGetClientTxn(req.Context())
	if ok {
		req.Header.Set(txHeaderName, fmt.Sprintf("%d", txn.ID()))
	}
	id := iIdentity.FromContext(req.Context())
	if !id.HasValue() {
		return nil
	}
	if tokenIdentity, ok := id.Value().(identity.TokenIdentity); ok {
		req.Header.Set(authHeaderName, fmt.Sprintf("%s%s", authSchemaPrefix, tokenIdentity.BearerToken()))
	}
	return nil
}

func (c *httpClient) request(req *http.Request) ([]byte, error) {
	err := c.setDefaultHeaders(req)
	if err != nil {
		return nil, err
	}
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
		return nil, fmt.Errorf("%v: %s", res.StatusCode, data)
	}
	// A discarded/committed transaction is indistinguishable from one that was never
	// created from the caller's perspective, so surface the same error in both cases.
	//
	// This is defensive against the following unlikely (but possible) situation:
	//
	// Goroutine A - UpdateDoc calls txs.Load(id)
	// Goroutine B - CommitTransaction completes, and the transaction is commited/discarded
	// Goroutine A - The handler executes, calling into a transaction object that is now stale
	if errRes.Error != nil && strings.Contains(errRes.Error.Error(), db.ErrTxnDiscarded.Error()) {
		return nil, ErrTransactionNotFound
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
