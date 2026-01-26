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

func TestBackupExport_JustUserCollection_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "John", "age": 30}`,
			},
			testUtils.BackupExport{
				Config: client.BackupConfig{
					Collections: []string{"User"},
				},
				ExpectedContent: `{"User":[{"_docID":"bae-3fc941b7-505c-5ce2-91a0-b180930ec8a9","_docIDNew":"bae-3fc941b7-505c-5ce2-91a0-b180930ec8a9","age":30,"name":"John"}]}`,
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupExport_AllCollectionsMultipleDocsAndDocUpdate_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "John", "age": 30}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob", "age": 31}`,
			},
			&action.CreateDoc{
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
				ExpectedContent: `{"Book":[{"_authorID":"bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4","_docID":"bae-a7dc8647-e224-5abc-a0df-0fe2d380c7a7","_docIDNew":"bae-5048dc2a-683b-5ff4-a4a6-8d25f01df2a3","name":"John and the sourcerers' stone"}],"User":[{"_docID":"bae-3fc941b7-505c-5ce2-91a0-b180930ec8a9","_docIDNew":"bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4","age":31,"name":"John"},{"_docID":"bae-be327e0b-a7fa-53ce-b29a-919cce5b5120","_docIDNew":"bae-be327e0b-a7fa-53ce-b29a-919cce5b5120","age":31,"name":"Bob"}]}`,
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupExport_AllCollectionsMultipleDocsAndMultipleDocUpdate_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "John", "age": 30}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob", "age": 31}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":   "John and the sourcerers' stone",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
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
				ExpectedContent: `{"Book":[{"_authorID":"bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4","_docID":"bae-a7dc8647-e224-5abc-a0df-0fe2d380c7a7","_docIDNew":"bae-5048dc2a-683b-5ff4-a4a6-8d25f01df2a3","name":"John and the sourcerers' stone"},{"_authorID":"bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4","_docID":"bae-db0363ee-4415-526f-8a49-97b7594d39f6","_docIDNew":"bae-bca6a77e-e915-51f4-b024-ecd5bc71ff20","name":"Game of chains"}],"User":[{"_docID":"bae-3fc941b7-505c-5ce2-91a0-b180930ec8a9","_docIDNew":"bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4","age":31,"name":"John"},{"_docID":"bae-be327e0b-a7fa-53ce-b29a-919cce5b5120","_docIDNew":"bae-be327e0b-a7fa-53ce-b29a-919cce5b5120","age":31,"name":"Bob"}]}`,
			},
		},
	}

	executeTestCase(t, test)
}
