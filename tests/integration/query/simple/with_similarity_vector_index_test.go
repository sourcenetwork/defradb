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

package simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// A _similarity + order:{_similarity:DESC} + limit query on a field with a ready @vectorIndex returns
// the k documents nearest to the query vector, in nearest-first order. The vectors are chosen so the
// nearest to [1,0,0] is clearly "x", then "xy", then "y"; the query asks for the two nearest.
func TestQuerySimple_WithSimilarityOnVectorIndex_ReturnsKNearest(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!] @vectorIndex(type: HNSW, metric: COSINE, dimensions: 3)
				}`,
			},
			&action.AddDoc{DocMap: map[string]any{"name": "x", "vector": []float32{1, 0, 0}}},
			&action.AddDoc{DocMap: map[string]any{"name": "y", "vector": []float32{0, 1, 0}}},
			&action.AddDoc{DocMap: map[string]any{"name": "xy", "vector": []float32{0.9, 0.4, 0}}},
			&action.AddDoc{DocMap: map[string]any{"name": "z", "vector": []float32{0, 0, 1}}},
			&action.Request{
				Request: `query {
					User(order: {_alias: {sim: DESC}}, limit: 2){
						name
						sim: SIMILARITY(vector: {vector: [1, 0, 0]})
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "x", "sim": float64(1)},
						{"name": "xy", "sim": float64(0.9138115423811489)},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// The routed query fetches only the k nearest documents, not the whole collection: with four
// documents and limit 2, the scan reads two. This proves the vector index narrowed the scan (a
// full-scan fallback would fetch all four).
func TestQuerySimple_WithSimilarityOnVectorIndex_FetchesOnlyKDocuments(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!] @vectorIndex(type: HNSW, metric: COSINE, dimensions: 3)
				}`,
			},
			&action.AddDoc{DocMap: map[string]any{"name": "x", "vector": []float32{1, 0, 0}}},
			&action.AddDoc{DocMap: map[string]any{"name": "y", "vector": []float32{0, 1, 0}}},
			&action.AddDoc{DocMap: map[string]any{"name": "xy", "vector": []float32{0.9, 0.4, 0}}},
			&action.AddDoc{DocMap: map[string]any{"name": "z", "vector": []float32{0, 0, 1}}},
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
