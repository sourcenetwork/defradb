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
				ExpectedContent: `{"User":[{"_docID":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","_docIDNew":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f","age":31,"name":"John"},{"_docID":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f","_docIDNew":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f","age":31,"name":"Bob"}],"Book":[{"_docID":"bae-f33a7110-fb6f-57aa-9501-df0111427315","_docIDNew":"bae-c9c1a385-afce-5ef7-8b98-9369b157fd97","author_id":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f","name":"John and the sourcerers' stone"}]}`,
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
				ExpectedContent: `{"User":[{"_docID":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","_docIDNew":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f","age":31,"name":"John"},{"_docID":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f","_docIDNew":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f","age":31,"name":"Bob"}],"Book":[{"_docID":"bae-ccf9da82-8ed6-5133-b64f-558c21bc8dfd","_docIDNew":"bae-27ae099a-fa7d-5a66-a919-6c3b0322d17c","author_id":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f","favourite_id":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f","name":"John and the sourcerers' stone"}]}`,
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
				ExpectedContent: `{"User":[{"_docID":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","_docIDNew":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f","age":31,"name":"John"},{"_docID":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f","_docIDNew":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f","age":31,"name":"Bob"}],"Book":[{"_docID":"bae-ccf9da82-8ed6-5133-b64f-558c21bc8dfd","_docIDNew":"bae-27ae099a-fa7d-5a66-a919-6c3b0322d17c","author_id":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f","favourite_id":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f","name":"John and the sourcerers' stone"},{"_docID":"bae-ffba7007-d4d4-5630-be53-d66f56da57fd","_docIDNew":"bae-ffba7007-d4d4-5630-be53-d66f56da57fd","name":"Game of chains"}]}`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
