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
// the k documents nearest to the query vector, in nearest-first order (the nearest to [1,0,0] is "x",
// then "xy"), AND fetches only those k documents rather than scanning the whole collection: the
// explain variant of the same query reports two doc fetches (a full-scan fallback would fetch four).
func TestQuerySimple_WithSimilarityOnVectorIndex_ReturnsKNearest(t *testing.T) {
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
						{"name": "x", "sim": testUtils.CosineSimilarity([]float64{1, 0, 0}, []float64{1, 0, 0})},
						{"name": "xy", "sim": testUtils.CosineSimilarity([]float64{0.9, 0.4, 0}, []float64{1, 0, 0})},
					},
				},
			},
			// The same query under explain reports one index fetch (the graph search) and fetches only
			// the two nearest documents, proving it routed to the vector index rather than full-scanning.
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

// Updating a document's vector re-indexes it: after moving "b" to sit right on the query vector, the
// nearest-neighbour query returns "b" first. This exercises the write-path graph maintenance (update
// = delete-then-insert) end to end through the query.
func TestQuerySimple_WithSimilarityOnVectorIndex_ReflectsUpdatedVector(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!] @vectorIndex(dimensions: 3, HNSW: {metric: COSINE})
				}`,
			},
			// a sits off the query axis; b starts even further off. After the update b lands exactly on
			// the query direction, making it uniquely nearest.
			&action.AddDoc{DocMap: map[string]any{"name": "a", "vector": []float32{0.6, 0.8, 0}}},
			&action.AddDoc{DocMap: map[string]any{"name": "b", "vector": []float32{0, 1, 0}}},
			&action.UpdateDoc{
				CollectionID: 0,
				DocID:        1,
				Doc:          `{"vector": [1, 0, 0]}`,
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
						{"name": "b", "sim": testUtils.CosineSimilarity([]float64{1, 0, 0}, []float64{1, 0, 0})},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Cosine ignores magnitude, so a very long vector pointing the same way as a short one is equally
// near. Here "long" points exactly along the query but with magnitude 10; it ties "unit" at
// similarity 1 and both beat the off-axis "off". This is the case the old unnormalised dot product
// got wrong (it would have ranked the long vector far above the unit one), so it guards the fix.
func TestQuerySimple_WithSimilarityOnVectorIndex_IsMagnitudeInvariant(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!] @vectorIndex(dimensions: 3, HNSW: {metric: COSINE})
				}`,
			},
			&action.AddDoc{DocMap: map[string]any{"name": "unit", "vector": []float32{1, 0, 0}}},
			&action.AddDoc{DocMap: map[string]any{"name": "long", "vector": []float32{10, 0, 0}}},
			&action.AddDoc{DocMap: map[string]any{"name": "off", "vector": []float32{0, 1, 0}}},
			&action.Request{
				Request: `query {
					User(order: {_alias: {sim: DESC}}, limit: 2){
						sim: SIMILARITY(vector: {vector: [1, 0, 0]})
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						// unit [1,0,0] and long [10,0,0] both point along the query, so both are 1.
						{"sim": testUtils.CosineSimilarity([]float64{1, 0, 0}, []float64{1, 0, 0})},
						{"sim": testUtils.CosineSimilarity([]float64{10, 0, 0}, []float64{1, 0, 0})},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// The vector index answers "nearest", which is order:DESC. An ascending order asks for the farthest
// documents, which HNSW cannot serve, so the query must fall back to a full scan and still return the
// correct farthest-first result. The explain variant confirms the fallback: no index fetch, and all
// four documents scanned.
func TestQuerySimple_WithSimilarityOnVectorIndex_AscendingOrderFullScans(t *testing.T) {
	docs := []any{
		&action.AddDoc{DocMap: map[string]any{"name": "near", "vector": []float32{1, 0, 0}}},
		&action.AddDoc{DocMap: map[string]any{"name": "mid", "vector": []float32{0.8, 0.6, 0}}},
		&action.AddDoc{DocMap: map[string]any{"name": "far", "vector": []float32{0.1, 1, 0}}},
		&action.AddDoc{DocMap: map[string]any{"name": "opp", "vector": []float32{-1, 0.2, 0}}},
	}

	test := testUtils.TestCase{
		Actions: append([]any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!] @vectorIndex(dimensions: 3, HNSW: {metric: COSINE})
				}`,
			},
		}, append(docs,
			// The two farthest from [1,0,0]: opp (negative similarity) then far.
			&action.Request{
				Request: `query {
					User(order: {_alias: {sim: ASC}}, limit: 2){
						name
						sim: SIMILARITY(vector: {vector: [1, 0, 0]})
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "opp", "sim": testUtils.CosineSimilarity([]float64{-1, 0.2, 0}, []float64{1, 0, 0})},
						{"name": "far", "sim": testUtils.CosineSimilarity([]float64{0.1, 1, 0}, []float64{1, 0, 0})},
					},
				},
			},
			&action.Request{
				Request: `query @explain(type: execute) {
					User(order: {_alias: {sim: ASC}}, limit: 2){
						sim: SIMILARITY(vector: {vector: [1, 0, 0]})
					}
				}`,
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0).WithDocFetches(4),
			},
		)...),
	}

	testUtils.ExecuteTestCase(t, test)
}

// Without an order clause a _similarity + limit query is not a nearest-neighbour query: it asks for
// any k documents plus their similarity, not the k nearest. Routing would wrongly return the nearest,
// so the query must not route. Explain confirms it: no index fetch. (docFetches cannot show this, as
// a limit reads only k documents whether or not it routed.)
func TestQuerySimple_WithSimilarityOnVectorIndex_NoOrderDoesNotRoute(t *testing.T) {
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
			&action.AddDoc{DocMap: map[string]any{"name": "z", "vector": []float32{0, 0, 1}}},
			&action.Request{
				Request: `query @explain(type: execute) {
					User(limit: 2){
						sim: SIMILARITY(vector: {vector: [1, 0, 0]})
					}
				}`,
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
