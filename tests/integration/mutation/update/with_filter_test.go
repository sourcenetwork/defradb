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

func TestMutationUpdate_WithBooleanFilter_ResultFilteredOut(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation with boolean equals filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						verified: Boolean
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"verified": true
				}`,
			},
			testUtils.Request{
				// The update will result in a record that no longer matches the filter
				Request: `mutation {
					update_Users(filter: {verified: {_eq: true}}, data: "{\"verified\":false}") {
						_key
						name
						verified
					}
				}`,
				// As the record no longer matches the filter it is not returned
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithBooleanFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation with boolean filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						verified: Boolean
						points: Float
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"verified": true,
					"points": 42.1
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Bob",
					"verified": false,
					"points": 66.6
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"verified": true,
					"points": 33
				}`,
			},
			testUtils.Request{
				Request: `mutation {
					update_Users(filter: {verified: {_eq: true}}, data: "{\"points\": 59}") {
						name
						points
					}
				}`,
				Results: []map[string]any{
					{
						"name":   "Fred",
						"points": float64(59),
					},
					{
						"name":   "John",
						"points": float64(59),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
