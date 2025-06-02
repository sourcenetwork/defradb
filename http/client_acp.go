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

func (c *Client) AddDACPolicy(
	ctx context.Context,
	policy string,
) (client.AddPolicyResult, error) {
	methodURL := c.http.apiURL.JoinPath("acp", "dac", "policy")

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

type addDACActorRelationshipRequest struct {
	CollectionName string
	DocID          string
	Relation       string
	TargetActor    string
}

func (c *Client) AddDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.AddActorRelationshipResult, error) {
	methodURL := c.http.apiURL.JoinPath("acp", "dac", "relationship")

	body, err := json.Marshal(
		addDACActorRelationshipRequest{
			CollectionName: collectionName,
			DocID:          docID,
			Relation:       relation,
			TargetActor:    targetActor,
		},
	)

	if err != nil {
		return client.AddActorRelationshipResult{}, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		methodURL.String(),
		bytes.NewBuffer(body),
	)

	if err != nil {
		return client.AddActorRelationshipResult{}, err
	}

	var addDocActorRelResult client.AddActorRelationshipResult
	if err := c.http.requestJson(req, &addDocActorRelResult); err != nil {
		return client.AddActorRelationshipResult{}, err
	}

	return addDocActorRelResult, nil
}

type deleteDACActorRelationshipRequest struct {
	CollectionName string
	DocID          string
	Relation       string
	TargetActor    string
}

func (c *Client) DeleteDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.DeleteActorRelationshipResult, error) {
	methodURL := c.http.apiURL.JoinPath("acp", "dac", "relationship")

	body, err := json.Marshal(
		deleteDACActorRelationshipRequest{
			CollectionName: collectionName,
			DocID:          docID,
			Relation:       relation,
			TargetActor:    targetActor,
		},
	)

	if err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		methodURL.String(),
		bytes.NewBuffer(body),
	)

	if err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}

	var deleteDocActorRelResult client.DeleteActorRelationshipResult
	if err := c.http.requestJson(req, &deleteDocActorRelResult); err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}

	return deleteDocActorRelResult, nil
}
