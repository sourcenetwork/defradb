// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package simple

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimple_WithDeletedField(t *testing.T) {
	test := testUtils.TestCase{
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
					"name": "John"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy"
				}`,
			},
			testUtils.DeleteDoc{
				DocID: 0,
			},
			testUtils.DeleteDoc{
				DocID: 1,
			},
			testUtils.Request{
				Request: `query {
						User(showDeleted: true) {
							_deleted
							name
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"_deleted": true,
							"name":     "John",
						},
						{
							"_deleted": true,
							"name":     "Andy",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
