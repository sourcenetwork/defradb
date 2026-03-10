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

func TestMutationDeletion_WithFilter(t *testing.T) {
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
					delete_User(filter: {name: {_eq: "Shahzad"}}) {
						name
					}
				}`,
				Results: map[string]any{
					"delete_User": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationDeletion_WithFilterMatchingMultipleDocs(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"age": 1
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"age": 2
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 3
				}`,
			},
			&action.Request{
				Request: `mutation {
					delete_User(filter: {name: {_eq: "Shahzad"}}) {
						age
					}
				}`,
				Results: map[string]any{
					"delete_User": []map[string]any{
						{
							"age": int64(1),
						},
						{
							"age": int64(2),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationDeletion_WithEmptyFilter(t *testing.T) {
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
					"name": "John"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Fred"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad"
				}`,
			},
			&action.Request{
				Request: `mutation {
					delete_User(filter: {}) {
						name
					}
				}`,
				Results: map[string]any{
					"delete_User": []map[string]any{
						{
							"name": "John",
						},
						{
							"name": "Fred",
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

func TestMutationDeletion_WithFilterNoMatch(t *testing.T) {
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
					delete_User(filter: {name: {_eq: "Lone"}}) {
						name
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

func TestMutationDeletion_WithFilterOnEmptyCollection(t *testing.T) {
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
					delete_User(filter: {name: {_eq: "Lone"}}) {
						name
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
