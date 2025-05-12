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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpdate_WithId(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation with document id",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						points: Float
					}
				`,
			},
			testUtils.CreateDoc{
				// bae-348f8357-9ad0-5ab0-9bf6-2985d123370d
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
					update_Users(docID: "bae-348f8357-9ad0-5ab0-9bf6-2985d123370d", input: {points: 59}) {
						name
						points
					}
				}`,
				Results: map[string]any{
					"update_Users": []map[string]any{
						{
							"name":   "John",
							"points": float64(59),
						},
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
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						points: Float
					}
				`,
			},
			testUtils.CreateDoc{
				// bae-348f8357-9ad0-5ab0-9bf6-2985d123370d
				Doc: `{
					"name": "John",
					"points": 42.1
				}`,
			},
			testUtils.Request{
				Request: `mutation {
					update_Users(docID: "bae-does-not-exist", input: {points: 59}) {
						_docID
						name
						points
					}
				}`,
				Results: map[string]any{
					"update_Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
