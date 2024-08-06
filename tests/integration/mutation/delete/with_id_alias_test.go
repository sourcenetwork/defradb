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

func TestMutationDeletion_WithIDAndAlias(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple delete mutation with an alias field name.",
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
				Request: `mutation {
					delete_User(docIDs: ["bae-22dacd35-4560-583a-9a80-8edbf28aa85c"]) {
						fancyKey: _docID
					}
				}`,
				Results: map[string]any{
					"delete_User": []map[string]any{
						{
							"fancyKey": "bae-22dacd35-4560-583a-9a80-8edbf28aa85c",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
