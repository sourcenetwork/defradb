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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestBackupImport_WithMultipleNoKeyAndMultipleCollections_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.BackupImport{
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
			testUtils.Request{
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
			},
			testUtils.Request{
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
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupImport_WithMultipleNoKeyAndMultipleCollectionsAndUpdatedDocs_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.BackupImport{
				ImportContent: `{
					"Book":[
						{
							"_docID":"bae-9828df35-b4cd-5d3a-acab-193c84e521c6",
							"_docIDNew":"bae-d7b5bc04-26af-570f-9aec-b9c5d923842f",
							"author_id":"bae-97f27fca-8b97-59f1-afa1-2e63140de933",
							"name":"Game of chains"
						},
						{
							"_docID":"bae-0fa157eb-c762-51af-859d-9d0eb941d2f4",
							"_docIDNew":"bae-8507cb9a-54ea-5db3-bb38-6b4e6e8f3dbf",
							"author_id":"bae-97f27fca-8b97-59f1-afa1-2e63140de933",
							"name":"John and the sourcerers' stone"
						}
					],
					"User":[
						{
							"_docID":"bae-88fea952-a678-5e05-9895-8a86ac6abc3b",
							"_docIDNew":"bae-88fea952-a678-5e05-9895-8a86ac6abc3b",
							"age":31,
							"name":"Bob"
						},
						{
							"_docID":"bae-a911f9cc-217a-58a3-a2f4-96548197403e",
							"_docIDNew":"bae-97f27fca-8b97-59f1-afa1-2e63140de933",
							"age":31,
							"name":"John"
						}
					]
				}`,
			},
			testUtils.Request{
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
			},
			testUtils.Request{
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
								"_docID": "bae-97f27fca-8b97-59f1-afa1-2e63140de933",
							},
						},
						{
							"name": "Game of chains",
							"author": map[string]any{
								"_docID": "bae-97f27fca-8b97-59f1-afa1-2e63140de933",
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
