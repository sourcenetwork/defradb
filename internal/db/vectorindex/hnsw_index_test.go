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
	bad.Metric = client.DistanceMetric("NOT_A_METRIC")
	_, err = Open(ctx, 1, 1, 1, bad)
	require.Error(t, err)
}

// The descriptor's metric reaches the engine and decides the ranking. The vectors are placed so each
// metric orders them differently, so no single ranking passes by accident. It also pins that only
// cosine normalizes: scaled to unit length, all three orders would match.
//
//	        cosine   squared L2   dot
//	(1, 0.5)  0.894      0.25      1.0
//	(3, 0.1)  0.999      4.01      3.0
//	(0.8,0.3) 0.936      0.13      0.8
func TestIndexSearch_MetricDecidesRanking(t *testing.T) {
	vectors := [][]float32{{1, 0.5}, {3, 0.1}, {0.8, 0.3}}

	testCases := []struct {
		metric   client.DistanceMetric
		expected []uint64
	}{
		{client.DistanceMetricCosine, []uint64{2, 3, 1}},
		{client.DistanceMetricEuclidean, []uint64{3, 1, 2}},
		{client.DistanceMetricDotProduct, []uint64{2, 1, 3}},
	}

	for _, testCase := range testCases {
		t.Run(string(testCase.metric), func(t *testing.T) {
			ctx := newStoreTestCtx(t)
			desc := testDesc()
			desc.Metric = testCase.metric

			index, err := Open(ctx, 1, 1, 1, desc)
			require.NoError(t, err)

			for i, v := range vectors {
				require.NoError(t, index.Insert(uint64(i+1), v))
			}

			hits, err := index.Search([]float32{1, 0}, len(vectors))
			require.NoError(t, err)
			require.Len(t, hits, len(vectors))

			actual := make([]uint64, len(hits))
			for i, h := range hits {
				actual[i] = h.NodeID
			}
			assert.Equal(t, testCase.expected, actual)
		})
	}
}
