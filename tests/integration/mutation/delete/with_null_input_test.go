// Copyright 2024 Democratized Data Foundation
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

func TestMutationDelete_WithNullFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple delete mutation, with null filter",
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
					delete_Users(filter: null) {
						name
					}
				}`,
				Results: map[string]any{
					"delete_Users": []map[string]any{
						{
							"name": "Bob",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationDelete_WithNullDocID_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple delete mutation, with null docID",
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
					delete_Users(docID: null) {
						name
					}
				}`,
				Results: map[string]any{
					"delete_Users": []map[string]any{
						{
							"name": "Bob",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationDelete_WithNullDocIDs_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple delete mutation, with null docIDs",
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
					delete_Users(docIDs: null) {
						name
					}
				}`,
				Results: map[string]any{
					"delete_Users": []map[string]any{
						{
							"name": "Bob",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
