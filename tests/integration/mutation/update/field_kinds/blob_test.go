// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package field_kinds

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpdate_WithBlobField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update of blob field",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						data: Blob
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"data": "00FE"
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"data": "00FF"
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Users {
							data
						}
					}
				`,
				Results: []map[string]any{
					{
						"data": "00FF",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
