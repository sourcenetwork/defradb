// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cursor

import (
	"encoding/base64"
	"encoding/json"
)

// CursorPayload holds the data encoded into a cursor token.
type CursorPayload struct {
	DocID     string         `json:"d"`
	Keys      map[string]any `json:"k,omitempty"`
	Direction string         `json:"o"`
}

// Encode serializes a CursorPayload into an opaque base64url string.
func Encode(p CursorPayload) (string, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return "", NewErrEncodingCursor(err)
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

// Decode deserializes a cursor token back into a CursorPayload.
func Decode(cursor string) (CursorPayload, error) {
	if cursor == "" {
		return CursorPayload{}, ErrInvalidCursor
	}

	data, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return CursorPayload{}, ErrInvalidCursor
	}

	var p CursorPayload
	if err := json.Unmarshal(data, &p); err != nil {
		return CursorPayload{}, ErrInvalidCursor
	}

	if p.DocID == "" {
		return CursorPayload{}, ErrInvalidCursor
	}

	return p, nil
}
