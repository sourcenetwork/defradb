// Copyright 2026 Democratized Data Foundation
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

func TestEncodeDecode_VeryLongDocID(t *testing.T) {
	longDocID := "bae-" + strings.Repeat("abcdefghij", 50)
	p := CursorPayload{
		DocID:     longDocID,
		Keys:      map[string]any{"age": float64(42)},
		Direction: "ASC",
	}
	encoded, err := Encode(p)
	require.NoError(t, err)

	decoded, err := Decode(encoded)
	require.NoError(t, err)
	assert.Equal(t, p, decoded)
	assert.Len(t, decoded.DocID, 504)
}

func TestEncodeDecode_SpecialCharactersInKeys(t *testing.T) {
	p := CursorPayload{
		DocID: "doc-special",
		Keys: map[string]any{
			"unicode_name": "Café 中文",
			"newline_val":  "line1\nline2\ttab",
			"quotes":       "He said \"hello\" and 'goodbye'",
		},
		Direction: "DESC",
	}
	encoded, err := Encode(p)
	require.NoError(t, err)

	decoded, err := Decode(encoded)
	require.NoError(t, err)
	assert.Equal(t, p, decoded)
	assert.Equal(t, "Café 中文", decoded.Keys["unicode_name"])
	assert.Equal(t, "🚀💫", decoded.Keys["emoji_field"])
	assert.Equal(t, "line1\nline2\ttab", decoded.Keys["newline_val"])
}

func TestEncodeDecode_NumericKeysPreservation(t *testing.T) {
	p := CursorPayload{
		DocID: "doc-numeric",
		Keys: map[string]any{
			"int_like":  float64(25),
			"decimal":   float64(3.14),
			"zero":      float64(0),
			"negative":  float64(-42.5),
			"large_int": float64(9007199254740991),
		},
		Direction: "ASC",
	}
	encoded, err := Encode(p)
	require.NoError(t, err)

	decoded, err := Decode(encoded)
	require.NoError(t, err)
	assert.Equal(t, p, decoded)

	assert.Equal(t, float64(25), decoded.Keys["int_like"])
	assert.Equal(t, float64(3.14), decoded.Keys["decimal"])
	assert.Equal(t, float64(0), decoded.Keys["zero"])
	assert.Equal(t, float64(-42.5), decoded.Keys["negative"])
	assert.Equal(t, float64(9007199254740991), decoded.Keys["large_int"])
}

func TestEncodeDecode_EmptyKeysVsNilKeys(t *testing.T) {
	// Empty Keys map
	pEmpty := CursorPayload{
		DocID:     "doc-empty-keys",
		Keys:      map[string]any{},
		Direction: "ASC",
	}
	encodedEmpty, err := Encode(pEmpty)
	require.NoError(t, err)

	decodedEmpty, err := Decode(encodedEmpty)
	require.NoError(t, err)
	assert.Nil(t, decodedEmpty.Keys, "empty Keys map becomes nil after JSON round-trip due to omitempty")

	// Nil Keys
	pNil := CursorPayload{
		DocID:     "doc-nil-keys",
		Keys:      nil,
		Direction: "ASC",
	}
	encodedNil, err := Encode(pNil)
	require.NoError(t, err)

	decodedNil, err := Decode(encodedNil)
	require.NoError(t, err)
	assert.Nil(t, decodedNil.Keys)

	// Keys with entries
	pWithKeys := CursorPayload{
		DocID:     "doc-with-keys",
		Keys:      map[string]any{"field": "value"},
		Direction: "ASC",
	}
	encodedWithKeys, err := Encode(pWithKeys)
	require.NoError(t, err)

	decodedWithKeys, err := Decode(encodedWithKeys)
	require.NoError(t, err)
	assert.NotNil(t, decodedWithKeys.Keys)
	assert.Equal(t, "value", decodedWithKeys.Keys["field"])
}
