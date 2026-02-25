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
			&action.AddDoc{
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
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "John", "age": 30}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob", "age": 31}`,
			},
			&action.AddDoc{
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
				ExpectedContent: `{"Book":[{"_authorID":"bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4","_docID":"bae-a7dc8647-e224-5abc-a0df-0fe2d380c7a7","_docIDNew":"bae-5048dc2a-683b-5ff4-a4a6-8d25f01df2a3","name":"John and the sourcerers' stone"}],"User":[{"_docID":"bae-3fc941b7-505c-5ce2-91a0-b180930ec8a9","_docIDNew":"bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4","age":31,"name":"John"},{"_docID":"bae-be327e0b-a7fa-53ce-b29a-919cce5b5120","_docIDNew":"bae-be327e0b-a7fa-53ce-b29a-919cce5b5120","age":31,"name":"Bob"}]}`,
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
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "John", "age": 30}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob", "age": 31}`,
			},
			&action.AddDoc{
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
				ExpectedContent: `{"User":[{"_docID":"bae-3fc941b7-505c-5ce2-91a0-b180930ec8a9","_docIDNew":"bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4","age":31,"name":"John"},{"_docID":"bae-be327e0b-a7fa-53ce-b29a-919cce5b5120","_docIDNew":"bae-be327e0b-a7fa-53ce-b29a-919cce5b5120","age":31,"name":"Bob"}],"Book":[{"_authorID":"bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4","_docID":"bae-64c6f714-5379-5aa0-be61-14c5fbbc2900","_docIDNew":"bae-041e43b9-93dd-5884-8075-f35b42f827ed","_favouriteID":"bae-be327e0b-a7fa-53ce-b29a-919cce5b5120","name":"John and the sourcerers' stone"}]}`,
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
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "John", "age": 30}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob", "age": 31}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "John and the sourcerers' stone",
					"author":    testUtils.NewDocIndex(0, 0),
					"favourite": testUtils.NewDocIndex(0, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc:          `{"name": "Game of chains"}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc:          `{"age": 31}`,
			},
			testUtils.BackupExport{
				ExpectedContent: `{"User":[{"_docID":"bae-3fc941b7-505c-5ce2-91a0-b180930ec8a9","_docIDNew":"bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4","age":31,"name":"John"},{"_docID":"bae-be327e0b-a7fa-53ce-b29a-919cce5b5120","_docIDNew":"bae-be327e0b-a7fa-53ce-b29a-919cce5b5120","age":31,"name":"Bob"}],"Book":[{"_docID":"bae-01e3546b-088f-5da6-b345-25cb5858f90b","_docIDNew":"bae-01e3546b-088f-5da6-b345-25cb5858f90b","name":"Game of chains"},{"_authorID":"bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4","_docID":"bae-64c6f714-5379-5aa0-be61-14c5fbbc2900","_docIDNew":"bae-041e43b9-93dd-5884-8075-f35b42f827ed","_favouriteID":"bae-be327e0b-a7fa-53ce-b29a-919cce5b5120","name":"John and the sourcerers' stone"}]}`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
