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

func TestMutationDeletion_WithUpdateAndIDsAndSelectAlias(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Delete multiple documents that exist, when given multiple keys with alias after update.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						age: Int
						points: Float
						verified: Boolean
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"age":  26,
					"points": 48.48,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age":  26,
					"points": 48.48,
					"verified": true
				}`,
			},
			testUtils.UpdateDoc{
				DocID: 0,
				Doc: `{
					"age":  27,
					"points": 48.2,
					"verified": false
				}`,
			},
			testUtils.Request{
				Request: `mutation {
					delete_User(ids: ["bae-6a6482a8-24e1-5c73-a237-ca569e41507d", "bae-3a1a496e-24eb-5ae3-9c17-524c146a393e"]) {
						AliasKey: _key
					}
				}`,
				Results: []map[string]any{
					{
						"AliasKey": "bae-3a1a496e-24eb-5ae3-9c17-524c146a393e",
					},
					{
						"AliasKey": "bae-6a6482a8-24e1-5c73-a237-ca569e41507d",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
