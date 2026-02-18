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
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/internal/utils"
)

var _ client.P2P = (*Client)(nil)

// CreateReplicatorParams contains the replicator fields that can be modified by the user.
type CreateReplicatorParams struct {
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

func (c *Client) PeerInfo(ctx context.Context, opts ...options.Enumerable[options.PeerInfoOptions]) ([]string, error) {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())

	methodURL := c.http.apiURL.JoinPath("p2p", "info")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var res []string
	if err := c.http.requestJson(req, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) ActivePeers(
	ctx context.Context,
	opts ...options.Enumerable[options.ActivePeersOptions],
) ([]string, error) {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())

	methodURL := c.http.apiURL.JoinPath("p2p", "active-peers")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var res []string
	if err := c.http.requestJson(req, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) Connect(
	ctx context.Context,
	addresses []string,
	opts ...options.Enumerable[options.ConnectOptions],
) error {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())

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

func (c *Client) CreateReplicator(
	ctx context.Context,
	addresses []string,
	opts ...options.Enumerable[options.CreateReplicatorOptions],
) error {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())

	methodURL := c.http.apiURL.JoinPath("p2p", "replicators")

	body, err := json.Marshal(CreateReplicatorParams{
		Addresses:   addresses,
		Collections: opt.CollectionNames,
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

func (c *Client) DeleteReplicator(
	ctx context.Context,
	id string,
	opts ...options.Enumerable[options.DeleteReplicatorOptions],
) error {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())

	methodURL := c.http.apiURL.JoinPath("p2p", "replicators")

	body, err := json.Marshal(DeleteReplicatorParams{
		ID:          id,
		Collections: opt.CollectionNames,
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

func (c *Client) ListReplicators(
	ctx context.Context,
	opts ...options.Enumerable[options.ListReplicatorsOptions],
) ([]client.Replicator, error) {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())

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

func (c *Client) CreateP2PCollections(
	ctx context.Context,
	collectionIDs []string,
	opts ...options.Enumerable[options.CreateP2PCollectionsOptions],
) error {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())

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

func (c *Client) DeleteP2PCollections(
	ctx context.Context,
	collectionIDs []string,
	opts ...options.Enumerable[options.DeleteP2PCollectionsOptions],
) error {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())

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

func (c *Client) ListP2PCollections(
	ctx context.Context,
	opts ...options.Enumerable[options.ListP2PCollectionsOptions],
) ([]string, error) {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())

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

func (c *Client) CreateP2PDocuments(
	ctx context.Context,
	docIDs []string,
	opts ...options.Enumerable[options.CreateP2PDocumentsOptions],
) error {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())

	methodURL := c.http.apiURL.JoinPath("p2p", "documents")

	body, err := json.Marshal(docIDs)
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

func (c *Client) DeleteP2PDocuments(
	ctx context.Context,
	docIDs []string,
	opts ...options.Enumerable[options.DeleteP2PDocumentsOptions],
) error {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())

	methodURL := c.http.apiURL.JoinPath("p2p", "documents")

	body, err := json.Marshal(docIDs)
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

func (c *Client) ListP2PDocuments(
	ctx context.Context,
	opts ...options.Enumerable[options.ListP2PDocumentsOptions],
) ([]string, error) {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())

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

	// Use a separate context for HTTP request with extra buffer time.
	// The server will use the timeout from the request body for the actual sync operation.
	// We add buffer time to account for HTTP overhead and response transmission.
	// This is necessary because the node handling this request will usually wait whole timeout
	// duration as it might receive responses from multiple peers.
	httpCtx := context.Background()
	if hasDeadline {
		var cancel context.CancelFunc
		httpCtx, cancel = context.WithTimeout(httpCtx, time.Until(deadline)+500*time.Millisecond)
		defer cancel()
	}

	httpReq, err := http.NewRequestWithContext(httpCtx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	_, err = c.http.request(httpReq)
	return err
}

func (c *Client) SyncCollectionVersions(
	ctx context.Context,
	versionIDs []string,
	opts ...options.Enumerable[options.SyncCollectionVersionsOptions],
) error {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())

	methodURL := c.http.apiURL.JoinPath("p2p", "collections", "sync-versions")

	req := map[string]any{
		"versionIDs": versionIDs,
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

func (c *Client) SyncBranchableCollection(
	ctx context.Context,
	collectionID string,
	opts ...options.Enumerable[options.SyncBranchableCollectionOptions],
) error {
	methodURL := c.http.apiURL.JoinPath("p2p", "collections", "sync-branchable")

	req := map[string]any{
		"collectionID": collectionID,
	}

	deadline, hasDeadline := ctx.Deadline()
	if hasDeadline {
		req["timeout"] = time.Until(deadline).String()
	}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	// Use a separate context for HTTP request with extra buffer time.
	// The server will use the timeout from the request body for the actual sync operation.
	// We add buffer time to account for HTTP overhead and response transmission.
	// This is necessary because the node handling this request will usually wait whole timeout
	// duration as it might receive responses from multiple peers.
	httpCtx := context.Background()
	opt := utils.NewOptions(opts...)
	httpCtx = identity.WithContext(httpCtx, opt.GetIdentity())
	if hasDeadline {
		var cancel context.CancelFunc
		httpCtx, cancel = context.WithTimeout(httpCtx, time.Until(deadline)+500*time.Millisecond)
		defer cancel()
	}

	httpReq, err := http.NewRequestWithContext(httpCtx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	_, err = c.http.request(httpReq)
	return err
}
