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

func TestMutationUpdate_IfFloatFieldSetToNull_ShouldBeNil(t *testing.T) {
	test := testUtils.TestCase{
		Description: "If float field is set to null, should set to nil",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						rate: Float
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"rate": 0.55
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"rate": null
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Users {
							rate
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"rate": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
