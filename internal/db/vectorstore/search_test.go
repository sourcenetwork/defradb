// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package vectorstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/internal/index/hnsw"
)

// SearchWithDistance returns hits nearest-first, and the distance decreases towards the query. This
// pins the ordering contract the query planner relies on (nearest first, smaller distance = nearer)
// at the graph level, independent of the docID resolution that vectorstore.Search adds on top.
func TestGraphSearchWithDistance_ReturnsNearestFirstByDistance(t *testing.T) {
	ctx := newStoreTestCtx(t)
	params := hnsw.DefaultParams(16)
	g := NewGraph(ctx, 1, 1, 1, hnsw.Cosine, params)

	// Three points spread around the unit circle. The query [1,0] is nearest to n1, then n2, then n3.
	require.NoError(t, g.Insert(1, []float32{1, 0}))
	require.NoError(t, g.Insert(2, []float32{0.7, 0.7}))
	require.NoError(t, g.Insert(3, []float32{0, 1}))

	hits, err := g.SearchWithDistance([]float32{1, 0}, 3, 64)
	require.NoError(t, err)
	require.Len(t, hits, 3)

	assert.Equal(t, hnsw.NodeID(1), hits[0].ID)
	assert.Equal(t, hnsw.NodeID(2), hits[1].ID)
	assert.Equal(t, hnsw.NodeID(3), hits[2].ID)

	// Distances are non-decreasing (nearest first).
	assert.LessOrEqual(t, hits[0].Distance, hits[1].Distance)
	assert.LessOrEqual(t, hits[1].Distance, hits[2].Distance)
	// The exact match has ~zero cosine distance.
	assert.InDelta(t, 0.0, hits[0].Distance, 1e-6)
}
