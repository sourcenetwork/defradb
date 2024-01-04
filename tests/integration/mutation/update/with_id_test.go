// Copyright 2023 Democratized Data Foundation
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

func TestMutationUpdate_WithId(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation with document id",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						points: Float
					}
				`,
			},
			testUtils.CreateDoc{
				// bae-cc36febf-4029-52b3-a876-c99c6293f588
				Doc: `{
					"name": "John",
					"points": 42.1
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Bob",
					"points": 66.6
				}`,
			},
			testUtils.Request{
				Request: `mutation {
					update_Users(docID: "bae-cc36febf-4029-52b3-a876-c99c6293f588", data: "{\"points\": 59}") {
						name
						points
					}
				}`,
				Results: []map[string]any{
					{
						"name":   "John",
						"points": float64(59),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithNonExistantId(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation with non existant document id",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						points: Float
					}
				`,
			},
			testUtils.CreateDoc{
				// bae-cc36febf-4029-52b3-a876-c99c6293f588
				Doc: `{
					"name": "John",
					"points": 42.1
				}`,
			},
			testUtils.Request{
				Request: `mutation {
					update_Users(docID: "bae-does-not-exist", data: "{\"points\": 59}") {
						_docID
						name
						points
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
