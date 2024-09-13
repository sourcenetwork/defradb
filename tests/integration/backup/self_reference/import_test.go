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
							"_docID":"bae-0dfbaf9f-3c58-5133-aa07-a9f25d792f4e",
							"age":31,
							"boss_id":"bae-8096e3d7-41ab-5afe-ad88-481150483db1",
							"name":"Bob"
						},
						{
							"_docID":"bae-8096e3d7-41ab-5afe-ad88-481150483db1",
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
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Bob",
							"boss": map[string]any{
								"name": "John",
							},
						},
						{
							"name": "John",
							"boss": nil,
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
		`"_docID":"bae-b9449db8-3894-5701-84ce-ee96a3eafc9c",` +
		`"_docIDNew":"bae-b9449db8-3894-5701-84ce-ee96a3eafc9c",` +
		`"age":31,` +
		`"boss_id":"bae-b9449db8-3894-5701-84ce-ee96a3eafc9c",` +
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
					"boss_id": "bae-b9449db8-3894-5701-84ce-ee96a3eafc9c"
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
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Bob",
							"boss": map[string]any{
								"name": "Bob",
							},
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
				Results: map[string]any{
					"Book": []map[string]any{
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
				Results: map[string]any{
					"Book": []map[string]any{
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
		`"_docID":"bae-bf1f16db-3c02-5759-8127-7d73346442cc",` +
		`"_docIDNew":"bae-bf1f16db-3c02-5759-8127-7d73346442cc",` +
		`"book_id":"bae-89136f56-3779-5656-b8a6-f76a1c262f37",` +
		`"name":"John"` +
		`}` +
		`],` +
		`"Book":[` +
		`{` +
		`"_docID":"bae-89136f56-3779-5656-b8a6-f76a1c262f37",` +
		`"_docIDNew":"bae-66b0f769-c743-5a50-ae6d-1dcd978e2404",` +
		`"name":"John and the sourcerers' stone",` +
		`"reviewedBy_id":"bae-bf1f16db-3c02-5759-8127-7d73346442cc"` +
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
				// bae-89136f56-3779-5656-b8a6-f76a1c262f37
				Doc: `{
					"name": "John and the sourcerers' stone"
				}`,
			},
			testUtils.CreateDoc{
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"book": "bae-89136f56-3779-5656-b8a6-f76a1c262f37"
				}`,
			},
			testUtils.UpdateDoc{
				NodeID:       immutable.Some(0),
				CollectionID: 1,
				DocID:        0,
				Doc: `{
					"reviewedBy_id": "bae-bf1f16db-3c02-5759-8127-7d73346442cc"
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
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":   "John and the sourcerers' stone",
							"author": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
