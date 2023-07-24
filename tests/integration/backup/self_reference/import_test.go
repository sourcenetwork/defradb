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
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.BackupImport{
				// Note: Whilst generating the `boss_id` field value manually would be really difficult,
				// the document may be exported with `_key`/`_newKey` values, resulting in an export that
				// could not be fully imported, although all the information required is there.
				ImportContent: `{
					"User":[
						{
							"_key":"bae-790e7e49-f2e3-5ad6-83d9-5dfb6d8ba81d",
							"age":31,
							"boss_id":"bae-790e7e49-f2e3-5ad6-83d9-5dfb6d8ba81d",
							"name":"Bob"
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
						"boss": nil,
					},
				},
			},
		},
	}

	executeTestCase(t, test)
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
func TestBackupSelfRefImport_SplitPrimaryRelationWithSecondCollection_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
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
			testUtils.BackupImport{
				// Note: Whilst generating the relation field values manually would be really difficult,
				// the document may be exported with `_key`/`_newKey` values, resulting in an export that
				// could not be fully imported, although all the information required is there.
				ImportContent: `{
					"Author":[
						{
							"_key":"bae-2570a325-7cd1-5f0a-9f00-197573a207e3",
							"name":"John",
							"book": "bae-69e58b41-3e41-5cc4-919d-b4f878dcdf21"
						}
					],
					"Book":[
						{
							"_key":"bae-69e58b41-3e41-5cc4-919d-b4f878dcdf21",
							"name":"John and the sourcerers' stone",
							"reviewedBy":"bae-2570a325-7cd1-5f0a-9f00-197573a207e3"
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
						"name":   "John and the sourcerers' stone",
						"author": nil,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Author", "Book"}, test)
}
