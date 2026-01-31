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

func TestCursorMultiRoundTrip_FullDatasetTraversal(t *testing.T) {
	var allDocs []map[string]any

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{Doc: `{"name": "John", "age": 10}`},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "age": 20}`},
			testUtils.CreateDoc{Doc: `{"name": "Fred", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Keenan", "age": 40}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "age": 50}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "age": 60}`},
			testUtils.CreateDoc{Doc: `{"name": "Shahzad", "age": 70}`},

			// Page 1: first 3 docs (ages 10, 20, 30)
			&action.Request{
				Request: `query {
					_cursor {
						User(first: 3, order: {age: ASC}) {
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
							{"name": "John", "age": int64(10)},
							{"name": "Addo", "age": int64(20)},
							{"name": "Fred", "age": int64(30)},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"hasPrev":   false,
							"endCursor": testUtils.CaptureCursor("page1End"),
						},
					},
				},
			},

			// Page 2: next 3 docs (ages 40, 50, 60)
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("page1End"),
				}),
				Request: `query($cursor: String) {
					_cursor {
						User(first: 3, after: $cursor, order: {age: ASC}) {
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
							{"name": "Keenan", "age": int64(40)},
							{"name": "Andy", "age": int64(50)},
							{"name": "Chris", "age": int64(60)},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"hasPrev":   true,
							"endCursor": testUtils.CaptureCursor("page2End"),
						},
					},
				},
			},

			// Page 3: final page (age 70), verify global properties
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("page2End"),
				}),
				Request: `query($cursor: String) {
					_cursor {
						User(first: 3, after: $cursor, order: {age: ASC}) {
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

					// Global verification: total count and order
					if len(allDocs) != 7 {
						return false, fmt.Sprintf("expected 7 docs total, got %d", len(allDocs))
					}
					for i := 1; i < len(allDocs); i++ {
						prevAge := allDocs[i-1]["age"].(int64)
						currAge := allDocs[i]["age"].(int64)
						if currAge <= prevAge {
							return false, fmt.Sprintf("order violation at index %d: %d <= %d", i, currAge, prevAge)
						}
					}
					return true, ""
				}),
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Shahzad", "age": int64(70)},
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
	executeTestCase(t, test)
}

func TestCursorMultiRoundTrip_TwoPages(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{Doc: `{"name": "John", "age": 15}`},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Fred", "age": 35}`},
			testUtils.CreateDoc{Doc: `{"name": "Keenan", "age": 45}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "age": 55}`},

			// Page 1: first 3 docs (ages 15, 25, 35)
			&action.Request{
				Request: `query {
					_cursor {
						User(first: 3, order: {age: ASC}) {
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
							{"name": "John", "age": int64(15)},
							{"name": "Addo", "age": int64(25)},
							{"name": "Fred", "age": int64(35)},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"hasPrev":   false,
							"endCursor": testUtils.CaptureCursor("page1End"),
						},
					},
				},
			},

			// Page 2: remaining 2 docs (ages 45, 55)
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("page1End"),
				}),
				Request: `query($cursor: String) {
					_cursor {
						User(first: 3, after: $cursor, order: {age: ASC}) {
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
							{"name": "Keenan", "age": int64(45)},
							{"name": "Andy", "age": int64(55)},
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
	executeTestCase(t, test)
}

func TestCursorMultiRoundTrip_ExactPageBoundary(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{Doc: `{"name": "John", "age": 10}`},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "age": 20}`},
			testUtils.CreateDoc{Doc: `{"name": "Fred", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Keenan", "age": 40}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "age": 50}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "age": 60}`},

			// Page 1: first 3 docs (ages 10, 20, 30)
			&action.Request{
				Request: `query {
					_cursor {
						User(first: 3, order: {age: ASC}) {
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
							{"name": "John", "age": int64(10)},
							{"name": "Addo", "age": int64(20)},
							{"name": "Fred", "age": int64(30)},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"hasPrev":   false,
							"endCursor": testUtils.CaptureCursor("page1End"),
						},
					},
				},
			},

			// Page 2: exactly 3 remaining docs (ages 40, 50, 60), hasNext=false at boundary
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("page1End"),
				}),
				Request: `query($cursor: String) {
					_cursor {
						User(first: 3, after: $cursor, order: {age: ASC}) {
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
							{"name": "Keenan", "age": int64(40)},
							{"name": "Andy", "age": int64(50)},
							{"name": "Chris", "age": int64(60)},
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
	executeTestCase(t, test)
}

func TestCursorMultiRoundTrip_VariablePageSizes(t *testing.T) {
	var allDocs []map[string]any

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{Doc: `{"name": "John", "age": 10}`},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "age": 20}`},
			testUtils.CreateDoc{Doc: `{"name": "Fred", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Keenan", "age": 40}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "age": 50}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "age": 60}`},
			testUtils.CreateDoc{Doc: `{"name": "Shahzad", "age": 70}`},
			testUtils.CreateDoc{Doc: `{"name": "Bruno", "age": 80}`},
			testUtils.CreateDoc{Doc: `{"name": "Islam", "age": 90}`},
			testUtils.CreateDoc{Doc: `{"name": "Roy", "age": 100}`},

			// Page 1: first 3 (ages 10, 20, 30)
			&action.Request{
				Request: `query {
					_cursor {
						User(first: 3, order: {age: ASC}) {
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
							{"name": "John", "age": int64(10)},
							{"name": "Addo", "age": int64(20)},
							{"name": "Fred", "age": int64(30)},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"hasPrev":   false,
							"endCursor": testUtils.CaptureCursor("page1End"),
						},
					},
				},
			},

			// Page 2: first 5 after cursor (ages 40, 50, 60, 70, 80)
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("page1End"),
				}),
				Request: `query($cursor: String) {
					_cursor {
						User(first: 5, after: $cursor, order: {age: ASC}) {
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
							{"name": "Keenan", "age": int64(40)},
							{"name": "Andy", "age": int64(50)},
							{"name": "Chris", "age": int64(60)},
							{"name": "Shahzad", "age": int64(70)},
							{"name": "Bruno", "age": int64(80)},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"hasPrev":   true,
							"endCursor": testUtils.CaptureCursor("page2End"),
						},
					},
				},
			},

			// Page 3: first 10 after cursor (only ages 90, 100 remaining)
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("page2End"),
				}),
				Request: `query($cursor: String) {
					_cursor {
						User(first: 10, after: $cursor, order: {age: ASC}) {
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

					// Global verification: total count and order
					if len(allDocs) != 10 {
						return false, fmt.Sprintf("expected 10 docs total, got %d", len(allDocs))
					}
					for i := 1; i < len(allDocs); i++ {
						prevAge := allDocs[i-1]["age"].(int64)
						currAge := allDocs[i]["age"].(int64)
						if currAge <= prevAge {
							return false, fmt.Sprintf("order violation at index %d: %d <= %d", i, currAge, prevAge)
						}
					}
					return true, ""
				}),
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Islam", "age": int64(90)},
							{"name": "Roy", "age": int64(100)},
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
	executeTestCase(t, test)
}

func TestCursorMultiRoundTrip_SingleDocPerPage(t *testing.T) {
	var allDocs []map[string]any

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{Doc: `{"name": "John", "age": 10}`},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "age": 20}`},
			testUtils.CreateDoc{Doc: `{"name": "Fred", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Keenan", "age": 40}`},

			// Page 1: first 1 (age 10)
			&action.Request{
				Request: `query {
					_cursor {
						User(first: 1, order: {age: ASC}) {
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
							{"name": "John", "age": int64(10)},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"hasPrev":   false,
							"endCursor": testUtils.CaptureCursor("page1End"),
						},
					},
				},
			},

			// Page 2: first 1 after cursor (age 20)
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("page1End"),
				}),
				Request: `query($cursor: String) {
					_cursor {
						User(first: 1, after: $cursor, order: {age: ASC}) {
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
							{"name": "Addo", "age": int64(20)},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"hasPrev":   true,
							"endCursor": testUtils.CaptureCursor("page2End"),
						},
					},
				},
			},

			// Page 3: first 1 after cursor (age 30)
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("page2End"),
				}),
				Request: `query($cursor: String) {
					_cursor {
						User(first: 1, after: $cursor, order: {age: ASC}) {
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
							{"name": "Fred", "age": int64(30)},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"hasPrev":   true,
							"endCursor": testUtils.CaptureCursor("page3End"),
						},
					},
				},
			},

			// Page 4: final page (age 40)
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("page3End"),
				}),
				Request: `query($cursor: String) {
					_cursor {
						User(first: 1, after: $cursor, order: {age: ASC}) {
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

					// Global verification: total count and order
					if len(allDocs) != 4 {
						return false, fmt.Sprintf("expected 4 docs total, got %d", len(allDocs))
					}
					for i := 1; i < len(allDocs); i++ {
						prevAge := allDocs[i-1]["age"].(int64)
						currAge := allDocs[i]["age"].(int64)
						if currAge <= prevAge {
							return false, fmt.Sprintf("order violation at index %d: %d <= %d", i, currAge, prevAge)
						}
					}
					return true, ""
				}),
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Keenan", "age": int64(40)},
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
	executeTestCase(t, test)
}

func TestCursorMultiRoundTrip_FilteredSubset(t *testing.T) {
	var allDocs []map[string]any

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{Doc: `{"name": "John", "age": 15}`},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Fred", "age": 35}`},
			testUtils.CreateDoc{Doc: `{"name": "Keenan", "age": 45}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "age": 55}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "age": 65}`},
			testUtils.CreateDoc{Doc: `{"name": "Shahzad", "age": 75}`},

			// Page 1: first 2 with filter age >= 40 (ages 45, 55)
			&action.Request{
				Request: `query {
					_cursor {
						User(first: 2, filter: {age: {_geq: 40}}, order: {age: ASC}) {
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
							{"name": "Keenan", "age": int64(45)},
							{"name": "Andy", "age": int64(55)},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"hasPrev":   false,
							"endCursor": testUtils.CaptureCursor("page1End"),
						},
					},
				},
			},

			// Page 2: remaining filtered docs (ages 65, 75)
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("page1End"),
				}),
				Request: `query($cursor: String) {
					_cursor {
						User(first: 2, filter: {age: {_geq: 40}}, after: $cursor, order: {age: ASC}) {
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

					// Global verification: total count and order
					if len(allDocs) != 4 {
						return false, fmt.Sprintf("expected 4 docs total, got %d", len(allDocs))
					}
					for i := 1; i < len(allDocs); i++ {
						prevAge := allDocs[i-1]["age"].(int64)
						currAge := allDocs[i]["age"].(int64)
						if currAge <= prevAge {
							return false, fmt.Sprintf("order violation at index %d: %d <= %d", i, currAge, prevAge)
						}
					}
					return true, ""
				}),
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Chris", "age": int64(65)},
							{"name": "Shahzad", "age": int64(75)},
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
	executeTestCase(t, test)
}

func TestCursorMultiRoundTrip_FilteredFewResults(t *testing.T) {
	var allDocs []map[string]any

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{Doc: `{"name": "John", "age": 10}`},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "age": 20}`},
			testUtils.CreateDoc{Doc: `{"name": "Fred", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Keenan", "age": 40}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "age": 50}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "age": 60}`},
			testUtils.CreateDoc{Doc: `{"name": "Shahzad", "age": 70}`},
			testUtils.CreateDoc{Doc: `{"name": "Bruno", "age": 80}`},
			testUtils.CreateDoc{Doc: `{"name": "Islam", "age": 90}`},
			testUtils.CreateDoc{Doc: `{"name": "Roy", "age": 100}`},

			// Page 1: first 2 with filter age >= 80 (ages 80, 90)
			&action.Request{
				Request: `query {
					_cursor {
						User(first: 2, filter: {age: {_geq: 80}}, order: {age: ASC}) {
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
							{"name": "Bruno", "age": int64(80)},
							{"name": "Islam", "age": int64(90)},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"hasPrev":   false,
							"endCursor": testUtils.CaptureCursor("page1End"),
						},
					},
				},
			},

			// Page 2: remaining filtered doc (age 100 only)
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("page1End"),
				}),
				Request: `query($cursor: String) {
					_cursor {
						User(first: 2, filter: {age: {_geq: 80}}, after: $cursor, order: {age: ASC}) {
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

					// Global verification: total count and order
					if len(allDocs) != 3 {
						return false, fmt.Sprintf("expected 3 docs total, got %d", len(allDocs))
					}
					for i := 1; i < len(allDocs); i++ {
						prevAge := allDocs[i-1]["age"].(int64)
						currAge := allDocs[i]["age"].(int64)
						if currAge <= prevAge {
							return false, fmt.Sprintf("order violation at index %d: %d <= %d", i, currAge, prevAge)
						}
					}
					return true, ""
				}),
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Roy", "age": int64(100)},
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
	executeTestCase(t, test)
}
