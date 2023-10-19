// Copyright 2022 Democratized Data Foundation
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

func TestMutationUpdateUnderscoredSchema(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update of schema with underscored name",

		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type My_User {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"name": "Fred"
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						My_User {
							name
						}
					}
				`,
				Results: []map[string]any{
					{
						"name": "Fred",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
