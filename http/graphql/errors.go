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

import (
	"encoding/json"
	"fmt"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
)

var (
	ErrInvalidSubscriptionTransport = errors.New("invalid subscription transport")
	ErrBadFormattedVariables        = errors.New("variable formatting")
	ErrReadTimeout                  = errors.New("read timeout")
	ErrWsConnClosed                 = errors.New("websocket connection closed")
	ErrInvalidMsg                   = errors.New("invalid message received")
	ErrUnableToUpgrade              = errors.New("unable to upgrade")
	ErrStreamingUnsupported         = errors.New("streaming unsupported")
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
	if msg, ok := out["error"].(string); ok {
		e.Error = client.ReviveError(msg)
	} else {
		e.Error = fmt.Errorf("%s", out)
	}
	return nil
}
