// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

func TestParseVectorIndex_OnField_ParsesArgsAndDefaults(t *testing.T) {
	cases := []indexTestCase{
		{
			description: "vector index with explicit args",
			sdl: `type user {
				embedding: [Float32!] @vectorIndex(type: HNSW, metric: COSINE, dimensions: 3)
			}`,
			targetDescriptions: []client.NewIndexRequest{
				{
					Fields: []client.IndexedFieldDescription{
						{Name: "embedding"},
					},
					Vector: &client.VectorIndexDescription{
						Algorithm:  client.VectorAlgorithmHNSW,
						Metric:     client.DistanceMetricCosine,
						Dimensions: 3,
						HNSW: &client.HNSWParams{
							M:              16,
							EfConstruction: 128,
							EfSearch:       64,
						},
					},
				},
			},
		},
		{
			description: "vector index with custom HNSW params",
			sdl: `type user {
				embedding: [Float32!] @vectorIndex(dimensions: 3, m: 32, efConstruction: 200, efSearch: 100)
			}`,
			targetDescriptions: []client.NewIndexRequest{
				{
					Fields: []client.IndexedFieldDescription{
						{Name: "embedding"},
					},
					Vector: &client.VectorIndexDescription{
						Algorithm:  client.VectorAlgorithmHNSW,
						Metric:     client.DistanceMetricCosine,
						Dimensions: 3,
						HNSW: &client.HNSWParams{
							M:              32,
							EfConstruction: 200,
							EfSearch:       100,
						},
					},
				},
			},
		},
	}

	for _, test := range cases {
		parseIndexAndTest(t, test)
	}
}

func TestParseVectorIndex_OnField_ProducesVectorKindIndex(t *testing.T) {
	schemaManager, err := NewSchemaManager(false)
	require.NoError(t, err)

	parseResult, err := schemaManager.ParseSDL(`type user {
		embedding: [Float32!] @vectorIndex(type: HNSW, metric: COSINE, dimensions: 3)
	}`)
	require.NoError(t, err)
	require.Len(t, parseResult, 1)
	require.Len(t, parseResult[0].NewIndexes, 1)

	newIndex := parseResult[0].NewIndexes[0]
	require.NotNil(t, newIndex.Vector)

	// Simulate turning the request into a descriptor, as processNewIndexRequest would.
	desc := client.IndexDescription{
		Fields: newIndex.Fields,
		Vector: newIndex.Vector,
	}
	assert.Equal(t, client.IndexKindVector, desc.Kind())
	assert.Equal(t, client.VectorAlgorithmHNSW, desc.Vector.Algorithm)
	assert.Equal(t, client.DistanceMetricCosine, desc.Vector.Metric)
	assert.Equal(t, uint32(3), desc.Vector.Dimensions)
	require.NotNil(t, desc.Vector.HNSW)
	assert.Equal(t, uint32(16), desc.Vector.HNSW.M)
	assert.Equal(t, uint32(128), desc.Vector.HNSW.EfConstruction)
	assert.Equal(t, uint32(64), desc.Vector.HNSW.EfSearch)
}

func TestParseVectorIndex_InvalidArgs_ReturnsError(t *testing.T) {
	cases := []invalidIndexTestCase{
		{
			description: "unsupported algorithm",
			sdl: `type user {
				embedding: [Float32!] @vectorIndex(type: FOO, dimensions: 3)
			}`,
			expectedErr: `Expected type "VectorIndexAlgorithm", found FOO`,
		},
		{
			description: "unsupported metric",
			sdl: `type user {
				embedding: [Float32!] @vectorIndex(metric: EUCLIDEAN, dimensions: 3)
			}`,
			expectedErr: `Expected type "VectorDistanceMetric", found EUCLIDEAN`,
		},
		{
			description: "unknown argument",
			sdl: `type user {
				embedding: [Float32!] @vectorIndex(unknown: "something", dimensions: 3)
			}`,
			expectedErr: `Unknown argument "unknown" on directive "@vectorIndex".`,
		},
	}

	for _, test := range cases {
		parseInvalidIndexAndTest(t, test)
	}
}
