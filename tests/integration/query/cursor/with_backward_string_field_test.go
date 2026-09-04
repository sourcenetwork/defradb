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

var userWithBackwardStringIndexSchema = `
	type User {
		name: String @index
		age: Int
	}
`

func TestCursorBackwardWithStringField_BasicPagination(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: userWithBackwardStringIndexSchema,
			},
			&action.AddDoc{Doc: `{"name": "Addo", "age": 25}`},
			&action.AddDoc{Doc: `{"name": "Andy", "age": 30}`},
			&action.AddDoc{Doc: `{"name": "Chris", "age": 35}`},
			&action.AddDoc{Doc: `{"name": "Fred", "age": 40}`},
			&action.AddDoc{Doc: `{"name": "Islam", "age": 45}`},
			&action.Request{
				Request: `query {
					_cursor {
						User(last: 2, order: {name: ASC}) {
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
							{"name": "Fred", "age": int64(40)},
							{"name": "Islam", "age": int64(45)},
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

func TestCursorBackwardWithStringField_MultiRoundTrip(t *testing.T) {
	page1End := testUtils.NewCapturedCursor()
	p2Start := testUtils.NewCapturedCursor()
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: userWithBackwardStringIndexSchema,
			},
			&action.AddDoc{Doc: `{"name": "Addo", "age": 25}`},
			&action.AddDoc{Doc: `{"name": "Andy", "age": 30}`},
			&action.AddDoc{Doc: `{"name": "Chris", "age": 35}`},
			&action.AddDoc{Doc: `{"name": "Fred", "age": 40}`},
			&action.AddDoc{Doc: `{"name": "Islam", "age": 45}`},

			// Forward page 1: first 3 -> [Addo, Andy, Chris]
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
							endCursor
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
							"hasNext":   true,
							"hasPrev":   false,
							"endCursor": page1End,
						},
					},
				},
			},

			// Forward page 2: first 2 after page1End -> [Fred, Islam]
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": page1End,
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
							startCursor
						}
					}
				}`,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Fred", "age": int64(40)},
							{"name": "Islam", "age": int64(45)},
						},
						"_pageInfo": map[string]any{
							"hasNext":     false,
							"hasPrev":     true,
							"startCursor": p2Start,
						},
					},
				},
			},

			// Backward: last 2 before p2Start -> [Andy, Chris]
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": p2Start,
				}),
				Request: `query($cursor: String) {
					_cursor {
						User(last: 2, before: $cursor, order: {name: ASC}) {
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
							{"name": "Andy", "age": int64(30)},
							{"name": "Chris", "age": int64(35)},
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
