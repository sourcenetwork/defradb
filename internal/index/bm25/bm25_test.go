// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package bm25

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

func TestTokenFrequencies(t *testing.T) {
	assert.Equal(t,
		map[string]int{"the": 2, "cat": 1, "sat": 1, "on": 1, "mat": 1, "étude": 1},
		TokenFrequencies("The cat sat on the mat; a ÉTUDE"),
	)
	assert.Empty(t, TokenFrequencies("--- a é"))
}

func TestScoreTerm(t *testing.T) {
	score := ScoreTerm(2, 2, 3, 4, 3, client.BM25Params{K1: 1.2, B: 0.75})
	require.Greater(t, score, 0.0)
	assert.InDelta(t, 0.5908617053, score, 0.0000000001)
	assert.Zero(t, ScoreTerm(0, 2, 3, 4, 3, client.BM25Params{K1: 1.2, B: 0.75}))
}
