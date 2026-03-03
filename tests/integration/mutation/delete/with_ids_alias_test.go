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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationDeletion_WithIDsAndSelectAlias(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int
						points: Float
						verified: Boolean
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"age":  26,
					"points": 48.48,
					"verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age":  26,
					"points": 48.48,
					"verified": true
				}`,
			},
			&action.Request{
				Request: `mutation {
					delete_User(docID: ["bae-3b39742b-cfff-5158-b6d5-d69cf79066b4", "bae-49057a46-bf84-5e83-b043-e6fa6ed5b70c"]) {
						AliasID: _docID
					}
				}`,
				Results: map[string]any{
					"delete_User": []map[string]any{
						{
							"AliasID": "bae-3b39742b-cfff-5158-b6d5-d69cf79066b4",
						},
						{
							"AliasID": "bae-49057a46-bf84-5e83-b043-e6fa6ed5b70c",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
