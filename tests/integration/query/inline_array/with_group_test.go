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

package inline_array

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryInlineArrayWithGroupByString(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteIntegers": [-1, 2, -1, 1, 0]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteIntegers": [1, -2, 1, -1, 0]
				}`,
			},
			&action.Request{
				Request: `query {
					Users (groupBy: [name]) {
						name
						GROUP {
							favouriteIntegers
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							"GROUP": []map[string]any{
								{
									"favouriteIntegers": []int64{-1, 2, -1, 1, 0},
								},
								{
									"favouriteIntegers": []int64{1, -2, 1, -1, 0},
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineArrayWithGroupByArray(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Andy",
					"favouriteIntegers": [-1, 2, -1, 1, 0]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteIntegers": [-1, 2, -1, 1, 0]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": [1, 2, 3]
				}`,
			},
			&action.Request{
				Request: `query {
					Users (groupBy: [favouriteIntegers]) {
						favouriteIntegers
						GROUP {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"favouriteIntegers": []int64{-1, 2, -1, 1, 0},
							"GROUP": []map[string]any{
								{"name": "Shahzad"},
								{"name": "Andy"},
							},
						},
						{
							"favouriteIntegers": []int64{1, 2, 3},
							"GROUP": []map[string]any{
								{"name": "John"},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}
