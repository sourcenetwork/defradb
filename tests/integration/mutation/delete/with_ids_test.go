// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package delete

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationDeletion_WithIDs(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			&action.Request{
				Request: `mutation {
					delete_User(docID: ["{{.DocID0_0}}", "{{.DocID0_1}}"]) {
						_docID
					}
				}`,
				Results: map[string]any{
					"delete_User": []map[string]any{
						{
							"_docID": "{{.DocID0_1}}",
						},
						{
							"_docID": "{{.DocID0_0}}",
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
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			&action.Request{
				Request: `mutation {
					delete_User(docID: []) {
						_docID
					}
				}`,
				Results: map[string]any{
					"delete_User": []map[string]any{},
				},
			},
			&action.Request{
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
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.Request{
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
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.Request{
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
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad"
				}`,
			},
			&action.Request{
				Request: `mutation {
					delete_User(docID: ["{{.DocID0_0}}", "{{.DocID0_1}}"]) {
						_docID
					}
				}`,
				Results: map[string]any{
					"delete_User": []map[string]any{
						{
							"_docID": "{{.DocID0_0}}",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
