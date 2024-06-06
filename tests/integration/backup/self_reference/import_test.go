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

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestBackupSelfRefImport_Simple_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.BackupImport{
				ImportContent: `{
					"User":[
						{
							"_docID":"bae-f4def2b3-2fe8-5e3b-838e-b9d9f8aca102",
							"age":31,
							"boss_id":"bae-a2162ff0-3257-50f1-ba2f-39c299921220",
							"name":"Bob"
						},
						{
							"_docID":"bae-a2162ff0-3257-50f1-ba2f-39c299921220",
							"age":30,
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
							boss {
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"name": "John",
						"boss": nil,
					},
					{
						"name": "Bob",
						"boss": map[string]any{
							"name": "John",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupSelfRefImport_SelfRef_NoError(t *testing.T) {
	expectedExportData := `{` +
		`"User":[` +
		`{` +
		`"_docID":"bae-20631b3d-1498-51f1-be29-5c0effbfa646",` +
		`"_docIDNew":"bae-20631b3d-1498-51f1-be29-5c0effbfa646",` +
		`"age":31,` +
		`"boss_id":"bae-20631b3d-1498-51f1-be29-5c0effbfa646",` +
		`"name":"Bob"` +
		`}` +
		`]` +
		`}`
	test := testUtils.TestCase{
		Actions: []any{
			// Configure 2 nodes for this test, we will export from the first
			// and import to the second.
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: schemas,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "Bob",
					"age": 31
				}`,
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"boss_id": "bae-20631b3d-1498-51f1-be29-5c0effbfa646"
				}`,
			},
			testUtils.BackupExport{
				NodeID:          immutable.Some(0),
				ExpectedContent: expectedExportData,
			},
			testUtils.BackupImport{
				NodeID:        immutable.Some(1),
				ImportContent: expectedExportData,
			},
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `
					query  {
						User {
							name
							boss {
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Bob",
						"boss": map[string]any{
							"name": "Bob",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestBackupSelfRefImport_PrimaryRelationWithSecondCollection_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Author {
						name: String
						book: Book @relation(name: "author_book")
						reviewed: Book @relation(name: "reviewedBy_reviewed")
					}
					type Book {
						name: String
						author: Author @primary @relation(name: "author_book")
						reviewedBy: Author @primary @relation(name: "reviewedBy_reviewed")
					}
				`,
			},
			testUtils.BackupImport{
				ImportContent: `{
					"Author":[
						{
							"name":"John"
						}
					],
					"Book":[
						{
							"name":"John and the sourcerers' stone",
							"author":"bae-da91935a-9176-57ea-ba68-afe05781da16",
							"reviewedBy":"bae-da91935a-9176-57ea-ba68-afe05781da16"
						}
					]
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Book {
							name
							author {
								name
								reviewed {
									name
								}
							}
						}
					}`,
				Results: []map[string]any{
					{
						"name": "John and the sourcerers' stone",
						"author": map[string]any{
							"name": "John",
							"reviewed": map[string]any{
								"name": "John and the sourcerers' stone",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestBackupSelfRefImport_PrimaryRelationWithSecondCollectionWrongOrder_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Author {
						name: String
						book: Book @relation(name: "author_book")
						reviewed: Book @relation(name: "reviewedBy_reviewed")
					}
					type Book {
						name: String
						author: Author @primary @relation(name: "author_book")
						reviewedBy: Author @primary @relation(name: "reviewedBy_reviewed")
					}
				`,
			},
			testUtils.BackupImport{
				ImportContent: `{
					"Book":[
						{
							"name":"John and the sourcerers' stone",
							"author":"bae-da91935a-9176-57ea-ba68-afe05781da16",
							"reviewedBy":"bae-da91935a-9176-57ea-ba68-afe05781da16"
						}
					],
					"Author":[
						{
							"name":"John"
						}
					]
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Book {
							name
							author {
								name
								reviewed {
									name
								}
							}
						}
					}`,
				Results: []map[string]any{
					{
						"name": "John and the sourcerers' stone",
						"author": map[string]any{
							"name": "John",
							"reviewed": map[string]any{
								"name": "John and the sourcerers' stone",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test documents undesirable behaviour, as the documents are not linked.
// https://github.com/sourcenetwork/defradb/issues/1704
func TestBackupSelfRefImport_SplitPrimaryRelationWithSecondCollection_NoError(t *testing.T) {
	expectedExportData := `{` +
		`"Author":[` +
		`{` +
		`"_docID":"bae-069af8c0-9728-5dde-84ff-ab2dd836f165",` +
		`"_docIDNew":"bae-f2e84aeb-decc-5e40-94ff-e365f0ed0f4b",` +
		`"book_id":"bae-006376a9-5ceb-5bd0-bfed-6ff5afd3eb93",` +
		`"name":"John"` +
		`}` +
		`],` +
		`"Book":[` +
		`{` +
		`"_docID":"bae-2b931633-22bf-576f-b788-d8098b213e5a",` +
		`"_docIDNew":"bae-c821a0a9-7afc-583b-accb-dc99a09c1ff8",` +
		`"name":"John and the sourcerers' stone",` +
		`"reviewedBy_id":"bae-069af8c0-9728-5dde-84ff-ab2dd836f165"` +
		`}` +
		`]` +
		`}`

	test := testUtils.TestCase{
		Actions: []any{
			// Configure 2 nodes for this test, we will export from the first
			// and import to the second.
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Author {
						name: String
						book: Book @primary @relation(name: "author_book")
						reviewed: Book @relation(name: "reviewedBy_reviewed")
					}
					type Book {
						name: String
						author: Author @relation(name: "author_book")
						reviewedBy: Author @primary @relation(name: "reviewedBy_reviewed")
					}
				`,
			},
			testUtils.CreateDoc{
				NodeID:       immutable.Some(0),
				CollectionID: 1,
				// bae-2b931633-22bf-576f-b788-d8098b213e5a
				Doc: `{
					"name": "John and the sourcerers' stone"
				}`,
			},
			testUtils.CreateDoc{
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"book": "bae-2b931633-22bf-576f-b788-d8098b213e5a"
				}`,
			},
			testUtils.UpdateDoc{
				NodeID:       immutable.Some(0),
				CollectionID: 1,
				DocID:        0,
				Doc: `{
					"reviewedBy_id": "bae-069af8c0-9728-5dde-84ff-ab2dd836f165"
				}`,
			},
			/*
				This fails due to the linked ticket.
				https://github.com/sourcenetwork/defradb/issues/1704
				testUtils.BackupExport{
					NodeID:          immutable.Some(0),
					ExpectedContent: expectedExportData,
				},
			*/
			testUtils.BackupImport{
				NodeID:        immutable.Some(1),
				ImportContent: expectedExportData,
			},
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `
					query {
						Book {
							name
							author {
								name
								reviewed {
									name
								}
							}
						}
					}`,
				Results: []map[string]any{
					{
						"name": "John and the sourcerers' stone",
						"author": map[string]any{
							"name":     "John",
							"reviewed": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
