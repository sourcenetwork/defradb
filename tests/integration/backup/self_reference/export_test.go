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
				Doc:          `{"name": "Bob", "age": 31, "boss": "bae-410a76c8-982f-5898-a509-b9e24bea4820"}`,
			},
			testUtils.BackupExport{
				Config: client.BackupConfig{
					Collections: []string{"User"},
				},
				ExpectedContent: `{"User":[{"_docID":"bae-410a76c8-982f-5898-a509-b9e24bea4820","_docIDNew":"bae-410a76c8-982f-5898-a509-b9e24bea4820","age":30,"name":"John"},{"_docID":"bae-e6b09a7a-47e9-5fbb-9cdc-638bf7bd1878","_docIDNew":"bae-e6b09a7a-47e9-5fbb-9cdc-638bf7bd1878","age":31,"boss_id":"bae-410a76c8-982f-5898-a509-b9e24bea4820","name":"Bob"}]}`,
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
				Doc:          `{"name": "Bob", "age": 31, "boss": "bae-410a76c8-982f-5898-a509-b9e24bea4820"}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc:          `{"age": 31}`,
			},
			testUtils.BackupExport{
				ExpectedContent: `{"User":[{"_docID":"bae-410a76c8-982f-5898-a509-b9e24bea4820","_docIDNew":"bae-73df44a4-8ac0-507e-bd76-87813298b503","age":31,"name":"John"},{"_docID":"bae-e6b09a7a-47e9-5fbb-9cdc-638bf7bd1878","_docIDNew":"bae-ad0c5cf1-8fbf-5616-82de-f65b7a6b756d","age":31,"boss_id":"bae-73df44a4-8ac0-507e-bd76-87813298b503","name":"Bob"}]}`,
			},
		},
	}

	executeTestCase(t, test)
}
