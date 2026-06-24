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

func TestDeleteWithFilter_Succeeds(t *testing.T) {
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
					"name": "John",
					"age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Fred",
					"age": 25
				}`,
			},
			testUtils.DeleteWithFilter{
				CollectionID: 0,
				Filter:       `{name: {_eq: "John"}}`,
			},
			&action.Request{
				Request: `query {
					User {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Fred"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDeleteWithFilter_WithMultipleMatchingDocs_DeletesAll(t *testing.T) {
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
					"name": "John",
					"age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Fred",
					"age": 25
				}`,
			},
			testUtils.DeleteWithFilter{
				CollectionID: 0,
				Filter:       `{name: {_eq: "John"}}`,
			},
			&action.Request{
				Request: `query {
					User {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Fred"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
