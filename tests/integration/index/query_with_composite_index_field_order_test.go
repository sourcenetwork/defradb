// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package index

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryWithCompositeIndex_WithDefaultOrder_ShouldFetchInDefaultOrder(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test composite index in default order",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(includes: [{field: "name"},  {field: "age"}]) {
						name: String
						age: Int
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alan",
						"age":	29
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						User(filter: {name: {_like: "Al%"}}) {
							name
							age
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alan",
							"age":  29,
						},
						{
							"name": "Alice",
							"age":  22,
						},
						{
							"name": "Alice",
							"age":  24,
						},
						{
							"name": "Alice",
							"age":  38,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithDefaultOrderCaseInsensitive_ShouldFetchInDefaultOrder(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test composite index in default order and case insensitive operator",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(includes: [{field: "name"},  {field: "age"}]) {
						name: String
						age: Int
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alan",
						"age":	29
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						User(filter: {name: {_ilike: "al%"}}) {
							name
							age
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alan",
							"age":  29,
						},
						{
							"name": "Alice",
							"age":  22,
						},
						{
							"name": "Alice",
							"age":  24,
						},
						{
							"name": "Alice",
							"age":  38,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithRevertedOrderOnFirstField_ShouldFetchInRevertedOrder(t *testing.T) {
	req := `query {
		User(filter: {name: {_like: "A%"}}) {
			name
			age
		}
	}`
	test := testUtils.TestCase{
		Description: "Test composite index with reverted order on first field",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(includes: [{field: "name", direction: DESC}, {field: "age", direction: ASC}]) {
						name: String
						age: Int
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alan",
						"age":	29
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Andy",
						"age":	24
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24
					}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Andy",
							"age":  24,
						},
						{
							"name": "Alice",
							"age":  22,
						},
						{
							"name": "Alice",
							"age":  24,
						},
						{
							"name": "Alice",
							"age":  38,
						},
						{
							"name": "Alan",
							"age":  29,
						},
					},
				},
			},
			testUtils.Request{
				Request: makeExplainQuery(req),
				// we fetch all available docs with index
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(5),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithRevertedOrderOnFirstFieldAndNoFilter_ShouldFetchInRevertedOrder(t *testing.T) {
	req := `query {
		User(order: {name: DESC}) {
			name
			age
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(includes: [{field: "name", direction: DESC}, {field: "age", direction: ASC}]) {
						name: String
						age: Int
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alan",
						"age":	29
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Andy",
						"age":	24
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24
					}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Andy",
							"age":  24,
						},
						{
							"name": "Alice",
							"age":  22,
						},
						{
							"name": "Alice",
							"age":  24,
						},
						{
							"name": "Alice",
							"age":  38,
						},
						{
							"name": "Alan",
							"age":  29,
						},
					},
				},
			},
			testUtils.Request{
				Request: makeExplainQuery(req),
				// we fetch all available docs with index
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(5),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithRevertedOrderOnFirstFieldCaseInsensitive_ShouldFetchInRevertedOrder(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test composite index with reverted order on first field and case insensitive operator",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(includes: [{field: "name", direction: DESC}, {field: "age", direction: ASC}]) {
						name: String
						age: Int
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alan",
						"age":	29
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Andy",
						"age":	24
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						User(filter: {name: {_ilike: "a%"}}) {
							name
							age
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Andy",
							"age":  24,
						},
						{
							"name": "Alice",
							"age":  22,
						},
						{
							"name": "Alice",
							"age":  24,
						},
						{
							"name": "Alice",
							"age":  38,
						},
						{
							"name": "Alan",
							"age":  29,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithRevertedOrderOnSecondField_ShouldFetchInRevertedOrder(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test composite index with reverted order on second field",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(includes: [{field: "name", direction: ASC}, {field: "age", direction: DESC}]) {
						name: String
						age: Int
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alan",
						"age":	29
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						User(filter: {name: {_like: "Al%"}}) {
							name
							age
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alan",
							"age":  29,
						},
						{
							"name": "Alice",
							"age":  38,
						},
						{
							"name": "Alice",
							"age":  24,
						},
						{
							"name": "Alice",
							"age":  22,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithRevertedOrderOnSecondFieldCaseInsensitive_ShouldFetchInRevertedOrder(
	t *testing.T,
) {
	test := testUtils.TestCase{
		Description: "Test composite index with reverted order on second field and case insensitive operator",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(includes: [{field: "name", direction: ASC}, {field: "age", direction: DESC}]) {
						name: String
						age: Int
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alan",
						"age":	29
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						User(filter: {name: {_ilike: "al%"}}) {
							name
							age
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alan",
							"age":  29,
						},
						{
							"name": "Alice",
							"age":  38,
						},
						{
							"name": "Alice",
							"age":  24,
						},
						{
							"name": "Alice",
							"age":  22,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_IfExactMatchWithRevertedOrderOnFirstField_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test composite index with reverted order on first field and filter with exact match",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(includes: [{field: "name", direction: DESC}, {field: "age", direction: ASC}]) {
						name: String
						age: Int
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alan",
						"age":	29
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						User(filter: {name: {_eq: "Alice"}, age: {_eq: 22}}) {
							name
							age
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alice",
							"age":  22,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_IfExactMatchWithRevertedOrderOnSecondField_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test composite index with reverted order on second field and filter with exact match",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(includes: [{field: "name", direction: ASC}, {field: "age", direction: DESC}]) {
						name: String
						age: Int
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alan",
						"age":	29
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						User(filter: {name: {_eq: "Alice"}, age: {_eq: 22}}) {
							name
							age
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alice",
							"age":  22,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithInFilterOnFirstFieldWithRevertedOrder_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test composite index with reverted order on first field and filtering with _in filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(includes: [{field: "name", direction: DESC}, {field: "age", direction: ASC}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: `query {
						User(filter: {name: {_in: ["Addo", "Andy", "Fred"]}}) {
							name
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Addo"},
						{"name": "Andy"},
						{"name": "Fred"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithInFilterOnSecondFieldWithRevertedOrder_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test composite index with reverted order on second field and filtering with _in filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(includes: [{field: "name", direction: ASC}, {field: "age", direction: DESC}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: `query {
						User(filter: {age: {_in: [20, 28, 33]}}) {
							name
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
						{"name": "Andy"},
						{"name": "Fred"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithRangeQueryOnFirstField_ShouldUseRangeOptimization(t *testing.T) {
	req := `
		query {
			User(filter: {age: {_gt: 25}}) {
				name
				age
			}
		}`

	test := testUtils.TestCase{
		Description: "Test composite index with range query on first field",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(includes: [{field: "age"}, {field: "name"}]) {
						name: String
						age: Int
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Bob",
						"age":	30
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Charlie",
						"age":	25
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"David",
						"age":	35
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Eve",
						"age":	28
					}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Eve",
							"age":  28,
						},
						{
							"name": "Bob",
							"age":  30,
						},
						{
							"name": "David",
							"age":  35,
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithRangeQueryOnFirstFieldWithMultipleFilters_ShouldUseMatchers(t *testing.T) {
	req := `
		query {
			User(filter: {age: {_gt: 25}, name: {_eq: "Bob"}}) {
				name
				age
			}
		}`

	test := testUtils.TestCase{
		Description: "Test composite index with range query and additional filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(includes: [{field: "age"}, {field: "name"}]) {
						name: String
						age: Int
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Bob",
						"age":	30
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Charlie",
						"age":	25
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"David",
						"age":	35
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Eve",
						"age":	28
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Bob",
						"age":	32
					}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Bob",
							"age":  30,
						},
						{
							"name": "Bob",
							"age":  32,
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				// Should fetch all entries with age > 25, then filter by name
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(4),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithDescendingFirstFieldAndRangeQuery_ShouldUseRangeOptimization(t *testing.T) {
	req := `
		query {
			User(filter: {age: {_le: 30}}) {
				name
				age
			}
		}`

	test := testUtils.TestCase{
		Description: "Test composite index with descending first field and range query",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(includes: [{field: "age", direction: DESC}, {field: "name"}]) {
						name: String
						age: Int
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Bob",
						"age":	30
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Charlie",
						"age":	25
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"David",
						"age":	35
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Eve",
						"age":	28
					}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Bob",
							"age":  30,
						},
						{
							"name": "Eve",
							"age":  28,
						},
						{
							"name": "Charlie",
							"age":  25,
						},
						{
							"name": "Alice",
							"age":  22,
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(4),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}