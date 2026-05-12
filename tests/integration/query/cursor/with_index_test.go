// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cursor

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// TestCursorWithIndex_UsesIndexWhenOrderOnIndexedField verifies index usage
// when ordering on an indexed field (indexFetches > 0 in explain).
func TestCursorWithIndex_UsesIndexWhenOrderOnIndexedField(t *testing.T) {
	req := `query {
		_cursor {
			User(first: 3, order: {age: ASC}) {
				name
				age
			}
			_pageInfo {
				hasNext
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: userCollectionGQLSchema,
			},
			testUtils.CreateDoc{Doc: `{"name": "Alice", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Bob", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Carol", "age": 35}`},
			testUtils.CreateDoc{Doc: `{"name": "Dave", "age": 40}`},
			testUtils.CreateDoc{Doc: `{"name": "Eve", "age": 45}`},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Alice", "age": int64(25)},
							{"name": "Bob", "age": int64(30)},
							{"name": "Carol", "age": int64(35)},
						},
						"_pageInfo": map[string]any{
							"hasNext": true,
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

// TestCursorWithIndex_FallsBackToNaiveWithoutOrder verifies naive iteration
// (by docID) when no order clause is specified (indexFetches = 0).
func TestCursorWithIndex_FallsBackToNaiveWithoutOrder(t *testing.T) {
	explainReq := `query @explain(type: execute) {
		_cursor {
			User(first: 3) {
				name
				age
			}
			_pageInfo {
				hasNext
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: userCollectionGQLSchema,
			},
			testUtils.CreateDoc{Doc: `{"name": "Alice", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Bob", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Carol", "age": 35}`},
			testUtils.CreateDoc{Doc: `{"name": "Dave", "age": 40}`},
			testUtils.CreateDoc{Doc: `{"name": "Eve", "age": 45}`},
			&action.Request{
				Request:  explainReq,
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// TestCursorWithIndex_RejectsOrderWithoutIndex verifies that ordering on
// a non-indexed field is rejected at query planning time.
func TestCursorWithIndex_RejectsOrderWithoutIndex(t *testing.T) {
	req := `query {
		_cursor {
			User(first: 3, order: {age: ASC}) {
				name
				age
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int
					}`,
			},
			testUtils.CreateDoc{Doc: `{"name": "Alice", "age": 25}`},
			&action.Request{
				Request:       req,
				ExpectedError: "no supporting index for cursor order field",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// TestCursorWithIndex_IndexPathReturnsSortedResults verifies that results
// are sorted by index key values when using the index-aware path.
func TestCursorWithIndex_IndexPathReturnsSortedResults(t *testing.T) {
	req := `query {
		_cursor {
			User(first: 5, order: {age: ASC}) {
				name
				age
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: userCollectionGQLSchema,
			},
			testUtils.CreateDoc{Doc: `{"name": "Eve", "age": 45}`},
			testUtils.CreateDoc{Doc: `{"name": "Dave", "age": 40}`},
			testUtils.CreateDoc{Doc: `{"name": "Carol", "age": 35}`},
			testUtils.CreateDoc{Doc: `{"name": "Bob", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Alice", "age": 25}`},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Alice", "age": int64(25)},
							{"name": "Bob", "age": int64(30)},
							{"name": "Carol", "age": int64(35)},
							{"name": "Dave", "age": int64(40)},
							{"name": "Eve", "age": int64(45)},
						},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(5),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// TestCursorWithIndex_DESCOrderSucceeds verifies that DESC order
// on an indexed ASC field succeeds (reverse iteration supported in Phase 10).
func TestCursorWithIndex_DESCOrderSucceeds(t *testing.T) {
	req := `query {
		_cursor {
			User(first: 3, order: {age: DESC}) {
				name
				age
			}
			_pageInfo {
				hasNext
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: userCollectionGQLSchema,
			},
			testUtils.CreateDoc{Doc: `{"name": "Alice", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Bob", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Carol", "age": 35}`},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Carol", "age": int64(35)},
							{"name": "Bob", "age": int64(30)},
							{"name": "Alice", "age": int64(25)},
						},
						"_pageInfo": map[string]any{
							"hasNext": false,
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
