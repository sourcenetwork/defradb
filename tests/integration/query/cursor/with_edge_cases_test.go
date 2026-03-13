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
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCursorEdgeCase_EmptyCollectionReturnsEmptyArray(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: userCollectionGQLSchema,
			},
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
							startCursor
							endCursor
						}
					}
				}`,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{},
						"_pageInfo": map[string]any{
							"hasNext":     false,
							"hasPrev":     false,
							"startCursor": nil,
							"endCursor":   nil,
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCursorEdgeCase_AfterCursorAtEndReturnsEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{Doc: `{"name": "Alice", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Bob", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Carol", "age": 35}`},

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
							"hasNext":     false,
							"hasPrev":     false,
							"startCursor": testUtils.ValidCursor(),
							"endCursor":   testUtils.CaptureCursor("page1End"),
						},
					},
				},
			},

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
							startCursor
							endCursor
						}
					}
				}`,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{},
						"_pageInfo": map[string]any{
							"hasNext":     false,
							"hasPrev":     true,
							"startCursor": nil,
							"endCursor":   nil,
						},
					},
				},
			},
		},
	}
	executeTestCase(t, test)
}

func TestCursorEdgeCase_MalformedBase64ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{Doc: `{"name": "Alice", "age": 25}`},
			&action.Request{
				Request: `query {
					_cursor {
						User(first: 1, after: "not-valid-base64!!!", order: {age: ASC}) {
							name
							age
						}
						_pageInfo {
							hasNext
							hasPrev
						}
					}
				}`,
				ExpectedError: "invalid cursor",
			},
		},
	}
	executeTestCase(t, test)
}

func TestCursorEdgeCase_InvalidJSONStructureReturnsError(t *testing.T) {
	invalidJSON := base64.StdEncoding.EncodeToString([]byte("not json"))

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{Doc: `{"name": "Alice", "age": 25}`},
			&action.Request{
				Request: fmt.Sprintf(`query {
					_cursor {
						User(first: 1, after: "%s", order: {age: ASC}) {
							name
							age
						}
						_pageInfo {
							hasNext
							hasPrev
						}
					}
				}`, invalidJSON),
				ExpectedError: "invalid cursor",
			},
		},
	}
	executeTestCase(t, test)
}

func TestCursorEdgeCase_MissingDocIDInCursorReturnsError(t *testing.T) {
	missingDocID := base64.StdEncoding.EncodeToString([]byte(`{"k":{"age":25},"o":""}`))

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{Doc: `{"name": "Alice", "age": 25}`},
			&action.Request{
				Request: fmt.Sprintf(`query {
					_cursor {
						User(first: 1, after: "%s", order: {age: ASC}) {
							name
							age
						}
						_pageInfo {
							hasNext
							hasPrev
						}
					}
				}`, missingDocID),
				ExpectedError: "invalid cursor",
			},
		},
	}
	executeTestCase(t, test)
}

func TestCursorEdgeCase_DeletedCursorDocContinuesToNext(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{Doc: `{"name": "Alice", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Bob", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Carol", "age": 35}`},

			&action.Request{
				Request: `query {
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
				}`,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Alice", "age": int64(25)},
							{"name": "Bob", "age": int64(30)},
						},
						"_pageInfo": map[string]any{
							"hasNext":   true,
							"endCursor": testUtils.CaptureCursor("bobCursor"),
						},
					},
				},
			},

			testUtils.DeleteDoc{
				CollectionID: 0,
				DocID:        1,
			},

			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("bobCursor"),
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
						}
					}
				}`,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"name": "Carol", "age": int64(35)},
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
	executeTestCase(t, test)
}

func TestCursorEdgeCase_DeletedResultDocExcludedFromResults(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{Doc: `{"name": "Alice", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Bob", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Carol", "age": 35}`},
			testUtils.CreateDoc{Doc: `{"name": "Dave", "age": 40}`},
			testUtils.CreateDoc{Doc: `{"name": "Eve", "age": 45}`},

			testUtils.DeleteDoc{
				CollectionID: 0,
				DocID:        2, // Carol
			},

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

func TestCursorEdgeCase_AllRemainingDocsDeletedReturnsEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{Doc: `{"name": "Alice", "age": 25}`},
			testUtils.CreateDoc{Doc: `{"name": "Bob", "age": 30}`},
			testUtils.CreateDoc{Doc: `{"name": "Carol", "age": 35}`},

			&action.Request{
				Request: `query {
					_cursor {
						User(first: 1, order: {age: ASC}) {
							name
							age
						}
						_pageInfo {
							hasNext
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
							"hasNext":   true,
							"endCursor": testUtils.CaptureCursor("aliceCursor"),
						},
					},
				},
			},

			testUtils.DeleteDoc{
				CollectionID: 0,
				DocID:        1,
			},
			testUtils.DeleteDoc{
				CollectionID: 0,
				DocID:        2,
			},

			&action.Request{
				Variables: immutable.Some(map[string]any{
					"cursor": testUtils.CapturedVar("aliceCursor"),
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
							startCursor
							endCursor
						}
					}
				}`,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{},
						"_pageInfo": map[string]any{
							"hasNext":     false,
							"hasPrev":     true,
							"startCursor": nil,
							"endCursor":   nil,
						},
					},
				},
			},
		},
	}
	executeTestCase(t, test)
}
