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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/internal/utils"
)

var _ client.Collection = (*Collection)(nil)

// Collection implements the client.Collection interface over HTTP.
type Collection struct {
	http *httpClient
	def  client.CollectionVersion
}

func (c *Collection) Version() client.CollectionVersion {
	return c.def
}

func (c *Collection) Name() string {
	return c.Version().Name
}

func (c *Collection) VersionID() string {
	return c.Version().VersionID
}

func (c *Collection) CollectionID() string {
	return c.Version().CollectionID
}

func (c *Collection) AddIndex(
	ctx context.Context,
	indexDesc client.IndexAddRequest,
	opts ...options.Enumerable[options.CollectionAddIndexOptions],
) (client.IndexDescription, error) {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())
	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name, "indexes")

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

func (c *Collection) DeleteIndex(
	ctx context.Context,
	indexName string,
	opts ...options.Enumerable[options.CollectionDeleteIndexOptions],
) error {
	if indexName == "" {
		return client.ErrIndexNameRequired
	}

	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())

	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name, "indexes", indexName)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), nil)
	if err != nil {
		return err
	}
	_, err = c.http.request(req)
	return err
}

func (c *Collection) ListIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.CollectionListIndexesOptions],
) ([]client.IndexDescription, error) {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())

	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name, "indexes")

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

func (c *Collection) AddEncryptedIndex(
	ctx context.Context,
	indexDesc client.EncryptedIndexDescription,
	opts ...options.Enumerable[options.AddEncryptedIndexOptions],
) (client.EncryptedIndexDescription, error) {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())
	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name, "encrypted-indexes")

	body, err := json.Marshal(&indexDesc)
	if err != nil {
		return client.EncryptedIndexDescription{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return client.EncryptedIndexDescription{}, err
	}
	var index client.EncryptedIndexDescription
	if err := c.http.requestJson(req, &index); err != nil {
		return client.EncryptedIndexDescription{}, err
	}
	return index, nil
}

func (c *Collection) ListEncryptedIndexes(
	ctx context.Context, opts ...options.Enumerable[options.CollectionListEncryptedIndexesOptions],
) ([]client.EncryptedIndexDescription, error) {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())
	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name, "encrypted-indexes")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var indexes []client.EncryptedIndexDescription
	if err := c.http.requestJson(req, &indexes); err != nil {
		return nil, err
	}
	return indexes, nil
}

func (c *Collection) DeleteEncryptedIndex(
	ctx context.Context,
	fieldName string,
	opts ...options.Enumerable[options.DeleteEncryptedIndexOptions],
) error {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())

	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name, "encrypted-indexes", fieldName)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), nil)
	if err != nil {
		return err
	}

	_, err = c.http.request(req)
	return err
}

func (c *Collection) Truncate(
	ctx context.Context, opts ...options.Enumerable[options.CollectionTruncateOptions],
) error {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())

	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name, "truncate")

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), nil)
	if err != nil {
		return err
	}

	_, err = c.http.request(req)
	return err
}
