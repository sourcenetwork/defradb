// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package json

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryJSON_WithAnyFilterWithAllTypes_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users {
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
			&action.AddDoc{
				Doc: `{
					"name": "Islam",
					"custom": null
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Keenan",
					"custom": 0
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Andy",
					"custom": ""
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"custom": true
				}`,
			},
			&action.Request{
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
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users {
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
					"custom": [null, false, "second", {"one": 1}, [1, [2, 3]]]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Islam",
					"custom": null
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Keenan",
					"custom": 3
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Bruno",
					"custom": [null, 3]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Andy",
					"custom": ""
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"custom": true
				}`,
			},
			&action.Request{
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
