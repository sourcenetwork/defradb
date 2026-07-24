// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package vectorindex

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

func testDesc() client.VectorIndexDescription {
	return client.VectorIndexDescription{
		Algorithm: client.VectorAlgorithmHNSW,
		Metric:    client.DistanceMetricCosine,
		HNSW: &client.HNSWParams{
			M:              client.DefaultHNSWM,
			EfConstruction: client.DefaultHNSWEfConstruction,
			EfSearch:       client.DefaultHNSWEfSearch,
		},
	}
}

// Search returns hits nearest-first with non-decreasing distance. This pins the ordering contract the
// query planner relies on (nearest first, smaller distance = nearer) at the binding level.
func TestIndexSearch_ReturnsNearestFirstByDistance(t *testing.T) {
	ctx := newStoreTestCtx(t)
	index, err := Open(ctx, 1, 1, 1, testDesc())
	require.NoError(t, err)

	// Three points spread around the unit circle. The query [1,0] is nearest to 1, then 2, then 3.
	require.NoError(t, index.Insert(1, []float32{1, 0}))
	require.NoError(t, index.Insert(2, []float32{0.7, 0.7}))
	require.NoError(t, index.Insert(3, []float32{0, 1}))

	hits, err := index.Search([]float32{1, 0}, 3)
	require.NoError(t, err)
	require.Len(t, hits, 3)

	assert.Equal(t, uint64(1), hits[0].NodeID)
	assert.Equal(t, uint64(2), hits[1].NodeID)
	assert.Equal(t, uint64(3), hits[2].NodeID)

	assert.LessOrEqual(t, hits[0].Distance, hits[1].Distance)
	assert.LessOrEqual(t, hits[1].Distance, hits[2].Distance)
	// The exact match has ~zero cosine distance.
	assert.InDelta(t, 0.0, hits[0].Distance, 1e-6)
}

// Delete removes a node from the results; Open rejects an unsupported metric.
func TestIndex_DeleteAndUnsupportedMetric(t *testing.T) {
	ctx := newStoreTestCtx(t)
	index, err := Open(ctx, 1, 1, 1, testDesc())
	require.NoError(t, err)

	require.NoError(t, index.Insert(1, []float32{1, 0}))
	require.NoError(t, index.Insert(2, []float32{0, 1}))
	require.NoError(t, index.Delete(1))

	hits, err := index.Search([]float32{1, 0}, 2)
	require.NoError(t, err)
	for _, h := range hits {
		assert.NotEqual(t, uint64(1), h.NodeID)
	}

	bad := testDesc()
	bad.Metric = client.DistanceMetric(99)
	_, err = Open(ctx, 1, 1, 1, bad)
	require.Error(t, err)
}
