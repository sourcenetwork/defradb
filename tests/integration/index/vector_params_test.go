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

// An HNSW parameter above its maximum is rejected at index creation. Without the cap an oversized M
// makes the first insert do a burst of work that any client can trigger when Node-ACP is off.
func TestVectorIndex_CreateWithOversizedM_IsRejected(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!] @index(vector: {dimensions: 3, hnsw: {metric: COSINE, M: 100000}})
				}`,
				ExpectedError: "vector index parameter is out of range",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestVectorIndex_CreateAsUnique_IsRejected(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!]
				}`,
			},
			&action.NewIndex{
				CollectionID: 0,
				FieldName:    "vector",
				Unique:       true,
				Vector: &client.VectorIndexDescription{
					Algorithm:  client.VectorAlgorithmHNSW,
					Metric:     client.DistanceMetricCosine,
					Dimensions: 3,
					HNSW:       &client.HNSWParams{},
				},
				ExpectedError: "only an ordered index can be unique",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestVectorIndex_CreateWithDirection_IsRejected(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!]
				}`,
			},
			&action.NewIndex{
				CollectionID: 0,
				Fields: []client.IndexedFieldDescription{
					{Name: "vector", Descending: true},
				},
				Vector: &client.VectorIndexDescription{
					Algorithm:  client.VectorAlgorithmHNSW,
					Metric:     client.DistanceMetricCosine,
					Dimensions: 3,
					HNSW:       &client.HNSWParams{},
				},
				ExpectedError: "vector index cannot have a direction",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// A vector index can be created through the index API (not just the SDL directive) on a populated
// collection, and a similarity query then routes to it. This exercises the vector path of the index
// API on every client, including the CLI and C bindings that carry the config as JSON.
func TestVectorIndex_CreateViaIndexAPI_ThenQueryRoutes(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!]
				}`,
			},
			&action.AddDoc{DocMap: map[string]any{"name": "x", "vector": []float32{1, 0, 0}}},
			&action.AddDoc{DocMap: map[string]any{"name": "y", "vector": []float32{0, 1, 0}}},
			&action.AddDoc{DocMap: map[string]any{"name": "xy", "vector": []float32{0.9, 0.4, 0}}},
			&action.NewIndex{
				CollectionID: 0,
				FieldName:    "vector",
				Vector: &client.VectorIndexDescription{
					Algorithm:  client.VectorAlgorithmHNSW,
					Metric:     client.DistanceMetricCosine,
					Dimensions: 3,
					HNSW: &client.HNSWParams{
						M:              client.DefaultHNSWM,
						EfConstruction: client.DefaultHNSWEfConstruction,
						EfSearch:       client.DefaultHNSWEfSearch,
					},
				},
			},
			&action.Request{
				Request: `query {
					User(order: {_alias: {sim: DESC}}, limit: 2){
						name
						sim: SIMILARITY(vector: {vector: [1, 0, 0]})
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "x", "sim": testUtils.CosineSimilarity([]float64{1, 0, 0}, []float64{1, 0, 0})},
						{"name": "xy", "sim": testUtils.CosineSimilarity([]float64{0.9, 0.4, 0}, []float64{1, 0, 0})},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// A vector index request that gives only the essentials (an empty HNSW config, no algorithm, metric,
// or params) is filled in with defaults and works. The algorithm is chosen by the config object being
// present, so the caller never names it. Guards against a sparse request creating an index that then
// fails when it is used.
func TestVectorIndex_CreateWithDefaultParams_Works(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!]
				}`,
			},
			&action.AddDoc{DocMap: map[string]any{"name": "x", "vector": []float32{1, 0, 0}}},
			&action.AddDoc{DocMap: map[string]any{"name": "y", "vector": []float32{0, 1, 0}}},
			&action.NewIndex{
				CollectionID: 0,
				FieldName:    "vector",
				// Only the config object and dimensions; no algorithm, metric, or params.
				Vector: &client.VectorIndexDescription{
					Dimensions: 3,
					HNSW:       &client.HNSWParams{},
				},
			},
			&action.Request{
				Request: `query {
					User(order: {_alias: {sim: DESC}}, limit: 1){
						name
						sim: SIMILARITY(vector: {vector: [1, 0, 0]})
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "x", "sim": testUtils.CosineSimilarity([]float64{1, 0, 0}, []float64{1, 0, 0})},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Both spellings set, naming different fields: rejected.
func TestIndexNew_FieldsDisagreeBetweenSpellings_IsRejected(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{SDL: `type User { name: String
				vector: [Float32!] }`},
			&action.NewIndex{
				CollectionID: 0,
				Fields:       []client.IndexedFieldDescription{{Name: "name"}},
				Vector: &client.VectorIndexDescription{
					Fields: []string{"vector"}, Metric: client.DistanceMetricCosine,
					Dimensions: 3, HNSW: &client.HNSWParams{},
				},
				ExpectedError: "naming different fields",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
