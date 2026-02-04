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
	"fmt"
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// userWithStringIndexSchema defines a User type with indexed String field for cursor pagination tests.
var userWithStringIndexSchema = `
	type User {
		name: String @index
		age: Int
	}
`

// TestCursorWithStringField_BasicPagination verifies basic cursor pagination
// works correctly when ordering on a String-indexed field.
func TestCursorWithStringField_BasicPagination(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: userWithStringIndexSchema,
			},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "age": 35}`},
			testUtils.CreateDoc{Doc: `{"name": "Fred", "age": 40}`},
			testUtils.CreateDoc{Doc: `{"name": "Islam", "age": 45}`},
			&action.Request{
				Request: `query {
					_cursor {
						User(first: 3, order: {name: ASC}) {
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
							{"name": "Addo", "age": int64(25)},
							{"name": "Andy", "age": int64(30)},
							{"name": "Chris", "age": int64(35)},
						},
						"_pageInfo": map[string]any{
							"hasNext": true,
							"hasPrev": false,
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// TestCursorWithStringField_MultiRoundTrip verifies that multi-page cursor pagination
// traverses all documents exactly once and maintains alphabetical order on String field.
func TestCursorWithStringField_MultiRoundTrip(t *testing.T) {
	var allDocs []map[string]any

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: userWithStringIndexSchema,
			},
			testUtils.CreateDoc{Doc: `{"name": "Islam", "age": 45}`},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Fred", "age": 40}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "age": 35}`},

			// Page 1: first 2 docs (alphabetically: Addo, Andy)
			&action.Request{
				Request: `query {
					_cursor {
						User(first: 2, order: {name: ASC}) {
							name
							age
						}
						_pageInfo {
							hasNext
							hasPrev
							endCursor
						}
					}
				}`,
				Asserter: testUtils.ResultAsserterFunc(func(t testing.TB, result map[string]any) (bool, string) {
					cursor := result["_cursor"].(map[string]any)
					users := extractUsers(cursor["User"])
					allDocs = append(allDocs, users...)
					return true, ""
				}),
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Addo", "age": int64(25)},
							{"name": "Andy", "age": int64(30)},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"hasPrev":   false,
							"endCursor": testUtils.CaptureCursor("page1End"),
						},
					},
				},
			},

			// Page 2: next 2 docs (alphabetically: Chris, Fred)
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("page1End"),
				}),
				Request: `query($cursor: String) {
					_cursor {
						User(first: 2, after: $cursor, order: {name: ASC}) {
							name
							age
						}
						_pageInfo {
							hasNext
							hasPrev
							endCursor
						}
					}
				}`,
				Asserter: testUtils.ResultAsserterFunc(func(t testing.TB, result map[string]any) (bool, string) {
					cursor := result["_cursor"].(map[string]any)
					users := extractUsers(cursor["User"])
					allDocs = append(allDocs, users...)
					return true, ""
				}),
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Chris", "age": int64(35)},
							{"name": "Fred", "age": int64(40)},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"hasPrev":   true,
							"endCursor": testUtils.CaptureCursor("page2End"),
						},
					},
				},
			},

			// Page 3: final page (alphabetically: Islam), verify global properties
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("page2End"),
				}),
				Request: `query($cursor: String) {
					_cursor {
						User(first: 2, after: $cursor, order: {name: ASC}) {
							name
							age
						}
						_pageInfo {
							hasNext
							hasPrev
							endCursor
						}
					}
				}`,
				Asserter: testUtils.ResultAsserterFunc(func(t testing.TB, result map[string]any) (bool, string) {
					cursor := result["_cursor"].(map[string]any)
					users := extractUsers(cursor["User"])
					allDocs = append(allDocs, users...)

					// Global verification: total count and alphabetical order
					if len(allDocs) != 5 {
						return false, fmt.Sprintf("expected 5 docs total, got %d", len(allDocs))
					}
					for i := 1; i < len(allDocs); i++ {
						prevName := allDocs[i-1]["name"].(string)
						currName := allDocs[i]["name"].(string)
						if currName <= prevName {
							return false, fmt.Sprintf("order violation at index %d: %s <= %s", i, currName, prevName)
						}
					}
					return true, ""
				}),
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Islam", "age": int64(45)},
						},
						"_pageInfo": map[string]any{
							"hasNext":   false,
							"hasPrev":   true,
							"endCursor": testUtils.ValidCursor(),
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// TestCursorWithStringField_IndexUsage verifies that queries ordering on
// String-indexed fields actually use the index (indexFetches > 0).
func TestCursorWithStringField_IndexUsage(t *testing.T) {
	req := `query {
		_cursor {
			User(first: 3, order: {name: ASC}) {
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
				Schema: userWithStringIndexSchema,
			},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "age": 35}`},
			testUtils.CreateDoc{Doc: `{"name": "Fred", "age": 40}`},
			testUtils.CreateDoc{Doc: `{"name": "Islam", "age": 45}`},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Addo", "age": int64(25)},
							{"name": "Andy", "age": int64(30)},
							{"name": "Chris", "age": int64(35)},
						},
						"_pageInfo": map[string]any{
							"hasNext": true,
						},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithCursor().WithIndexFetches(4),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
