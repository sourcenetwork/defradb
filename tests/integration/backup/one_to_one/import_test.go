// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package backup

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestBackupImport_WithMultipleNoKeyAndMultipleCollections_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.BackupImport{
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
			testUtils.Request{
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
			},
			testUtils.Request{
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
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupImport_WithMultipleNoKeyAndMultipleCollectionsAndUpdatedDocs_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.BackupImport{
				ImportContent: `{
					"Book":[
						{
							"_docID":"bae-af59fdc4-e495-5fd3-a9a6-386249aafdbb",
							"_docIDNew":"bae-d374c406-c6ea-51cd-9e9b-dd44a97b499c",
							"author_id":"bae-97f27fca-8b97-59f1-afa1-2e63140de933",
							"name":"John and the sourcerers' stone"
						}
					],
					"User":[
						{
							"_docID":"bae-88fea952-a678-5e05-9895-8a86ac6abc3b",
							"_docIDNew":"bae-88fea952-a678-5e05-9895-8a86ac6abc3b",
							"age":31,
							"name":"Bob"
						},
						{
							"_docID":"bae-a911f9cc-217a-58a3-a2f4-96548197403e",
							"_docIDNew":"bae-97f27fca-8b97-59f1-afa1-2e63140de933",
							"age":31,
							"name":"John"
						}
					]
				}`,
			},
			testUtils.Request{
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
			},
			testUtils.Request{
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
								"_docID": "bae-97f27fca-8b97-59f1-afa1-2e63140de933",
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
			testUtils.BackupImport{
				ImportContent: `{
					"Book":[
						{
							"_docID":"bae-4399f189-138d-5d49-9e25-82e78463677b",
							"_docIDNew":"bae-78a40f28-a4b8-5dca-be44-392b0f96d0ff",
							"author_id":"bae-97f27fca-8b97-59f1-afa1-2e63140de933",
							"name":"Game of chains"
						},
						{
							"_docID":"bae-af59fdc4-e495-5fd3-a9a6-386249aafdbb",
							"_docIDNew":"bae-d374c406-c6ea-51cd-9e9b-dd44a97b499c",
							"author_id":"bae-97f27fca-8b97-59f1-afa1-2e63140de933",
							"name":"John and the sourcerers' stone"
						}
					],
					"User":[
						{
							"_docID":"bae-88fea952-a678-5e05-9895-8a86ac6abc3b",
							"_docIDNew":"bae-88fea952-a678-5e05-9895-8a86ac6abc3b",
							"age":31,
							"name":"Bob"
						},
						{
							"_docID":"bae-a911f9cc-217a-58a3-a2f4-96548197403e",
							"_docIDNew":"bae-97f27fca-8b97-59f1-afa1-2e63140de933",
							"age":31,
							"name":"John"
						}
					]
				}`,
				ExpectedError: "target document is already linked to another document.",
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupImport_DoubleRelationshipWithUpdate_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
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
			testUtils.BackupImport{
				ImportContent: `{"User":[{"_docID":"bae-88fea952-a678-5e05-9895-8a86ac6abc3b","_docIDNew":"bae-88fea952-a678-5e05-9895-8a86ac6abc3b","age":31,"name":"Bob"},{"_docID":"bae-a911f9cc-217a-58a3-a2f4-96548197403e","_docIDNew":"bae-97f27fca-8b97-59f1-afa1-2e63140de933","age":31,"name":"John"}],"Book":[{"_docID":"bae-4cb9a1d2-eef3-564d-8695-1ce61c596e5a","_docIDNew":"bae-4cb9a1d2-eef3-564d-8695-1ce61c596e5a","name":"Game of chains"},{"_docID":"bae-556ece21-bf45-5652-8f32-c8a40373e8b5","_docIDNew":"bae-ccdd6d22-7339-5978-b0cb-f25d3d95c06d","author_id":"bae-97f27fca-8b97-59f1-afa1-2e63140de933","favourite_id":"bae-97f27fca-8b97-59f1-afa1-2e63140de933","name":"John and the sourcerers' stone"}]}`,
			},
			testUtils.Request{
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
