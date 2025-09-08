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
				ExpectedContent: `{"User":[{"_docID":"bae-a911f9cc-217a-58a3-a2f4-96548197403e","_docIDNew":"bae-a911f9cc-217a-58a3-a2f4-96548197403e","age":30,"name":"John"}]}`,
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
				ExpectedContent: `{"User":[{"_docID":"bae-88fea952-a678-5e05-9895-8a86ac6abc3b","_docIDNew":"bae-88fea952-a678-5e05-9895-8a86ac6abc3b","age":31,"name":"Bob"},{"_docID":"bae-a911f9cc-217a-58a3-a2f4-96548197403e","_docIDNew":"bae-97f27fca-8b97-59f1-afa1-2e63140de933","age":31,"name":"John"}],"Book":[{"_docID":"bae-f97cb90a-20db-5595-b193-89bdf50bdee8","_docIDNew":"bae-8a319b41-e061-5d19-a847-388fa51f732c","author_id":"bae-97f27fca-8b97-59f1-afa1-2e63140de933","name":"John and the sourcerers' stone"}]}`,
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupExport_AllCollectionsMultipleDocsAndMultipleDocUpdate_NoError(t *testing.T) {
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
				DocMap: map[string]any{
					"name":   "John and the sourcerers' stone",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":   "Game of chains",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc:          `{"age": 31}`,
			},
			testUtils.BackupExport{
				ExpectedContent: `{"User":[{"_docID":"bae-88fea952-a678-5e05-9895-8a86ac6abc3b","_docIDNew":"bae-88fea952-a678-5e05-9895-8a86ac6abc3b","age":31,"name":"Bob"},{"_docID":"bae-a911f9cc-217a-58a3-a2f4-96548197403e","_docIDNew":"bae-97f27fca-8b97-59f1-afa1-2e63140de933","age":31,"name":"John"}],"Book":[{"_docID":"bae-191238ef-acd2-5d9f-8d95-dcc15415fc75","_docIDNew":"bae-b5ea3d12-0519-5a5f-a7bc-9f5fee62c12f","author_id":"bae-97f27fca-8b97-59f1-afa1-2e63140de933","name":"Game of chains"},{"_docID":"bae-f97cb90a-20db-5595-b193-89bdf50bdee8","_docIDNew":"bae-8a319b41-e061-5d19-a847-388fa51f732c","author_id":"bae-97f27fca-8b97-59f1-afa1-2e63140de933","name":"John and the sourcerers' stone"}]}`,
			},
		},
	}

	executeTestCase(t, test)
}
