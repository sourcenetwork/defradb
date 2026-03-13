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

var userWithBackwardFloatIndexSchema = `
	type User {
		name: String
		score: Float @index
	}
`

func TestCursorBackwardWithFloatField_BasicPagination(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: userWithBackwardFloatIndexSchema,
			},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "score": 1.5}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "score": 2.7}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "score": 3.9}`},
			testUtils.CreateDoc{Doc: `{"name": "Fred", "score": 5.1}`},
			testUtils.CreateDoc{Doc: `{"name": "Islam", "score": 6.3}`},
			&action.Request{
				Request: `query {
					_cursor {
						User(last: 2, order: {score: ASC}) {
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
							{"name": "Fred", "score": 5.1},
							{"name": "Islam", "score": 6.3},
						},
						"_pageInfo": map[string]any{
							"hasNext": false,
							"hasPrev": true,
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCursorBackwardWithFloatField_MultiRoundTrip(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: userWithBackwardFloatIndexSchema,
			},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "score": 1.5}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "score": 2.7}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "score": 3.9}`},
			testUtils.CreateDoc{Doc: `{"name": "Fred", "score": 5.1}`},
			testUtils.CreateDoc{Doc: `{"name": "Islam", "score": 6.3}`},

			// Forward page 1: first 3 -> [Addo/1.5, Andy/2.7, Chris/3.9]
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
							endCursor
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
							"hasNext":   true,
							"hasPrev":   false,
							"endCursor": testUtils.CaptureCursor("page1End"),
						},
					},
				},
			},

			// Forward page 2: first 2 after page1End -> [Fred/5.1, Islam/6.3]
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("page1End"),
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
							startCursor
						}
					}
				}`,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Fred", "score": 5.1},
							{"name": "Islam", "score": 6.3},
						},
						"_pageInfo": map[string]any{
							"hasNext":     false,
							"hasPrev":     true,
							"startCursor": testUtils.CaptureCursor("p2Start"),
						},
					},
				},
			},

			// Backward: last 2 before p2Start -> [Andy/2.7, Chris/3.9]
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("p2Start"),
				}),
				Request: `query($cursor: String) {
					_cursor {
						User(last: 2, before: $cursor, order: {score: ASC}) {
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
							{"name": "Andy", "score": 2.7},
							{"name": "Chris", "score": 3.9},
						},
						"_pageInfo": map[string]any{
							"hasNext": true,
							"hasPrev": true,
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
