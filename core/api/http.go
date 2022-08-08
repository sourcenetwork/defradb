// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package api

type DataResponse struct {
	Data interface{} `json:"data"`
}

type ErrorResponse struct {
	Errors []ErrorItem `json:"errors"`
}

type ErrorItem struct {
	Message    string      `json:"message"`
	Extensions *Extensions `json:"extensions,omitempty"`
}

type Extensions struct {
	Status    int    `json:"status"`
	HTTPError string `json:"httpError"`
	Stack     string `json:"stack,omitempty"`
}
