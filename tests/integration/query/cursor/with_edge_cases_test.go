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
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

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
	capturedEndCursor := new(string)

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
				Asserter: testUtils.ResultAsserterFunc(func(t testing.TB, result map[string]any) (bool, string) {
					cursor, ok := result["_cursor"].(map[string]any)
					require.True(t, ok, "expected _cursor field")

					usersRaw := cursor["User"]
					var userCount int
					switch users := usersRaw.(type) {
					case []any:
						userCount = len(users)
					case []map[string]any:
						userCount = len(users)
					default:
						t.Fatalf("unexpected User type: %T", usersRaw)
					}
					require.Equal(t, 3, userCount, "expected 3 users")

					pageInfo, ok := cursor["_pageInfo"].(map[string]any)
					require.True(t, ok, "expected _pageInfo field")

					endCursor, ok := pageInfo["endCursor"].(string)
					require.True(t, ok, "expected endCursor to be a string")
					require.NotEmpty(t, endCursor, "expected endCursor to be non-empty")
					*capturedEndCursor = endCursor

					require.Equal(t, false, pageInfo["hasNext"], "hasNext should be false")
					require.Equal(t, false, pageInfo["hasPrev"], "hasPrev should be false")

					return true, ""
				}),
			},
			&dynamicRequest{
				requestFn: func() string {
					return fmt.Sprintf(`query {
						_cursor {
							User(first: 3, after: "%s", order: {age: ASC}) {
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
					}`, *capturedEndCursor)
				},
				asserter: testUtils.ResultAsserterFunc(func(t testing.TB, result map[string]any) (bool, string) {
					cursor, ok := result["_cursor"].(map[string]any)
					require.True(t, ok, "expected _cursor field")

					usersRaw := cursor["User"]
					var userCount int
					switch users := usersRaw.(type) {
					case []any:
						userCount = len(users)
					case []map[string]any:
						userCount = len(users)
					default:
						t.Fatalf("unexpected User type: %T", usersRaw)
					}
					require.Equal(t, 0, userCount, "expected empty results after cursor at end")

					pageInfo, ok := cursor["_pageInfo"].(map[string]any)
					require.True(t, ok, "expected _pageInfo field")

					require.Equal(t, false, pageInfo["hasNext"], "hasNext should be false")
					require.Equal(t, true, pageInfo["hasPrev"], "hasPrev should be true")
					require.Nil(t, pageInfo["startCursor"], "startCursor should be nil for empty results")
					require.Nil(t, pageInfo["endCursor"], "endCursor should be nil for empty results")

					return true, ""
				}),
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
	capturedEndCursor := new(string)

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
							hasPrev
							startCursor
							endCursor
						}
					}
				}`,
				Asserter: testUtils.ResultAsserterFunc(func(t testing.TB, result map[string]any) (bool, string) {
					cursor, ok := result["_cursor"].(map[string]any)
					require.True(t, ok, "expected _cursor field")

					pageInfo, ok := cursor["_pageInfo"].(map[string]any)
					require.True(t, ok, "expected _pageInfo field")

					endCursor, ok := pageInfo["endCursor"].(string)
					require.True(t, ok, "expected endCursor to be a string")
					*capturedEndCursor = endCursor

					require.Equal(t, true, pageInfo["hasNext"], "hasNext should be true")

					return true, ""
				}),
			},
			testUtils.DeleteDoc{
				CollectionID: 0,
				DocID:        1,
			},
			&dynamicRequest{
				requestFn: func() string {
					return fmt.Sprintf(`query {
						_cursor {
							User(first: 2, after: "%s", order: {age: ASC}) {
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
					}`, *capturedEndCursor)
				},
				asserter: testUtils.ResultAsserterFunc(func(t testing.TB, result map[string]any) (bool, string) {
					cursor, ok := result["_cursor"].(map[string]any)
					require.True(t, ok, "expected _cursor field")

					usersRaw := cursor["User"]
					var users []map[string]any
					switch u := usersRaw.(type) {
					case []any:
						for _, item := range u {
							if m, ok := item.(map[string]any); ok {
								users = append(users, m)
							}
						}
					case []map[string]any:
						users = u
					}
					require.Len(t, users, 1, "expected 1 user (Carol)")
					require.Equal(t, "Carol", users[0]["name"], "expected Carol")
					require.Equal(t, int64(35), users[0]["age"], "expected age 35")

					pageInfo, ok := cursor["_pageInfo"].(map[string]any)
					require.True(t, ok, "expected _pageInfo field")

					require.Equal(t, false, pageInfo["hasNext"], "hasNext should be false")
					require.Equal(t, true, pageInfo["hasPrev"], "hasPrev should be true")

					return true, ""
				}),
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
				DocID:        2,
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
	capturedEndCursor := new(string)

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
							hasPrev
							startCursor
							endCursor
						}
					}
				}`,
				Asserter: testUtils.ResultAsserterFunc(func(t testing.TB, result map[string]any) (bool, string) {
					cursor, ok := result["_cursor"].(map[string]any)
					require.True(t, ok, "expected _cursor field")

					pageInfo, ok := cursor["_pageInfo"].(map[string]any)
					require.True(t, ok, "expected _pageInfo field")

					endCursor, ok := pageInfo["endCursor"].(string)
					require.True(t, ok, "expected endCursor to be a string")
					*capturedEndCursor = endCursor

					require.Equal(t, true, pageInfo["hasNext"], "hasNext should be true")

					return true, ""
				}),
			},
			testUtils.DeleteDoc{
				CollectionID: 0,
				DocID:        1,
			},
			testUtils.DeleteDoc{
				CollectionID: 0,
				DocID:        2,
			},
			&dynamicRequest{
				requestFn: func() string {
					return fmt.Sprintf(`query {
						_cursor {
							User(first: 10, after: "%s", order: {age: ASC}) {
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
					}`, *capturedEndCursor)
				},
				asserter: testUtils.ResultAsserterFunc(func(t testing.TB, result map[string]any) (bool, string) {
					cursor, ok := result["_cursor"].(map[string]any)
					require.True(t, ok, "expected _cursor field")

					usersRaw := cursor["User"]
					var userCount int
					switch users := usersRaw.(type) {
					case []any:
						userCount = len(users)
					case []map[string]any:
						userCount = len(users)
					default:
						t.Fatalf("unexpected User type: %T", usersRaw)
					}
					require.Equal(t, 0, userCount, "expected empty results")

					pageInfo, ok := cursor["_pageInfo"].(map[string]any)
					require.True(t, ok, "expected _pageInfo field")

					require.Equal(t, false, pageInfo["hasNext"], "hasNext should be false")
					require.Equal(t, true, pageInfo["hasPrev"], "hasPrev should be true")
					require.Nil(t, pageInfo["startCursor"], "startCursor should be nil for empty results")
					require.Nil(t, pageInfo["endCursor"], "endCursor should be nil for empty results")

					return true, ""
				}),
			},
		},
	}
	executeTestCase(t, test)
}
