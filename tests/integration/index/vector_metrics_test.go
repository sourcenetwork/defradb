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

// vectorIndexWithMetric is the index the SDL directive below produces, for ListIndexes to assert.
func vectorIndexWithMetric(id uint32, metric client.DistanceMetric) []client.IndexDescription {
	return []client.IndexDescription{
		{
			Name:   "User_vector_ASC",
			ID:     id,
			Fields: []client.IndexedFieldDescription{{Name: "vector"}},
			Kind:   client.IndexKindVector,
			KindDescription: &client.VectorIndexDescription{
				Algorithm:  client.VectorAlgorithmHNSW,
				Metric:     metric,
				Dimensions: 3,
				HNSW: &client.HNSWParams{
					M:              client.DefaultHNSWM,
					EfConstruction: client.DefaultHNSWEfConstruction,
					EfSearch:       client.DefaultHNSWEfSearch,
				},
			},
		},
	}
}

// A non-cosine metric survives the whole path: the directive parses it, it is stored, and writes
// maintain the graph. A metric the engine could not use would show up as a failed build.
//
// Results stay cosine-scored: `_similarity` is cosine regardless of any index, and the planner leaves
// a non-cosine index alone (the explain assertion below).
func TestVectorIndex_EuclideanMetric_IndexIsBuiltAndMaintained(t *testing.T) {
	testUtils.ExecuteTestCase(t, metricLifecycleTest("EUCLIDEAN", client.DistanceMetricEuclidean))
}

func TestVectorIndex_DotProductMetric_IndexIsBuiltAndMaintained(t *testing.T) {
	testUtils.ExecuteTestCase(t, metricLifecycleTest("DOT", client.DistanceMetricDotProduct))
}

func metricLifecycleTest(sdlMetric string, metric client.DistanceMetric) testUtils.TestCase {
	req := `query {
		User(order: {_alias: {sim: DESC}}, limit: 2){
			name
			sim: SIMILARITY(vector: {vector: [1, 0, 0]})
		}
	}`

	return testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!] @vectorIndex(dimensions: 3, HNSW: {metric: ` + sdlMetric + `})
				}`,
			},
			&action.AddDoc{DocMap: map[string]any{"name": "x", "vector": []float32{1, 0, 0}}},
			&action.AddDoc{DocMap: map[string]any{"name": "y", "vector": []float32{0, 1, 0}}},
			&action.AddDoc{DocMap: map[string]any{"name": "xy", "vector": []float32{0.9, 0.4, 0}}},
			// An update and a delete cover the maintenance path, not just the insert.
			&action.UpdateDoc{
				DocID: 1,
				Doc:   `{"vector": [0, 0.5, 0]}`,
			},
			testUtils.DeleteDoc{DocID: 2},
			&action.WaitForIndexReady{CollectionID: 0},
			&action.ListIndexes{
				CollectionID:    0,
				ExpectedIndexes: vectorIndexWithMetric(1, metric),
			},
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
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0).WithDocFetches(2),
			},
		},
	}
}

// A cosine index still routes. Without this, the two tests above would pass even if the planner
// never routed to any vector index.
func TestVectorIndex_CosineMetric_StillRoutesToIndex(t *testing.T) {
	req := `query {
		User(order: {_alias: {sim: DESC}}, limit: 2){
			name
			sim: SIMILARITY(vector: {vector: [1, 0, 0]})
		}
	}`

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!] @vectorIndex(dimensions: 3, HNSW: {metric: COSINE})
				}`,
			},
			&action.AddDoc{DocMap: map[string]any{"name": "x", "vector": []float32{1, 0, 0}}},
			&action.AddDoc{DocMap: map[string]any{"name": "y", "vector": []float32{0, 1, 0}}},
			&action.AddDoc{DocMap: map[string]any{"name": "xy", "vector": []float32{0.9, 0.4, 0}}},
			&action.WaitForIndexReady{CollectionID: 0},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1).WithDocFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
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
				CollectionID:    0,
				ExpectedIndexes: vectorIndexWithMetric(2, client.DistanceMetricEuclidean),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
