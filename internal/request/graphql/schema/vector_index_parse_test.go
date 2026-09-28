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
			description: "alg selector uses the default HNSW configuration",
			sdl: `type user {
				embedding: [Float32!] @index(vector: {dimensions: 3, alg: hnsw})
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
			description: "matching alg and HNSW config are allowed",
			sdl: `type user {
				embedding: [Float32!] @index(
					name: "embeddingIndex",
					vector: {dimensions: 3, alg: hnsw, hnsw: {metric: EUCLIDEAN, M: 32, efConstruction: 200, efSearch: 100}}
				)
			}`,
			targetDescriptions: []client.NewIndexRequest{
				{
					Name: "embeddingIndex",
					Fields: []client.IndexedFieldDescription{
						{Name: "embedding"},
					},
					Vector: &client.VectorIndexDescription{
						Algorithm:  client.VectorAlgorithmHNSW,
						Metric:     client.DistanceMetricEuclidean,
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
		embedding: [Float32!] @index(vector: {dimensions: 3, hnsw: {metric: COSINE}})
	}`)
	require.NoError(t, err)
	require.Len(t, parseResult, 1)
	require.Len(t, parseResult[0].NewIndexes, 1)

	newIndex := parseResult[0].NewIndexes[0]
	require.NotNil(t, newIndex.Vector)

	// Simulate turning the request into a descriptor, as processNewIndexRequest would: a request with
	// a Vector becomes a vector-kind descriptor carrying the vector config.
	desc := client.IndexDescription{
		Fields:          newIndex.Fields,
		Kind:            client.IndexKindVector,
		KindDescription: newIndex.Vector,
	}
	require.True(t, desc.IsVector())
	vec, _ := desc.GetVector()
	assert.Equal(t, client.VectorAlgorithmHNSW, vec.Algorithm)
	assert.Equal(t, client.DistanceMetricCosine, vec.Metric)
	assert.Equal(t, uint32(3), vec.Dimensions)
	require.NotNil(t, vec.HNSW)
	assert.Equal(t, uint32(16), vec.HNSW.M)
	assert.Equal(t, uint32(128), vec.HNSW.EfConstruction)
	assert.Equal(t, uint32(64), vec.HNSW.EfSearch)
}

func TestParseVectorIndex_InvalidArgs_ReturnsError(t *testing.T) {
	cases := []invalidIndexTestCase{
		{
			description: "vector is not an index kind",
			sdl: `type user {
				embedding: [Float32!] @index(kind: vector)
			}`,
			expectedErr: `Expected type "IndexKind", found vector`,
		},
		{
			description: "unknown algorithm enum",
			sdl: `type user {
				embedding: [Float32!] @index(vector: {dimensions: 3, alg: IVFFlat})
			}`,
			expectedErr: `Expected type "VectorIndexAlgorithm", found IVFFlat`,
		},
		{
			description: "unknown algorithm config field",
			sdl: `type user {
				embedding: [Float32!] @index(vector: {dimensions: 3, IVFFlat: {}})
			}`,
			expectedErr: `In field "IVFFlat": Unknown field.`,
		},
		{
			description: "unsupported metric inside the HNSW config",
			sdl: `type user {
				embedding: [Float32!] @index(vector: {dimensions: 3, hnsw: {metric: MANHATTAN}})
			}`,
			expectedErr: `Expected type "VectorDistanceMetric", found MANHATTAN`,
		},
		{
			description: "unknown top-level argument",
			sdl: `type user {
				embedding: [Float32!] @index(unknown: "something", vector: {dimensions: 3})
			}`,
			expectedErr: `Unknown argument "unknown" on directive "@index".`,
		},
		{
			description: "unknown field inside the vector config",
			sdl: `type user {
				embedding: [Float32!] @index(vector: {dimensions: 3, unknown: 1})
			}`,
			expectedErr: `In field "unknown": Unknown field.`,
		},
		{
			description: "unknown field inside the HNSW config",
			sdl: `type user {
				embedding: [Float32!] @index(vector: {dimensions: 3, hnsw: {unknown: 1}})
			}`,
			expectedErr: `In field "unknown": Unknown field.`,
		},
		{
			description: "vector and legacy ordered configs are competing kind selectors",
			sdl: `type user {
				embedding: [Float32!] @index(vector: {}, unique: false)
			}`,
			expectedErr: errIndexInvalidArgument,
		},
		{
			description: "vector config is invalid on an object directive",
			sdl: `type user @index(vector: {dimensions: 3, alg: hnsw}) {
				embedding: [Float32!]
			}`,
			expectedErr: errIndexInvalidArgument,
		},
	}

	for _, test := range cases {
		parseInvalidIndexAndTest(t, test)
	}
}
