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

var userWithBackwardDateTimeIndexSchema = `
	type User {
		name: String
		birthday: DateTime @index
	}
`

func TestCursorBackwardWithDateTimeField_BasicPagination(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: userWithBackwardDateTimeIndexSchema,
			},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "birthday": "2000-01-15T00:00:00-00:00"}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "birthday": "2001-03-20T00:00:00-00:00"}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "birthday": "2002-06-10T00:00:00-00:00"}`},
			testUtils.CreateDoc{Doc: `{"name": "Fred", "birthday": "2003-09-25T00:00:00-00:00"}`},
			testUtils.CreateDoc{Doc: `{"name": "Islam", "birthday": "2004-12-01T00:00:00-00:00"}`},
			&action.Request{
				Request: `query {
					_cursor {
						User(last: 2, order: {birthday: ASC}) {
							name
							birthday
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
							{"name": "Fred", "birthday": mustParseTime("2003-09-25T00:00:00Z")},
							{"name": "Islam", "birthday": mustParseTime("2004-12-01T00:00:00Z")},
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

func TestCursorBackwardWithDateTimeField_MultiRoundTrip(t *testing.T) {
	page1End := testUtils.NewCapturedCursor()
	p2Start := testUtils.NewCapturedCursor()
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: userWithBackwardDateTimeIndexSchema,
			},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "birthday": "2000-01-15T00:00:00-00:00"}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "birthday": "2001-03-20T00:00:00-00:00"}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "birthday": "2002-06-10T00:00:00-00:00"}`},
			testUtils.CreateDoc{Doc: `{"name": "Fred", "birthday": "2003-09-25T00:00:00-00:00"}`},
			testUtils.CreateDoc{Doc: `{"name": "Islam", "birthday": "2004-12-01T00:00:00-00:00"}`},

			// Forward page 1: first 3 -> [Addo/2000, Andy/2001, Chris/2002]
			&action.Request{
				Request: `query {
					_cursor {
						User(first: 3, order: {birthday: ASC}) {
							name
							birthday
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
							{"name": "Addo", "birthday": mustParseTime("2000-01-15T00:00:00Z")},
							{"name": "Andy", "birthday": mustParseTime("2001-03-20T00:00:00Z")},
							{"name": "Chris", "birthday": mustParseTime("2002-06-10T00:00:00Z")},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"hasPrev":   false,
							"endCursor": page1End,
						},
					},
				},
			},

			// Forward page 2: first 2 after page1End -> [Fred/2003, Islam/2004]
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": page1End,
				}),
				Request: `query($cursor: String) {
					_cursor {
						User(first: 2, after: $cursor, order: {birthday: ASC}) {
							name
							birthday
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
							{"name": "Fred", "birthday": mustParseTime("2003-09-25T00:00:00Z")},
							{"name": "Islam", "birthday": mustParseTime("2004-12-01T00:00:00Z")},
						},
						"_pageInfo": map[string]any{
							"hasNext":     false,
							"hasPrev":     true,
							"startCursor": p2Start,
						},
					},
				},
			},

			// Backward: last 2 before p2Start -> [Andy/2001, Chris/2002]
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": p2Start,
				}),
				Request: `query($cursor: String) {
					_cursor {
						User(last: 2, before: $cursor, order: {birthday: ASC}) {
							name
							birthday
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
							{"name": "Andy", "birthday": mustParseTime("2001-03-20T00:00:00Z")},
							{"name": "Chris", "birthday": mustParseTime("2002-06-10T00:00:00Z")},
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
