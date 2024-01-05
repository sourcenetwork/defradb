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
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob", "age": 31, "boss_id": "bae-e933420a-988a-56f8-8952-6c245aebd519"}`,
			},
			testUtils.BackupExport{
				Config: client.BackupConfig{
					Collections: []string{"User"},
				},
				ExpectedContent: `{"User":[{"_docID":"bae-790e7e49-f2e3-5ad6-83d9-5dfb6d8ba81d","_docIDNew":"bae-790e7e49-f2e3-5ad6-83d9-5dfb6d8ba81d","age":31,"boss_id":"bae-e933420a-988a-56f8-8952-6c245aebd519","name":"Bob"},{"_docID":"bae-e933420a-988a-56f8-8952-6c245aebd519","_docIDNew":"bae-e933420a-988a-56f8-8952-6c245aebd519","age":30,"name":"John"}]}`,
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupExport_MultipleDocsAndDocUpdate_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "John", "age": 30}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob", "age": 31, "boss_id": "bae-e933420a-988a-56f8-8952-6c245aebd519"}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc:          `{"age": 31}`,
			},
			testUtils.BackupExport{
				ExpectedContent: `{"User":[{"_docID":"bae-790e7e49-f2e3-5ad6-83d9-5dfb6d8ba81d","_docIDNew":"bae-067fd15e-32a1-5681-8f41-c423f563e21b","age":31,"boss_id":"bae-807ea028-6c13-5f86-a72b-46e8b715a162","name":"Bob"},{"_docID":"bae-e933420a-988a-56f8-8952-6c245aebd519","_docIDNew":"bae-807ea028-6c13-5f86-a72b-46e8b715a162","age":31,"name":"John"}]}`,
			},
		},
	}

	executeTestCase(t, test)
}
