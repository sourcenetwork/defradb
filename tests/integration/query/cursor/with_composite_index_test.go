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

// TestCursorWithCompositeIndex_OrderByAllFields verifies that cursor pagination
// works correctly when ordering by all fields of a composite index.
// Note: Ordering by only a prefix of composite index fields is not currently supported
// for multi-page cursor pagination.
func TestCursorWithCompositeIndex_OrderByAllFields(t *testing.T) {
	req := `query {
		_cursor {
			User(first: 3, order: [{name: ASC}, {age: ASC}]) {
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
				Schema: `
					type User @index(includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						email: String
					}
				`,
			},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "age": 25, "email": "addo@test.com"}`},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "age": 35, "email": "addo2@test.com"}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "age": 30, "email": "andy@test.com"}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "age": 40, "email": "chris@test.com"}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "age": 45, "email": "chris2@test.com"}`},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Addo", "age": int64(25)},
							{"name": "Addo", "age": int64(35)},
							{"name": "Andy", "age": int64(30)},
						},
						"_pageInfo": map[string]any{
							"hasNext": true,
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// TestCursorWithCompositeIndex_MultiRoundTrip verifies full dataset traversal
// using cursor pagination with a composite index ordering by all index fields.
func TestCursorWithCompositeIndex_MultiRoundTrip(t *testing.T) {
	var allDocs []map[string]any

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						email: String
					}
				`,
			},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "age": 25, "email": "addo@test.com"}`},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "age": 35, "email": "addo2@test.com"}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "age": 30, "email": "andy@test.com"}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "age": 40, "email": "chris@test.com"}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "age": 45, "email": "chris2@test.com"}`},

			// Page 1: first 2 docs (Addo 25, Addo 35)
			&action.Request{
				Request: `query {
					_cursor {
						User(first: 2, order: [{name: ASC}, {age: ASC}]) {
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
							{"name": "Addo", "age": int64(35)},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"hasPrev":   false,
							"endCursor": testUtils.CaptureCursor("page1End"),
						},
					},
				},
			},

			// Page 2: next 2 docs (Andy 30, Chris 40)
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("page1End"),
				}),
				Request: `query($cursor: String) {
					_cursor {
						User(first: 2, after: $cursor, order: [{name: ASC}, {age: ASC}]) {
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
							{"name": "Andy", "age": int64(30)},
							{"name": "Chris", "age": int64(40)},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"hasPrev":   true,
							"endCursor": testUtils.CaptureCursor("page2End"),
						},
					},
				},
			},

			// Page 3: final page (Chris 45), verify global properties
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("page2End"),
				}),
				Request: `query($cursor: String) {
					_cursor {
						User(first: 2, after: $cursor, order: [{name: ASC}, {age: ASC}]) {
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

					// Global verification: total count and order (name ASC, then age ASC)
					if len(allDocs) != 5 {
						return false, fmt.Sprintf("expected 5 docs total, got %d", len(allDocs))
					}
					for i := 1; i < len(allDocs); i++ {
						prevName := allDocs[i-1]["name"].(string)
						currName := allDocs[i]["name"].(string)
						prevAge := allDocs[i-1]["age"].(int64)
						currAge := allDocs[i]["age"].(int64)
						// For same name, age must increase. For different names, name must increase.
						if currName < prevName {
							return false, fmt.Sprintf("name order violation at index %d: %s < %s", i, currName, prevName)
						}
						if currName == prevName && currAge <= prevAge {
							return false, fmt.Sprintf("age order violation at index %d: %d <= %d for name %s", i, currAge, prevAge, currName)
						}
					}
					return true, ""
				}),
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Chris", "age": int64(45)},
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

// TestCursorWithCompositeIndex_IndexUsage verifies that the composite index
// is actually used when ordering by all index fields (indexFetches > 0).
func TestCursorWithCompositeIndex_IndexUsage(t *testing.T) {
	req := `query {
		_cursor {
			User(first: 3, order: [{name: ASC}, {age: ASC}]) {
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
				Schema: `
					type User @index(includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						email: String
					}
				`,
			},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "age": 25, "email": "addo@test.com"}`},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "age": 35, "email": "addo2@test.com"}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "age": 30, "email": "andy@test.com"}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "age": 40, "email": "chris@test.com"}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "age": 45, "email": "chris2@test.com"}`},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Addo", "age": int64(25)},
							{"name": "Addo", "age": int64(35)},
							{"name": "Andy", "age": int64(30)},
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

// TestCursorWithCompositeIndex_OrderByNonIndexedFieldFails verifies that ordering
// by a field not covered by any index fails with an appropriate error.
func TestCursorWithCompositeIndex_OrderByNonIndexedFieldFails(t *testing.T) {
	req := `query {
		_cursor {
			User(first: 3, order: {email: ASC}) {
				name
				age
				email
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						email: String
					}
				`,
			},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "age": 25, "email": "addo@test.com"}`},
			&action.Request{
				Request:       req,
				ExpectedError: "no supporting index for cursor order field",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
