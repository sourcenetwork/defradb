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
				ExpectedContent: `{"User":[{"_docID":"bae-e933420a-988a-56f8-8952-6c245aebd519","_docIDNew":"bae-e933420a-988a-56f8-8952-6c245aebd519","age":30,"name":"John"}]}`,
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
				Doc:          `{"name": "John and the sourcerers' stone", "author": "bae-e933420a-988a-56f8-8952-6c245aebd519"}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc:          `{"age": 31}`,
			},
			testUtils.BackupExport{
				ExpectedContent: `{"User":[{"_docID":"bae-0648f44e-74e8-593b-a662-3310ec278927","_docIDNew":"bae-0648f44e-74e8-593b-a662-3310ec278927","age":31,"name":"Bob"},{"_docID":"bae-e933420a-988a-56f8-8952-6c245aebd519","_docIDNew":"bae-807ea028-6c13-5f86-a72b-46e8b715a162","age":31,"name":"John"}],"Book":[{"_docID":"bae-5cf2fec3-d8ed-50d5-8286-39109853d2da","_docIDNew":"bae-edeade01-2d21-5d6d-aadf-efc5a5279de5","author_id":"bae-807ea028-6c13-5f86-a72b-46e8b715a162","name":"John and the sourcerers' stone"}]}`,
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
				Doc:          `{"name": "John and the sourcerers' stone", "author": "bae-e933420a-988a-56f8-8952-6c245aebd519", "favourite": "bae-0648f44e-74e8-593b-a662-3310ec278927"}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc:          `{"age": 31}`,
			},
			testUtils.BackupExport{
				ExpectedContent: `{"User":[{"_docID":"bae-0648f44e-74e8-593b-a662-3310ec278927","_docIDNew":"bae-0648f44e-74e8-593b-a662-3310ec278927","age":31,"name":"Bob"},{"_docID":"bae-e933420a-988a-56f8-8952-6c245aebd519","_docIDNew":"bae-807ea028-6c13-5f86-a72b-46e8b715a162","age":31,"name":"John"}],"Book":[{"_docID":"bae-45b1def4-4e63-5a93-a1b8-f7b08e682164","_docIDNew":"bae-add2ccfe-84a1-519c-ab7d-c54b43909532","author_id":"bae-807ea028-6c13-5f86-a72b-46e8b715a162","favourite_id":"bae-0648f44e-74e8-593b-a662-3310ec278927","name":"John and the sourcerers' stone"}]}`,
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
				Doc:          `{"name": "John and the sourcerers' stone", "author": "bae-e933420a-988a-56f8-8952-6c245aebd519", "favourite": "bae-0648f44e-74e8-593b-a662-3310ec278927"}`,
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
				ExpectedContent: `{"User":[{"_docID":"bae-0648f44e-74e8-593b-a662-3310ec278927","_docIDNew":"bae-0648f44e-74e8-593b-a662-3310ec278927","age":31,"name":"Bob"},{"_docID":"bae-e933420a-988a-56f8-8952-6c245aebd519","_docIDNew":"bae-807ea028-6c13-5f86-a72b-46e8b715a162","age":31,"name":"John"}],"Book":[{"_docID":"bae-45b1def4-4e63-5a93-a1b8-f7b08e682164","_docIDNew":"bae-add2ccfe-84a1-519c-ab7d-c54b43909532","author_id":"bae-807ea028-6c13-5f86-a72b-46e8b715a162","favourite_id":"bae-0648f44e-74e8-593b-a662-3310ec278927","name":"John and the sourcerers' stone"},{"_docID":"bae-da7f2d88-05c4-528a-846a-0d18ab26603b","_docIDNew":"bae-da7f2d88-05c4-528a-846a-0d18ab26603b","name":"Game of chains"}]}`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
