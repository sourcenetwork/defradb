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

func TestQueryInlineIntegerArrayWithAverageAndNullArray(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array with no filter, average of nil integer array",
		Query: `query {
					users {
						Name
						_avg(FavouriteIntegers: {})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"FavouriteIntegers": null
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"Name": "John",
				"_avg": float64(0),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArrayWithAverageAndEmptyArray(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array with no filter, average of empty integer array",
		Query: `query {
					users {
						Name
						_avg(FavouriteIntegers: {})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"FavouriteIntegers": []
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"Name": "John",
				"_avg": float64(0),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArrayWithAverageAndZeroArray(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array with no filter, average of zero integer array",
		Query: `query {
					users {
						Name
						_avg(FavouriteIntegers: {})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"FavouriteIntegers": [0, 0, 0]
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"Name": "John",
				"_avg": float64(0),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArrayWithAverageAndPopulatedArray(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array with no filter, average of populated integer array",
		Query: `query {
					users {
						Name
						_avg(FavouriteIntegers: {})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"FavouriteIntegers": [-1, 0, 9, 0]
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"Name": "John",
				"_avg": float64(2),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableIntegerArrayWithAverageAndPopulatedArray(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array with no filter, average of populated nillable integer array",
		Query: `query {
					users {
						Name
						_avg(TestScores: {})
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
		Results: []map[string]interface{}{
			{
				"Name": "John",
				"_avg": float64(4),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArrayWithAverageAndNullArray(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array with no filter, average of nil float array",
		Query: `query {
					users {
						Name
						_avg(FavouriteFloats: {})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"FavouriteFloats": null
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"Name": "John",
				"_avg": float64(0),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArrayWithAverageAndEmptyArray(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array with no filter, average of empty float array",
		Query: `query {
					users {
						Name
						_avg(FavouriteFloats: {})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"FavouriteFloats": []
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"Name": "John",
				"_avg": float64(0),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArrayWithAverageAndZeroArray(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array with no filter, average of zero float array",
		Query: `query {
					users {
						Name
						_avg(FavouriteFloats: {})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"FavouriteFloats": [0, 0, 0]
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"Name": "John",
				"_avg": float64(0),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArrayWithAverageAndPopulatedArray(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array with no filter, average of populated float array",
		Query: `query {
					users {
						Name
						_avg(FavouriteFloats: {})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"FavouriteFloats": [-0.1, 0, 0.9, 0]
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"Name": "John",
				"_avg": float64(0.2),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableFloatArrayWithAverageAndPopulatedArray(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array with no filter, average of populated nillable float array",
		Query: `query {
					users {
						Name
						_avg(PageRatings: {})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"PageRatings": [-0.1, 0, 0.9, 0, null]
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"Name": "John",
				"_avg": float64(0.2),
			},
		},
	}

	executeTestCase(t, test)
}
