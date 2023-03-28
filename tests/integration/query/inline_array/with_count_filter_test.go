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
	test := testUtils.RequestTestCase{
		Description: "Simple inline array, filtered count of bool array",
		Request: `query {
					Users {
						name
						_count(likedIndexes: {filter: {_eq: true}})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "Shahzad",
					"likedIndexes": [true, true, false, true]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name":   "Shahzad",
				"_count": 3,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableBoolArrayWithCountWithFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array with filter, count of nillable bool array",
		Request: `query {
					Users {
						name
						_count(indexLikesDislikes: {filter: {_eq: true}})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "John",
					"indexLikesDislikes": [true, true, false, null]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name":   "John",
				"_count": 2,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArrayWithCountWithFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array, filtered count of integer array",
		Request: `query {
					Users {
						name
						_count(favouriteIntegers: {filter: {_gt: 0}})
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
		Results: []map[string]any{
			{
				"name":   "Shahzad",
				"_count": 2,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableIntegerArrayWithCountWithFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array, filtered count of integer array",
		Request: `query {
					Users {
						name
						_count(testScores: {filter: {_gt: 0}})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "Shahzad",
					"testScores": [-1, 2, 1, null, 0]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name":   "Shahzad",
				"_count": 2,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArrayWithsWithCountWithAndFilterAndPopulatedArray(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array, filtered count of integer array",
		Request: `query {
					Users {
						name
						_count(favouriteIntegers: {filter: {_and: [{_gt: -2}, {_lt: 2}]}})
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"name": "Shahzad",
				"favouriteIntegers": [-1, 2, -1, 1, 0, -2]
			}`)},
		},
		Results: []map[string]any{
			{
				"name":   "Shahzad",
				"_count": 4,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArrayWithCountWithFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array, filtered count of float array",
		Request: `query {
					Users {
						name
						_count(favouriteFloats: {filter: {_lt: 9}})
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
		Results: []map[string]any{
			{
				"name":   "Shahzad",
				"_count": 2,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableFloatArrayWithCountWithFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array, filtered count of nillable float array",
		Request: `query {
					Users {
						name
						_count(pageRatings: {filter: {_lt: 9}})
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
		Results: []map[string]any{
			{
				"name":   "Shahzad",
				"_count": 2,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineStringArrayWithCountWithFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array, filtered count of string array",
		Request: `query {
					Users {
						name
						_count(preferredStrings: {filter: {_in: ["", "the first"]}})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "Shahzad",
					"preferredStrings": ["", "the previous", "the first", "empty string"]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name":   "Shahzad",
				"_count": 2,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableStringArrayWithCountWithFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array, filtered count of string array",
		Request: `query {
					Users {
						name
						_count(pageHeaders: {filter: {_in: ["", "the first", "empty string"]}})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "Shahzad",
					"pageHeaders": ["", "the previous", null, "empty string"]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name":   "Shahzad",
				"_count": 2,
			},
		},
	}

	executeTestCase(t, test)
}
