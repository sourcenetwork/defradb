// Copyright 2024 Democratized Data Foundation
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
	"strings"

	"github.com/sourcenetwork/defradb/client"
)

func (c *Client) AddPolicy(
	ctx context.Context,
	policy string,
) (client.AddPolicyResult, error) {
	methodURL := c.http.baseURL.JoinPath("acp", "policy")

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		methodURL.String(),
		strings.NewReader(policy),
	)

	if err != nil {
		return client.AddPolicyResult{}, err
	}

	var policyResult client.AddPolicyResult
	if err := c.http.requestJson(req, &policyResult); err != nil {
		return client.AddPolicyResult{}, err
	}

	return policyResult, nil
}

type addDocActorRelationshipRequest struct {
	CollectionName string
	DocID          string
	Relation       string
	TargetActor    string
}

func (c *Client) AddDocActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.AddDocActorRelationshipResult, error) {
	methodURL := c.http.baseURL.JoinPath("acp", "relationship")

	body, err := json.Marshal(
		addDocActorRelationshipRequest{
			CollectionName: collectionName,
			DocID:          docID,
			Relation:       relation,
			TargetActor:    targetActor,
		},
	)

	if err != nil {
		return client.AddDocActorRelationshipResult{}, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		methodURL.String(),
		bytes.NewBuffer(body),
	)

	if err != nil {
		return client.AddDocActorRelationshipResult{}, err
	}

	var addDocActorRelResult client.AddDocActorRelationshipResult
	if err := c.http.requestJson(req, &addDocActorRelResult); err != nil {
		return client.AddDocActorRelationshipResult{}, err
	}

	return addDocActorRelResult, nil
}

type deleteDocActorRelationshipRequest struct {
	CollectionName string
	DocID          string
	Relation       string
	TargetActor    string
}

func (c *Client) DeleteDocActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.DeleteDocActorRelationshipResult, error) {
	methodURL := c.http.baseURL.JoinPath("acp", "relationship")

	body, err := json.Marshal(
		deleteDocActorRelationshipRequest{
			CollectionName: collectionName,
			DocID:          docID,
			Relation:       relation,
			TargetActor:    targetActor,
		},
	)

	if err != nil {
		return client.DeleteDocActorRelationshipResult{}, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		methodURL.String(),
		bytes.NewBuffer(body),
	)

	if err != nil {
		return client.DeleteDocActorRelationshipResult{}, err
	}

	var deleteDocActorRelResult client.DeleteDocActorRelationshipResult
	if err := c.http.requestJson(req, &deleteDocActorRelResult); err != nil {
		return client.DeleteDocActorRelationshipResult{}, err
	}

	return deleteDocActorRelResult, nil
}
