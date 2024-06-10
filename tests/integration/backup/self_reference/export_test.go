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
				Doc:          `{"name": "Bob", "age": 31, "boss": "bae-a2162ff0-3257-50f1-ba2f-39c299921220"}`,
			},
			testUtils.BackupExport{
				Config: client.BackupConfig{
					Collections: []string{"User"},
				},
				ExpectedContent: `{"User":[{"_docID":"bae-a2162ff0-3257-50f1-ba2f-39c299921220","_docIDNew":"bae-a2162ff0-3257-50f1-ba2f-39c299921220","age":30,"name":"John"},{"_docID":"bae-f4def2b3-2fe8-5e3b-838e-b9d9f8aca102","_docIDNew":"bae-f4def2b3-2fe8-5e3b-838e-b9d9f8aca102","age":31,"boss_id":"bae-a2162ff0-3257-50f1-ba2f-39c299921220","name":"Bob"}]}`,
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
				Doc:          `{"name": "Bob", "age": 31, "boss": "bae-a2162ff0-3257-50f1-ba2f-39c299921220"}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc:          `{"age": 31}`,
			},
			testUtils.BackupExport{
				ExpectedContent: `{"User":[{"_docID":"bae-a2162ff0-3257-50f1-ba2f-39c299921220","_docIDNew":"bae-99fbc678-167f-5325-bdf1-79fa76039125","age":31,"name":"John"},{"_docID":"bae-f4def2b3-2fe8-5e3b-838e-b9d9f8aca102","_docIDNew":"bae-98531af8-dda5-5993-b140-1495fa8f1576","age":31,"boss_id":"bae-99fbc678-167f-5325-bdf1-79fa76039125","name":"Bob"}]}`,
			},
		},
	}

	executeTestCase(t, test)
}
