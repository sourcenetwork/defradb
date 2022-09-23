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
	test := testUtils.QueryTestCase{
		Description: "Simple inline array with no filter, mixed integers, group by string",
		Query: `query {
					users (groupBy: [Name]) {
						Name
						_group {
							FavouriteIntegers
						}
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "Shahzad",
					"FavouriteIntegers": [-1, 2, -1, 1, 0]
				}`,
				`{
					"Name": "Shahzad",
					"FavouriteIntegers": [1, -2, 1, -1, 0]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "Shahzad",
				"_group": []map[string]any{
					{
						"FavouriteIntegers": []int64{-1, 2, -1, 1, 0},
					},
					{
						"FavouriteIntegers": []int64{1, -2, 1, -1, 0},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineArrayWithGroupByArray(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array with no filter, mixed integers, group by array",
		Query: `query {
					users (groupBy: [FavouriteIntegers]) {
						FavouriteIntegers
						_group {
							Name
						}
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "Andy",
					"FavouriteIntegers": [-1, 2, -1, 1, 0]
				}`,
				`{
					"Name": "Shahzad",
					"FavouriteIntegers": [-1, 2, -1, 1, 0]
				}`,
				`{
					"Name": "John",
					"FavouriteIntegers": [1, 2, 3]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"FavouriteIntegers": []int64{-1, 2, -1, 1, 0},
				"_group": []map[string]any{
					{
						"Name": "Shahzad",
					},
					{
						"Name": "Andy",
					},
				},
			},
			{
				"FavouriteIntegers": []int64{1, 2, 3},
				"_group": []map[string]any{
					{
						"Name": "John",
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
