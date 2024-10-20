// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package update

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpdate_WithNullFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation, with null filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Bob",
				},
			},
			testUtils.Request{
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
		Description: "Simple update mutation, with null docID",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Bob",
				},
			},
			testUtils.Request{
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
		Description: "Simple update mutation, with null docIDs",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Bob",
				},
			},
			testUtils.Request{
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
