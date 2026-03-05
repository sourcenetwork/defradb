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

package backup

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestBackupImport_WithMultipleNoKeyAndMultipleCollections_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.ImportBackup{
				ImportContent: `{
					"User":[
						{"age":30,"name":"John"},
						{"age":31,"name":"Smith"},
						{"age":32,"name":"Bob"}
					],
					"Book":[
						{"name":"John and the sourcerers' stone"},
						{"name":"Game of chains"}
					]
				}`,
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
							"name": "Smith",
							"age":  int64(31),
						},
						{
							"name": "John",
							"age":  int64(30),
						},
						{
							"name": "Bob",
							"age":  int64(32),
						},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				Request: `
					query  {
						Book {
							name
						}
					}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name": "Game of chains",
						},
						{
							"name": "John and the sourcerers' stone",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupImport_WithMultipleNoKeyAndMultipleCollectionsAndUpdatedDocs_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.ImportBackup{
				ImportContent: `{
					"Book":[
						{
							"_authorID":"bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4",
							"name":"John and the sourcerers' stone"
						}
					],
					"User":[
						{
							"age":31,
							"name":"Bob"
						},
						{
							"age":31,
							"name":"John"
						}
					]
				}`,
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
							"name": "Bob",
							"age":  int64(31),
						},
						{
							"name": "John",
							"age":  int64(31),
						},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				Request: `
					query  {
						Book {
							name
							author {
								_docID
							}
						}
					}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name": "John and the sourcerers' stone",
							"author": map[string]any{
								"_docID": "bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4",
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupImport_WithMultipleNoKeyAndMultipleCollectionsAndMultipleUpdatedDocs_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.ImportBackup{
				ImportContent: `{
					"Book":[
						{
							"_authorID":"bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4",
							"name":"Game of chains"
						},
						{
							"_authorID":"bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4",
							"name":"John and the sourcerers' stone"
						}
					],
					"User":[
						{
							"age":31,
							"name":"Bob"
						},
						{
							"age":31,
							"name":"John"
						}
					]
				}`,
				ExpectedError: "can not index a doc's field(s) that violates unique index",
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupImport_DoubleRelationshipWithUpdate_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
				type User {
					name: String
					age: Int
					book: Book @relation(name: "written_books")
					favouriteBook: Book @relation(name: "favourite_books")
				}
				type Book {
					name: String
					author: User @relation(name: "written_books") @primary
					favourite: User @relation(name: "favourite_books") @primary
				}
				`,
			},
			testUtils.ImportBackup{
				ImportContent: `{"User":[{"age":31,"name":"Bob"},{"age":31,"name":"John"}],"Book":[{"name":"Game of chains"},{"_authorID":"bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4","_favouriteID":"bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4","name":"John and the sourcerers' stone"}]}`,
			},
			&action.Request{
				Request: `
					query  {
						Book {
							name
							author {
								name
								favouriteBook {
									name
								}
							}
						}
					}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":   "Game of chains",
							"author": nil,
						},
						{
							"name": "John and the sourcerers' stone",
							"author": map[string]any{
								"name": "John",
								"favouriteBook": map[string]any{
									"name": "John and the sourcerers' stone",
								},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
