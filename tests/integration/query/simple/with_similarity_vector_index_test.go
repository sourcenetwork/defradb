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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// A _similarity + order DESC + limit query on a ready @index(vector: {...}) returns the k nearest documents
// (nearest to [1,0,0] is "x", then "xy") and reads only those k, not the whole collection. The
// explain variant asserts two doc fetches (a full scan would read four).
func TestQuerySimple_WithSimilarityOnVectorIndex_ReturnsKNearest(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!] @index(vector: {dimensions: 3, hnsw: {metric: COSINE}})
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
			// Under explain: one index fetch and two doc fetches, proving it used the index.
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

// Updating a vector re-indexes it: after moving "b" onto the query vector, the query returns "b"
// first. Exercises the update path (delete then insert) end to end.
func TestQuerySimple_WithSimilarityOnVectorIndex_ReflectsUpdatedVector(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!] @index(vector: {dimensions: 3, hnsw: {metric: COSINE}})
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

// Deleting a document removes it from the index: after deleting the nearest doc "x", the same query
// returns the next-nearest "xy" instead. The delete is a soft delete in the graph (the node is
// tombstoned, not unlinked), but search skips tombstoned nodes, so from a query's point of view the
// document is gone immediately. Exercises the delete path end to end through the planner.
func TestQuerySimple_WithSimilarityOnVectorIndex_ExcludesDeletedDoc(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!] @index(vector: {dimensions: 3, hnsw: {metric: COSINE}})
				}`,
			},
			// x is nearest to the query [1,0,0], xy is second nearest.
			&action.AddDoc{DocMap: map[string]any{"name": "x", "vector": []float32{1, 0, 0}}},
			&action.AddDoc{DocMap: map[string]any{"name": "y", "vector": []float32{0, 1, 0}}},
			&action.AddDoc{DocMap: map[string]any{"name": "xy", "vector": []float32{0.9, 0.4, 0}}},
			// Delete the nearest doc "x" (added first, so DocID 0).
			testUtils.DeleteDoc{CollectionID: 0, DocID: 0},
			// The query must now skip the deleted "x" and return "xy" as nearest.
			&action.Request{
				Request: `query {
					User(order: {_alias: {sim: DESC}}, limit: 1){
						name
						sim: SIMILARITY(vector: {vector: [1, 0, 0]})
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "xy", "sim": testUtils.CosineSimilarity([]float64{0.9, 0.4, 0}, []float64{1, 0, 0})},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Cosine ignores magnitude: "long" [10,0,0] ties "unit" [1,0,0] at similarity 1, both ahead of the
// off-axis "off". Guards the normalization fix — the old raw dot product ranked "long" far above
// "unit".
func TestQuerySimple_WithSimilarityOnVectorIndex_IsMagnitudeInvariant(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!] @index(vector: {dimensions: 3, hnsw: {metric: COSINE}})
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

// The index answers "nearest" (order DESC). Ascending order asks for the farthest documents, which
// HNSW cannot serve, so the query full-scans and still returns the correct farthest-first result.
// Explain confirms the fallback: no index fetch, all four documents scanned.
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
					vector: [Float32!] @index(vector: {dimensions: 3, hnsw: {metric: COSINE}})
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
				ExpectedWarnings: ascendingFallbackWarning(),
			},
			&action.Request{
				Request: `query @explain(type: execute) {
					User(order: {_alias: {sim: ASC}}, limit: 2){
						sim: SIMILARITY(vector: {vector: [1, 0, 0]})
					}
				}`,
				Asserter:         testUtils.NewExplainAsserter().WithIndexFetches(0).WithDocFetches(4),
				ExpectedWarnings: ascendingFallbackWarning(),
			},
		)...),
	}

	testUtils.ExecuteTestCase(t, test)
}

// With an offset, the graph must return Limit+Offset neighbours so the offset can skip the nearest
// ones and still leave a full page. Requesting only Limit would let the offset skip the whole page
// and return too few. Here offset 1 drops the nearest ("x"), so the page is the 2nd and 3rd nearest.
func TestQuerySimple_WithSimilarityOnVectorIndex_RespectsOffset(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!] @index(vector: {dimensions: 3, hnsw: {metric: COSINE}})
				}`,
			},
			&action.AddDoc{DocMap: map[string]any{"name": "x", "vector": []float32{1, 0, 0}}},
			&action.AddDoc{DocMap: map[string]any{"name": "a", "vector": []float32{0.9, 0.4, 0}}},
			&action.AddDoc{DocMap: map[string]any{"name": "b", "vector": []float32{0.6, 0.8, 0}}},
			&action.AddDoc{DocMap: map[string]any{"name": "c", "vector": []float32{0, 1, 0}}},
			&action.Request{
				Request: `query {
					User(order: {_alias: {sim: DESC}}, limit: 2, offset: 1){
						name
						sim: SIMILARITY(vector: {vector: [1, 0, 0]})
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "a", "sim": testUtils.CosineSimilarity([]float64{0.9, 0.4, 0}, []float64{1, 0, 0})},
						{"name": "b", "sim": testUtils.CosineSimilarity([]float64{0.6, 0.8, 0}, []float64{1, 0, 0})},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// A query vector whose length differs from the index dimensions would be scored on only its shared
// leading elements, giving wrong results. Both the indexed and full-scan paths must error on this.
func TestQuerySimple_WithSimilarityOnVectorIndex_WrongLengthQueryErrors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!] @index(vector: {dimensions: 3, hnsw: {metric: COSINE}})
				}`,
			},
			&action.AddDoc{DocMap: map[string]any{"name": "x", "vector": []float32{1, 0, 0}}},
			&action.AddDoc{DocMap: map[string]any{"name": "y", "vector": []float32{0, 1, 0}}},
			&action.Request{
				Request: `query {
					User(order: {_alias: {sim: DESC}}, limit: 1){
						name
						sim: SIMILARITY(vector: {vector: [1, 0]})
					}
				}`,
				ExpectedError: "source and vector must be of the same length",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Without an order clause, a _similarity + limit query asks for any k documents, not the k nearest,
// so the index must not be used (it would wrongly return the nearest). Explain asserts no index fetch.
// (docFetches cannot show this: a limit reads only k documents whether or not the index was used.)
func TestQuerySimple_WithSimilarityOnVectorIndex_NoOrderDoesNotUseIndex(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!] @index(vector: {dimensions: 3, hnsw: {metric: COSINE}})
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
				Asserter:         testUtils.NewExplainAsserter().WithIndexFetches(0),
				ExpectedWarnings: ascendingFallbackWarning(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// ascendingFallbackWarning is the warning reported when a similarity query on an indexed field is
// not ordered by that similarity descending, so it reads the whole collection.
func ascendingFallbackWarning() []client.GQLWarning {
	return []client.GQLWarning{
		{
			Code:   client.WarningCodeVectorIndexUnused,
			Detail: map[string]any{"field": "vector", "reason": "notOrderedBySimilarityDesc"},
		},
	}
}
