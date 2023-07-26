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

func TestQuerySimpleWithRestart(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with no filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			testUtils.Restart{},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"age": 30
				}`,
			},
			testUtils.Restart{},
			testUtils.Request{
				Request: ` query {
					Users {
						name
						age
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Shahzad",
						"age":  uint64(30),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
