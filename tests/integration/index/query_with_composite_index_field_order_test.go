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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryWithCompositeIndex_WithDefaultOrder_ShouldFetchInDefaultOrder(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name"},  {field: "age"}]) {
						name: String
						age: Int
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alan",
						"age":	29
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24
					}`,
			},
			&action.Request{
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
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name"},  {field: "age"}]) {
						name: String
						age: Int
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alan",
						"age":	29
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24
					}`,
			},
			&action.Request{
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
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name", direction: DESC}, {field: "age", direction: ASC}]) {
						name: String
						age: Int
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alan",
						"age":	29
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Andy",
						"age":	24
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24
					}`,
			},
			&action.Request{
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
			&action.Request{
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
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name", direction: DESC}, {field: "age", direction: ASC}]) {
						name: String
						age: Int
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alan",
						"age":	29
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Andy",
						"age":	24
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24
					}`,
			},
			&action.Request{
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
			&action.Request{
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
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name", direction: DESC}, {field: "age", direction: ASC}]) {
						name: String
						age: Int
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alan",
						"age":	29
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Andy",
						"age":	24
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24
					}`,
			},
			&action.Request{
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
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name", direction: ASC}, {field: "age", direction: DESC}]) {
						name: String
						age: Int
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alan",
						"age":	29
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24
					}`,
			},
			&action.Request{
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
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name", direction: ASC}, {field: "age", direction: DESC}]) {
						name: String
						age: Int
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alan",
						"age":	29
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24
					}`,
			},
			&action.Request{
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
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name", direction: DESC}, {field: "age", direction: ASC}]) {
						name: String
						age: Int
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alan",
						"age":	29
					}`,
			},
			&action.Request{
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
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name", direction: ASC}, {field: "age", direction: DESC}]) {
						name: String
						age: Int
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alan",
						"age":	29
					}`,
			},
			&action.Request{
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
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name", direction: DESC}, {field: "age", direction: ASC}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.AddPredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
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

// TODO: This test documents incorrect behaviour. https://github.com/sourcenetwork/defradb/issues/3780
func TestQueryWithCompositeIndex_WithInFilterOnSecondFieldWithRevertedOrder_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name", direction: ASC}, {field: "age", direction: DESC}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.AddPredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
				Request: `query {
						User(filter: {age: {_in: [20, 28, 33]}}) {
							name
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						/* Expected results
						{"name": "Andy"},
						{"name": "Fred"},
						{"name": "Shahzad"},
						*/
						// Actual results
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
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "age"}, {field: "name"}]) {
						name: String
						age: Int
					}`,
			},
			&action.AddDoc{
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			&action.AddDoc{
				Doc: `
					{
						"name":	"Bob",
						"age":	30
					}`,
			},
			&action.AddDoc{
				Doc: `
					{
						"name":	"Charlie",
						"age":	25
					}`,
			},
			&action.AddDoc{
				Doc: `
					{
						"name":	"David",
						"age":	35
					}`,
			},
			&action.AddDoc{
				Doc: `
					{
						"name":	"Eve",
						"age":	28
					}`,
			},
			&action.Request{
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
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithRangeQueryOnFirstFieldWithMultipleFilters_ShouldUseRangeOptimization(t *testing.T) {
	req := `
		query {
			User(filter: {age: {_gt: 25}, name: {_eq: "Bob"}}) {
				name
				age
			}
		}`

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "age"}, {field: "name"}]) {
						name: String
						age: Int
					}`,
			},
			&action.AddDoc{
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			&action.AddDoc{
				Doc: `
					{
						"name":	"Bob",
						"age":	30
					}`,
			},
			&action.AddDoc{
				Doc: `
					{
						"name":	"Charlie",
						"age":	25
					}`,
			},
			&action.AddDoc{
				Doc: `
					{
						"name":	"David",
						"age":	35
					}`,
			},
			&action.AddDoc{
				Doc: `
					{
						"name":	"Eve",
						"age":	28
					}`,
			},
			&action.AddDoc{
				Doc: `
					{
						"name":	"Bob",
						"age":	32
					}`,
			},
			&action.Request{
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
			&action.Request{
				Request: makeExplainQuery(req),
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
			User(filter: {age: {_leq: 30}}) {
				name
				age
			}
		}`

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "age", direction: DESC}, {field: "name"}]) {
						name: String
						age: Int
					}`,
			},
			&action.AddDoc{
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			&action.AddDoc{
				Doc: `
					{
						"name":	"Bob",
						"age":	30
					}`,
			},
			&action.AddDoc{
				Doc: `
					{
						"name":	"Charlie",
						"age":	25
					}`,
			},
			&action.AddDoc{
				Doc: `
					{
						"name":	"David",
						"age":	35
					}`,
			},
			&action.AddDoc{
				Doc: `
					{
						"name":	"Eve",
						"age":	28
					}`,
			},
			&action.Request{
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
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(4),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
