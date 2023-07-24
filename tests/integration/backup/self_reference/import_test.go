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
							"_key":"bae-790e7e49-f2e3-5ad6-83d9-5dfb6d8ba81d",
							"age":31,
							"boss_id":"bae-e933420a-988a-56f8-8952-6c245aebd519",
							"name":"Bob"
						},
						{
							"_key":"bae-e933420a-988a-56f8-8952-6c245aebd519",
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
	}

	executeTestCase(t, test)
}

// This test documents undesirable behaviour, as the document is not linked to itself.
// https://github.com/sourcenetwork/defradb/issues/1697
func TestBackupSelfRefImport_SelfRef_NoError(t *testing.T) {
	expectedExportData := `{` +
		`"User":[` +
		`{` +
		`"_key":"bae-0648f44e-74e8-593b-a662-3310ec278927",` +
		`"_newKey":"bae-0648f44e-74e8-593b-a662-3310ec278927",` +
		`"age":31,` +
		`"boss_id":"bae-0648f44e-74e8-593b-a662-3310ec278927",` +
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
					"boss_id": "bae-0648f44e-74e8-593b-a662-3310ec278927"
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
						"boss": nil,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"User"}, test)
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
							"author":"bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad",
							"reviewedBy":"bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad"
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

	testUtils.ExecuteTestCase(t, []string{"Author", "Book"}, test)
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
							"author":"bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad",
							"reviewedBy":"bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad"
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

	testUtils.ExecuteTestCase(t, []string{"Author", "Book"}, test)
}

// This test documents undesirable behaviour, as the documents are not linked.
// https://github.com/sourcenetwork/defradb/issues/1697
// https://github.com/sourcenetwork/defradb/issues/1704
func TestBackupSelfRefImport_SplitPrimaryRelationWithSecondCollection_NoError(t *testing.T) {
	expectedExportData := `{` +
		`"Author":[` +
		`{` +
		`"_key":"bae-d760e445-22ef-5956-9947-26de226891f6",` +
		`"_newKey":"bae-e3a6ff01-33ff-55f4-88f9-d13db26274c8",` +
		`"book_id":"bae-c821a0a9-7afc-583b-accb-dc99a09c1ff8",` +
		`"name":"John"` +
		`}` +
		`],` +
		`"Book":[` +
		`{` +
		`"_key":"bae-4059cb15-2b30-5049-b0df-64cc7ad9b5e4",` +
		`"_newKey":"bae-c821a0a9-7afc-583b-accb-dc99a09c1ff8",` +
		`"name":"John and the sourcerers' stone",` +
		`"reviewedBy_id":"bae-e3a6ff01-33ff-55f4-88f9-d13db26274c8"` +
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
				// bae-4059cb15-2b30-5049-b0df-64cc7ad9b5e4
				Doc: `{
					"name": "John and the sourcerers' stone"
				}`,
			},
			testUtils.CreateDoc{
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"book": "bae-4059cb15-2b30-5049-b0df-64cc7ad9b5e4"
				}`,
			},
			testUtils.UpdateDoc{
				NodeID:       immutable.Some(0),
				CollectionID: 1,
				DocID:        0,
				Doc: `{
					"reviewedBy_id": "bae-d760e445-22ef-5956-9947-26de226891f6"
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

	testUtils.ExecuteTestCase(t, []string{"Author", "Book"}, test)
}
