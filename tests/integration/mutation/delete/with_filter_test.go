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

func TestMutationDeletion_WithFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Delete using filter - One matching document, that exists.",
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
					delete_User(filter: {name: {_eq: "Shahzad"}}) {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Shahzad",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationDeletion_WithFilterMatchingMultipleDocs(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Delete using filter - Multiple matching documents that exist.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"age": 1
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"age": 2
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": 3
				}`,
			},
			testUtils.Request{
				Request: `mutation {
					delete_User(filter: {name: {_eq: "Shahzad"}}) {
						age
					}
				}`,
				Results: []map[string]any{
					{
						"age": int64(2),
					},
					{
						"age": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationDeletion_WithEmptyFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Delete using filter - Match everything in this collection.",
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
					"name": "Fred"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.Request{
				Request: `mutation {
					delete_User(filter: {}) {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Fred",
					},
					{
						"name": "Shahzad",
					},
					{
						"name": "John",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationDeletion_WithFilterNoMatch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "No delete with filter: because no document matches filter.",
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
					delete_User(filter: {name: {_eq: "Lone"}}) {
						name
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationDeletion_WithFilterOnEmptyCollection(t *testing.T) {
	test := testUtils.TestCase{
		Description: "No delete with filter: because the collection is empty.",
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
					delete_User(filter: {name: {_eq: "Lone"}}) {
						name
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
