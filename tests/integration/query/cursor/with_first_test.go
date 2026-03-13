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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCursorWithFirst_ReturnsExactlyNResults(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{Doc: `{"name": "Alice", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Bob", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Carol", "age": 35}`},
			testUtils.CreateDoc{Doc: `{"name": "Dave", "age": 40}`},
			testUtils.CreateDoc{Doc: `{"name": "Eve", "age": 45}`},
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
							{"name": "Alice", "age": int64(25)},
							{"name": "Bob", "age": int64(30)},
							{"name": "Carol", "age": int64(35)},
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

func TestCursorWithFirst_ReturnsAllWhenFewerExist(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{Doc: `{"name": "Alice", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Bob", "age": 30}`},
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
							"hasNext":     false,
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

func TestCursorWithFirst_ReturnsPageInfoWithHasNextTrue(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{Doc: `{"name": "Alice", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Bob", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Carol", "age": 35}`},
			testUtils.CreateDoc{Doc: `{"name": "Dave", "age": 40}`},
			testUtils.CreateDoc{Doc: `{"name": "Eve", "age": 45}`},
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

func TestCursorWithFirst_ReturnsPageInfoWithHasNextFalse(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{Doc: `{"name": "Alice", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Bob", "age": 30}`},
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
							"hasNext":     false,
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
