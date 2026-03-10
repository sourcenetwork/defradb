// Copyright 2025 Democratized Data Foundation
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

type dataMap = map[string]any

func TestCursorBackwardExplain_SimpleShowsLastBeforeFields(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{Doc: `{"name": "Alice", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Bob", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Carol", "age": 35}`},

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
				Results: dataMap{
					"_cursor": dataMap{
						"User": []dataMap{
							{"name": "Bob", "age": int64(30)},
							{"name": "Carol", "age": int64(35)},
						},
						"_pageInfo": dataMap{
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
	executeTestCase(t, test)
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
			testUtils.CreateDoc{Doc: `{"name": "Alice", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Bob", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Carol", "age": 35}`},

			&action.Request{
				Request: req,
				Results: dataMap{
					"_cursor": dataMap{
						"User": []dataMap{
							{"name": "Bob", "age": int64(30)},
							{"name": "Carol", "age": int64(35)},
						},
						"_pageInfo": dataMap{
							"hasNext": false,
							"hasPrev": true,
						},
					},
				},
			},

			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithCursor().WithIndexFetches(3),
			},
		},
	}
	executeTestCase(t, test)
}

func TestCursorBackwardExplain_ExecuteUsesReverseSeekForBeforeCursor(t *testing.T) {
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
			testUtils.CreateDoc{Doc: `{"name": "Alice", "age": 10}`},
			testUtils.CreateDoc{Doc: `{"name": "Bob", "age": 20}`},
			testUtils.CreateDoc{Doc: `{"name": "Carol", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Dave", "age": 40}`},
			testUtils.CreateDoc{Doc: `{"name": "Eve", "age": 50}`},

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
				Results: dataMap{
					"_cursor": dataMap{
						"User": []dataMap{
							{"name": "Alice", "age": int64(10)},
							{"name": "Bob", "age": int64(20)},
							{"name": "Carol", "age": int64(30)},
							{"name": "Dave", "age": int64(40)},
							{"name": "Eve", "age": int64(50)},
						},
						"_pageInfo": dataMap{
							"endCursor": testUtils.CaptureCursor("end"),
						},
					},
				},
			},

			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("end"),
				}),
				Request: req,
				Results: dataMap{
					"_cursor": dataMap{
						"User": []dataMap{
							{"name": "Carol", "age": int64(30)},
							{"name": "Dave", "age": int64(40)},
						},
						"_pageInfo": dataMap{
							"hasNext": true,
							"hasPrev": true,
						},
					},
				},
			},

			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("end"),
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
				Asserter: testUtils.NewExplainAsserter().WithCursor().WithIndexFetches(3),
			},
		},
	}

	executeTestCase(t, test)
}
