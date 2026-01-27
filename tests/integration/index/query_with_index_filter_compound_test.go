// Copyright 2026 Democratized Data Foundation
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

func TestQueryWithIndex_WithOrFilter_ShouldFetchCorrectDocs(t *testing.T) {
	req := `query {
		User(filter: {_or: [{age: {_eq: 55}}, {age: {_eq: 19}}]}) {
			name
			age
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int @index
					}`,
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Bob",
					"age":  32,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Carlo",
					"age":  55,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Alice",
					"age":  19,
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alice",
							"age":  int64(19),
						},
						{
							"name": "Carlo",
							"age":  int64(55),
						},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithOrFilterWithThreeBranches_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {_or: [{age: {_eq: 55}}, {age: {_eq: 19}}, {age: {_eq: 21}}]}) {
			name
			age
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int @index
					}`,
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Bob",
					"age":  32,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Carlo",
					"age":  55,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Alice",
					"age":  19,
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
							"age":  int64(21),
						},
						{
							"name": "Alice",
							"age":  int64(19),
						},
						{
							"name": "Carlo",
							"age":  int64(55),
						},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithOrFilterWithRangeConditions_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {_or: [{age: {_gt: 50}}, {age: {_lt: 20}}]}) {
			name
			age
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int @index
					}`,
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Bob",
					"age":  32,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Carlo",
					"age":  55,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Alice",
					"age":  19,
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alice",
							"age":  int64(19),
						},
						{
							"name": "Carlo",
							"age":  int64(55),
						},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithOrFilterWithOverlappingConditions_ShouldDeduplicateResults(t *testing.T) {
	// Test that documents matching multiple OR branches are not returned multiple times
	req := `query {
		User(filter: {_or: [{age: {_geq: 30}}, {age: {_eq: 32}}]}) {
			name
			age
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int @index
					}`,
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Bob",
					"age":  32,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Carlo",
					"age":  55,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Alice",
					"age":  19,
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Bob",
							"age":  int64(32),
						},
						{
							"name": "Carlo",
							"age":  int64(55),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithOrFilterOnTwoDifferentIndexedFields_ShouldNotUseIndex(t *testing.T) {
	// Test OR filter where each branch filters on a different indexed field.
	// Currently, the system can only use one index per query, so OR filters
	// on different indexed fields fall back to a full scan.
	req := `query {
		User(filter: {_or: [{age: {_eq: 55}}, {score: {_eq: 100}}]}) {
			name
			age
			score
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int @index
						score: Int @index
					}`,
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name":  "John",
					"age":   21,
					"score": 100,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name":  "Bob",
					"age":   32,
					"score": 80,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name":  "Carlo",
					"age":   55,
					"score": 90,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name":  "Alice",
					"age":   19,
					"score": 75,
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name":  "John",
							"age":   int64(21),
							"score": int64(100),
						},
						{
							"name":  "Carlo",
							"age":   int64(55),
							"score": int64(90),
						},
					},
				},
				NonOrderedResults: true,
			},
			// Falls back to full scan since OR branches reference different indexed fields
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithOrFilterOnTwoDifferentIndexedFields_WithOverlap_ShouldReturnCorrectResults(t *testing.T) {
	// Test OR filter on different indexed fields where a document matches both branches.
	// Falls back to full scan since different indexes are involved.
	req := `query {
		User(filter: {_or: [{age: {_eq: 21}}, {score: {_eq: 100}}]}) {
			name
			age
			score
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int @index
						score: Int @index
					}`,
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name":  "John",
					"age":   21,
					"score": 100, // Matches both age=21 AND score=100
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name":  "Bob",
					"age":   32,
					"score": 80,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name":  "Carlo",
					"age":   55,
					"score": 90,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name":  "Alice",
					"age":   19,
					"score": 75,
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name":  "John",
							"age":   int64(21),
							"score": int64(100),
						},
					},
				},
			},
			// Falls back to full scan since OR branches reference different indexed fields
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithOrFilterOnIndexedAndNonIndexedField_ShouldFallbackToFullScan(t *testing.T) {
	// When one OR branch uses an indexed field and another uses a non-indexed field,
	// the system falls back to a full scan since not all branches can use the index.
	req := `query {
		User(filter: {_or: [{age: {_eq: 55}}, {name: {_eq: "John"}}]}) {
			name
			age
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int @index
					}`,
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Bob",
					"age":  32,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Carlo",
					"age":  55,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Alice",
					"age":  19,
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
							"age":  int64(21),
						},
						{
							"name": "Carlo",
							"age":  int64(55),
						},
					},
				},
				NonOrderedResults: true,
			},
			// Falls back to full scan since one branch references a non-indexed field
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithOrFilterOnOnlyNonIndexedFields_ShouldNotUseIndex(t *testing.T) {
	// When all OR branches use non-indexed fields, no index should be used
	req := `query {
		User(filter: {_or: [{name: {_eq: "Carlo"}}, {name: {_eq: "John"}}]}) {
			name
			age
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int @index
					}`,
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Bob",
					"age":  32,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Carlo",
					"age":  55,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Alice",
					"age":  19,
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
							"age":  int64(21),
						},
						{
							"name": "Carlo",
							"age":  int64(55),
						},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithOrFilterWithAnyAndNoneOnSameArrayField_ShouldFallbackToFullScan(t *testing.T) {
	// Test OR filter combining _any and _none on the same indexed array field.
	// Falls back to full scan since _none cannot use the index.
	req := `query {
		User(filter: {_or: [{numbers: {_any: {_gt: 20}}}, {numbers: {_none: {_eq: 10}}}]}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						numbers: [Int!] @index
					}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 10, 20]
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Bob",
					"numbers": [30, 40, 50]
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Carlo",
					"numbers": [20, 20, 20]
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Alice",
					"numbers": [5, 15, 25]
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Bob"},
						{"name": "Alice"},
						{"name": "Carlo"},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithOrFilterOnTwoDifferentIndexedFields_WithRangeConditions_ShouldNotUseIndex(t *testing.T) {
	// Test OR filter with range conditions on different indexed fields.
	// Falls back to full scan since different indexes are involved.
	req := `query {
		User(filter: {_or: [{age: {_gt: 50}}, {score: {_lt: 80}}]}) {
			name
			age
			score
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int @index
						score: Int @index
					}`,
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name":  "John",
					"age":   21,
					"score": 100,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name":  "Bob",
					"age":   32,
					"score": 80,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name":  "Carlo",
					"age":   55,
					"score": 90,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name":  "Alice",
					"age":   19,
					"score": 75,
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name":  "Carlo",
							"age":   int64(55),
							"score": int64(90),
						},
						{
							"name":  "Alice",
							"age":   int64(19),
							"score": int64(75),
						},
					},
				},
				NonOrderedResults: true,
			},
			// Falls back to full scan since OR branches reference different indexed fields
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
