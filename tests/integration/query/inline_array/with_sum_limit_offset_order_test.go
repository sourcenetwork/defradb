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

func TestQueryInlineIntegerArrayWithSumWithOffsetWithLimitWithOrderAsc(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array, ordered offsetted limited sum of integer array",
		Query: `query {
					users {
						Name
						_sum(FavouriteIntegers: {offset: 1, limit: 3, order: ASC})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "Shahzad",
					"FavouriteIntegers": [-1, 2, 5, 1, 0, 7]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "Shahzad",
				// 0 + 1 + 2
				"_sum": int64(3),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArrayWithSumWithOffsetWithLimitWithOrderDesc(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array, ordered offsetted limited sum of integer array",
		Query: `query {
					users {
						Name
						_sum(FavouriteIntegers: {offset: 1, limit: 3, order: DESC})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "Shahzad",
					"FavouriteIntegers": [-1, 2, 5, 1, 0, 7]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "Shahzad",
				// 5 + 2 + 1
				"_sum": int64(8),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableIntegerArrayWithSumWithOffsetWithLimitWithOrderAsc(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array, ordered offsetted limited sum of integer array",
		Query: `query {
					users {
						Name
						_sum(TestScores: {offset: 1, limit: 3, order: ASC})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "Shahzad",
					"TestScores": [2, null, 5, 1, 0, 7]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "Shahzad",
				// 0 + 1 + 2
				"_sum": int64(3),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableIntegerArrayWithSumWithOffsetWithLimitWithOrderDesc(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array, ordered offsetted limited sum of integer array",
		Query: `query {
					users {
						Name
						_sum(TestScores: {offset: 1, limit: 3, order: DESC})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "Shahzad",
					"TestScores": [null, 2, 5, 1, 0, 7]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "Shahzad",
				// 5 + 2 + 1
				"_sum": int64(8),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArrayWithSumWithOffsetWithLimitWithOrderAsc(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array, ordered offsetted limited sum of integer array",
		Query: `query {
					users {
						Name
						_sum(FavouriteFloats: {offset: 1, limit: 3, order: ASC})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "Shahzad",
					"FavouriteFloats": [3.1425, 0.00000000001, 10, 2.718, 0.577, 6.283]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "Shahzad",
				// 0.577 + 2.718 + 3.1425
				"_sum": float64(6.4375),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArrayWithSumWithOffsetWithLimitWithOrderDesc(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array, ordered offsetted limited sum of integer array",
		Query: `query {
					users {
						Name
						_sum(FavouriteFloats: {offset: 1, limit: 3, order: DESC})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "Shahzad",
					"FavouriteFloats": [3.1425, 0.00000000001, 10, 2.718, 0.577, 6.283]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "Shahzad",
				// 6.283 + 3.1425 + 2.718
				"_sum": float64(12.1435),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableFloatArrayWithSumWithOffsetWithLimitWithOrderAsc(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array, ordered offsetted limited sum of integer array",
		Query: `query {
					users {
						Name
						_sum(PageRatings: {offset: 1, limit: 3, order: ASC})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "Shahzad",
					"PageRatings": [3.1425, null, 10, 2.718, 0.577, 6.283]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "Shahzad",
				// 0.577 + 2.718 + 3.1425
				"_sum": float64(6.4375),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableFloatArrayWithSumWithOffsetWithLimitWithOrderDesc(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array, ordered offsetted limited sum of integer array",
		Query: `query {
					users {
						Name
						_sum(PageRatings: {offset: 1, limit: 3, order: DESC})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "Shahzad",
					"PageRatings": [3.1425, null, 10, 2.718, 0.577, 6.283]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "Shahzad",
				// 6.283 + 3.1425 + 2.718
				"_sum": float64(12.1435),
			},
		},
	}

	executeTestCase(t, test)
}
