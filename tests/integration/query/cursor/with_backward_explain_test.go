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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCursorBackwardExplain_SimpleShowsLastBeforeFields(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: userCollectionGQLSchema,
			},
			&action.AddDoc{Doc: `{"name": "Alice", "age": 25}`},
			&action.AddDoc{Doc: `{"name": "Bob", "age": 30}`},
			&action.AddDoc{Doc: `{"name": "Carol", "age": 35}`},

			&action.Request{
				Request: `query {
					_cursor {
						User(last: 2, order: {age: ASC}) {
							name
							age
						}
						_pageInfo {
							hasNext
							hasPrev
						}
					}
				}`,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Bob", "age": int64(30)},
							{"name": "Carol", "age": int64(35)},
						},
						"_pageInfo": map[string]any{
							"hasNext": false,
							"hasPrev": true,
						},
					},
				},
			},

			// Simple explain verifies cursorNode shows last/before fields.
			&action.Request{
				Request: `query @explain {
					_cursor {
						User(last: 2, order: {age: ASC}) {
							name
							age
						}
						_pageInfo {
							hasNext
							hasPrev
						}
					}
				}`,
				Asserter: testUtils.ResultAsserterFunc(func(t testing.TB, result map[string]any) (bool, string) {
					explain, ok := result["explain"].(map[string]any)
					if !ok {
						return false, "missing explain key"
					}
					ops, ok := explain["operationNode"].([]any)
					if !ok || len(ops) == 0 {
						return false, "missing operationNode"
					}
					op := ops[0].(map[string]any)
					selectTop := op["selectTopNode"].(map[string]any)
					cursorN := selectTop["cursorNode"].(map[string]any)

					if cursorN["last"] != uint64(2) {
						return false, "cursorNode.last should be 2"
					}
					if cursorN["before"] != nil {
						return false, "cursorNode.before should be nil for last-only query"
					}
					if cursorN["first"] != nil {
						return false, "cursorNode.first should be nil for backward query"
					}
					if cursorN["after"] != nil {
						return false, "cursorNode.after should be nil for backward query"
					}
					return true, ""
				}),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCursorBackwardExplain_ExecuteShowsIndexFetches(t *testing.T) {
	req := `query {
		_cursor {
			User(last: 2, order: {age: ASC}) {
				name
				age
			}
			_pageInfo {
				hasNext
				hasPrev
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: userCollectionGQLSchema,
			},
			&action.AddDoc{Doc: `{"name": "Alice", "age": 25}`},
			&action.AddDoc{Doc: `{"name": "Bob", "age": 30}`},
			&action.AddDoc{Doc: `{"name": "Carol", "age": 35}`},

			&action.Request{
				Request: req,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Bob", "age": int64(30)},
							{"name": "Carol", "age": int64(35)},
						},
						"_pageInfo": map[string]any{
							"hasNext": false,
							"hasPrev": true,
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

func TestCursorBackwardExplain_ExecuteUsesReverseSeekForBeforeCursor(t *testing.T) {
	end := testUtils.NewCapturedCursor()
	req := `query($cursor: String) {
		_cursor {
			User(last: 2, before: $cursor, order: {age: ASC}) {
				name
				age
			}
			_pageInfo {
				hasNext
				hasPrev
			}
		}
	}`

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: userCollectionGQLSchema,
			},
			&action.AddDoc{Doc: `{"name": "Alice", "age": 10}`},
			&action.AddDoc{Doc: `{"name": "Bob", "age": 20}`},
			&action.AddDoc{Doc: `{"name": "Carol", "age": 30}`},
			&action.AddDoc{Doc: `{"name": "Dave", "age": 40}`},
			&action.AddDoc{Doc: `{"name": "Eve", "age": 50}`},

			&action.Request{
				Request: `query {
					_cursor {
						User(first: 5, order: {age: ASC}) {
							name
							age
						}
						_pageInfo {
							endCursor
						}
					}
				}`,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Alice", "age": int64(10)},
							{"name": "Bob", "age": int64(20)},
							{"name": "Carol", "age": int64(30)},
							{"name": "Dave", "age": int64(40)},
							{"name": "Eve", "age": int64(50)},
						},
						"_pageInfo": map[string]any{
							"endCursor": end,
						},
					},
				},
			},

			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": end,
				}),
				Request: req,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Carol", "age": int64(30)},
							{"name": "Dave", "age": int64(40)},
						},
						"_pageInfo": map[string]any{
							"hasNext": true,
							"hasPrev": true,
						},
					},
				},
			},

			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": end,
				}),
				Request: `query($cursor: String) @explain(type: execute) {
					_cursor {
						User(last: 2, before: $cursor, order: {age: ASC}) {
							name
							age
						}
						_pageInfo {
							hasNext
							hasPrev
						}
					}
				}`,
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
