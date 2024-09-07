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
							"name": "John",
							"age":  int64(30),
						},
						{
							"name": "Bob",
							"age":  int64(32),
						},
						{
							"name": "Smith",
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
							"_docID":"bae-4a28c746-ccbf-5511-91a9-391036f42f80",
							"_docIDNew":"bae-d821f684-47de-5b63-b9c7-6eccec368e52",
							"author_id":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f",
							"name":"Game of chains"
						},
						{
							"_docID":"bae-8c8be5c6-d26b-50d4-9378-2acd5fe6959d",
							"_docIDNew":"bae-c94e52f8-6e91-522c-b6a6-38346a06b3d2",
							"author_id":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f",
							"name":"John and the sourcerers' stone"
						}
					],
					"User":[
						{
							"_docID":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f",
							"_docIDNew":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f",
							"age":31,
							"name":"Bob"
						},
						{
							"_docID":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d",
							"_docIDNew":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f",
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
							"name": "John",
							"age":  int64(31),
						},
						{
							"name": "Bob",
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
								"_docID": "bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f",
							},
						},
						{
							"name": "Game of chains",
							"author": map[string]any{
								"_docID": "bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f",
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
