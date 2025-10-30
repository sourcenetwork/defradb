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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationDeletion_WithIDs(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
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
					delete_User(docID: ["bae-390b4419-fe1c-506b-98bd-20847cdab2d9", "bae-7f4197fe-c647-5cc6-91bb-5f32229fd4cd"]) {
						_docID
					}
				}`,
				Results: map[string]any{
					"delete_User": []map[string]any{
						{
							"_docID": "bae-7f4197fe-c647-5cc6-91bb-5f32229fd4cd",
						},
						{
							"_docID": "bae-390b4419-fe1c-506b-98bd-20847cdab2d9",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationDeletion_WithEmptyIDs(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
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
					delete_User(docID: []) {
						_docID
					}
				}`,
				Results: map[string]any{
					"delete_User": []map[string]any{},
				},
			},
			testUtils.Request{
				// Make sure no documents have been deleted
				Request: `query {
						User {
							name
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
						},
						{
							"name": "Shahzad",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationDeletion_WithIDsSingleUnknownID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			testUtils.Request{
				Request: `mutation {
					delete_User(docID: ["bae-6a6482a8-24e1-5c73-a237-ca569e41507e"]) {
						_docID
					}
				}`,
				Results: map[string]any{
					"delete_User": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationDeletion_WithIDsMultipleUnknownID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			testUtils.Request{
				Request: `mutation {
					delete_User(docID: ["bae-028383cc-d6ba-5df7-959f-2bdce3536a05", "bae-028383cc-d6ba-5df7-959f-2bdce3536a03"]) {
						_docID
					}
				}`,
				Results: map[string]any{
					"delete_User": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationDeletion_WithIDsKnownAndUnknown(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
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
					delete_User(docID: ["bae-390b4419-fe1c-506b-98bd-20847cdab2d9", "bae-7f4197fe-c647-5cc6-91bb-5f32229fd4cd"]) {
						_docID
					}
				}`,
				Results: map[string]any{
					"delete_User": []map[string]any{
						{
							"_docID": "bae-390b4419-fe1c-506b-98bd-20847cdab2d9",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
