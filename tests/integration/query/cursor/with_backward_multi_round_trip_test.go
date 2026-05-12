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

func TestCursorBackwardMultiRoundTrip_ForwardThenBackward(t *testing.T) {
	p1Start := testUtils.NewCapturedCursor()
	p1End := testUtils.NewCapturedCursor()
	p2Start := testUtils.NewCapturedCursor()
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: userCollectionGQLSchema,
			},
			&action.AddDoc{Doc: `{"name": "John", "age": 10}`},
			&action.AddDoc{Doc: `{"name": "Addo", "age": 20}`},
			&action.AddDoc{Doc: `{"name": "Fred", "age": 30}`},
			&action.AddDoc{Doc: `{"name": "Keenan", "age": 40}`},
			&action.AddDoc{Doc: `{"name": "Andy", "age": 50}`},
			&action.AddDoc{Doc: `{"name": "Chris", "age": 60}`},

			// Forward page 1: first 3 -> [10, 20, 30]
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
							startCursor
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
							"hasNext":     true,
							"hasPrev":     false,
							"startCursor": p1Start,
							"endCursor":   p1End,
						},
					},
				},
			},

			// Forward page 2: next 3 after p1End -> [40, 50, 60]
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": p1End,
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
							startCursor
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
							"hasNext":     false,
							"hasPrev":     true,
							"startCursor": p2Start,
							"endCursor":   testUtils.ValidCursor(),
						},
					},
				},
			},

			// Backward from p2Start: last 2 before age 40 -> [20, 30]
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": p2Start,
				}),
				Request: `query($cursor: String) {
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
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Addo", "age": int64(20)},
							{"name": "Fred", "age": int64(30)},
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

func TestCursorBackwardMultiRoundTrip_BackwardThenForward(t *testing.T) {
	fEnd := testUtils.NewCapturedCursor()
	f2Start := testUtils.NewCapturedCursor()
	f2End := testUtils.NewCapturedCursor()
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: userCollectionGQLSchema,
			},
			&action.AddDoc{Doc: `{"name": "John", "age": 10}`},
			&action.AddDoc{Doc: `{"name": "Addo", "age": 20}`},
			&action.AddDoc{Doc: `{"name": "Fred", "age": 30}`},
			&action.AddDoc{Doc: `{"name": "Keenan", "age": 40}`},
			&action.AddDoc{Doc: `{"name": "Andy", "age": 50}`},
			&action.AddDoc{Doc: `{"name": "Chris", "age": 60}`},

			// Forward page 1: first 2 -> [10, 20]
			&action.Request{
				Request: `query {
					_cursor {
						User(first: 2, order: {age: ASC}) {
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
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"hasPrev":   false,
							"endCursor": fEnd,
						},
					},
				},
			},

			// Forward page 2: first 2 after fEnd -> [30, 40]
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": fEnd,
				}),
				Request: `query($cursor: String) {
					_cursor {
						User(first: 2, after: $cursor, order: {age: ASC}) {
							name
							age
						}
						_pageInfo {
							hasNext
							hasPrev
							startCursor
							endCursor
						}
					}
				}`,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Fred", "age": int64(30)},
							{"name": "Keenan", "age": int64(40)},
						},
						"_pageInfo": map[string]any{
							"hasNext":     true,
							"hasPrev":     true,
							"startCursor": f2Start,
							"endCursor":   f2End,
						},
					},
				},
			},

			// Backward: last 2 before f2Start (age 30) -> [10, 20]
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": f2Start,
				}),
				Request: `query($cursor: String) {
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
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "John", "age": int64(10)},
							{"name": "Addo", "age": int64(20)},
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

func TestCursorBackwardMultiRoundTrip_FullBackwardTraversal(t *testing.T) {
	b1Start := testUtils.NewCapturedCursor()
	b2Start := testUtils.NewCapturedCursor()
	var allDocs []map[string]any

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: userCollectionGQLSchema,
			},
			&action.AddDoc{Doc: `{"name": "John", "age": 10}`},
			&action.AddDoc{Doc: `{"name": "Addo", "age": 20}`},
			&action.AddDoc{Doc: `{"name": "Fred", "age": 30}`},
			&action.AddDoc{Doc: `{"name": "Keenan", "age": 40}`},
			&action.AddDoc{Doc: `{"name": "Andy", "age": 50}`},
			&action.AddDoc{Doc: `{"name": "Chris", "age": 60}`},

			// Backward page 1: last 2 -> [50, 60]
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
							startCursor
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
							{"name": "Andy", "age": int64(50)},
							{"name": "Chris", "age": int64(60)},
						},
						"_pageInfo": map[string]any{
							"hasNext":     false,
							"hasPrev":     true,
							"startCursor": b1Start,
						},
					},
				},
			},

			// Backward page 2: last 2 before b1Start -> [30, 40]
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": b1Start,
				}),
				Request: `query($cursor: String) {
					_cursor {
						User(last: 2, before: $cursor, order: {age: ASC}) {
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
				Asserter: testUtils.ResultAsserterFunc(func(t testing.TB, result map[string]any) (bool, string) {
					cursor := result["_cursor"].(map[string]any)
					users := extractUsers(cursor["User"])
					// Prepend since we're going backward
					allDocs = append(users, allDocs...)
					return true, ""
				}),
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Fred", "age": int64(30)},
							{"name": "Keenan", "age": int64(40)},
						},
						"_pageInfo": map[string]any{
							"hasNext":     true,
							"hasPrev":     true,
							"startCursor": b2Start,
						},
					},
				},
			},

			// Backward page 3: last 2 before b2Start -> [10, 20], hasPrev=false
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": b2Start,
				}),
				Request: `query($cursor: String) {
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
				Asserter: testUtils.ResultAsserterFunc(func(t testing.TB, result map[string]any) (bool, string) {
					cursor := result["_cursor"].(map[string]any)
					users := extractUsers(cursor["User"])
					allDocs = append(users, allDocs...)

					if len(allDocs) != 6 {
						return false, fmt.Sprintf("expected 6 docs total, got %d", len(allDocs))
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
							{"name": "John", "age": int64(10)},
							{"name": "Addo", "age": int64(20)},
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
