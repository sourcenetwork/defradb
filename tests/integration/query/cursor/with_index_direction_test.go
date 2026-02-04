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

// TestCursorWithIndexDirection_DescIndexAscQueryRejects verifies that a query
// with ASC order on a DESC index is rejected (reverse scan not supported).
// Per decision [05-01]: Cursor pagination does NOT support reverse scans.
func TestCursorWithIndexDirection_DescIndexAscQueryRejects(t *testing.T) {
	req := `query {
		_cursor {
			User(first: 3, order: {age: ASC}) {
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
					type User {
						name: String
						age: Int @index(direction: DESC)
					}
				`,
			},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "age": 35}`},
			&action.Request{
				Request:       req,
				ExpectedError: "cursor index does not support required scan direction",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// TestCursorWithIndexDirection_AscIndexDescQueryRejects verifies that a query
// with DESC order on an ASC index (default) is rejected.
// Per decision [05-01]: Cursor pagination does NOT support reverse scans.
func TestCursorWithIndexDirection_AscIndexDescQueryRejects(t *testing.T) {
	req := `query {
		_cursor {
			User(first: 3, order: {age: DESC}) {
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
					type User {
						name: String
						age: Int @index
					}
				`,
			},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "age": 35}`},
			&action.Request{
				Request:       req,
				ExpectedError: "cursor index does not support required scan direction",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// TestCursorWithIndexDirection_MatchingDirectionSucceeds verifies that a DESC order
// query on a DESC index succeeds (matching direction works).
func TestCursorWithIndexDirection_MatchingDirectionSucceeds(t *testing.T) {
	req := `query {
		_cursor {
			User(first: 3, order: {age: DESC}) {
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
					type User {
						name: String
						age: Int @index(direction: DESC)
					}
				`,
			},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "age": 35}`},
			testUtils.CreateDoc{Doc: `{"name": "Fred", "age": 40}`},
			testUtils.CreateDoc{Doc: `{"name": "Islam", "age": 45}`},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Islam", "age": int64(45)},
							{"name": "Fred", "age": int64(40)},
							{"name": "Chris", "age": int64(35)},
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

// TestCursorWithIndexDirection_CompositeIndexMismatch verifies that a query with
// ASC order on a composite index with DESC direction on the leading field is rejected.
// Note: The error message is "no supporting index" because the direction mismatch means
// the index doesn't match the query ordering, not because of scan direction.
func TestCursorWithIndexDirection_CompositeIndexMismatch(t *testing.T) {
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
					type User @index(includes: [{field: "name", direction: DESC}, {field: "age"}]) {
						name: String
						age: Int
					}
				`,
			},
			testUtils.CreateDoc{Doc: `{"name": "Addo", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Andy", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Chris", "age": 35}`},
			&action.Request{
				Request:       req,
				ExpectedError: "no supporting index for cursor order field",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
