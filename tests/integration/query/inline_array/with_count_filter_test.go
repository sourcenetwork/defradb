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
	test := testUtils.TestCase{
		Description: "Simple inline array, filtered count of bool array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"likedIndexes": [true, true, false, true]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_count(likedIndexes: {filter: {_eq: true}})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "Shahzad",
							"_count": 3,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableBoolArrayWithCountWithFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with filter, count of nillable bool array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"indexLikesDislikes": [true, true, false, null]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_count(indexLikesDislikes: {filter: {_eq: true}})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "John",
							"_count": 2,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArrayWithCountWithFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array, filtered count of integer array",
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
						_count(favouriteIntegers: {filter: {_gt: 0}})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "Shahzad",
							"_count": 2,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableIntegerArrayWithCountWithFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array, filtered count of integer array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"testScores": [-1, 2, 1, null, 0]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_count(testScores: {filter: {_gt: 0}})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "Shahzad",
							"_count": 2,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArrayWithsWithCountWithAndFilterAndPopulatedArray(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array, filtered count of integer array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteIntegers": [-1, 2, -1, 1, 0, -2]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_count(favouriteIntegers: {filter: {_and: [{_gt: -2}, {_lt: 2}]}})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "Shahzad",
							"_count": 4,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArrayWithCountWithFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array, filtered count of float array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteFloats": [3.1425, 0.00000000001, 10]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_count(favouriteFloats: {filter: {_lt: 9}})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "Shahzad",
							"_count": 2,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableFloatArrayWithCountWithFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array, filtered count of nillable float array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"pageRatings": [3.1425, 0.00000000001, 10, null]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_count(pageRatings: {filter: {_lt: 9}})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "Shahzad",
							"_count": 2,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineStringArrayWithCountWithFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array, filtered count of string array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"preferredStrings": ["", "the previous", "the first", "empty string"]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_count(preferredStrings: {filter: {_in: ["", "the first"]}})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "Shahzad",
							"_count": 2,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableStringArrayWithCountWithFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array, filtered count of string array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"pageHeaders": ["", "the previous", null, "empty string"]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_count(pageHeaders: {filter: {_in: ["", "the first", "empty string"]}})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "Shahzad",
							"_count": 2,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
