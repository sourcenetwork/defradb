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
				ExpectedContent: `{"User":[{"_docID":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","_docIDNew":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","age":30,"name":"John"}]}`,
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
				ExpectedContent: `{"User":[{"_docID":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","_docIDNew":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f","age":31,"name":"John"},{"_docID":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f","_docIDNew":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f","age":31,"name":"Bob"}],"Book":[{"_docID":"bae-45c92e9c-4d31-5e96-8bd7-3d532734e117","_docIDNew":"bae-3ca9a4c3-6240-5e86-a00f-9590d2f2ecf3","author_id":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f","name":"John and the sourcerers' stone"}]}`,
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
				ExpectedContent: `{"User":[{"_docID":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","_docIDNew":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f","age":31,"name":"John"},{"_docID":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f","_docIDNew":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f","age":31,"name":"Bob"}],"Book":[{"_docID":"bae-45c92e9c-4d31-5e96-8bd7-3d532734e117","_docIDNew":"bae-3ca9a4c3-6240-5e86-a00f-9590d2f2ecf3","author_id":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f","name":"John and the sourcerers' stone"},{"_docID":"bae-8fc3d148-869b-5629-ae22-5423c73f709b","_docIDNew":"bae-33c136bd-4375-54a0-81ff-54ca560c7bb8","author_id":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f","name":"Game of chains"}]}`,
			},
		},
	}

	executeTestCase(t, test)
}
