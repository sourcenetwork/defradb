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
				Results: []map[string]any{
					{
						"name": "Smith",
						"age":  int64(31),
					},
					{
						"name": "Bob",
						"age":  int64(32),
					},
					{
						"name": "John",
						"age":  int64(30),
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
				Results: []map[string]any{
					{
						"name": "John and the sourcerers' stone",
					},
					{
						"name": "Game of chains",
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
							"_docID":"bae-4399f189-138d-5d49-9e25-82e78463677b",
							"_docIDNew":"bae-78a40f28-a4b8-5dca-be44-392b0f96d0ff",
							"author_id":"bae-807ea028-6c13-5f86-a72b-46e8b715a162",
							"name":"Game of chains"
						},
						{
							"_docID":"bae-5cf2fec3-d8ed-50d5-8286-39109853d2da",
							"_docIDNew":"bae-edeade01-2d21-5d6d-aadf-efc5a5279de5",
							"author_id":"bae-807ea028-6c13-5f86-a72b-46e8b715a162",
							"name":"John and the sourcerers' stone"
						}
					],
					"User":[
						{
							"_docID":"bae-0648f44e-74e8-593b-a662-3310ec278927",
							"_docIDNew":"bae-0648f44e-74e8-593b-a662-3310ec278927",
							"age":31,
							"name":"Bob"
						},
						{
							"_docID":"bae-e933420a-988a-56f8-8952-6c245aebd519",
							"_docIDNew":"bae-807ea028-6c13-5f86-a72b-46e8b715a162",
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
				Results: []map[string]any{
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
				Results: []map[string]any{
					{
						"name": "Game of chains",
						"author": map[string]any{
							"_docID": "bae-807ea028-6c13-5f86-a72b-46e8b715a162",
						},
					},
					{
						"name": "John and the sourcerers' stone",
						"author": map[string]any{
							"_docID": "bae-807ea028-6c13-5f86-a72b-46e8b715a162",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
