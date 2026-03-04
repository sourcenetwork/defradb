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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestBackupImport_WithMultipleNoKeyAndMultipleCollections_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.ImportBackup{
				ImportContent: `{
					"User":[
						{"age":30,"name":"John"},
						{"age":31,"name":"Smith"},
						{"age":32,"name":"Bob"}
					],
					"Book":[
						{"name":"John and the sourcerers' stone"},
						{"name":"Game of chains"}
					]
				}`,
			},
			&action.Request{
				Request: `
					query  {
						User {
							name
							age
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Smith",
							"age":  int64(31),
						},
						{
							"name": "John",
							"age":  int64(30),
						},
						{
							"name": "Bob",
							"age":  int64(32),
						},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				Request: `
					query  {
						Book {
							name
						}
					}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name": "Game of chains",
						},
						{
							"name": "John and the sourcerers' stone",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupImport_WithMultipleNoKeyAndMultipleCollectionsAndUpdatedDocs_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.ImportBackup{
				ImportContent: `{
					"Book":[
						{
							"_docID":"bae-9828df35-b4cd-5d3a-acab-193c84e521c6",
							"_docIDNew":"bae-d7b5bc04-26af-570f-9aec-b9c5d923842f",
							"_authorID":"bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4",
							"name":"Game of chains"
						},
						{
							"_docID":"bae-0fa157eb-c762-51af-859d-9d0eb941d2f4",
							"_docIDNew":"bae-8507cb9a-54ea-5db3-bb38-6b4e6e8f3dbf",
							"_authorID":"bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4",
							"name":"John and the sourcerers' stone"
						}
					],
					"User":[
						{
							"_docID":"bae-be327e0b-a7fa-53ce-b29a-919cce5b5120",
							"_docIDNew":"bae-be327e0b-a7fa-53ce-b29a-919cce5b5120",
							"age":31,
							"name":"Bob"
						},
						{
							"_docID":"bae-3fc941b7-505c-5ce2-91a0-b180930ec8a9",
							"_docIDNew":"bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4",
							"age":31,
							"name":"John"
						}
					]
				}`,
			},
			&action.Request{
				Request: `
					query  {
						User {
							name
							age
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Bob",
							"age":  int64(31),
						},
						{
							"name": "John",
							"age":  int64(31),
						},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				Request: `
					query  {
						Book {
							name
							author {
								_docID
							}
						}
					}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name": "John and the sourcerers' stone",
							"author": map[string]any{
								"_docID": "bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4",
							},
						},
						{
							"name": "Game of chains",
							"author": map[string]any{
								"_docID": "bae-1552bcf5-6b3b-5cd0-bdaf-33bb43f74ab4",
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}
