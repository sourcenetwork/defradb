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
	"context"
	"fmt"
	"net/http"

	"github.com/sourcenetwork/defradb/datastore"
)

var _ datastore.Txn = (*TxClient)(nil)

type TxClient struct {
	id   uint64
	http *httpClient
}

func (c *TxClient) ID() uint64 {
	return c.id
}

func (c *TxClient) Commit(ctx context.Context) error {
	methodURL := c.http.baseURL.JoinPath("tx", fmt.Sprintf("%d", c.id))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), nil)
	if err != nil {
		return err
	}
	return c.http.request(req)
}

func (c *TxClient) Discard(ctx context.Context) {
	methodURL := c.http.baseURL.JoinPath("tx", fmt.Sprintf("%d", c.id))

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), nil)
	if err != nil {
		return
	}
	c.http.request(req) //nolint:errcheck
}

func (c *TxClient) OnSuccess(fn func()) {
	panic("client side transaction")
}

func (c *TxClient) OnError(fn func()) {
	panic("client side transaction")
}

func (c *TxClient) OnDiscard(fn func()) {
	panic("client side transaction")
}

func (c *TxClient) Rootstore() datastore.DSReaderWriter {
	panic("client side transaction")
}

func (c *TxClient) Datastore() datastore.DSReaderWriter {
	panic("client side transaction")
}

func (c *TxClient) Headstore() datastore.DSReaderWriter {
	panic("client side transaction")
}

func (c *TxClient) DAGstore() datastore.DAGStore {
	panic("client side transaction")
}

func (c *TxClient) Systemstore() datastore.DSReaderWriter {
	panic("client side transaction")
}
