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

	internalCursor "github.com/sourcenetwork/defradb/internal/cursor"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCursorWithLastBefore_ReturnsItemsBeforeCursor(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{Doc: `{"name": "Alice", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Bob", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Carol", "age": 35}`},
			testUtils.CreateDoc{Doc: `{"name": "Dave", "age": 40}`},
			testUtils.CreateDoc{Doc: `{"name": "Eve", "age": 45}`},

			// Step 1: Forward query to get first 3 items, capture endCursor
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
							{"name": "Alice", "age": int64(25)},
							{"name": "Bob", "age": int64(30)},
							{"name": "Carol", "age": int64(35)},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"hasPrev":   false,
							"endCursor": testUtils.CaptureCursor("p1End"),
						},
					},
				},
			},

			// Step 2: Forward query from p1End to get remaining, capture startCursor
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("p1End"),
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
						}
					}
				}`,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Dave", "age": int64(40)},
							{"name": "Eve", "age": int64(45)},
						},
						"_pageInfo": map[string]any{
							"hasNext":     false,
							"hasPrev":     true,
							"startCursor": testUtils.CaptureCursor("p2Start"),
						},
					},
				},
			},

			// Step 3: Backward query last:2, before: p2Start
			// p2Start points to age 40. last:2 before that should return
			// the 2 items immediately before age 40 in the dataset.
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("p2Start"),
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
							endCursor
						}
					}
				}`,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Bob", "age": int64(30)},
							{"name": "Carol", "age": int64(35)},
						},
						"_pageInfo": map[string]any{
							"hasNext":     true,
							"hasPrev":     true,
							"startCursor": testUtils.ValidCursor(),
							"endCursor":   testUtils.ValidCursor(),
						},
					},
				},
			},
		},
	}
	executeTestCase(t, test)
}

func TestCursorWithLastBefore_BeforeCursorAtFirstItem(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{Doc: `{"name": "Alice", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Bob", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Carol", "age": 35}`},
			testUtils.CreateDoc{Doc: `{"name": "Dave", "age": 40}`},
			testUtils.CreateDoc{Doc: `{"name": "Eve", "age": 45}`},

			// Capture startCursor from a full query (points to first item, age 25)
			&action.Request{
				Request: `query {
					_cursor {
						User(first: 5, order: {age: ASC}) {
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
							{"name": "Alice", "age": int64(25)},
							{"name": "Bob", "age": int64(30)},
							{"name": "Carol", "age": int64(35)},
							{"name": "Dave", "age": int64(40)},
							{"name": "Eve", "age": int64(45)},
						},
						"_pageInfo": map[string]any{
							"hasNext":     false,
							"hasPrev":     false,
							"startCursor": testUtils.CaptureCursor("firstItemCursor"),
						},
					},
				},
			},

			// Backward query before the first item: should return empty
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("firstItemCursor"),
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
						"User": []map[string]any{},
						"_pageInfo": map[string]any{
							"hasNext": true,
							"hasPrev": false,
						},
					},
				},
			},
		},
	}
	executeTestCase(t, test)
}

func TestCursorWithLastBefore_ReturnsFewerWhenNotEnough(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{Doc: `{"name": "Alice", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Bob", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Carol", "age": 35}`},
			testUtils.CreateDoc{Doc: `{"name": "Dave", "age": 40}`},
			testUtils.CreateDoc{Doc: `{"name": "Eve", "age": 45}`},

			// Forward first:3, capture endCursor (points to age 35)
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
							{"name": "Alice", "age": int64(25)},
							{"name": "Bob", "age": int64(30)},
							{"name": "Carol", "age": int64(35)},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"hasPrev":   false,
							"endCursor": testUtils.CaptureCursor("page1End"),
						},
					},
				},
			},

			// Backward last:10 before page1End (age 35).
			// Only 2 items exist before age 35: [25, 30].
			// Requesting 10 but only 2 available.
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("page1End"),
				}),
				Request: `query($cursor: String) {
					_cursor {
						User(last: 10, before: $cursor, order: {age: ASC}) {
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
							{"name": "Alice", "age": int64(25)},
							{"name": "Bob", "age": int64(30)},
						},
						"_pageInfo": map[string]any{
							"hasNext":     true,
							"hasPrev":     false,
							"startCursor": testUtils.ValidCursor(),
							"endCursor":   testUtils.ValidCursor(),
						},
					},
				},
			},
		},
	}
	executeTestCase(t, test)
}

func TestCursorWithLastBefore_DocIDOnlyCursorFallsBackToDrain(t *testing.T) {
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
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Alice", "age": int64(10)},
							{"name": "Bob", "age": int64(20)},
							{"name": "Carol", "age": int64(30)},
							{"name": "Dave", "age": int64(40)},
							{"name": "Eve", "age": int64(50)},
						},
						"_pageInfo": map[string]any{
							"endCursor": captureDocIDOnlyCursor("legacyEnd"),
						},
					},
				},
			},

			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("legacyEnd"),
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
							{"name": "Carol", "age": int64(30)},
							{"name": "Dave", "age": int64(40)},
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
	executeTestCase(t, test)
}

func captureDocIDOnlyCursor(name string) *docIDOnlyCursorCapture {
	return &docIDOnlyCursorCapture{name: name}
}

type docIDOnlyCursorCapture struct {
	s           testUtils.TestState
	name        string
	failureText string
}

var _ testUtils.TestStateMatcher = (*docIDOnlyCursorCapture)(nil)
var _ testUtils.StatefulMatcher = (*docIDOnlyCursorCapture)(nil)

func (m *docIDOnlyCursorCapture) SetTestState(s testUtils.TestState) {
	m.s = s
}

func (m *docIDOnlyCursorCapture) ResetMatcherState() {
	m.failureText = ""
}

func (m *docIDOnlyCursorCapture) Match(actual any) (bool, error) {
	cursorValue, ok := actual.(string)
	if !ok {
		m.failureText = "expected cursor string"
		return false, nil
	}

	payload, err := internalCursor.Decode(cursorValue)
	if err != nil {
		m.failureText = err.Error()
		return false, nil
	}

	legacyCursor, err := internalCursor.Encode(internalCursor.CursorPayload{
		DocID:     payload.DocID,
		Direction: payload.Direction,
	})
	if err != nil {
		return false, err
	}

	m.s.SetCapturedVariable(m.name, legacyCursor)
	return true, nil
}

func (m *docIDOnlyCursorCapture) FailureMessage(actual any) string {
	if m.failureText == "" {
		return "failed to capture docID-only cursor"
	}
	return m.failureText
}

func (m *docIDOnlyCursorCapture) NegatedFailureMessage(actual any) string {
	return "expected cursor capture to fail"
}
