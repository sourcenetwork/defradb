// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

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
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "John", "age": 30}`,
			},
			testUtils.ExportBackup{
				Config: client.BackupConfig{
					Collections: []string{"User"},
				},
				ExpectedContent: `{"User":[{"_docID":"{{.DocID0_0}}","_docIDNew":"{{.DocID0_0}}","age":30,"name":"John"}]}`,
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupExport_AllCollectionsMultipleDocsAndDocUpdate_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "John", "age": 30}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob", "age": 31}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":   "John and the sourcerers' stone",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc:          `{"age": 31}`,
			},
			testUtils.ExportBackup{
				ExpectedContent: `{"Book":[{"_authorID":"{{.DocID0_0}}","_docID":"{{.DocID1_0}}","_docIDNew":"{{.DocID1_0}}","name":"John and the sourcerers' stone"}],"User":[{"_docID":"{{.DocID0_0}}","_docIDNew":"{{.DocID0_0}}","age":31,"name":"John"},{"_docID":"{{.DocID0_1}}","_docIDNew":"{{.DocID0_1}}","age":31,"name":"Bob"}]}`,
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupExport_AllCollectionsMultipleDocsAndMultipleDocUpdate_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "John", "age": 30}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob", "age": 31}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":   "John and the sourcerers' stone",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":   "Game of chains",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc:          `{"age": 31}`,
			},
			testUtils.ExportBackup{
				ExpectedContent: `{"Book":[{"_authorID":"{{.DocID0_0}}","_docID":"{{.DocID1_0}}","_docIDNew":"{{.DocID1_0}}","name":"John and the sourcerers' stone"},{"_authorID":"{{.DocID0_0}}","_docID":"{{.DocID1_1}}","_docIDNew":"{{.DocID1_1}}","name":"Game of chains"}],"User":[{"_docID":"{{.DocID0_0}}","_docIDNew":"{{.DocID0_0}}","age":31,"name":"John"},{"_docID":"{{.DocID0_1}}","_docIDNew":"{{.DocID0_1}}","age":31,"name":"Bob"}]}`,
			},
		},
	}

	executeTestCase(t, test)
}
