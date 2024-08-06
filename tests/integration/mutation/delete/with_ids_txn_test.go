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
		Description: "Simple multi-docIDs delete mutation with one ID that exists and txn.",
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
					delete_User(docIDs: ["bae-959725a4-17cb-5e04-8908-98bc78fd06dd"]) {
						_docID
					}
				}`,
				Results: map[string]any{
					"delete_User": []map[string]any{
						{
							"_docID": "bae-959725a4-17cb-5e04-8908-98bc78fd06dd",
						},
					},
				},
			},
			testUtils.Request{
				TransactionID: immutable.Some(0),
				Request: `query {
					User(docIDs: ["bae-959725a4-17cb-5e04-8908-98bc78fd06dd"]) {
						_docID
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
