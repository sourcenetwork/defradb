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

	"github.com/sourcenetwork/defradb/client"
)

// AddPolicyResult wraps the result of successfully adding/registering a Policy.
type AddPolicyRequest struct {
	// Policy body in JSON or YAML format.
	Policy string `json:"policy"`
	// Policy Creator's identity.
	Identity string `json:"identity"`
}

func (c *Client) AddPolicy(
	ctx context.Context,
	creator string,
	policy string,
) (client.AddPolicyResult, error) {
	methodURL := c.http.baseURL.JoinPath("acp", "policy")

	addPolicyRequest := AddPolicyRequest{
		Policy:   policy,
		Identity: creator,
	}

	addPolicyBody, err := json.Marshal(addPolicyRequest)
	if err != nil {
		return client.AddPolicyResult{}, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		methodURL.String(),
		bytes.NewBuffer(addPolicyBody),
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
