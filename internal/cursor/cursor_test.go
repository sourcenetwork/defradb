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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecode_RoundtripWithKeys(t *testing.T) {
	p := CursorPayload{
		DocID:     "abc123",
		Keys:      map[string]any{"age": float64(25), "name": "Alice"},
		Direction: "ASC",
	}
	encoded, err := Encode(p)
	require.NoError(t, err)

	decoded, err := Decode(encoded)
	require.NoError(t, err)
	assert.Equal(t, p, decoded)
}

func TestEncodeDecode_RoundtripWithoutKeys(t *testing.T) {
	p := CursorPayload{
		DocID:     "doc1",
		Direction: "DESC",
	}
	encoded, err := Encode(p)
	require.NoError(t, err)

	decoded, err := Decode(encoded)
	require.NoError(t, err)
	assert.Equal(t, p.DocID, decoded.DocID)
	assert.Equal(t, p.Direction, decoded.Direction)
	assert.Nil(t, decoded.Keys)
}

func TestEncodeDecode_RoundtripDocIDOnly(t *testing.T) {
	p := CursorPayload{
		DocID:     "x",
		Direction: "ASC",
	}
	encoded, err := Encode(p)
	require.NoError(t, err)

	decoded, err := Decode(encoded)
	require.NoError(t, err)
	assert.Equal(t, p.DocID, decoded.DocID)
}

func TestDecode_EmptyString(t *testing.T) {
	_, err := Decode("")
	assert.ErrorIs(t, err, ErrInvalidCursor)
}

func TestDecode_InvalidBase64(t *testing.T) {
	_, err := Decode("not-base64!!!")
	assert.ErrorIs(t, err, ErrInvalidCursor)
}

func TestDecode_InvalidJSON(t *testing.T) {
	encoded := base64.RawURLEncoding.EncodeToString([]byte("not json"))
	_, err := Decode(encoded)
	assert.ErrorIs(t, err, ErrInvalidCursor)
}

func TestDecode_MissingDocID(t *testing.T) {
	encoded := base64.RawURLEncoding.EncodeToString([]byte(`{"k":{"age":25},"o":"ASC"}`))
	_, err := Decode(encoded)
	assert.ErrorIs(t, err, ErrInvalidCursor)
}

func TestEncode_URLSafeCharacters(t *testing.T) {
	p := CursorPayload{
		DocID:     "bae-abc123-def456",
		Keys:      map[string]any{"age": float64(25), "name": "Alice+Bob/Charlie="},
		Direction: "ASC",
	}
	encoded, err := Encode(p)
	require.NoError(t, err)

	assert.False(t, strings.Contains(encoded, "+"), "encoded cursor must not contain +")
	assert.False(t, strings.Contains(encoded, "/"), "encoded cursor must not contain /")
	assert.False(t, strings.Contains(encoded, "="), "encoded cursor must not contain =")
}
