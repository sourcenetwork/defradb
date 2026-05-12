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
	"fmt"
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// userWithFloatIndexSchema defines a User type with indexed Float field for cursor pagination tests.
var userWithFloatIndexSchema = `
	type User {
		name: String
		score: Float @index
	}
`

// TestCursorWithFloatField_BasicPagination verifies basic cursor pagination
// works correctly when ordering on a Float-indexed field.
func TestCursorWithFloatField_BasicPagination(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: userWithFloatIndexSchema,
			},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "score": 1.5}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "score": 2.7}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "score": 3.9}`},
			testUtils.CreateDoc{Doc: `{"name": "Fred", "score": 4.2}`},
			testUtils.CreateDoc{Doc: `{"name": "Islam", "score": 5.8}`},
			&action.Request{
				Request: `query {
					_cursor {
						User(first: 3, order: {score: ASC}) {
							name
							score
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
							{"name": "Addo", "score": 1.5},
							{"name": "Andy", "score": 2.7},
							{"name": "Chris", "score": 3.9},
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

// TestCursorWithFloatField_MultiRoundTrip verifies that multi-page cursor pagination
// traverses all documents exactly once and maintains numeric order on Float field.
func TestCursorWithFloatField_MultiRoundTrip(t *testing.T) {
	page1End := testUtils.NewCapturedCursor()
	page2End := testUtils.NewCapturedCursor()
	var allDocs []map[string]any

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: userWithFloatIndexSchema,
			},
			testUtils.CreateDoc{Doc: `{"name": "Islam", "score": 5.8}`},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "score": 1.5}`},
			testUtils.CreateDoc{Doc: `{"name": "Fred", "score": 4.2}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "score": 2.7}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "score": 3.9}`},

			// Page 1: first 2 docs (by score: 1.5, 2.7)
			&action.Request{
				Request: `query {
					_cursor {
						User(first: 2, order: {score: ASC}) {
							name
							score
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
							{"name": "Addo", "score": 1.5},
							{"name": "Andy", "score": 2.7},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"hasPrev":   false,
							"endCursor": page1End,
						},
					},
				},
			},

			// Page 2: next 2 docs (by score: 3.9, 4.2)
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": page1End,
				}),
				Request: `query($cursor: String) {
					_cursor {
						User(first: 2, after: $cursor, order: {score: ASC}) {
							name
							score
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
							{"name": "Chris", "score": 3.9},
							{"name": "Fred", "score": 4.2},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"hasPrev":   true,
							"endCursor": page2End,
						},
					},
				},
			},

			// Page 3: final page (score: 5.8), verify global properties
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": page2End,
				}),
				Request: `query($cursor: String) {
					_cursor {
						User(first: 2, after: $cursor, order: {score: ASC}) {
							name
							score
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

					// Global verification: total count and numeric order
					if len(allDocs) != 5 {
						return false, fmt.Sprintf("expected 5 docs total, got %d", len(allDocs))
					}
					for i := 1; i < len(allDocs); i++ {
						prevScore := allDocs[i-1]["score"].(float64)
						currScore := allDocs[i]["score"].(float64)
						if currScore <= prevScore {
							return false, fmt.Sprintf("order violation at index %d: %f <= %f", i, currScore, prevScore)
						}
					}
					return true, ""
				}),
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Islam", "score": 5.8},
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

// TestCursorWithFloatField_IndexUsage verifies that queries ordering on
// Float-indexed fields actually use the index (indexFetches > 0).
func TestCursorWithFloatField_IndexUsage(t *testing.T) {
	req := `query {
		_cursor {
			User(first: 3, order: {score: ASC}) {
				name
				score
			}
			_pageInfo {
				hasNext
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: userWithFloatIndexSchema,
			},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "score": 1.5}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "score": 2.7}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "score": 3.9}`},
			testUtils.CreateDoc{Doc: `{"name": "Fred", "score": 4.2}`},
			testUtils.CreateDoc{Doc: `{"name": "Islam", "score": 5.8}`},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Addo", "score": 1.5},
							{"name": "Andy", "score": 2.7},
							{"name": "Chris", "score": 3.9},
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
