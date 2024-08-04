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
							"name": "John",
							"age":  int64(30),
						},
						{
							"name": "Bob",
							"age":  int64(32),
						},
						{
							"name": "Smith",
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
							"_docID":"bae-f33a7110-fb6f-57aa-9501-df0111427315",
							"_docIDNew":"bae-c9c1a385-afce-5ef7-8b98-9369b157fd97",
							"author_id":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f",
							"name":"John and the sourcerers' stone"
						}
					],
					"User":[
						{
							"_docID":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f",
							"_docIDNew":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f",
							"age":31,
							"name":"Bob"
						},
						{
							"_docID":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d",
							"_docIDNew":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f",
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
							"name": "John",
							"age":  int64(31),
						},
						{
							"name": "Bob",
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
								"_docID": "bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f",
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
							"author_id":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f",
							"name":"Game of chains"
						},
						{
							"_docID":"bae-f33a7110-fb6f-57aa-9501-df0111427315",
							"_docIDNew":"bae-c9c1a385-afce-5ef7-8b98-9369b157fd97",
							"author_id":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f",
							"name":"John and the sourcerers' stone"
						}
					],
					"User":[
						{
							"_docID":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f",
							"_docIDNew":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f",
							"age":31,
							"name":"Bob"
						},
						{
							"_docID":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d",
							"_docIDNew":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f",
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
			testUtils.SchemaUpdate{
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
				ImportContent: `{"Book":[{"_docID":"bae-236c14bd-4621-5d43-bc03-4442f3b8719e","_docIDNew":"bae-6dbb3738-d3db-5121-acee-6fbdd97ff7a8","author_id":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f","favourite_id":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f","name":"John and the sourcerers' stone"},{"_docID":"bae-ffba7007-d4d4-5630-be53-d66f56da57fd","_docIDNew":"bae-ffba7007-d4d4-5630-be53-d66f56da57fd","name":"Game of chains"}],"User":[{"_docID":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f","_docIDNew":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f","age":31,"name":"Bob"},{"_docID":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","_docIDNew":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f","age":31,"name":"John"}]}`,
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
							"name": "John and the sourcerers' stone",
							"author": map[string]any{
								"name": "John",
								"favouriteBook": map[string]any{
									"name": "John and the sourcerers' stone",
								},
							},
						},
						{
							"name":   "Game of chains",
							"author": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
