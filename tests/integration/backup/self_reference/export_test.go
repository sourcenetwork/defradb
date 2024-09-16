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
				Doc:          `{"name": "Bob", "age": 31, "boss": "bae-8096e3d7-41ab-5afe-ad88-481150483db1"}`,
			},
			testUtils.BackupExport{
				Config: client.BackupConfig{
					Collections: []string{"User"},
				},
				ExpectedContent: `{"User":[{"_docID":"bae-0dfbaf9f-3c58-5133-aa07-a9f25d792f4e","_docIDNew":"bae-0dfbaf9f-3c58-5133-aa07-a9f25d792f4e","age":31,"boss_id":"bae-8096e3d7-41ab-5afe-ad88-481150483db1","name":"Bob"},{"_docID":"bae-8096e3d7-41ab-5afe-ad88-481150483db1","_docIDNew":"bae-8096e3d7-41ab-5afe-ad88-481150483db1","age":30,"name":"John"}]}`,
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
				Doc:          `{"name": "Bob", "age": 31, "boss": "bae-8096e3d7-41ab-5afe-ad88-481150483db1"}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc:          `{"age": 31}`,
			},
			testUtils.BackupExport{
				ExpectedContent: `{"User":[{"_docID":"bae-0dfbaf9f-3c58-5133-aa07-a9f25d792f4e","_docIDNew":"bae-f3c5fc81-300f-5dd7-aeb3-20dd15883930","age":31,"boss_id":"bae-e4423c73-b867-511b-a5f1-565bd87d9c53","name":"Bob"},{"_docID":"bae-8096e3d7-41ab-5afe-ad88-481150483db1","_docIDNew":"bae-e4423c73-b867-511b-a5f1-565bd87d9c53","age":31,"name":"John"}]}`,
			},
		},
	}

	executeTestCase(t, test)
}
