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
	"github.com/sourcenetwork/defradb/tests/action"
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
				ExpectedContent: `{"User":[{"_docID":"bae-a911f9cc-217a-58a3-a2f4-96548197403e","_docIDNew":"bae-a911f9cc-217a-58a3-a2f4-96548197403e","age":30,"name":"John"}]}`,
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
				ExpectedContent: `{"User":[{"_docID":"bae-88fea952-a678-5e05-9895-8a86ac6abc3b","_docIDNew":"bae-88fea952-a678-5e05-9895-8a86ac6abc3b","age":31,"name":"Bob"},{"_docID":"bae-a911f9cc-217a-58a3-a2f4-96548197403e","_docIDNew":"bae-97f27fca-8b97-59f1-afa1-2e63140de933","age":31,"name":"John"}],"Book":[{"_docID":"bae-f97cb90a-20db-5595-b193-89bdf50bdee8","_docIDNew":"bae-8a319b41-e061-5d19-a847-388fa51f732c","author_id":"bae-97f27fca-8b97-59f1-afa1-2e63140de933","name":"John and the sourcerers' stone"}]}`,
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupExport_DoubleReletionship_NoError(t *testing.T) {
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
				ExpectedContent: `{"User":[{"_docID":"bae-88fea952-a678-5e05-9895-8a86ac6abc3b","_docIDNew":"bae-88fea952-a678-5e05-9895-8a86ac6abc3b","age":31,"name":"Bob"},{"_docID":"bae-a911f9cc-217a-58a3-a2f4-96548197403e","_docIDNew":"bae-97f27fca-8b97-59f1-afa1-2e63140de933","age":31,"name":"John"}],"Book":[{"_docID":"bae-556ece21-bf45-5652-8f32-c8a40373e8b5","_docIDNew":"bae-ccdd6d22-7339-5978-b0cb-f25d3d95c06d","author_id":"bae-97f27fca-8b97-59f1-afa1-2e63140de933","favourite_id":"bae-88fea952-a678-5e05-9895-8a86ac6abc3b","name":"John and the sourcerers' stone"}]}`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestBackupExport_DoubleReletionshipWithUpdate_NoError(t *testing.T) {
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
				ExpectedContent: `{"User":[{"_docID":"bae-88fea952-a678-5e05-9895-8a86ac6abc3b","_docIDNew":"bae-88fea952-a678-5e05-9895-8a86ac6abc3b","age":31,"name":"Bob"},{"_docID":"bae-a911f9cc-217a-58a3-a2f4-96548197403e","_docIDNew":"bae-97f27fca-8b97-59f1-afa1-2e63140de933","age":31,"name":"John"}],"Book":[{"_docID":"bae-4cb9a1d2-eef3-564d-8695-1ce61c596e5a","_docIDNew":"bae-4cb9a1d2-eef3-564d-8695-1ce61c596e5a","name":"Game of chains"},{"_docID":"bae-556ece21-bf45-5652-8f32-c8a40373e8b5","_docIDNew":"bae-ccdd6d22-7339-5978-b0cb-f25d3d95c06d","author_id":"bae-97f27fca-8b97-59f1-afa1-2e63140de933","favourite_id":"bae-88fea952-a678-5e05-9895-8a86ac6abc3b","name":"John and the sourcerers' stone"}]}`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
