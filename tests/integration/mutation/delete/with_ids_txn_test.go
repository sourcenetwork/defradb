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

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationDeletion_WithIDsAndTxn(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple multi-key delete mutation with one key that exists and txn.",
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
			testUtils.Request{
				TransactionID: immutable.Some(0),
				Request: `mutation {
					delete_User(ids: ["bae-6a6482a8-24e1-5c73-a237-ca569e41507d"]) {
						_key
					}
				}`,
				Results: []map[string]any{
					{
						"_key": "bae-6a6482a8-24e1-5c73-a237-ca569e41507d",
					},
				},
			},
			testUtils.Request{
				TransactionID: immutable.Some(0),
				Request: `query {
					User(dockeys: ["bae-6a6482a8-24e1-5c73-a237-ca569e41507d"]) {
						_key
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
