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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestBackupSelfRefImport_Simple_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.BackupImport{
				ImportContent: `{
					"User":[
						{
							"_docID":"bae-692a9178-a258-5224-990f-9ad703a2bbea",
							"age":31,
							"_bossID":"bae-1635f80b-612a-5378-a185-cad7a3018354",
							"name":"Bob"
						},
						{
							"_docID":"bae-1635f80b-612a-5378-a185-cad7a3018354",
							"age":30,
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
							boss {
								name
							}
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
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
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestBackupSelfRefImport_SelfRef_NoError(t *testing.T) {
	expectedExportData := `{` +
		`"User":[` +
		`{` +
		`"_bossID":"bae-0a85be75-1f76-5dcd-b31a-4798f65e45e9",` +
		`"_docID":"bae-0a85be75-1f76-5dcd-b31a-4798f65e45e9",` +
		`"_docIDNew":"bae-0a85be75-1f76-5dcd-b31a-4798f65e45e9",` +
		`"age":31,` +
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
			&action.AddCollection{
				SDL: userCollection,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "Bob",
					"age": 31
				}`,
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"_bossID": "bae-0a85be75-1f76-5dcd-b31a-4798f65e45e9"
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
			&action.Request{
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
			&action.AddCollection{
				SDL: `
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
							"author":"bae-ca99414a-8336-537d-87d7-a7c4d90903b4",
							"reviewedBy":"bae-ca99414a-8336-537d-87d7-a7c4d90903b4"
						}
					]
				}`,
			},
			&action.Request{
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
			&action.AddCollection{
				SDL: `
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
							"author":"bae-ca99414a-8336-537d-87d7-a7c4d90903b4",
							"reviewedBy":"bae-ca99414a-8336-537d-87d7-a7c4d90903b4"
						}
					],
					"Author":[
						{
							"name":"John"
						}
					]
				}`,
			},
			&action.Request{
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
		`"_bookID":"bae-89136f56-3779-5656-b8a6-f76a1c262f37",` +
		`"name":"John"` +
		`}` +
		`],` +
		`"Book":[` +
		`{` +
		`"_docID":"bae-89136f56-3779-5656-b8a6-f76a1c262f37",` +
		`"_docIDNew":"bae-66b0f769-c743-5a50-ae6d-1dcd978e2404",` +
		`"name":"John and the sourcerers' stone",` +
		`"_reviewedByID":"bae-bf1f16db-3c02-5759-8127-7d73346442cc"` +
		`}` +
		`]` +
		`}`

	test := testUtils.TestCase{
		Actions: []any{
			// Configure 2 nodes for this test, we will export from the first
			// and import to the second.
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				NodeID:       immutable.Some(0),
				CollectionID: 1,
				// bae-89136f56-3779-5656-b8a6-f76a1c262f37
				Doc: `{
					"name": "John and the sourcerers' stone"
				}`,
			},
			&action.AddDoc{
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
					"_reviewedByID": "bae-bf1f16db-3c02-5759-8127-7d73346442cc"
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
			&action.Request{
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
