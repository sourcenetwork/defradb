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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestUpdateSave_DeletedDoc_DoesNothing(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Save existing, deleted document",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name":	"John"
				}`,
			},
			testUtils.DeleteDoc{
				DocID: 0,
			},
			testUtils.UpdateDoc{
				DocID: 0,
				Doc: `{
					"name": "Fred"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(showDeleted: true) {
						_deleted
						name
					}
				}`,
				Results: []map[string]any{
					{
						"_deleted": true,
						"name":     "John",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
