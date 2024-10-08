// Copyright 2024 Democratized Data Foundation
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

func TestQuerySimple_WithNotLikeOpOnJSONField_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":   "John",
					"custom": "{\"tree\": \"maple\", \"age\": 250}",
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":   "Andy",
					"custom": "{\"tree\": \"oak\", \"age\": 450}",
				},
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {custom: {_nlike: "%maple%"}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "Andy"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_WithNotLikeOpOnJSONFieldAllTypes_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":   "Andy",
					"custom": "{\"tree\": \"oak\", \"age\": 450}",
				},
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": [1, 2]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"custom": {"one": 1}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": false
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "David",
					"custom": 32
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {custom: {_nlike: "%maple%"}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "John"},
						{"name": "Andy"},
						{"name": "Fred"},
						{"name": "David"},
						{"name": "Shahzad"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
