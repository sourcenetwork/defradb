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

package index

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestIndexDelete_ShouldNotHinderQuerying(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String @index
						age: Int 
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.DeleteIndex{
				IndexName: "User_name_ASC",
			},
			&action.Request{
				Request: `
					query  {
						User {
							name
							age
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
							"age":  int64(21),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndexDelete_ShouldRemoveIndexFromCollection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String @index
						age: Int @index
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.DeleteIndex{
				IndexName: "User_age_ASC",
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						ID:   1,
						Name: "User_name_ASC",
						Fields: []client.IndexedFieldDescription{
							{Name: "name"},
						},
					},
				},
			},
			&action.DeleteIndex{
				IndexName: "User_name_ASC",
			},
			&action.ListIndexes{
				CollectionID:    0,
				ExpectedIndexes: []client.IndexDescription{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestDocDeleteWithFilter_WithUniqueIndex_CleansUpIndexEntries verifies that a filter-based
// document deletion removes the document's index entries. It does this by adding a document,
// deleting the document with a filter, and then re-inserting a document with the same name.
// If the index entries were not cleaned up, the re-insertion would fail with a "violates unique index" error.
func TestDocDeleteWithFilter_WithUniqueIndex_CleansUpIndexEntries(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String @index(unique: true)
						age: Int
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"age": 30
				}`,
			},
			testUtils.DeleteWithFilter{
				CollectionID: 0,
				Filter:       `{name: {_eq: "John"}}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"age": 31
				}`,
			},
			&action.Request{
				Request: `query {
					User {
						name
						age
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John", "age": int64(31)},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestDocDeleteWithFilter_WithMultipleMatchingDocs_CleansUpAllIndexEntries verifies that a
// filter-based deletion removes index entries for every matched document, even when multiple
// documents share the same non-unique index value. It does this by adding multiple documents,
// deleting them, then re-inserting multiple documents with the same email values.
func TestDocDeleteWithFilter_WithMultipleMatchingDocs_CleansUpAllIndexEntries(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						city:  String @index
						email: String @index(unique: true)
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"city": "London", "email": "a@example.com"}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"city": "London", "email": "b@example.com"}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"city": "London", "email": "c@example.com"}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"city": "London", "email": "d@example.com"}`,
			},
			testUtils.DeleteWithFilter{
				CollectionID: 0,
				Filter:       `{city: {_eq: "London"}}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"city": "Paris", "email": "a@example.com"}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"city": "Paris", "email": "b@example.com"}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"city": "Paris", "email": "c@example.com"}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"city": "Paris", "email": "d@example.com"}`,
			},
			&action.Request{
				Request: `query {
					User {
						city
						email
					}
				}`,
				NonOrderedResults: true,
				Results: map[string]any{
					"User": []map[string]any{
						{"city": "Paris", "email": "a@example.com"},
						{"city": "Paris", "email": "b@example.com"},
						{"city": "Paris", "email": "c@example.com"},
						{"city": "Paris", "email": "d@example.com"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndexDelete_IfIndexDoesNotExist_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.DeleteIndex{
				CollectionID:  0,
				IndexName:     "non_existing_index",
				ExpectedError: "index with name doesn't exists. Name: non_existing_index",
			},
			&action.Request{
				Request: `
					query  {
						User {
							name
							age
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
							"age":  int64(21),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
