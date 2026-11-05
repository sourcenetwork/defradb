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

// newAsyncVectorIndex creates an async HNSW/cosine vector index with default params on the "vector"
// field of collection 0. Async so the build runs on the worker and can be held at the build gate.
func newAsyncVectorIndex() *action.NewIndex {
	return &action.NewIndex{
		CollectionID: 0,
		FieldName:    "vector",
		Async:        true,
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
	}
}

// Creating a vector index on a populated collection backfills the graph via the async worker. The
// explain assertion is what proves the backfill built a usable graph: a broken build would leave the
// index non-ready and the query would full-scan (indexFetches 0, docFetches 4) instead of using the index.
func TestVectorIndex_AsyncBackfillOverExistingDocs_BuildsUsableGraph(t *testing.T) {
	release, cleanup := installBuildGate(t)
	defer cleanup()

	vectorIndex := newAsyncVectorIndex()

	expectedIndexes := []client.IndexDescription{
		{
			Name:   "User_vector_ASC",
			ID:     1,
			Fields: []client.IndexedFieldDescription{{Name: "vector"}},
			Kind:   client.IndexKindVector,
			KindDescription: &client.VectorIndexDescription{
				Fields:     []string{"vector"},
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
	}

	test := testUtils.TestCase{
		// The build gate is an in-process hook, so this runs only on the Go client (see the var docs).
		SupportedClientTypes:   goClientTypes,
		SupportedDatabaseTypes: inProcessDBType,
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
			&action.AddDoc{DocMap: map[string]any{"name": "z", "vector": []float32{0, 0, 1}}},
			vectorIndex,
			// The gate holds the build, so the index is reliably caught in-progress here.
			&action.ListIndexes{
				CollectionID:    0,
				ExpectedIndexes: expectedIndexes,
				ExpectedStatuses: map[string]client.ActionExecution{
					"User_vector_ASC": {
						Status: client.InProgressActionStatus,
						Action: client.BackfillIndexAction,
					},
				},
			},
			// Release the gate so the build can finish; the wait below would otherwise time out.
			action.NewRunFunc(release),
			&action.WaitForIndexReady{CollectionID: 0},
			&action.ListIndexes{
				CollectionID:    0,
				ExpectedIndexes: expectedIndexes,
				ExpectedStatuses: map[string]client.ActionExecution{
					"User_vector_ASC": {Status: client.CompletedActionStatus},
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
			&action.Request{
				Request: `query @explain(type: execute) {
					User(order: {_alias: {sim: DESC}}, limit: 2){
						name
						sim: SIMILARITY(vector: {vector: [1, 0, 0]})
					}
				}`,
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1).WithDocFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// While a vector index is building, the planner must not use it: the query full-scans (indexFetches
// 0) but still returns correct results. Once ready, the same query uses the index (indexFetches
// 1).
func TestVectorIndex_QueryWhileBuilding_FullScansThenUsesIndex(t *testing.T) {
	release, cleanup := installBuildGate(t)
	defer cleanup()

	req := `query {
		User(order: {_alias: {sim: DESC}}, limit: 2){
			name
			sim: SIMILARITY(vector: {vector: [1, 0, 0]})
		}
	}`

	test := testUtils.TestCase{
		SupportedClientTypes:   goClientTypes,
		SupportedDatabaseTypes: inProcessDBType,
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
			&action.AddDoc{DocMap: map[string]any{"name": "z", "vector": []float32{0, 0, 1}}},
			newAsyncVectorIndex(),
			// Building: full-scan (no index fetch), but still correct — nearest to [1,0,0] are x then xy.
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "x", "sim": testUtils.CosineSimilarity([]float64{1, 0, 0}, []float64{1, 0, 0})},
						{"name": "xy", "sim": testUtils.CosineSimilarity([]float64{0.9, 0.4, 0}, []float64{1, 0, 0})},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0),
			},
			action.NewRunFunc(release),
			&action.WaitForIndexReady{CollectionID: 0},
			// Ready: the same query now uses the index.
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
