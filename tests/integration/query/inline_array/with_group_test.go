// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package inline_array

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryInlineArrayWithGroupByString(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, mixed integers, group by string",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteIntegers": [-1, 2, -1, 1, 0]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteIntegers": [1, -2, 1, -1, 0]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users (groupBy: [name]) {
						name
						_group {
							favouriteIntegers
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							"_group": []map[string]any{
								{
									"favouriteIntegers": []int64{1, -2, 1, -1, 0},
								},
								{
									"favouriteIntegers": []int64{-1, 2, -1, 1, 0},
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
		Description: "Simple inline array with no filter, mixed integers, group by array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"favouriteIntegers": [-1, 2, -1, 1, 0]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteIntegers": [-1, 2, -1, 1, 0]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": [1, 2, 3]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users (groupBy: [favouriteIntegers]) {
						favouriteIntegers
						_group {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"favouriteIntegers": []int64{1, 2, 3},
							"_group": []map[string]any{
								{
									"name": "John",
								},
							},
						},
						{
							"favouriteIntegers": []int64{-1, 2, -1, 1, 0},
							"_group": []map[string]any{
								{
									"name": "Andy",
								},
								{

									"name": "Shahzad",
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
