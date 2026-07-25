// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// A number written as an integer literal in a schema parses to an int32, while the same option
// read back from the JSON the description is stored as gives a float64, and a request built in Go
// carries whatever the caller wrote. All of them are the same option.
func TestBM25Option_AcceptsIntAndFloat(t *testing.T) {
	fromLiteral, ok := BM25Option(map[string]any{"k1": int32(2)}, "k1", BM25DefaultK1)
	assert.True(t, ok)
	fromStore, ok := BM25Option(map[string]any{"k1": 2.0}, "k1", BM25DefaultK1)
	assert.True(t, ok)
	fromGo, ok := BM25Option(map[string]any{"k1": 2}, "k1", BM25DefaultK1)
	assert.True(t, ok)

	assert.Equal(t, 2.0, fromLiteral)
	assert.Equal(t, fromLiteral, fromStore)
	assert.Equal(t, fromLiteral, fromGo)

	missing, ok := BM25Option(nil, "b", BM25DefaultB)
	assert.True(t, ok)
	assert.Equal(t, BM25DefaultB, missing)

	// A value that is not a number reports itself as unusable and reads as the default, so a
	// caller with no error to give scores with the default rather than with a zero.
	notANumber, ok := BM25Option(map[string]any{"b": "fast"}, "b", BM25DefaultB)
	assert.False(t, ok)
	assert.Equal(t, BM25DefaultB, notANumber)
}

// The parts of the keyspace must stay distinct once encoded, since a term and a document short
// ID both follow the part in the same key.
func TestBM25Key_SeparatesTheParts(t *testing.T) {
	posting := BM25Key(1, 2, 3, BM25PostingPart, "cat", 4)
	length := BM25Key(1, 2, 3, BM25LengthPart, "", 4)
	totals := BM25Key(1, 2, 3, BM25TotalsPart, "", 0)

	assert.NotEqual(t, posting.Bytes(), length.Bytes())
	assert.NotEqual(t, length.Bytes(), totals.Bytes())
	assert.Equal(t, uint64(4), posting.DocShortID)
	assert.Equal(t, uint64(0), totals.DocShortID)
}
