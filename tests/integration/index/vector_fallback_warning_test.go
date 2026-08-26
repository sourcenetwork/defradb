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

// vectorWarningSetup builds a collection with a cosine vector index and four documents. Every test
// here uses the same data and differs only in the shape of the request.
func vectorWarningSetup() []any {
	return []any{
		&action.AddCollection{
			SDL: `type User {
				name: String
				age: Int
				vector: [Float32!] @vectorIndex(dimensions: 3, HNSW: {metric: COSINE})
			}`,
		},
		&action.AddDoc{DocMap: map[string]any{"name": "x", "age": 10, "vector": []float32{1, 0, 0}}},
		&action.AddDoc{DocMap: map[string]any{"name": "y", "age": 20, "vector": []float32{0, 1, 0}}},
		&action.AddDoc{DocMap: map[string]any{"name": "xy", "age": 30, "vector": []float32{0.9, 0.4, 0}}},
		&action.AddDoc{DocMap: map[string]any{"name": "z", "age": 40, "vector": []float32{0, 0, 1}}},
	}
}

// unusedIndexWarning builds the warning a fallback is expected to report.
func unusedIndexWarning(reason string) []client.GQLWarning {
	return []client.GQLWarning{
		{
			Code: client.WarningCodeVectorIndexUnused,
			Detail: map[string]any{
				"field":  "vector",
				"reason": reason,
			},
		},
	}
}

// The control. Without it, tests that only check the fallbacks would pass even if every query
// warned.
func TestVectorIndexWarning_QueryUsesIndex_ReportsNoWarning(t *testing.T) {
	test := testUtils.TestCase{
		Actions: append(vectorWarningSetup(),
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
		),
	}

	testUtils.ExecuteTestCase(t, test)
}

// Ascending asks for the farthest documents, which the graph cannot serve. This is the easiest
// mistake to make: anyone thinking in distances rather than similarities reaches for ASC.
func TestVectorIndexWarning_OrderedAscending_ReportsWarning(t *testing.T) {
	test := testUtils.TestCase{
		Actions: append(vectorWarningSetup(),
			&action.Request{
				Request: `query {
					User(order: {_alias: {sim: ASC}}, limit: 2){
						name
						sim: SIMILARITY(vector: {vector: [1, 0, 0]})
					}
				}`,
				// "y" and "z" both score 0, so their relative order is not defined.
				NonOrderedResults: true,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "z", "sim": testUtils.CosineSimilarity([]float64{0, 0, 1}, []float64{1, 0, 0})},
						{"name": "y", "sim": testUtils.CosineSimilarity([]float64{0, 1, 0}, []float64{1, 0, 0})},
					},
				},
				ExpectedWarnings: unusedIndexWarning("notOrderedBySimilarityDesc"),
			},
		),
	}

	testUtils.ExecuteTestCase(t, test)
}

// Without a limit there is no k for the graph to return, so the whole collection is read.
func TestVectorIndexWarning_NoLimit_ReportsWarning(t *testing.T) {
	test := testUtils.TestCase{
		Actions: append(vectorWarningSetup(),
			&action.Request{
				Request: `query {
					User(order: {_alias: {sim: DESC}}){
						name
						sim: SIMILARITY(vector: {vector: [1, 0, 0]})
					}
				}`,
				// "y" and "z" both score 0, so their relative order is not defined.
				NonOrderedResults: true,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "x", "sim": testUtils.CosineSimilarity([]float64{1, 0, 0}, []float64{1, 0, 0})},
						{"name": "xy", "sim": testUtils.CosineSimilarity([]float64{0.9, 0.4, 0}, []float64{1, 0, 0})},
						{"name": "y", "sim": testUtils.CosineSimilarity([]float64{0, 1, 0}, []float64{1, 0, 0})},
						{"name": "z", "sim": testUtils.CosineSimilarity([]float64{0, 0, 1}, []float64{1, 0, 0})},
					},
				},
				ExpectedWarnings: unusedIndexWarning("noLimit"),
			},
		),
	}

	testUtils.ExecuteTestCase(t, test)
}

// Lifting this is https://github.com/sourcenetwork/defradb/issues/5071
func TestVectorIndexWarning_WithFilter_ReportsWarning(t *testing.T) {
	test := testUtils.TestCase{
		Actions: append(vectorWarningSetup(),
			&action.Request{
				Request: `query {
					User(filter: {age: {_gt: 15}}, order: {_alias: {sim: DESC}}, limit: 2){
						name
						sim: SIMILARITY(vector: {vector: [1, 0, 0]})
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "xy", "sim": testUtils.CosineSimilarity([]float64{0.9, 0.4, 0}, []float64{1, 0, 0})},
						{"name": "y", "sim": testUtils.CosineSimilarity([]float64{0, 1, 0}, []float64{1, 0, 0})},
					},
				},
				ExpectedWarnings: unusedIndexWarning("filter"),
			},
		),
	}

	testUtils.ExecuteTestCase(t, test)
}

// Filtering the k nearest after the fact can reject all of them and return nothing while matching
// documents exist. Only "z" matches age > 35 and it is the farthest from the query, so asking the
// graph for k=1 would find "x", drop it, and return an empty result.
func TestVectorIndexWarning_SelectiveFilter_StillReturnsMatchingDocs(t *testing.T) {
	test := testUtils.TestCase{
		Actions: append(vectorWarningSetup(),
			&action.Request{
				Request: `query {
					User(filter: {age: {_gt: 35}}, order: {_alias: {sim: DESC}}, limit: 1){
						name
						sim: SIMILARITY(vector: {vector: [1, 0, 0]})
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "z", "sim": testUtils.CosineSimilarity([]float64{0, 0, 1}, []float64{1, 0, 0})},
					},
				},
				ExpectedWarnings: unusedIndexWarning("filter"),
			},
		),
	}

	testUtils.ExecuteTestCase(t, test)
}

// With two similarity fields, which one drives the search is ambiguous, so the query full-scans.
func TestVectorIndexWarning_MultipleSimilarityFields_ReportsWarning(t *testing.T) {
	test := testUtils.TestCase{
		Actions: append(vectorWarningSetup(),
			&action.Request{
				Request: `query {
					User(order: {_alias: {sim: DESC}}, limit: 2){
						name
						sim: SIMILARITY(vector: {vector: [1, 0, 0]})
						other: SIMILARITY(vector: {vector: [0, 1, 0]})
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name":  "x",
							"sim":   testUtils.CosineSimilarity([]float64{1, 0, 0}, []float64{1, 0, 0}),
							"other": testUtils.CosineSimilarity([]float64{1, 0, 0}, []float64{0, 1, 0}),
						},
						{
							"name":  "xy",
							"sim":   testUtils.CosineSimilarity([]float64{0.9, 0.4, 0}, []float64{1, 0, 0}),
							"other": testUtils.CosineSimilarity([]float64{0.9, 0.4, 0}, []float64{0, 1, 0}),
						},
					},
				},
				ExpectedWarnings: unusedIndexWarning("multipleSimilarityFields"),
			},
		),
	}

	testUtils.ExecuteTestCase(t, test)
}

// Nothing to fall back from, so no warning. Otherwise every similarity query in a schema without
// vector indexes would report one.
func TestVectorIndexWarning_NoVectorIndexOnField_ReportsNoWarning(t *testing.T) {
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
			&action.Request{
				Request: `query {
					User(order: {_alias: {sim: ASC}}, limit: 1){
						name
						sim: SIMILARITY(vector: {vector: [1, 0, 0]})
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "y", "sim": testUtils.CosineSimilarity([]float64{0, 1, 0}, []float64{1, 0, 0})},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
