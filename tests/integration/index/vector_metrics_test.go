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

// Routing and scoring must both follow the index's metric: doing one without the other would rank by
// one metric and score by another. A metric the engine cannot use fails the build instead.
func TestVectorIndex_QueryOnAnyMetric_RoutesAndScoresByIndexMetric(t *testing.T) {
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
						vector: [Float32!] @vectorIndex(dimensions: 3, HNSW: {metric: ` +
						testCase.sdlMetric + `})
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

// Routing must not change the answer. The vectors are ranked differently by each metric, so a wrong
// metric cannot pass by coincidence:
//
//	           cosine   -L2      dot
//	short      1.0      -0.25    0.5
//	long       1.0      -4.0     3.0
//	diag       0.914    -0.17    0.9
//
// Both runs share one indexed collection: dropping the index would not full-scan the same search, it
// would make it a cosine one. Omitting the limit full-scans while leaving the metric intact.
func TestVectorIndex_SameQueryRoutedAndFullScanned_ReturnsSameResults(t *testing.T) {
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

	// Routing needs a limit, so dropping it is the least invasive way to force the full scan. Any of
	// the other conditions in tryRouteSimilarityToVectorIndex would do just as well.
	routedReq := `query {
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
						vector: [Float32!] @vectorIndex(dimensions: 3, HNSW: {metric: ` +
						testCase.sdlMetric + `})
					}`,
				},
				&action.AddDoc{DocMap: map[string]any{"name": "short", "vector": vectors["short"]}},
				&action.AddDoc{DocMap: map[string]any{"name": "long", "vector": vectors["long"]}},
				&action.AddDoc{DocMap: map[string]any{"name": "diag", "vector": vectors["diag"]}},
				&action.WaitForIndexReady{CollectionID: 0},

				&action.Request{Request: routedReq, Results: map[string]any{"User": expected[:2]}},
				&action.Request{
					Request:  makeExplainQuery(routedReq),
					Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1).WithDocFetches(2),
				},

				// The same order, reached without the graph.
				&action.Request{Request: fullScanReq, Results: map[string]any{"User": expected}},
				&action.Request{
					Request:  makeExplainQuery(fullScanReq),
					Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0).WithDocFetches(3),
				},
			},
		}

		t.Run(testCase.sdlMetric, func(t *testing.T) { testUtils.ExecuteTestCase(t, test) })
	}
}

// Changing the metric would need the graph rebuilt, so the request is rejected instead.
func TestVectorIndex_ChangeMetricOnExistingIndex_IsRejected(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!] @vectorIndex(dimensions: 3, HNSW: {metric: COSINE})
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
				ExpectedError: "cannot change the distance metric of an existing vector index",
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
					vector: [Float32!] @vectorIndex(dimensions: 3, HNSW: {metric: COSINE})
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
