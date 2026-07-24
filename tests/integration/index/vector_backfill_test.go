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

// A vector index created on a populated collection is backfilled by the same async worker that builds
// scalar indexes. The build is held at the gate to observe the building state, then released; once
// ready, a nearest-neighbour query fetches only the two nearest of the four pre-existing documents.
//
// The docFetches assertion is what proves the backfill actually built a usable graph: a failed or
// empty backfill would leave the index non-ready, the query would fall back to a full scan, and
// docFetches would be four, not two.
func TestVectorIndex_AsyncBackfillOverExistingDocs_BuildsUsableGraph(t *testing.T) {
	release, cleanup := installBuildGate(t)
	defer cleanup()

	vectorIndex := &action.NewIndex{
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

	expectedIndexes := []client.IndexDescription{
		{
			Name:   "User_vector_ASC",
			ID:     1,
			Fields: []client.IndexedFieldDescription{{Name: "vector"}},
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
	}

	test := testUtils.TestCase{
		// The build gate is an in-process hook, so it only takes effect with the Go client running the
		// DB in-process, on a single backend (see the var declarations for the full reasoning).
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
			// The gate this test installed holds the build at its first batch, so the index is caught
			// in-progress here deterministically rather than racing a build that might already be done.
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
			// Releasing that gate is what lets the held build run to completion; without it the wait
			// below would time out. (A build with no gate installed would just finish on its own.)
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
				Asserter: testUtils.NewExplainAsserter().WithDocFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
