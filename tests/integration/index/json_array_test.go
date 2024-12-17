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

func TestJSONArrayIndex_WithDifferentElementValuesAndTypes_ShouldFetchCorrectlyUsingIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {numbers: {_any: {_eq: 4}}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"custom": map[string]any{
						"numbers": []int{3, 5, 7},
					},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Islam",
					"custom": map[string]any{
						"numbers": []int{3},
					},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Shahzad",
					"custom": map[string]any{
						"numbers": []int{4, 8, 4, 4, 5, 4},
					},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Keenan",
					"custom": map[string]any{
						"numbers": []any{8, nil},
					},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Bruno",
					"custom": map[string]any{
						"numbers": []any{10, "str", true},
					},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"custom": map[string]any{
						"numbers": 4,
					},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Chris",
					"custom": map[string]any{
						"numbers": nil,
					},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Fred",
					"custom": map[string]any{
						"height": 198,
					},
				},
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONArrayIndex_WithNestedArrays_ShouldNotConsiderThem(t *testing.T) {
	req := `query {
		User(filter: {custom: {numbers: {_any: {_eq: 4}}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"custom": map[string]any{
						"numbers": []any{3, 5, []int{9, 4}, 7},
					},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Islam",
					"custom": map[string]any{
						"numbers": []any{0, []int{2, 6}, 9},
					},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Fred",
					"custom": map[string]any{
						"numbers": []any{3, 5, []any{1, 0, []int{9, 4, 6}}, 7},
					},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"custom": map[string]any{
						"numbers": []any{1, 2, []int{8, 6}, 10},
					},
				},
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONArrayIndex_WithNoneFilterOnDifferentElementValues_ShouldFetchCorrectlyUsingIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {numbers: {_none: {_eq: 4}}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"custom": map[string]any{
						"numbers": []int{3, 5, 7},
					},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Shahzad",
					"custom": map[string]any{
						"numbers": []int{4, 8},
					},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Islam",
					"custom": map[string]any{
						"numbers": []any{8, nil},
					},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Fred",
					"custom": map[string]any{
						"numbers": []any{1, []int{4}},
					},
				},
			},
			// TODO: This document should be part of the query result, but it needs additional work
			// with json encoding https://github.com/sourcenetwork/defradb/issues/3329
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"custom": map[string]any{
						"numbers": 4,
					},
				},
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Islam"},
						{"name": "Fred"},
						{"name": "John"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(9),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONArrayIndex_WithAllFilterOnDifferentElementValues_ShouldFetchCorrectlyUsingIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {numbers: {_all: {_eq: 4}}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"custom": map[string]any{
						"numbers": []int{3, 4},
					},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Shahzad",
					"custom": map[string]any{
						"numbers": []any{4, []int{4, 8}},
					},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Islam",
					"custom": map[string]any{
						"numbers": 4,
					},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Fred",
					"custom": map[string]any{
						"numbers": []any{4, []any{4, []int{4}}},
					},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Bruno",
					"custom": map[string]any{
						"numbers": []any{4, 4, 4},
					},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"custom": map[string]any{
						"numbers": 3,
					},
				},
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Bruno"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(5),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
