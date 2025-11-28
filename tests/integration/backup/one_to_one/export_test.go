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
				ExpectedContent: `{"User":[{"_docID":"bae-3fc941b7-505c-5ce2-91a0-b180930ec8a9","_docIDNew":"bae-3fc941b7-505c-5ce2-91a0-b180930ec8a9","age":30,"name":"John"}]}`,
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
				ExpectedContent: `{"Book":[{"_docID":"bae-49229a73-9634-558d-9cad-2392f9b7dab5","_docIDNew":"bae-80133f4e-aee1-56c0-a4e3-e145af32aed1","author_id":"bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4","name":"John and the sourcerers' stone"}],"User":[{"_docID":"bae-3fc941b7-505c-5ce2-91a0-b180930ec8a9","_docIDNew":"bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4","age":31,"name":"John"},{"_docID":"bae-be327e0b-a7fa-53ce-b29a-919cce5b5120","_docIDNew":"bae-be327e0b-a7fa-53ce-b29a-919cce5b5120","age":31,"name":"Bob"}]}`,
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
				ExpectedContent: `{"Book":[{"_docID":"bae-d336efaf-171a-596c-bd0a-80208e5e4576","_docIDNew":"bae-06584acb-65a5-52f2-b71d-3ddb5354d7e3","author_id":"bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4","favourite_id":"bae-be327e0b-a7fa-53ce-b29a-919cce5b5120","name":"John and the sourcerers' stone"}],"User":[{"_docID":"bae-3fc941b7-505c-5ce2-91a0-b180930ec8a9","_docIDNew":"bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4","age":31,"name":"John"},{"_docID":"bae-be327e0b-a7fa-53ce-b29a-919cce5b5120","_docIDNew":"bae-be327e0b-a7fa-53ce-b29a-919cce5b5120","age":31,"name":"Bob"}]}`,
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
				ExpectedContent: `{"Book":[{"_docID":"bae-ce5806b0-e773-5135-8acd-090c68bc1c38","_docIDNew":"bae-ce5806b0-e773-5135-8acd-090c68bc1c38","name":"Game of chains"},{"_docID":"bae-d336efaf-171a-596c-bd0a-80208e5e4576","_docIDNew":"bae-06584acb-65a5-52f2-b71d-3ddb5354d7e3","author_id":"bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4","favourite_id":"bae-be327e0b-a7fa-53ce-b29a-919cce5b5120","name":"John and the sourcerers' stone"}],"User":[{"_docID":"bae-3fc941b7-505c-5ce2-91a0-b180930ec8a9","_docIDNew":"bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4","age":31,"name":"John"},{"_docID":"bae-be327e0b-a7fa-53ce-b29a-919cce5b5120","_docIDNew":"bae-be327e0b-a7fa-53ce-b29a-919cce5b5120","age":31,"name":"Bob"}]}`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
