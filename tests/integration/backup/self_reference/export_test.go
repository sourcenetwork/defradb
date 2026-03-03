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

func TestBackupExport_Simple_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "John", "age": 30}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob", "age": 31, "boss": "bae-1635f80b-612a-5378-a185-cad7a3018354"}`,
			},
			testUtils.ExportBackup{
				Config: client.BackupConfig{
					Collections: []string{"User"},
				},
				ExpectedContent: `{"User":[{"_docID":"bae-1635f80b-612a-5378-a185-cad7a3018354","_docIDNew":"bae-1635f80b-612a-5378-a185-cad7a3018354","age":30,"name":"John"},{"_bossID":"bae-1635f80b-612a-5378-a185-cad7a3018354","_docID":"bae-692a9178-a258-5224-990f-9ad703a2bbea","_docIDNew":"bae-692a9178-a258-5224-990f-9ad703a2bbea","age":31,"name":"Bob"}]}`,
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupExport_MultipleDocsAndDocUpdate_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "John", "age": 30}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob", "age": 31, "boss": "bae-1635f80b-612a-5378-a185-cad7a3018354"}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc:          `{"age": 31}`,
			},
			testUtils.ExportBackup{
				ExpectedContent: `{"User":[{"_docID":"bae-1635f80b-612a-5378-a185-cad7a3018354","_docIDNew":"bae-32c15b83-186c-565f-be06-caa21431c38b","age":31,"name":"John"},{"_bossID":"bae-32c15b83-186c-565f-be06-caa21431c38b","_docID":"bae-692a9178-a258-5224-990f-9ad703a2bbea","_docIDNew":"bae-69f87811-246f-5203-ae83-ff043c6fce10","age":31,"name":"Bob"}]}`,
			},
		},
	}

	executeTestCase(t, test)
}
