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

// The six documents below are strictly ordered by their cosine similarity to [1,0,0]: n1 is nearest
// and n6 farthest. The "group" field splits them so a filter can drop the nearest ones, forcing the
// scan to widen the graph search to find the requested number of matches.
func similarityFilterDocs() []any {
	return []any{
		&action.AddDoc{DocMap: map[string]any{"name": "n1", "group": "a", "vector": []float32{1, 0, 0}}},
		&action.AddDoc{DocMap: map[string]any{"name": "n2", "group": "a", "vector": []float32{0.95, 0.31, 0}}},
		&action.AddDoc{DocMap: map[string]any{"name": "n3", "group": "a", "vector": []float32{0.9, 0.44, 0}}},
		&action.AddDoc{DocMap: map[string]any{"name": "n4", "group": "b", "vector": []float32{0.8, 0.6, 0}}},
		&action.AddDoc{DocMap: map[string]any{"name": "n5", "group": "b", "vector": []float32{0.6, 0.8, 0}}},
		&action.AddDoc{DocMap: map[string]any{"name": "n6", "group": "b", "vector": []float32{0, 1, 0}}},
	}
}

func similarityFilterTestCase(actions ...any) testUtils.TestCase {
	return testUtils.TestCase{
		Actions: append(append([]any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					group: String
					vector: [Float32!] @vectorIndex(dimensions: 3, HNSW: {metric: COSINE})
				}`,
			},
		}, similarityFilterDocs()...), actions...),
	}
}

// A filter every document passes changes nothing: the first graph search of k documents already
// yields k results, so the scan never widens. Explain confirms it routed (one index fetch) and read
// only the two documents the search returned.
func TestQuerySimple_WithSimilarityOnVectorIndexAndFilter_PassingAllDocs(t *testing.T) {
	test := similarityFilterTestCase(
		&action.Request{
			Request: `query {
				User(filter: {group: {_in: ["a", "b"]}}, order: {_alias: {sim: DESC}}, limit: 2){
					name
					sim: SIMILARITY(vector: {vector: [1, 0, 0]})
				}
			}`,
			Results: map[string]any{
				"User": []map[string]any{
					{"name": "n1", "sim": testUtils.CosineSimilarity([]float64{1, 0, 0}, []float64{1, 0, 0})},
					{"name": "n2", "sim": testUtils.CosineSimilarity([]float64{0.95, 0.31, 0}, []float64{1, 0, 0})},
				},
			},
		},
		&action.Request{
			Request: `query @explain(type: execute) {
				User(filter: {group: {_in: ["a", "b"]}}, order: {_alias: {sim: DESC}}, limit: 2){
					name
					sim: SIMILARITY(vector: {vector: [1, 0, 0]})
				}
			}`,
			Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1).WithDocFetches(2),
		},
	)

	testUtils.ExecuteTestCase(t, test)
}

// The filter drops the three nearest documents, so the first search of k=2 yields nothing. The scan
// must widen the search until two matching documents have been found, rather than returning the
// short result the plain filter-after-search would give.
func TestQuerySimple_WithSimilarityOnVectorIndexAndFilter_WidensSearch(t *testing.T) {
	test := similarityFilterTestCase(
		&action.Request{
			Request: `query {
				User(filter: {group: {_eq: "b"}}, order: {_alias: {sim: DESC}}, limit: 2){
					name
					sim: SIMILARITY(vector: {vector: [1, 0, 0]})
				}
			}`,
			Results: map[string]any{
				"User": []map[string]any{
					{"name": "n4", "sim": testUtils.CosineSimilarity([]float64{0.8, 0.6, 0}, []float64{1, 0, 0})},
					{"name": "n5", "sim": testUtils.CosineSimilarity([]float64{0.6, 0.8, 0}, []float64{1, 0, 0})},
				},
			},
		},
		// Still routed to the vector index, and each document is fetched at most once across the
		// widening rounds: six documents exist, six are read.
		&action.Request{
			Request: `query @explain(type: execute) {
				User(filter: {group: {_eq: "b"}}, order: {_alias: {sim: DESC}}, limit: 2){
					name
					sim: SIMILARITY(vector: {vector: [1, 0, 0]})
				}
			}`,
			Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1).WithDocFetches(6),
		},
	)

	testUtils.ExecuteTestCase(t, test)
}

// Fewer matching documents than asked for: the search widens until the graph is exhausted and
// returns the one document that matches, not an error or a repeat.
func TestQuerySimple_WithSimilarityOnVectorIndexAndFilter_FewerMatchesThanLimit(t *testing.T) {
	test := similarityFilterTestCase(
		&action.Request{
			Request: `query {
				User(filter: {name: {_eq: "n5"}}, order: {_alias: {sim: DESC}}, limit: 3){
					name
					sim: SIMILARITY(vector: {vector: [1, 0, 0]})
				}
			}`,
			Results: map[string]any{
				"User": []map[string]any{
					{"name": "n5", "sim": testUtils.CosineSimilarity([]float64{0.6, 0.8, 0}, []float64{1, 0, 0})},
				},
			},
		},
	)

	testUtils.ExecuteTestCase(t, test)
}

// No document matches, so the widening stops when the graph is exhausted and the query returns
// nothing.
func TestQuerySimple_WithSimilarityOnVectorIndexAndFilter_NoMatches(t *testing.T) {
	test := similarityFilterTestCase(
		&action.Request{
			Request: `query {
				User(filter: {group: {_eq: "c"}}, order: {_alias: {sim: DESC}}, limit: 2){
					name
					sim: SIMILARITY(vector: {vector: [1, 0, 0]})
				}
			}`,
			Results: map[string]any{
				"User": []map[string]any{},
			},
		},
	)

	testUtils.ExecuteTestCase(t, test)
}

// An offset skips matching documents, so the search must widen until limit+offset of them have
// passed the filter. Here the three group "b" documents are all needed to serve a page of two after
// skipping one.
func TestQuerySimple_WithSimilarityOnVectorIndexAndFilter_WithOffset(t *testing.T) {
	test := similarityFilterTestCase(
		&action.Request{
			Request: `query {
				User(filter: {group: {_eq: "b"}}, order: {_alias: {sim: DESC}}, limit: 2, offset: 1){
					name
					sim: SIMILARITY(vector: {vector: [1, 0, 0]})
				}
			}`,
			Results: map[string]any{
				"User": []map[string]any{
					{"name": "n5", "sim": testUtils.CosineSimilarity([]float64{0.6, 0.8, 0}, []float64{1, 0, 0})},
					{"name": "n6", "sim": testUtils.CosineSimilarity([]float64{0, 1, 0}, []float64{1, 0, 0})},
				},
			},
		},
	)

	testUtils.ExecuteTestCase(t, test)
}

// A condition on the similarity itself is applied by the similarity node, above the scan, so the
// scan cannot count the documents that passed and the query keeps the full-scan path. Widening the
// search would stop after the two nearest group "a" documents had passed the scan filter, and the
// nearest of those is then dropped for being too similar, leaving one result instead of two.
func TestQuerySimple_WithSimilarityOnVectorIndexAndSimilarityFilter_DoesNotRoute(t *testing.T) {
	test := similarityFilterTestCase(
		&action.Request{
			Request: `query {
				User(
					filter: {group: {_eq: "a"}, _alias: {sim: {_lt: 0.99}}},
					order: {_alias: {sim: DESC}},
					limit: 2
				){
					name
					sim: SIMILARITY(vector: {vector: [1, 0, 0]})
				}
			}`,
			Results: map[string]any{
				"User": []map[string]any{
					{"name": "n2", "sim": testUtils.CosineSimilarity([]float64{0.95, 0.31, 0}, []float64{1, 0, 0})},
					{"name": "n3", "sim": testUtils.CosineSimilarity([]float64{0.9, 0.44, 0}, []float64{1, 0, 0})},
				},
			},
		},
		// No index fetch: the query full-scanned rather than routing to the vector index.
		&action.Request{
			Request: `query @explain(type: execute) {
				User(
					filter: {group: {_eq: "a"}, _alias: {sim: {_lt: 0.99}}},
					order: {_alias: {sim: DESC}},
					limit: 2
				){
					name
					sim: SIMILARITY(vector: {vector: [1, 0, 0]})
				}
			}`,
			Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0).WithDocFetches(6),
		},
	)

	testUtils.ExecuteTestCase(t, test)
}

// The filtered field has its own index, so the query resolves the filter through that index and
// scores the documents it yields, instead of widening a graph search. Explain shows the difference:
// only the three matching documents are read, where a widened search would have read all six.
func TestQuerySimple_WithSimilarityOnVectorIndexAndFilter_IndexedFieldDoesNotRoute(t *testing.T) {
	test := testUtils.TestCase{
		Actions: append(append([]any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					group: String @index
					vector: [Float32!] @vectorIndex(dimensions: 3, HNSW: {metric: COSINE})
				}`,
			},
		}, similarityFilterDocs()...),
			&action.Request{
				Request: `query {
					User(filter: {group: {_eq: "b"}}, order: {_alias: {sim: DESC}}, limit: 2){
						name
						sim: SIMILARITY(vector: {vector: [1, 0, 0]})
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "n4", "sim": testUtils.CosineSimilarity([]float64{0.8, 0.6, 0}, []float64{1, 0, 0})},
						{"name": "n5", "sim": testUtils.CosineSimilarity([]float64{0.6, 0.8, 0}, []float64{1, 0, 0})},
					},
				},
			},
			&action.Request{
				Request: `query @explain(type: execute) {
					User(filter: {group: {_eq: "b"}}, order: {_alias: {sim: DESC}}, limit: 2){
						name
						sim: SIMILARITY(vector: {vector: [1, 0, 0]})
					}
				}`,
				Asserter: testUtils.NewExplainAsserter().WithDocFetches(3),
			},
		),
	}

	testUtils.ExecuteTestCase(t, test)
}

// The same six documents, split across two teams: group "a" is in team "red" and group "b" in team
// "blue". Any query that joins to the team runs the scan under the join, which moves filter
// conditions above it and re-drives it once per parent document.
func similarityRelationTestCase(actions ...any) testUtils.TestCase {
	return testUtils.TestCase{
		Actions: append([]any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					group: String
					vector: [Float32!] @vectorIndex(dimensions: 3, HNSW: {metric: COSINE})
					team: Team
				}

				type Team {
					name: String
					members: [User]
				}`,
			},
			&action.AddDoc{CollectionID: 1, DocMap: map[string]any{"name": "red"}},
			&action.AddDoc{CollectionID: 1, DocMap: map[string]any{"name": "blue"}},
			&action.AddDoc{DocMap: map[string]any{"name": "n1", "group": "a",
				"vector": []float32{1, 0, 0}, "team": testUtils.NewDocIndex(1, 0)}},
			&action.AddDoc{DocMap: map[string]any{"name": "n2", "group": "a",
				"vector": []float32{0.95, 0.31, 0}, "team": testUtils.NewDocIndex(1, 0)}},
			&action.AddDoc{DocMap: map[string]any{"name": "n3", "group": "a",
				"vector": []float32{0.9, 0.44, 0}, "team": testUtils.NewDocIndex(1, 0)}},
			&action.AddDoc{DocMap: map[string]any{"name": "n4", "group": "b",
				"vector": []float32{0.8, 0.6, 0}, "team": testUtils.NewDocIndex(1, 1)}},
			&action.AddDoc{DocMap: map[string]any{"name": "n5", "group": "b",
				"vector": []float32{0.6, 0.8, 0}, "team": testUtils.NewDocIndex(1, 1)}},
			&action.AddDoc{DocMap: map[string]any{"name": "n6", "group": "b",
				"vector": []float32{0, 1, 0}, "team": testUtils.NewDocIndex(1, 1)}},
		}, actions...),
	}
}

// A condition on a related object is split out to a join, which drops documents above the scan, so
// the scan cannot tell how many passed and the query keeps the full-scan path. The result is still
// the two nearest members of the team.
func TestQuerySimple_WithSimilarityOnVectorIndexAndRelationFilter_DoesNotRoute(t *testing.T) {
	test := similarityRelationTestCase(
		&action.Request{
			Request: `query {
				User(filter: {team: {name: {_eq: "blue"}}}, order: {_alias: {sim: DESC}}, limit: 2){
					name
					sim: SIMILARITY(vector: {vector: [1, 0, 0]})
				}
			}`,
			Results: map[string]any{
				"User": []map[string]any{
					{"name": "n4", "sim": testUtils.CosineSimilarity([]float64{0.8, 0.6, 0}, []float64{1, 0, 0})},
					{"name": "n5", "sim": testUtils.CosineSimilarity([]float64{0.6, 0.8, 0}, []float64{1, 0, 0})},
				},
			},
		},
	)

	testUtils.ExecuteTestCase(t, test)
}

// Selecting a related object builds a join over the scan even though the filter is a plain scalar
// one. The join moves conditions above the scan and re-drives it per parent document, so the query
// keeps the full-scan path.
func TestQuerySimple_WithSimilarityOnVectorIndexAndRelationSelection_DoesNotRoute(t *testing.T) {
	test := similarityRelationTestCase(
		&action.Request{
			Request: `query {
				User(filter: {group: {_eq: "b"}}, order: {_alias: {sim: DESC}}, limit: 2){
					name
					sim: SIMILARITY(vector: {vector: [1, 0, 0]})
					team {
						name
					}
				}
			}`,
			Results: map[string]any{
				"User": []map[string]any{
					{
						"name": "n4",
						"sim":  testUtils.CosineSimilarity([]float64{0.8, 0.6, 0}, []float64{1, 0, 0}),
						"team": map[string]any{"name": "blue"},
					},
					{
						"name": "n5",
						"sim":  testUtils.CosineSimilarity([]float64{0.6, 0.8, 0}, []float64{1, 0, 0}),
						"team": map[string]any{"name": "blue"},
					},
				},
			},
		},
		&action.Request{
			Request: `query @explain(type: execute) {
				User(filter: {group: {_eq: "b"}}, order: {_alias: {sim: DESC}}, limit: 2){
					name
					sim: SIMILARITY(vector: {vector: [1, 0, 0]})
					team {
						name
					}
				}
			}`,
			Asserter: testUtils.NewExplainAsserter("root").WithIndexFetches(0),
		},
	)

	testUtils.ExecuteTestCase(t, test)
}

// A child selection can carry its own nearest-neighbour query. It is scanned under the join too, so
// it keeps the full-scan path and still returns the nearest matching member.
func TestQuerySimple_WithSimilarityOnVectorIndexInChildSelection_DoesNotRoute(t *testing.T) {
	test := similarityRelationTestCase(
		&action.Request{
			Request: `query {
				Team(filter: {name: {_eq: "blue"}}){
					name
					members(filter: {group: {_eq: "b"}}, order: {_alias: {sim: DESC}}, limit: 1){
						name
						sim: SIMILARITY(vector: {vector: [1, 0, 0]})
					}
				}
			}`,
			Results: map[string]any{
				"Team": []map[string]any{
					{
						"name": "blue",
						"members": []map[string]any{
							{"name": "n4", "sim": testUtils.CosineSimilarity([]float64{0.8, 0.6, 0}, []float64{1, 0, 0})},
						},
					},
				},
			},
		},
	)

	testUtils.ExecuteTestCase(t, test)
}

// A grouped query limits groups, not documents, so asking the graph for limit documents is wrong.
// The query keeps the full-scan path and returns the nearest group.
func TestQuerySimple_WithSimilarityOnVectorIndexAndGroupBy_DoesNotRoute(t *testing.T) {
	test := similarityFilterTestCase(
		&action.Request{
			Request: `query {
				User(groupBy: [group], filter: {group: {_in: ["a", "b"]}}, order: {_alias: {sim: DESC}}, limit: 1){
					group
					sim: SIMILARITY(vector: {vector: [1, 0, 0]})
				}
			}`,
			Results: map[string]any{
				"User": []map[string]any{
					// A group renders the similarity of the last document read into it, n3 for group "a".
					// Routing would have read only the one nearest document, rendering n1 instead.
					{"group": "a", "sim": testUtils.CosineSimilarity([]float64{0.9, 0.44, 0}, []float64{1, 0, 0})},
				},
			},
		},
	)

	testUtils.ExecuteTestCase(t, test)
}

// Deleting a document tombstones it out of the graph, so a routed query could never return it. A
// showDeleted query keeps the full-scan path and returns the deleted document with the rest.
func TestQuerySimple_WithSimilarityOnVectorIndexAndShowDeleted_DoesNotRoute(t *testing.T) {
	test := similarityFilterTestCase(
		// n4 is the nearest of group "b" (added fourth, so DocID 3).
		testUtils.DeleteDoc{CollectionID: 0, DocID: 3},
		&action.Request{
			Request: `query {
				User(showDeleted: true, filter: {group: {_eq: "b"}}, order: {_alias: {sim: DESC}}, limit: 2){
					name
					_deleted
					sim: SIMILARITY(vector: {vector: [1, 0, 0]})
				}
			}`,
			Results: map[string]any{
				"User": []map[string]any{
					{"name": "n4", "_deleted": true,
						"sim": testUtils.CosineSimilarity([]float64{0.8, 0.6, 0}, []float64{1, 0, 0})},
					{"name": "n5", "_deleted": false,
						"sim": testUtils.CosineSimilarity([]float64{0.6, 0.8, 0}, []float64{1, 0, 0})},
				},
			},
		},
	)

	testUtils.ExecuteTestCase(t, test)
}
