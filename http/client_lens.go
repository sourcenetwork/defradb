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

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/immutable/enumerable"

	"github.com/sourcenetwork/defradb/client"
)

var _ client.LensRegistry = (*LensRegistry)(nil)

// LensRegistry implements the client.LensRegistry interface over HTTP.
type LensRegistry struct {
	http *httpClient
}

type setMigrationRequest struct {
	CollectionID uint32
	Config       model.Lens
}

func (w *LensRegistry) Init(txnSource client.TxnSource) {}

func (c *LensRegistry) SetMigration(ctx context.Context, collectionID uint32, config model.Lens) error {
	methodURL := c.http.baseURL.JoinPath("lens", "registry")

	body, err := json.Marshal(setMigrationRequest{
		CollectionID: collectionID,
		Config:       config,
	})
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

func (c *LensRegistry) ReloadLenses(ctx context.Context) error {
	methodURL := c.http.baseURL.JoinPath("lens", "registry", "reload")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), nil)
	if err != nil {
		return err
	}
	_, err = c.http.request(req)
	return err
}

type migrateRequest struct {
	CollectionID uint32
	Data         []map[string]any
}

func (c *LensRegistry) MigrateUp(
	ctx context.Context,
	src enumerable.Enumerable[map[string]any],
	collectionID uint32,
) (enumerable.Enumerable[map[string]any], error) {
	methodURL := c.http.baseURL.JoinPath("lens", "registry", fmt.Sprint(collectionID), "up")

	var data []map[string]any
	err := enumerable.ForEach(src, func(item map[string]any) {
		data = append(data, item)
	})
	if err != nil {
		return nil, err
	}

	request := migrateRequest{
		CollectionID: collectionID,
		Data:         data,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	var result []map[string]any
	if err := c.http.requestJson(req, &result); err != nil {
		return nil, err
	}
	return enumerable.New(result), nil
}

func (c *LensRegistry) MigrateDown(
	ctx context.Context,
	src enumerable.Enumerable[map[string]any],
	collectionID uint32,
) (enumerable.Enumerable[map[string]any], error) {
	methodURL := c.http.baseURL.JoinPath("lens", "registry", fmt.Sprint(collectionID), "down")

	var data []map[string]any
	err := enumerable.ForEach(src, func(item map[string]any) {
		data = append(data, item)
	})
	if err != nil {
		return nil, err
	}

	request := migrateRequest{
		CollectionID: collectionID,
		Data:         data,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	var result []map[string]any
	if err := c.http.requestJson(req, &result); err != nil {
		return nil, err
	}
	return enumerable.New(result), nil
}
