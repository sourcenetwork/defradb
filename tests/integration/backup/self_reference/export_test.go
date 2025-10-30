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
				Doc:          `{"name": "Bob", "age": 31, "boss": "bae-59a4a7b9-1ce9-557f-bbb8-48485ee44f35"}`,
			},
			testUtils.BackupExport{
				Config: client.BackupConfig{
					Collections: []string{"User"},
				},
				ExpectedContent: `{"User":[{"_docID":"bae-59a4a7b9-1ce9-557f-bbb8-48485ee44f35","_docIDNew":"bae-59a4a7b9-1ce9-557f-bbb8-48485ee44f35","age":30,"name":"John"},{"_docID":"bae-7da3959b-0a8f-54e1-bf62-cfa35699e627","_docIDNew":"bae-7da3959b-0a8f-54e1-bf62-cfa35699e627","age":31,"boss_id":"bae-59a4a7b9-1ce9-557f-bbb8-48485ee44f35","name":"Bob"}]}`,
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
				Doc:          `{"name": "Bob", "age": 31, "boss": "bae-59a4a7b9-1ce9-557f-bbb8-48485ee44f35"}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc:          `{"age": 31}`,
			},
			testUtils.BackupExport{
				ExpectedContent: `{"User":[{"_docID":"bae-59a4a7b9-1ce9-557f-bbb8-48485ee44f35","_docIDNew":"bae-c5d4a120-9447-5f7f-8344-bb1e7c2b7e3c","age":31,"name":"John"},{"_docID":"bae-7da3959b-0a8f-54e1-bf62-cfa35699e627","_docIDNew":"bae-238d6566-b9da-5205-8dee-f4825b077213","age":31,"boss_id":"bae-c5d4a120-9447-5f7f-8344-bb1e7c2b7e3c","name":"Bob"}]}`,
			},
		},
	}

	executeTestCase(t, test)
}
