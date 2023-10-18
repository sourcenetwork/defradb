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

var _ datastore.Txn = (*Transaction)(nil)

// Transaction implements the datastore.Txn interface over HTTP.
type Transaction struct {
	id   uint64
	http *httpClient
}

func NewTransaction(rawURL string, id uint64) (*Transaction, error) {
	httpClient, err := newHttpClient(rawURL)
	if err != nil {
		return nil, err
	}
	return &Transaction{id, httpClient}, nil
}

func (c *Transaction) ID() uint64 {
	return c.id
}

func (c *Transaction) Commit(ctx context.Context) error {
	methodURL := c.http.baseURL.JoinPath("tx", fmt.Sprintf("%d", c.id))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), nil)
	if err != nil {
		return err
	}
	_, err = c.http.request(req)
	return err
}

func (c *Transaction) Discard(ctx context.Context) {
	methodURL := c.http.baseURL.JoinPath("tx", fmt.Sprintf("%d", c.id))

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), nil)
	if err != nil {
		return
	}
	c.http.request(req) //nolint:errcheck
}

func (c *Transaction) OnSuccess(fn func()) {
	panic("client side transaction")
}

func (c *Transaction) OnError(fn func()) {
	panic("client side transaction")
}

func (c *Transaction) OnDiscard(fn func()) {
	panic("client side transaction")
}

func (c *Transaction) Rootstore() datastore.DSReaderWriter {
	panic("client side transaction")
}

func (c *Transaction) Datastore() datastore.DSReaderWriter {
	panic("client side transaction")
}

func (c *Transaction) Headstore() datastore.DSReaderWriter {
	panic("client side transaction")
}

func (c *Transaction) Peerstore() datastore.DSBatching {
	panic("client side transaction")
}

func (c *Transaction) DAGstore() datastore.DAGStore {
	panic("client side transaction")
}

func (c *Transaction) Systemstore() datastore.DSReaderWriter {
	panic("client side transaction")
}
