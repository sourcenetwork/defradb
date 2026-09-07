// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package cursor

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCursorWithFirstAfter_CursorsAreNonEmptyStrings(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: userCollectionGQLSchema,
			},
			&action.AddDoc{Doc: `{"name": "Alice", "age": 25}`},
			&action.AddDoc{Doc: `{"name": "Bob", "age": 30}`},
			&action.AddDoc{Doc: `{"name": "Carol", "age": 35}`},
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
	testUtils.ExecuteTestCase(t, test)
}

func TestCursorWithFirstAfter_AfterNullStartsFromBeginning(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: userCollectionGQLSchema,
			},
			&action.AddDoc{Doc: `{"name": "Alice", "age": 25}`},
			&action.AddDoc{Doc: `{"name": "Bob", "age": 30}`},
			&action.AddDoc{Doc: `{"name": "Carol", "age": 35}`},
			&action.Request{
				Request: `query {
					_cursor {
						User(first: 2, after: null, order: {age: ASC}) {
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
	testUtils.ExecuteTestCase(t, test)
}

func TestCursorWithFirstAfter_HasPrevFalseOnFirstPage(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: userCollectionGQLSchema,
			},
			&action.AddDoc{Doc: `{"name": "Alice", "age": 25}`},
			&action.AddDoc{Doc: `{"name": "Bob", "age": 30}`},
			&action.AddDoc{Doc: `{"name": "Carol", "age": 35}`},
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
	testUtils.ExecuteTestCase(t, test)
}

func TestCursorWithFirstAfter_CursorsEncodeIndexKeyValues(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: userCollectionGQLSchema,
			},
			&action.AddDoc{Doc: `{"name": "Alice", "age": 25}`},
			&action.AddDoc{Doc: `{"name": "Bob", "age": 30}`},
			&action.AddDoc{Doc: `{"name": "Carol", "age": 35}`},
			&action.Request{
				Request: `query {
					_cursor {
						User(first: 2, order: {age: ASC}) {
							name
							age
						}
						_pageInfo {
							hasNext
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
							"startCursor": testUtils.ValidCursor().WithKeys(map[string]any{"age": int64(25)}),
							"endCursor":   testUtils.ValidCursor().WithKeys(map[string]any{"age": int64(30)}),
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
