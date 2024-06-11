// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package delete

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationDeletion_WithIDs(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Delete multiple documents that exist, when given multiple IDs.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.Request{
				Request: `mutation {
					delete_User(docIDs: ["bae-22dacd35-4560-583a-9a80-8edbf28aa85c", "bae-1ef746f8-821e-586f-99b2-4cb1fb9b782f"]) {
						_docID
					}
				}`,
				Results: []map[string]any{
					{
						"_docID": "bae-1ef746f8-821e-586f-99b2-4cb1fb9b782f",
					},
					{
						"_docID": "bae-22dacd35-4560-583a-9a80-8edbf28aa85c",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationDeletion_WithEmptyIDs(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Deletion of using ids, empty ids set.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.Request{
				Request: `mutation {
					delete_User(docIDs: []) {
						_docID
					}
				}`,
				Results: []map[string]any{},
			},
			testUtils.Request{
				// Make sure no documents have been deleted
				Request: `query {
						User {
							name
						}
					}`,
				Results: []map[string]any{
					{
						"name": "John",
					},
					{
						"name": "Shahzad",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationDeletion_WithIDsSingleUnknownID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Deletion of using ids, single unknown item.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			testUtils.Request{
				Request: `mutation {
					delete_User(docIDs: ["bae-6a6482a8-24e1-5c73-a237-ca569e41507e"]) {
						_docID
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationDeletion_WithIDsMultipleUnknownID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Deletion of using ids, single unknown item.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			testUtils.Request{
				Request: `mutation {
					delete_User(docIDs: ["bae-028383cc-d6ba-5df7-959f-2bdce3536a05", "bae-028383cc-d6ba-5df7-959f-2bdce3536a03"]) {
						_docID
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationDeletion_WithIDsKnownAndUnknown(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Deletion of using ids, known and unknown items.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad"
				}`,
			},
			testUtils.Request{
				Request: `mutation {
					delete_User(docIDs: ["bae-22dacd35-4560-583a-9a80-8edbf28aa85c", "bae-1ef746f8-821e-586f-99b2-4cb1fb9b782f"]) {
						_docID
					}
				}`,
				Results: []map[string]any{
					{
						"_docID": "bae-22dacd35-4560-583a-9a80-8edbf28aa85c",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
