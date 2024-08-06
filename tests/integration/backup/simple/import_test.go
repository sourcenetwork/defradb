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

func TestBackupImport_Simple_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.BackupImport{
				ImportContent: `{"User":[{"_docID":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","_docIDNew":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","age":30,"name":"John"}]}`,
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
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupImport_WithInvalidFilePath_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.BackupImport{
				Filepath:      t.TempDir() + "/some/test.json",
				ExpectedError: "failed to open file",
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupImport_WithInvalidCollection_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.BackupImport{
				ImportContent: `{"Invalid":[{"_docID":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","_docIDNew":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","age":30,"name":"John"}]}`,
				ExpectedError: "failed to get collection: datastore: key not found. Name: Invalid",
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupImport_WithDocAlreadyExists_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "John", "age": 30}`,
			},
			testUtils.BackupImport{
				ImportContent: `{"User":[{"_docID":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","_docIDNew":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","age":30,"name":"John"}]}`,
				ExpectedError: "a document with the given ID already exists",
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupImport_WithNoKeys_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.BackupImport{
				ImportContent: `{"User":[{"age":30,"name":"John"}]}`,
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
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupImport_WithMultipleNoKeys_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.BackupImport{
				ImportContent: `{"User":[
					{"age":30,"name":"John"},
					{"age":31,"name":"Smith"},
					{"age":32,"name":"Bob"}
				]}`,
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
		},
	}

	executeTestCase(t, test)
}

func TestBackupImport_EmptyObject_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.BackupImport{
				ImportContent: `{"User":[{}]}`,
			},
			testUtils.Request{
				Request: `
					query  {
						User {
							name
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": nil,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupImport_WithMultipleNoKeysAndInvalidField_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.BackupImport{
				ImportContent: `{"User":[
					{"age":30,"name":"John"},
					{"INVALID":31,"name":"Smith"},
					{"age":32,"name":"Bob"}
				]}`,
				ExpectedError: "The given field does not exist. Name: INVALID",
			},
			testUtils.Request{
				Request: `
					query  {
						User {
							name
							age
						}
					}`,
				// No documents should have been commited
				Results: map[string]any{
					"User": []map[string]any{},
				},
			},
		},
	}

	executeTestCase(t, test)
}
