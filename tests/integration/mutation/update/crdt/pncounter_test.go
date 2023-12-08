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

func TestMutationUpdate_PNCounter_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation of a PN Counter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						points: Int @crdt(type: "pncounter")
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"points": 0
				}`,
			},
			testUtils.Request{
				Request: `mutation {
					update_Users(id: "bae-7d3bc1c9-b467-5ad0-979c-2ecfa06f2184", data: "{\"points\": 10}") {
						name
						points
					}
				}`,
				Results: []map[string]any{
					{
						"name":   "John",
						"points": int64(10),
					},
				},
			},
			testUtils.Request{
				Request: `mutation {
					update_Users(id: "bae-7d3bc1c9-b467-5ad0-979c-2ecfa06f2184", data: "{\"points\": 10}") {
						name
						points
					}
				}`,
				Results: []map[string]any{
					{
						"name":   "John",
						"points": int64(20),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
