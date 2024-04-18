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
	"context"
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
