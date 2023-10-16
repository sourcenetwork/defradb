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

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/sourcenetwork/defradb/client"
)

func (c *Client) PeerInfo() peer.AddrInfo {
	methodURL := c.http.baseURL.JoinPath("p2p", "info")

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return peer.AddrInfo{}
	}
	var res peer.AddrInfo
	if err := c.http.requestJson(req, &res); err != nil {
		return peer.AddrInfo{}
	}
	return res
}

func (c *Client) SetReplicator(ctx context.Context, rep client.Replicator) error {
	methodURL := c.http.baseURL.JoinPath("p2p", "replicators")

	body, err := json.Marshal(rep)
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

func (c *Client) DeleteReplicator(ctx context.Context, rep client.Replicator) error {
	methodURL := c.http.baseURL.JoinPath("p2p", "replicators")

	body, err := json.Marshal(rep)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	_, err = c.http.request(req)
	return err
}

func (c *Client) GetAllReplicators(ctx context.Context) ([]client.Replicator, error) {
	methodURL := c.http.baseURL.JoinPath("p2p", "replicators")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var reps []client.Replicator
	if err := c.http.requestJson(req, &reps); err != nil {
		return nil, err
	}
	return reps, nil
}

func (c *Client) AddP2PCollections(ctx context.Context, collectionIDs []string) error {
	methodURL := c.http.baseURL.JoinPath("p2p", "collections")

	body, err := json.Marshal(collectionIDs)
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

func (c *Client) RemoveP2PCollections(ctx context.Context, collectionIDs []string) error {
	methodURL := c.http.baseURL.JoinPath("p2p", "collections")

	body, err := json.Marshal(collectionIDs)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	_, err = c.http.request(req)
	return err
}

func (c *Client) GetAllP2PCollections(ctx context.Context) ([]string, error) {
	methodURL := c.http.baseURL.JoinPath("p2p", "collections")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var cols []string
	if err := c.http.requestJson(req, &cols); err != nil {
		return nil, err
	}
	return cols, nil
}
