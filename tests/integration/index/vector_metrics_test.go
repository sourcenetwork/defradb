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

package index

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// Using the index and scoring must both follow its metric: doing one without the other would rank by
// one metric and score by another. A metric the engine cannot use fails the build instead.
func TestVectorIndex_QueryOnAnyMetric_ShouldUseIndexAndScoreByItsMetric(t *testing.T) {
	testCases := []struct {
		sdlMetric string
		metric    client.DistanceMetric
	}{
		{"COSINE", client.DistanceMetricCosine},
		{"EUCLIDEAN", client.DistanceMetricEuclidean},
		{"DOT", client.DistanceMetricDotProduct},
	}

	req := `query {
		User(order: {_alias: {sim: DESC}}, limit: 2){
			name
			sim: SIMILARITY(vector: {vector: [1, 0, 0]})
		}
	}`

	for _, testCase := range testCases {
		test := testUtils.TestCase{
			Actions: []any{
				&action.AddCollection{
					SDL: `type User {
						name: String
						vector: [Float32!] @index(vector: {dimensions: 3, hnsw: {metric: ` +
						testCase.sdlMetric + `}})
					}`,
				},
				&action.AddDoc{DocMap: map[string]any{"name": "x", "vector": []float32{1, 0, 0}}},
				&action.AddDoc{DocMap: map[string]any{"name": "y", "vector": []float32{0, 1, 0}}},
				&action.AddDoc{DocMap: map[string]any{"name": "xy", "vector": []float32{0.9, 0.4, 0}}},
				// An update and a delete cover the maintenance path, not just the insert.
				&action.UpdateDoc{DocID: 1, Doc: `{"vector": [0, 0.5, 0]}`},
				testUtils.DeleteDoc{DocID: 2},
				&action.WaitForIndexReady{CollectionID: 0},
				&action.Request{
					Request: req,
					Results: map[string]any{
						"User": []map[string]any{
							{
								"name": "x",
								"sim": testUtils.SimilarityScore(
									testCase.metric, []float64{1, 0, 0}, []float64{1, 0, 0}),
							},
							{
								"name": "y",
								"sim": testUtils.SimilarityScore(
									testCase.metric, []float64{0, 0.5, 0}, []float64{1, 0, 0}),
							},
						},
					},
				},
				&action.Request{
					Request: makeExplainQuery(req),
					Asserter: testUtils.NewExplainAsserter().
						WithIndexFetches(1).
						WithDocFetches(2),
				},
			},
		}

		t.Run(testCase.sdlMetric, func(t *testing.T) { testUtils.ExecuteTestCase(t, test) })
	}
}

// Using the index must not change the answer. Each metric ranks these vectors differently, so a wrong
// metric cannot pass by coincidence.
//
// Both runs share one indexed collection, since dropping the index would make the query a cosine one
// instead of the same search. Omitting the limit is what makes it a full scan.
func TestVectorIndex_SameQueryUsingIndexAndFullScan_ReturnsSameResults(t *testing.T) {
	testCases := []struct {
		sdlMetric string
		metric    client.DistanceMetric
		// Cosine is absent: it scores "short" and "long" equally, so their order is an arbitrary
		// tiebreak nothing guarantees.
		order []string
	}{
		{"EUCLIDEAN", client.DistanceMetricEuclidean, []string{"diag", "short", "long"}},
		{"DOT", client.DistanceMetricDotProduct, []string{"long", "diag", "short"}},
	}

	vectors := map[string][]float32{
		"short": {0.5, 0, 0},
		"long":  {3, 0, 0},
		"diag":  {0.9, 0.4, 0},
	}

	// The index is only used with a limit, so dropping it is the least invasive way to force a full
	// scan: the index, its metric and the ordering all stay as they are.
	indexedReq := `query {
		User(order: {_alias: {sim: DESC}}, limit: 2){
			name
			sim: SIMILARITY(vector: {vector: [1, 0, 0]})
		}
	}`
	fullScanReq := `query {
		User(order: {_alias: {sim: DESC}}){
			name
			sim: SIMILARITY(vector: {vector: [1, 0, 0]})
		}
	}`

	for _, testCase := range testCases {
		expected := make([]map[string]any, 0, len(testCase.order))
		for _, name := range testCase.order {
			source := make([]float64, len(vectors[name]))
			for i, v := range vectors[name] {
				source[i] = float64(v)
			}
			expected = append(expected, map[string]any{
				"name": name,
				"sim":  testUtils.SimilarityScore(testCase.metric, source, []float64{1, 0, 0}),
			})
		}

		test := testUtils.TestCase{
			Actions: []any{
				&action.AddCollection{
					SDL: `type User {
						name: String
						vector: [Float32!] @index(vector: {dimensions: 3, hnsw: {metric: ` +
						testCase.sdlMetric + `}})
					}`,
				},
				&action.AddDoc{DocMap: map[string]any{"name": "short", "vector": vectors["short"]}},
				&action.AddDoc{DocMap: map[string]any{"name": "long", "vector": vectors["long"]}},
				&action.AddDoc{DocMap: map[string]any{"name": "diag", "vector": vectors["diag"]}},
				&action.WaitForIndexReady{CollectionID: 0},

				&action.Request{Request: indexedReq, Results: map[string]any{"User": expected[:2]}},
				&action.Request{
					Request:  makeExplainQuery(indexedReq),
					Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1).WithDocFetches(2),
				},

				// The same order, reached without the index. Dropping the limit is what forces the
				// full scan, so this reports the unused-index warning.
				&action.Request{
					Request: fullScanReq,
					Results: map[string]any{"User": expected},
					ExpectedWarnings: []client.GQLWarning{
						{Code: client.WarningCodeVectorIndexUnused, Detail: map[string]any{"field": "vector", "reason": "noLimit"}},
					},
				},
				&action.Request{
					Request:  makeExplainQuery(fullScanReq),
					Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0).WithDocFetches(3),
					// Explaining a query still plans it, so the warning is reported here too.
					ExpectedWarnings: []client.GQLWarning{
						{Code: client.WarningCodeVectorIndexUnused, Detail: map[string]any{"field": "vector", "reason": "noLimit"}},
					},
				},
			},
		}

		t.Run(testCase.sdlMetric, func(t *testing.T) { testUtils.ExecuteTestCase(t, test) })
	}
}

// Two indexes on one field with different metrics would leave it arbitrary which one answers a
// query, so the second is rejected.
func TestVectorIndex_SecondIndexOnFieldWithDifferentMetric_IsRejected(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!] @index(vector: {dimensions: 3, hnsw: {metric: COSINE}})
				}`,
			},
			&action.NewIndex{
				CollectionID: 0,
				FieldName:    "vector",
				Vector: &client.VectorIndexDescription{
					Metric:     client.DistanceMetricEuclidean,
					Dimensions: 3,
					HNSW:       &client.HNSWParams{},
				},
				ExpectedError: "field already has a vector index with a different distance metric",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Dropping and recreating is the rebuild the rejection above asks for, so the new metric is accepted.
// This is also the index API's path for a metric, which the CLI and C clients carry as JSON.
func TestVectorIndex_DropThenRecreateWithDifferentMetric_IsAllowed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!] @index(vector: {dimensions: 3, hnsw: {metric: COSINE}})
				}`,
			},
			&action.AddDoc{DocMap: map[string]any{"name": "x", "vector": []float32{1, 0, 0}}},
			&action.WaitForIndexReady{CollectionID: 0},
			&action.DeleteIndex{CollectionID: 0, IndexName: "User_vector_ASC"},
			&action.NewIndex{
				CollectionID: 0,
				IndexName:    "User_vector_ASC",
				FieldName:    "vector",
				Vector: &client.VectorIndexDescription{
					Metric:     client.DistanceMetricEuclidean,
					Dimensions: 3,
					HNSW: &client.HNSWParams{
						M:              client.DefaultHNSWM,
						EfConstruction: client.DefaultHNSWEfConstruction,
						EfSearch:       client.DefaultHNSWEfSearch,
					},
				},
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name: "User_vector_ASC",
						// The recreated index gets a fresh id; the dropped one had 1.
						ID:     2,
						Fields: []client.IndexedFieldDescription{{Name: "vector"}},
						Kind:   client.IndexKindVector,
						KindDescription: &client.VectorIndexDescription{
							Fields:     []string{"vector"},
							Algorithm:  client.VectorAlgorithmHNSW,
							Metric:     client.DistanceMetricEuclidean,
							Dimensions: 3,
							HNSW: &client.HNSWParams{
								M:              client.DefaultHNSWM,
								EfConstruction: client.DefaultHNSWEfConstruction,
								EfSearch:       client.DefaultHNSWEfSearch,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Vectors whose squared distance exceeds a float32 are searchable and scored correctly. The graph's
// own ordering is covered by the engine tests; this covers the query path end to end, where the
// score is computed separately and was never affected.
// https://github.com/sourcenetwork/defradb/issues/5220
func TestVectorIndex_EuclideanOnLargeVectors_OrdersByDistance(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!] @index(vector: {dimensions: 1, hnsw: {metric: EUCLIDEAN}})
				}`,
			},
			&action.AddDoc{DocMap: map[string]any{"name": "mid", "vector": []float32{2e19}}},
			&action.AddDoc{DocMap: map[string]any{"name": "far", "vector": []float32{3e19}}},
			&action.AddDoc{DocMap: map[string]any{"name": "farthest", "vector": []float32{4e19}}},
			&action.WaitForIndexReady{CollectionID: 0},
			&action.Request{
				Request: `query {
					User(order: {_alias: {sim: DESC}}, limit: 1){
						name
						sim: SIMILARITY(vector: {vector: [0]})
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "mid", "sim": -3.999999984405158e+38},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// The dot product of two large vectors also exceeds a float32, so it needs the same float64 handling
// as euclidean. Cosine cannot overflow because its vectors are normalised first, but it is covered
// alongside so a future change to either metric is caught here.
// https://github.com/sourcenetwork/defradb/issues/5220
func TestVectorIndex_DotProductOnLargeVectors_OrdersByDistance(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!] @index(vector: {dimensions: 1, hnsw: {metric: DOT}})
				}`,
			},
			&action.AddDoc{DocMap: map[string]any{"name": "near", "vector": []float32{4e19}}},
			&action.AddDoc{DocMap: map[string]any{"name": "mid", "vector": []float32{3e19}}},
			&action.AddDoc{DocMap: map[string]any{"name": "far", "vector": []float32{2e19}}},
			&action.WaitForIndexReady{CollectionID: 0},
			&action.Request{
				Request: `query {
					User(order: {_alias: {sim: DESC}}, limit: 3){
						name
						sim: SIMILARITY(vector: {vector: [1e20]})
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "near", "sim": 4.0000000723660884e+39},
						{"name": "mid", "sim": 3.000000164225731e+39},
						{"name": "far", "sim": 2.0000000361830442e+39},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestVectorIndex_CosineOnLargeVectors_OrdersByDistance(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!] @index(vector: {dimensions: 2, hnsw: {metric: COSINE}})
				}`,
			},
			&action.AddDoc{DocMap: map[string]any{"name": "aligned", "vector": []float32{2e19, 0}}},
			&action.AddDoc{DocMap: map[string]any{"name": "diagonal", "vector": []float32{2e19, 2e19}}},
			&action.AddDoc{DocMap: map[string]any{"name": "orthogonal", "vector": []float32{0, 2e19}}},
			&action.WaitForIndexReady{CollectionID: 0},
			&action.Request{
				Request: `query {
					User(order: {_alias: {sim: DESC}}, limit: 3){
						name
						sim: SIMILARITY(vector: {vector: [1e20, 0]})
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "aligned", "sim": 1.0},
						{"name": "diagonal", "sim": 0.7071067811865476},
						{"name": "orthogonal", "sim": 0.0},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
