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

func TestQueryInlineBoolArrayWithCountWithFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array, filtered count of bool array",
		Query: `query {
					users {
						Name
						_count(LikedIndexes: {filter: {_eq: true}})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "Shahzad",
					"LikedIndexes": [true, true, false, true]
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"Name":   "Shahzad",
				"_count": 3,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableBoolArrayWithCountWithFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array with filter, count of nillable bool array",
		Query: `query {
					users {
						Name
						_count(IndexLikesDislikes: {filter: {_eq: true}})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"IndexLikesDislikes": [true, true, false, null]
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"Name":   "John",
				"_count": 2,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArrayWithCountWithFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array, filtered count of integer array",
		Query: `query {
					users {
						Name
						_count(FavouriteIntegers: {filter: {_gt: 0}})
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
		Results: []map[string]interface{}{
			{
				"Name":   "Shahzad",
				"_count": 2,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableIntegerArrayWithCountWithFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array, filtered count of integer array",
		Query: `query {
					users {
						Name
						_count(TestScores: {filter: {_gt: 0}})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "Shahzad",
					"TestScores": [-1, 2, 1, null, 0]
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"Name":   "Shahzad",
				"_count": 2,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArrayWithCountWithFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array, filtered count of float array",
		Query: `query {
					users {
						Name
						_count(FavouriteFloats: {filter: {_lt: 9}})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "Shahzad",
					"FavouriteFloats": [3.1425, 0.00000000001, 10]
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"Name":   "Shahzad",
				"_count": 2,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineStringArrayWithCountWithFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array, filtered count of string array",
		Query: `query {
					users {
						Name
						_count(PreferredStrings: {filter: {_in: ["", "the first"]}})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "Shahzad",
					"PreferredStrings": ["", "the previous", "the first", "empty string"]
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"Name":   "Shahzad",
				"_count": 2,
			},
		},
	}

	executeTestCase(t, test)
}
