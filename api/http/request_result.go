// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package http

import "github.com/sourcenetwork/defradb/client"

type GQLResult struct {
	Errors []string `json:"errors,omitempty"`

	Data any `json:"data"`
}

func newGQLResult(r client.GQLResult) *GQLResult {
	errors := make([]string, len(r.Errors))
	for i := range r.Errors {
		errors[i] = r.Errors[i].Error()
	}

	return &GQLResult{
		Errors: errors,
		Data:   r.Data,
	}
}
