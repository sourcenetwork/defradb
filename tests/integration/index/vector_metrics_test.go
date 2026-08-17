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

// Every metric builds a graph that writes maintain, but only a cosine one answers the similarity
// query. `_similarity` scores by cosine, so a graph built for another metric holds a different set of
// k nearest; the planner leaves it alone and full-scans (indexFetches 0) rather than return the wrong
// documents. Results are identical either way, which is the point: the metric changes the cost, never
// the answer.
//
// A metric the engine could not use would surface here as a failed build instead of a ready index.
func TestVectorIndex_MetricDecidesWhetherQueryRoutes(t *testing.T) {
	testCases := []struct {
		sdlMetric string
		// indexFetches is 1 when the planner routed to the graph, 0 when it full-scanned.
		indexFetches int
	}{
		{"COSINE", 1},
		{"EUCLIDEAN", 0},
		{"DOT", 0},
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
							{"name": "x", "sim": testUtils.CosineSimilarity([]float64{1, 0, 0}, []float64{1, 0, 0})},
							{"name": "y", "sim": testUtils.CosineSimilarity([]float64{0, 0.5, 0}, []float64{1, 0, 0})},
						},
					},
				},
				&action.Request{
					Request: makeExplainQuery(req),
					Asserter: testUtils.NewExplainAsserter().
						WithIndexFetches(testCase.indexFetches).
						WithDocFetches(2),
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
