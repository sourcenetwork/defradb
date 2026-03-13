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

// TestCursorWithIndexDirection_DescIndexAscQuerySucceeds verifies that a query
// with ASC order on a DESC index succeeds (reverse iteration supported).
// Per Phase 10: Reverse direction cursor pagination is now supported.
func TestCursorWithIndexDirection_DescIndexAscQuerySucceeds(t *testing.T) {
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
				Request: req,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Addo", "age": int64(25)},
							{"name": "Andy", "age": int64(30)},
							{"name": "Chris", "age": int64(35)},
						},
						"_pageInfo": map[string]any{
							"hasNext": false,
						},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithCursor().WithIndexFetches(3),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// TestCursorWithIndexDirection_AscIndexDescQuerySucceeds verifies that a query
// with DESC order on an ASC index (default) succeeds (reverse iteration supported).
// Per Phase 10: Reverse direction cursor pagination is now supported.
func TestCursorWithIndexDirection_AscIndexDescQuerySucceeds(t *testing.T) {
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
				Request: req,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Chris", "age": int64(35)},
							{"name": "Andy", "age": int64(30)},
							{"name": "Addo", "age": int64(25)},
						},
						"_pageInfo": map[string]any{
							"hasNext": false,
						},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithCursor().WithIndexFetches(3),
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
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithCursor().WithIndexFetches(4),
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
				Request: `query {
							_cursor {
								User(first: 3, order: [{name: ASC}, {age: ASC}]) {
									name
									age
								}
								_pageInfo {
									hasNext
								}
							}
						}`,
				ExpectedError: "no supporting index for cursor order field",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// TestCursorWithIndexDirection_AscIndexDescQueryMultiPage verifies multi-page
// cursor pagination works correctly when querying DESC on an ASC index.
// 5 documents, page size 2: page 1 (45, 40), page 2 (35, 30), page 3 (25).
func TestCursorWithIndexDirection_AscIndexDescQueryMultiPage(t *testing.T) {
	req := `query {
		_cursor {
			User(first: 2, order: {age: DESC}) {
				name
				age
			}
			_pageInfo {
				hasNext
				endCursor
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
			testUtils.CreateDoc{Doc: `{"name": "Fred", "age": 40}`},
			testUtils.CreateDoc{Doc: `{"name": "Islam", "age": 45}`},
			// Page 1: first 2 in DESC order
			&action.Request{
				Request: req,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Islam", "age": int64(45)},
							{"name": "Fred", "age": int64(40)},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"endCursor": testUtils.CaptureCursor("cursor1"),
						},
					},
				},
			},
			// Explain: verify page 1 uses index
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithCursor().WithIndexFetches(3),
			},
			// Page 2: next 2
			&action.Request{
				Request: `query($cursor: String) {
					_cursor {
						User(first: 2, after: $cursor, order: {age: DESC}) {
							name
							age
						}
						_pageInfo {
							hasNext
							endCursor
						}
					}
				}`,
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("cursor1"),
				}),
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Chris", "age": int64(35)},
							{"name": "Andy", "age": int64(30)},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"endCursor": testUtils.CaptureCursor("cursor2"),
						},
					},
				},
			},
			// Page 3: last 1
			&action.Request{
				Request: `query($cursor: String) {
					_cursor {
						User(first: 2, after: $cursor, order: {age: DESC}) {
							name
							age
						}
						_pageInfo {
							hasNext
						}
					}
				}`,
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("cursor2"),
				}),
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Addo", "age": int64(25)},
						},
						"_pageInfo": map[string]any{
							"hasNext": false,
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// TestCursorWithIndexDirection_DescIndexAscQueryMultiPage verifies multi-page
// cursor pagination works correctly when querying ASC on a DESC index.
// 5 documents, page size 2: page 1 (25, 30), page 2 (35, 40), page 3 (45).
func TestCursorWithIndexDirection_DescIndexAscQueryMultiPage(t *testing.T) {
	req := `query {
		_cursor {
			User(first: 2, order: {age: ASC}) {
				name
				age
			}
			_pageInfo {
				hasNext
				endCursor
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
			// Page 1: first 2 in ASC order
			&action.Request{
				Request: req,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Addo", "age": int64(25)},
							{"name": "Andy", "age": int64(30)},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"endCursor": testUtils.CaptureCursor("cursor1"),
						},
					},
				},
			},
			// Explain: verify page 1 uses index
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithCursor().WithIndexFetches(3),
			},
			// Page 2: next 2
			&action.Request{
				Request: `query($cursor: String) {
					_cursor {
						User(first: 2, after: $cursor, order: {age: ASC}) {
							name
							age
						}
						_pageInfo {
							hasNext
							endCursor
						}
					}
				}`,
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("cursor1"),
				}),
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Chris", "age": int64(35)},
							{"name": "Fred", "age": int64(40)},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"endCursor": testUtils.CaptureCursor("cursor2"),
						},
					},
				},
			},
			// Page 3: last 1
			&action.Request{
				Request: `query($cursor: String) {
					_cursor {
						User(first: 2, after: $cursor, order: {age: ASC}) {
							name
							age
						}
						_pageInfo {
							hasNext
						}
					}
				}`,
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("cursor2"),
				}),
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Islam", "age": int64(45)},
						},
						"_pageInfo": map[string]any{
							"hasNext": false,
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
