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

func TestMutationDeletion_WithIDAndTxn(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple delete mutation where one element exists.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad"
				}`,
			},
			testUtils.Request{
				TransactionID: immutable.Some(0),
				Request: `mutation {
					delete_User(docIDs: ["bae-d7546ac1-c133-5853-b866-9b9f926fe7e5"]) {
						_docID
					}
				}`,
				Results: []map[string]any{
					{
						"_docID": "bae-d7546ac1-c133-5853-b866-9b9f926fe7e5",
					},
				},
			},
			testUtils.Request{
				TransactionID: immutable.Some(0),
				Request: `query {
					User {
						_docID
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
