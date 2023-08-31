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
	"encoding/json"
	"errors"
)

const (
	errInvalidRequestBody    = "invalid request body"
	errDocKeyDoesNotMatch    = "document key does not match"
	errStreamingNotSupported = "streaming not supported"
	errMigrationNotFound     = "migration not found"
	errMissingRequest        = "missing request"
	errInvalidTransactionId  = "invalid transaction id"
)

var (
	ErrInvalidRequestBody    = errors.New(errInvalidRequestBody)
	ErrDocKeyDoesNotMatch    = errors.New(errDocKeyDoesNotMatch)
	ErrStreamingNotSupported = errors.New(errStreamingNotSupported)
	ErrMigrationNotFound     = errors.New(errMigrationNotFound)
	ErrMissingRequest        = errors.New(errMissingRequest)
	ErrInvalidTransactionId  = errors.New(errInvalidTransactionId)
)

type errorResponse struct {
	Error error `json:"error"`
}

func (e errorResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{"error": e.Error.Error()})
}

func (e *errorResponse) UnmarshalJSON(data []byte) error {
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}
	e.Error = parseError(out["error"])
	return nil
}
