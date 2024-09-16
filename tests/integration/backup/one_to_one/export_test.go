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

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestBackupExport_JustUserCollection_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "John", "age": 30}`,
			},
			testUtils.BackupExport{
				Config: client.BackupConfig{
					Collections: []string{"User"},
				},
				ExpectedContent: `{"User":[{"_docID":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","_docIDNew":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","age":30,"name":"John"}]}`,
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupExport_AllCollectionsMultipleDocsAndDocUpdate_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "John", "age": 30}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob", "age": 31}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":   "John and the sourcerers' stone",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc:          `{"age": 31}`,
			},
			testUtils.BackupExport{
				ExpectedContent: `{"User":[{"_docID":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","_docIDNew":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f","age":31,"name":"John"},{"_docID":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f","_docIDNew":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f","age":31,"name":"Bob"}],"Book":[{"_docID":"bae-af59fdc4-e495-5fd3-a9a6-386249aafdbb","_docIDNew":"bae-d374c406-c6ea-51cd-9e9b-dd44a97b499c","author_id":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f","name":"John and the sourcerers' stone"}]}`,
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupExport_DoubleReletionship_NoError(t *testing.T) {
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
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "John", "age": 30}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob", "age": 31}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "John and the sourcerers' stone",
					"author":    testUtils.NewDocIndex(0, 0),
					"favourite": testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc:          `{"age": 31}`,
			},
			testUtils.BackupExport{
				ExpectedContent: `{"User":[{"_docID":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","_docIDNew":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f","age":31,"name":"John"},{"_docID":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f","_docIDNew":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f","age":31,"name":"Bob"}],"Book":[{"_docID":"bae-bddb7139-7035-5fff-a118-3fc2033723b3","_docIDNew":"bae-6972e51f-a8cd-59eb-9a34-8a37058ddf4e","author_id":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f","favourite_id":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f","name":"John and the sourcerers' stone"}]}`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestBackupExport_DoubleReletionshipWithUpdate_NoError(t *testing.T) {
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
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "John", "age": 30}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob", "age": 31}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "John and the sourcerers' stone",
					"author":    testUtils.NewDocIndex(0, 0),
					"favourite": testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc:          `{"name": "Game of chains"}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc:          `{"age": 31}`,
			},
			testUtils.BackupExport{
				ExpectedContent: `{"User":[{"_docID":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","_docIDNew":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f","age":31,"name":"John"},{"_docID":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f","_docIDNew":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f","age":31,"name":"Bob"}],"Book":[{"_docID":"bae-0aa10275-4f6e-5b38-9915-5664dd4c7802","_docIDNew":"bae-0aa10275-4f6e-5b38-9915-5664dd4c7802","name":"Game of chains"},{"_docID":"bae-bddb7139-7035-5fff-a118-3fc2033723b3","_docIDNew":"bae-6972e51f-a8cd-59eb-9a34-8a37058ddf4e","author_id":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f","favourite_id":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f","name":"John and the sourcerers' stone"}]}`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
