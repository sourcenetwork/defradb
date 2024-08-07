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
	test := testUtils.TestCase{
		Description: "Simple inline array, filtered average of integer array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteIntegers": [-1, 2, -1, 1, 0]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_avg(favouriteIntegers: {filter: {_gt: 0}})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							"_avg": float64(1.5),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableIntegerArrayWithAverageWithFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with filter, average of populated nillable integer array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"testScores": [-1, null, 13, 0]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_avg(testScores: {filter: {_gt: -1}})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"_avg": float64(6.5),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArrayWithAverageWithFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array, filtered average of float array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteFloats": [3.4, 3.6, 10]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_avg(favouriteFloats: {filter: {_lt: 9}})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							"_avg": 3.5,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableFloatArrayWithAverageWithFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array, filtered average of nillable float array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"pageRatings": [3.4, 3.6, 10, null]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_avg(pageRatings: {filter: {_lt: 9}})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							"_avg": 3.5,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
