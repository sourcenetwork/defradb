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

// Errors returnable from this package.
//
// This list is incomplete. Undefined errors may also be returned.
// Errors returned from this package may be tested against these errors with errors.Is.
var (
	ErrNoListener            = errors.New("cannot serve with no listener")
	ErrSchema                = errors.New("base must start with the http or https scheme")
	ErrDatabaseNotAvailable  = errors.New("no database available")
	ErrFormNotSupported      = errors.New("content type application/x-www-form-urlencoded not yet supported")
	ErrBodyEmpty             = errors.New("body cannot be empty")
	ErrMissingGQLRequest     = errors.New("missing GraphQL request")
	ErrPeerIdUnavailable     = errors.New("no PeerID available. P2P might be disabled")
	ErrStreamingUnsupported  = errors.New("streaming unsupported")
	ErrNoEmail               = errors.New("email address must be specified for tls with autocert")
	ErrPayloadFormat         = errors.New("invalid payload format")
	ErrMissingNewKey         = errors.New("missing _newKey for imported doc")
	ErrInvalidRequestBody    = errors.New("invalid request body")
	ErrDocKeyDoesNotMatch    = errors.New("document key does not match")
	ErrStreamingNotSupported = errors.New("streaming not supported")
	ErrMigrationNotFound     = errors.New("migration not found")
	ErrMissingRequest        = errors.New("missing request")
	ErrInvalidTransactionId  = errors.New("invalid transaction id")
	ErrP2PDisabled           = errors.New("p2p network is disabled")
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
