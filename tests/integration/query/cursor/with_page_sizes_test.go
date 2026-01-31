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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// TestCursorWithPageSize_SingleItemPage verifies first: 1 returns exactly
// one result with correct hasNext/hasPrev values.
func TestCursorWithPageSize_SingleItemPage(t *testing.T) {
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
						User(first: 1, order: {age: ASC}) {
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

// TestCursorWithPageSize_LargeFirstFewerDocs verifies first: 100 with only
// 10 documents returns all documents and hasNext: false.
func TestCursorWithPageSize_LargeFirstFewerDocs(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{Doc: `{"name": "User01", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "User02", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "User03", "age": 35}`},
			testUtils.CreateDoc{Doc: `{"name": "User04", "age": 40}`},
			testUtils.CreateDoc{Doc: `{"name": "User05", "age": 45}`},
			testUtils.CreateDoc{Doc: `{"name": "User06", "age": 50}`},
			testUtils.CreateDoc{Doc: `{"name": "User07", "age": 55}`},
			testUtils.CreateDoc{Doc: `{"name": "User08", "age": 60}`},
			testUtils.CreateDoc{Doc: `{"name": "User09", "age": 65}`},
			testUtils.CreateDoc{Doc: `{"name": "User10", "age": 70}`},
			&action.Request{
				Request: `query {
					_cursor {
						User(first: 100, order: {age: ASC}) {
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
							{"name": "User01", "age": int64(25)},
							{"name": "User02", "age": int64(30)},
							{"name": "User03", "age": int64(35)},
							{"name": "User04", "age": int64(40)},
							{"name": "User05", "age": int64(45)},
							{"name": "User06", "age": int64(50)},
							{"name": "User07", "age": int64(55)},
							{"name": "User08", "age": int64(60)},
							{"name": "User09", "age": int64(65)},
							{"name": "User10", "age": int64(70)},
						},
						"_pageInfo": map[string]any{
							"hasNext": false,
							"hasPrev": false,
						},
					},
				},
			},
		},
	}
	executeTestCase(t, test)
}

// TestCursorWithPageSize_FirstEqualsDocCount verifies first: N where N equals
// the exact document count returns all results with hasNext: false.
func TestCursorWithPageSize_FirstEqualsDocCount(t *testing.T) {
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
						User(first: 5, order: {age: ASC}) {
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
							{"name": "Alice", "age": int64(25)},
							{"name": "Bob", "age": int64(30)},
							{"name": "Carol", "age": int64(35)},
							{"name": "Dave", "age": int64(40)},
							{"name": "Eve", "age": int64(45)},
						},
						"_pageInfo": map[string]any{
							"hasNext": false,
							"hasPrev": false,
						},
					},
				},
			},
		},
	}
	executeTestCase(t, test)
}

// TestCursorWithPageSize_FirstExceedsDocCount verifies first: N where N exceeds
// document count returns all available results with hasNext: false.
func TestCursorWithPageSize_FirstExceedsDocCount(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{Doc: `{"name": "Alice", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Bob", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Carol", "age": 35}`},
			&action.Request{
				Request: `query {
					_cursor {
						User(first: 10, order: {age: ASC}) {
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
							{"name": "Alice", "age": int64(25)},
							{"name": "Bob", "age": int64(30)},
							{"name": "Carol", "age": int64(35)},
						},
						"_pageInfo": map[string]any{
							"hasNext": false,
							"hasPrev": false,
						},
					},
				},
			},
		},
	}
	executeTestCase(t, test)
}

// TestCursorWithPageSize_ZeroFirst verifies first: 0 returns zero results
// with hasNext: true if documents exist (consistent with SQL LIMIT 0).
func TestCursorWithPageSize_ZeroFirst(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{Doc: `{"name": "Alice", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Bob", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Carol", "age": 35}`},
			&action.Request{
				Request: `query {
					_cursor {
						User(first: 0, order: {age: ASC}) {
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
