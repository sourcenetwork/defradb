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
	"time"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// userWithDateTimeIndexSchema defines a User type with indexed DateTime field for cursor pagination tests.
var userWithDateTimeIndexSchema = `
	type User {
		name: String
		birthday: DateTime @index
	}
`

func TestCursorWithDateTimeField_BasicPagination(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: userWithDateTimeIndexSchema,
			},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "birthday": "2000-01-15T00:00:00-00:00"}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "birthday": "2001-03-20T00:00:00-00:00"}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "birthday": "2002-06-10T00:00:00-00:00"}`},
			testUtils.CreateDoc{Doc: `{"name": "Fred", "birthday": "2003-09-25T00:00:00-00:00"}`},
			testUtils.CreateDoc{Doc: `{"name": "Islam", "birthday": "2004-12-01T00:00:00-00:00"}`},
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
							"endCursor": testUtils.ValidCursor(),
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCursorWithDateTimeField_MultiRoundTrip(t *testing.T) {
	var allDocs []map[string]any

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: userWithDateTimeIndexSchema,
			},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "birthday": "2000-01-15T00:00:00-00:00"}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "birthday": "2001-03-20T00:00:00-00:00"}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "birthday": "2002-06-10T00:00:00-00:00"}`},
			testUtils.CreateDoc{Doc: `{"name": "Fred", "birthday": "2003-09-25T00:00:00-00:00"}`},
			testUtils.CreateDoc{Doc: `{"name": "Islam", "birthday": "2004-12-01T00:00:00-00:00"}`},

			// Page 1: first 2 docs (2000, 2001)
			&action.Request{
				Request: `query {
					_cursor {
						User(first: 2, order: {birthday: ASC}) {
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
				Asserter: testUtils.ResultAsserterFunc(func(t testing.TB, result map[string]any) (bool, string) {
					cursor := result["_cursor"].(map[string]any)
					users := extractUsers(cursor["User"])
					allDocs = append(allDocs, users...)
					return true, ""
				}),
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Addo", "birthday": mustParseTime("2000-01-15T00:00:00Z")},
							{"name": "Andy", "birthday": mustParseTime("2001-03-20T00:00:00Z")},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"hasPrev":   false,
							"endCursor": testUtils.CaptureCursor("page1End"),
						},
					},
				},
			},

			// Page 2: next 2 docs (2002, 2003)
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("page1End"),
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
							{"name": "Chris", "birthday": mustParseTime("2002-06-10T00:00:00Z")},
							{"name": "Fred", "birthday": mustParseTime("2003-09-25T00:00:00Z")},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"hasPrev":   true,
							"endCursor": testUtils.CaptureCursor("page2End"),
						},
					},
				},
			},

			// Page 3: final doc (2004), verify global chronological order
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("page2End"),
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
							endCursor
						}
					}
				}`,
				Asserter: testUtils.ResultAsserterFunc(func(t testing.TB, result map[string]any) (bool, string) {
					cursor := result["_cursor"].(map[string]any)
					users := extractUsers(cursor["User"])
					allDocs = append(allDocs, users...)

					// Global verification: total count and chronological order
					if len(allDocs) != 5 {
						return false, fmt.Sprintf("expected 5 docs total, got %d", len(allDocs))
					}
					for i := 1; i < len(allDocs); i++ {
						prevBirthday := allDocs[i-1]["birthday"].(time.Time)
						currBirthday := allDocs[i]["birthday"].(time.Time)
						if !currBirthday.After(prevBirthday) {
							return false, fmt.Sprintf("chronological order violation at index %d: %v not after %v",
								i, currBirthday, prevBirthday)
						}
					}
					return true, ""
				}),
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Islam", "birthday": mustParseTime("2004-12-01T00:00:00Z")},
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

func TestCursorWithDateTimeField_IndexUsage(t *testing.T) {
	req := `query {
		_cursor {
			User(first: 3, order: {birthday: ASC}) {
				name
				birthday
			}
			_pageInfo {
				hasNext
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: userWithDateTimeIndexSchema,
			},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "birthday": "2000-01-15T00:00:00-00:00"}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "birthday": "2001-03-20T00:00:00-00:00"}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "birthday": "2002-06-10T00:00:00-00:00"}`},
			testUtils.CreateDoc{Doc: `{"name": "Fred", "birthday": "2003-09-25T00:00:00-00:00"}`},
			testUtils.CreateDoc{Doc: `{"name": "Islam", "birthday": "2004-12-01T00:00:00-00:00"}`},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Addo", "birthday": mustParseTime("2000-01-15T00:00:00Z")},
							{"name": "Andy", "birthday": mustParseTime("2001-03-20T00:00:00Z")},
							{"name": "Chris", "birthday": mustParseTime("2002-06-10T00:00:00Z")},
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

func TestCursorWithDateTimeField_SameYearDifferentDates(t *testing.T) {
	var allDocs []map[string]any

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: userWithDateTimeIndexSchema,
			},
			// All born in 2000 but different months/days
			testUtils.CreateDoc{Doc: `{"name": "Addo", "birthday": "2000-01-15T00:00:00-00:00"}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "birthday": "2000-03-20T00:00:00-00:00"}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "birthday": "2000-06-10T00:00:00-00:00"}`},
			testUtils.CreateDoc{Doc: `{"name": "Fred", "birthday": "2000-09-25T00:00:00-00:00"}`},

			// Page 1: first 2 docs (Jan, Mar)
			&action.Request{
				Request: `query {
					_cursor {
						User(first: 2, order: {birthday: ASC}) {
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
				Asserter: testUtils.ResultAsserterFunc(func(t testing.TB, result map[string]any) (bool, string) {
					cursor := result["_cursor"].(map[string]any)
					users := extractUsers(cursor["User"])
					allDocs = append(allDocs, users...)
					return true, ""
				}),
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Addo", "birthday": mustParseTime("2000-01-15T00:00:00Z")},
							{"name": "Andy", "birthday": mustParseTime("2000-03-20T00:00:00Z")},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"hasPrev":   false,
							"endCursor": testUtils.CaptureCursor("page1End"),
						},
					},
				},
			},

			// Page 2: remaining docs (Jun, Sep), verify fine-grained chronological order
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("page1End"),
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
							endCursor
						}
					}
				}`,
				Asserter: testUtils.ResultAsserterFunc(func(t testing.TB, result map[string]any) (bool, string) {
					cursor := result["_cursor"].(map[string]any)
					users := extractUsers(cursor["User"])
					allDocs = append(allDocs, users...)

					// Global verification: total count and fine-grained chronological order
					if len(allDocs) != 4 {
						return false, fmt.Sprintf("expected 4 docs total, got %d", len(allDocs))
					}
					for i := 1; i < len(allDocs); i++ {
						prevBirthday := allDocs[i-1]["birthday"].(time.Time)
						currBirthday := allDocs[i]["birthday"].(time.Time)
						if !currBirthday.After(prevBirthday) {
							return false, fmt.Sprintf("chronological order violation at index %d: %v not after %v",
								i, currBirthday, prevBirthday)
						}
					}
					return true, ""
				}),
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Chris", "birthday": mustParseTime("2000-06-10T00:00:00Z")},
							{"name": "Fred", "birthday": mustParseTime("2000-09-25T00:00:00Z")},
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

// mustParseTime parses a time string in RFC3339 format and panics on error.
func mustParseTime(timeString string) time.Time {
	t, err := time.Parse(time.RFC3339, timeString)
	if err != nil {
		panic(err)
	}
	return t
}
