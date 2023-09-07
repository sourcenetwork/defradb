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

func TestMutationDeletion_WithIDsAndEmptyFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Deletion of using ids and filter, known id and empty filter.",
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
					delete_User(ids: ["bae-d7546ac1-c133-5853-b866-9b9f926fe7e5"], filter: {}) {
						_key
					}
				}`,
				Results: []map[string]any{
					{
						"_key": "bae-d7546ac1-c133-5853-b866-9b9f926fe7e5",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
