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

func TestQueryInlineIntegerArrayWithAverageWithFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array, filtered average of integer array",
		Query: `query {
					users {
						Name
						_avg(FavouriteIntegers: {filter: {_gt: 0}})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "Shahzad",
					"FavouriteIntegers": [-1, 2, -1, 1, 0]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "Shahzad",
				"_avg": float64(1.5),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableIntegerArrayWithAverageWithFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array with filter, average of populated nillable integer array",
		Query: `query {
					users {
						Name
						_avg(TestScores: {filter: {_gt: -1}})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"TestScores": [-1, null, 13, 0]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "John",
				"_avg": float64(6.5),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArrayWithAverageWithFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array, filtered average of float array",
		Query: `query {
					users {
						Name
						_avg(FavouriteFloats: {filter: {_lt: 9}})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "Shahzad",
					"FavouriteFloats": [3.4, 3.6, 10]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "Shahzad",
				"_avg": 3.5,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableFloatArrayWithAverageWithFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array, filtered average of nillable float array",
		Query: `query {
					users {
						Name
						_avg(PageRatings: {filter: {_lt: 9}})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "Shahzad",
					"PageRatings": [3.4, 3.6, 10, null]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "Shahzad",
				"_avg": 3.5,
			},
		},
	}

	executeTestCase(t, test)
}
