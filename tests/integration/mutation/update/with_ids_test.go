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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpdate_WithIds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						points: Float
					}
				`,
			},
			testUtils.CreateDoc{
				// bae-9466cfe3-c011-5d44-b1cd-f0c5a46d9202
				Doc: `{
					"name": "John",
					"points": 42.1
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Bob",
					"points": 66.6
				}`,
			},
			testUtils.CreateDoc{
				// bae-b76814bb-7ac8-5430-bac9-fbd7fc86db40
				Doc: `{
					"name": "Fred",
					"points": 33
				}`,
			},
			&action.Request{
				Request: `mutation {
					update_Users(
						docID: ["bae-9466cfe3-c011-5d44-b1cd-f0c5a46d9202", "bae-b76814bb-7ac8-5430-bac9-fbd7fc86db40"],
						input: {points: 59}
					) {
						name
						points
					}
				}`,
				Results: map[string]any{
					"update_Users": []map[string]any{
						{
							"name":   "John",
							"points": float64(59),
						},
						{
							"name":   "Fred",
							"points": float64(59),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
