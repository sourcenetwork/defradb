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
	test := testUtils.RequestTestCase{
		Description: "Simple inline array with no filter, mixed integers, group by string",
		Request: `query {
					Users (groupBy: [name]) {
						name
						_group {
							favouriteIntegers
						}
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "Shahzad",
					"favouriteIntegers": [-1, 2, -1, 1, 0]
				}`,
				`{
					"name": "Shahzad",
					"favouriteIntegers": [1, -2, 1, -1, 0]
				}`,
			},
		},
		Results: []map[string]any{
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
	}

	executeTestCase(t, test)
}

func TestQueryInlineArrayWithGroupByArray(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array with no filter, mixed integers, group by array",
		Request: `query {
					Users (groupBy: [favouriteIntegers]) {
						favouriteIntegers
						_group {
							name
						}
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "Andy",
					"favouriteIntegers": [-1, 2, -1, 1, 0]
				}`,
				`{
					"name": "Shahzad",
					"favouriteIntegers": [-1, 2, -1, 1, 0]
				}`,
				`{
					"name": "John",
					"favouriteIntegers": [1, 2, 3]
				}`,
			},
		},
		Results: []map[string]any{
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
	}

	executeTestCase(t, test)
}
