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

// A collection carrying a valid @vectorIndex on a raw [Float32!] field is created end-to-end: the
// index is registered as a vector-kind index with its parsed algorithm/metric/dimensions/HNSW
// params. The index performs no graph work yet (Phase 3 wires the HNSW engine); this asserts the
// schema surface + descriptor plumbing only.
func TestCollectionVersion_VectorIndexOnRawFloat32Array_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						embedding: [Float32!] @vectorIndex(type: HNSW, metric: COSINE, dimensions: 3)
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
						Vector: &client.VectorIndexDescription{
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
						embedding: [Float32!] @vectorIndex
					}
				`,
				ExpectedError: "vector index requires dimensions when the field is not a generated embedding",
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
						embedding: String @vectorIndex(dimensions: 3)
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
						embedding: [Float64!] @vectorIndex(dimensions: 3)
					}
				`,
				ExpectedError: "unsupported field type for vector index",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersion_VectorIndexWithUnsupportedAlgorithm_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						embedding: [Float32!] @vectorIndex(type: FOO, dimensions: 3)
					}
				`,
				ExpectedError: `Expected type "VectorIndexAlgorithm", found FOO`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersion_VectorIndexWithUnsupportedMetric_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						embedding: [Float32!] @vectorIndex(metric: EUCLIDEAN, dimensions: 3)
					}
				`,
				ExpectedError: `Expected type "VectorDistanceMetric", found EUCLIDEAN`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
