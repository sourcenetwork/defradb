// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package collection_version

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// A collection carrying a valid @index(vector: {...}) on a raw [Float32!] field is created end-to-end: the
// index is registered as a vector-kind index with its parsed algorithm/metric/dimensions/HNSW
// params. The index performs no graph work yet (Phase 3 wires the HNSW engine); this asserts the
// schema surface + descriptor plumbing only.
func TestCollectionVersion_VectorIndexOnRawFloat32Array_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						embedding: [Float32!] @index(vector: {dimensions: 3, hnsw: {metric: COSINE}})
					}
				`,
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name: "Users_embedding_ASC",
						ID:   1,
						Fields: []client.IndexedFieldDescription{
							{Name: "embedding"},
						},
						Kind: client.IndexKindVector,
						KindDescription: &client.VectorIndexDescription{
							Fields:     []string{"embedding"},
							Algorithm:  client.VectorAlgorithmHNSW,
							Metric:     client.DistanceMetricCosine,
							Dimensions: 3,
							HNSW:       &client.HNSWParams{M: 16, EfConstruction: 128, EfSearch: 64},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersion_VectorIndexOnFloat32ArrayWithoutDimensionsOrEmbedding_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						embedding: [Float32!] @index(vector: {})
					}
				`,
				ExpectedError: "vector index dimensions must be greater than zero",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersion_VectorIndexOnStringField_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						embedding: String @index(vector: {dimensions: 3})
					}
				`,
				ExpectedError: "unsupported field type for vector index",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersion_VectorIndexOnFloat64ArrayField_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						embedding: [Float64!] @index(vector: {dimensions: 3})
					}
				`,
				ExpectedError: "unsupported field type for vector index",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersion_VectorIndexWithUnsupportedAlgorithm_ShouldError(t *testing.T) {
	// An unknown algorithm config is an unknown field in the vector config, so GraphQL rejects it
	// before the parser runs.
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						embedding: [Float32!] @index(vector: {dimensions: 3, IVFFlat: {}})
					}
				`,
				ExpectedError: `In field "IVFFlat": Unknown field.`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// The directive accepts each supported metric and stores it on the descriptor. The test above covers
// COSINE; these cover the two that do not scale vectors to unit length.
func TestCollectionVersion_VectorIndexWithEuclideanMetric_ShouldSucceed(t *testing.T) {
	testUtils.ExecuteTestCase(t, vectorIndexMetricTest("EUCLIDEAN", client.DistanceMetricEuclidean))
}

func TestCollectionVersion_VectorIndexWithDotProductMetric_ShouldSucceed(t *testing.T) {
	testUtils.ExecuteTestCase(t, vectorIndexMetricTest("DOT", client.DistanceMetricDotProduct))
}

func vectorIndexMetricTest(sdlMetric string, expected client.DistanceMetric) testUtils.TestCase {
	return testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						embedding: [Float32!] @index(vector: {dimensions: 3, hnsw: {metric: ` + sdlMetric + `}})
					}
				`,
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "Users_embedding_ASC",
						ID:     1,
						Fields: []client.IndexedFieldDescription{{Name: "embedding"}},
						Kind:   client.IndexKindVector,
						KindDescription: &client.VectorIndexDescription{
							Fields:     []string{"embedding"},
							Algorithm:  client.VectorAlgorithmHNSW,
							Metric:     expected,
							Dimensions: 3,
							HNSW:       &client.HNSWParams{M: 16, EfConstruction: 128, EfSearch: 64},
						},
					},
				},
			},
		},
	}
}

func TestCollectionVersion_VectorIndexWithUnsupportedMetric_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						embedding: [Float32!] @index(vector: {dimensions: 3, hnsw: {metric: MANHATTAN}})
					}
				`,
				ExpectedError: `Expected type "VectorDistanceMetric", found MANHATTAN`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// A vector index is searched by nearness, not read in key order, so it cannot honour a descending
// direction. The direction reaches the db layer so the error can say that, rather than only that an
// argument was invalid.
func TestCollectionVersion_VectorIndexWithDescendingDirection_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						embedding: [Float32!] @index(vector: {dimensions: 3, hnsw: {metric: COSINE}}, direction: DESC)
					}
				`,
				ExpectedError: "vector index cannot have a direction",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Ascending is what a vector index already does, so asking for it explicitly is accepted.
func TestCollectionVersion_VectorIndexWithAscendingDirection_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						embedding: [Float32!] @index(vector: {dimensions: 3, hnsw: {metric: COSINE}}, direction: ASC)
					}
				`,
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "Users_embedding_ASC",
						ID:     1,
						Fields: []client.IndexedFieldDescription{{Name: "embedding"}},
						Kind:   client.IndexKindVector,
						KindDescription: &client.VectorIndexDescription{
							Fields:     []string{"embedding"},
							Algorithm:  client.VectorAlgorithmHNSW,
							Metric:     client.DistanceMetricCosine,
							Dimensions: 3,
							HNSW:       &client.HNSWParams{M: 16, EfConstruction: 128, EfSearch: 64},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
