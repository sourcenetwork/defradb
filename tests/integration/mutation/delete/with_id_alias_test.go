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

func TestMutationDeletion_WithIDAndAlias(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple delete mutation with an alias field name.",
		Actions: []any{
			&action.AddSchema{
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
					delete_User(docID: ["bae-9287a692-50a2-5724-98ae-c1a944e29ef1"]) {
						fancyKey: _docID
					}
				}`,
				Results: map[string]any{
					"delete_User": []map[string]any{
						{
							"fancyKey": "bae-9287a692-50a2-5724-98ae-c1a944e29ef1",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
