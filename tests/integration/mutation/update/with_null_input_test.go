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

package update

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpdate_WithNullFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Bob",
				},
			},
			&action.Request{
				Request: `mutation {
					update_Users(filter: null, input: {name: "Alice"}) {
						name
					}
				}`,
				Results: map[string]any{
					"update_Users": []map[string]any{
						{
							"name": "Alice",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithNullDocID_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Bob",
				},
			},
			&action.Request{
				Request: `mutation {
					update_Users(docID: null, input: {name: "Alice"}) {
						name
					}
				}`,
				Results: map[string]any{
					"update_Users": []map[string]any{
						{
							"name": "Alice",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithNullDocIDs_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Bob",
				},
			},
			&action.Request{
				Request: `mutation {
					update_Users(docID: null, input: {name: "Alice"}) {
						name
					}
				}`,
				Results: map[string]any{
					"update_Users": []map[string]any{
						{
							"name": "Alice",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
