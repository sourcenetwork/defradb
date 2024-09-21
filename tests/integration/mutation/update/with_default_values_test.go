// Copyright 2024 Democratized Data Foundation
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

func TestMutationUpdate_WithDefaultValues_DoesNotOverwrite(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation with default value does not overwrite",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String 
						score: Int @default(int: 100)
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"score": 0
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"name": "Fred"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						score
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":  "Fred",
							"score": int64(0),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
