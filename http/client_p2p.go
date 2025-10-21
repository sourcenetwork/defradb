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
	"time"

	"github.com/sourcenetwork/defradb/client"
)

var _ client.P2P = (*Client)(nil)

// SetReplicatorParams contains the replicator fields that can be modified by the user.
type SetReplicatorParams struct {
	// Addresses list of peer addresses.
	Addresses []string
	// Collections is the list of collection names to replicate.
	Collections []string
}

// DeleteReplicatorParams contains the params needed to delete a replicator.
type DeleteReplicatorParams struct {
	// ID is the ID of the replicator to delete.
	ID string
	// Collections is the list of collection names to replicate.
	Collections []string
}

func (c *Client) PeerInfo() ([]string, error) {
	methodURL := c.http.apiURL.JoinPath("p2p", "info")

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var res []string
	if err := c.http.requestJson(req, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) Connect(ctx context.Context, addresses []string) error {
	methodURL := c.http.apiURL.JoinPath("p2p", "connect")

	body, err := json.Marshal(addresses)
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

func (c *Client) SetReplicator(ctx context.Context, addresses []string, collections ...string) error {
	methodURL := c.http.apiURL.JoinPath("p2p", "replicators")

	body, err := json.Marshal(SetReplicatorParams{
		Addresses:   addresses,
		Collections: collections,
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

func (c *Client) DeleteReplicator(ctx context.Context, id string, collections ...string) error {
	methodURL := c.http.apiURL.JoinPath("p2p", "replicators")

	body, err := json.Marshal(DeleteReplicatorParams{
		ID:          id,
		Collections: collections,
	})
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
	methodURL := c.http.apiURL.JoinPath("p2p", "replicators")

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

func (c *Client) AddP2PCollections(ctx context.Context, collectionIDs ...string) error {
	methodURL := c.http.apiURL.JoinPath("p2p", "collections")

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

func (c *Client) RemoveP2PCollections(ctx context.Context, collectionIDs ...string) error {
	methodURL := c.http.apiURL.JoinPath("p2p", "collections")

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
	methodURL := c.http.apiURL.JoinPath("p2p", "collections")

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

func (c *Client) AddP2PDocuments(ctx context.Context, collectionIDs ...string) error {
	methodURL := c.http.apiURL.JoinPath("p2p", "documents")

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

func (c *Client) RemoveP2PDocuments(ctx context.Context, collectionIDs ...string) error {
	methodURL := c.http.apiURL.JoinPath("p2p", "documents")

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

func (c *Client) GetAllP2PDocuments(ctx context.Context) ([]string, error) {
	methodURL := c.http.apiURL.JoinPath("p2p", "documents")

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

func (c *Client) SyncDocuments(
	ctx context.Context,
	collectionName string,
	docIDs []string,
) error {
	methodURL := c.http.apiURL.JoinPath("p2p", "documents", "sync")

	req := map[string]any{
		"collectionName": collectionName,
		"docIDs":         docIDs,
	}

	deadline, hasDeadline := ctx.Deadline()
	if hasDeadline {
		req["timeout"] = time.Until(deadline).String()
	}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	_, err = c.http.request(httpReq)
	return err
}
