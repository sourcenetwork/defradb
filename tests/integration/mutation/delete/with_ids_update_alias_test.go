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
		Description: "Delete multiple documents that exist, when given multiple IDs with alias after update.",
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
					delete_User(docIDs: ["bae-959725a4-17cb-5e04-8908-98bc78fd06dd", "bae-3eed37ed-5c7b-53ff-b125-d04fb173f6c0"]) {
						AliasID: _docID
					}
				}`,
				Results: map[string]any{
					"delete_User": []map[string]any{
						{
							"AliasID": "bae-3eed37ed-5c7b-53ff-b125-d04fb173f6c0",
						},
						{
							"AliasID": "bae-959725a4-17cb-5e04-8908-98bc78fd06dd",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
