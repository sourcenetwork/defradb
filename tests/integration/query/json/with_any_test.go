// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package json

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryJSON_WithAnyFilterWithAllTypes_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple JSON array, filtered any of all types array",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `type Users {
					name: String
					custom: JSON
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": [1, false, "second", {"one": 1}, [1, 2]]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"custom": [null, false, "second", {"one": 1}, [1, 2]]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": null
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"custom": 0
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"custom": ""
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": true
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {custom: {_any: {_eq: null}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithAnyFilterAndNestedArray_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple JSON array, filtered any of all types array",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `type Users {
					name: String
					custom: JSON
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": [1, false, "second", {"one": 1}, [1, 2]]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"custom": [null, false, "second", {"one": 1}, [1, [2, 3]]]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": null
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"custom": 3
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Bruno",
					"custom": [null, 3]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"custom": ""
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": true
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {custom: {_any: {_eq: 3}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "Bruno"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
