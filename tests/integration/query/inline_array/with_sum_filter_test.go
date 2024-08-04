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

func TestQueryInlineIntegerArrayWithSumWithFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array, filtered sum of integer array",
		Request: `query {
					Users {
						name
						_sum(favouriteIntegers: {filter: {_gt: 0}})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "Shahzad",
					"favouriteIntegers": [-1, 2, -1, 1, 0]
				}`,
			},
		},
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"name": "Shahzad",
					"_sum": int64(3),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableIntegerArrayWithSumWithFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array with filter, sum of nillable integer array",
		Request: `query {
					Users {
						name
						_sum(testScores: {filter: {_gt: 0}})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "Shahzad",
					"testScores": [-1, 2, null, 1, 0]
				}`,
			},
		},
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"name": "Shahzad",
					"_sum": int64(3),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArrayWithSumWithFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array, filtered sum of float array",
		Request: `query {
					Users {
						name
						_sum(favouriteFloats: {filter: {_lt: 9}})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "Shahzad",
					"favouriteFloats": [3.1425, 0.00000000001, 10]
				}`,
			},
		},
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"name": "Shahzad",
					"_sum": 3.14250000001,
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableFloatArrayWithSumWithFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array with filter, sum of nillable float array",
		Request: `query {
					Users {
						name
						_sum(pageRatings: {filter: {_lt: 9}})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "Shahzad",
					"pageRatings": [3.1425, 0.00000000001, 10, null]
				}`,
			},
		},
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"name": "Shahzad",
					"_sum": float64(3.14250000001),
				},
			},
		},
	}

	executeTestCase(t, test)
}
