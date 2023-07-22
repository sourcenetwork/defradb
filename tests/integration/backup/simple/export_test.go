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

func TestBackupExport_Simple_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "John", "age": 30}`,
			},
			testUtils.BackupExport{
				ExpectedContent: `{"User":[{"_key":"bae-e933420a-988a-56f8-8952-6c245aebd519","_newKey":"bae-e933420a-988a-56f8-8952-6c245aebd519","age":30,"name":"John"}]}`,
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupExport_WithInvalidFilePath_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "John", "age": 30}`,
			},
			testUtils.BackupExport{
				Config: client.BackupConfig{
					Filepath: t.TempDir() + "/some/test.json",
				},
				ExpectedError: "no such file or directory",
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupExport_WithInvalidCollection_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "John", "age": 30}`,
			},
			testUtils.BackupExport{
				Config: client.BackupConfig{
					Collections: []string{"Invalid"},
				},
				ExpectedError: "failed to get collection: datastore: key not found. Name: Invalid",
			},
		},
	}

	executeTestCase(t, test)
}

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
				ExpectedContent: `{"User":[{"_key":"bae-e933420a-988a-56f8-8952-6c245aebd519","_newKey":"bae-e933420a-988a-56f8-8952-6c245aebd519","age":30,"name":"John"}]}`,
			},
		},
	}

	executeTestCase(t, test)
}
