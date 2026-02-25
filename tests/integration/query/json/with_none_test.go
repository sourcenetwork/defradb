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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryJSON_WithNoneFilter_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `type Users {
					name: String
					custom: JSON
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": [1, false, "second", {"one": 1}, [1, 2]]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Fred",
					"custom": [null, false, "second", {"one": 1}, [1, 2]]
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {custom: {_none: {_eq: null}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithNoneFilterAndNestedArray_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `type Users {
					name: String
					custom: JSON
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": [1, false, "second", {"one": 3}, [1, 3]]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Fred",
					"custom": [null, false, "second", 3, {"one": 1}, [1, 2]]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Islam",
					"custom": 3
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Bruno",
					"custom": null
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"custom": false
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {custom: {_none: {_eq: 3}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "Shahzad"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
