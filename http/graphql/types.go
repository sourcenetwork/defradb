// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package graphql

// GraphQL Request represents a GraphQL request body.
type Request struct {
	Query         string         `json:"query"`
	OperationName string         `json:"operationName"`
	Variables     map[string]any `json:"variables"`
}

// GraphQL Response represents a GraphQL response body.
type Response struct {
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
	Data map[string]any `json:"data"`
}
