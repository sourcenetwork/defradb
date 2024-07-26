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

func TestQueryInlineIntegerArrayWithSumAndNullArray(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array with no filter, sum of nil integer array",
		Request: `query {
					Users {
						name
						_sum(favouriteIntegers: {})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "John",
					"favouriteIntegers": null
				}`,
			},
		},
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"name": "John",
					"_sum": int64(0),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArrayWithSumAndEmptyArray(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array with no filter, sum of empty integer array",
		Request: `query {
					Users {
						name
						_sum(favouriteIntegers: {})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "John",
					"favouriteIntegers": []
				}`,
			},
		},
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"name": "John",
					"_sum": int64(0),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArrayWithSumAndPopulatedArray(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array with no filter, sum of integer array",
		Request: `query {
					Users {
						name
						_sum(favouriteIntegers: {})
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
					"_sum": int64(1),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableIntegerArrayWithSumAndPopulatedArray(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array with no filter, sum of nillable integer array",
		Request: `query {
					Users {
						name
						_sum(testScores: {})
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
					"_sum": int64(2),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArrayWithSumAndNullArray(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array with no filter, sum of nil float array",
		Request: `query {
					Users {
						name
						_sum(favouriteFloats: {})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "John",
					"favouriteFloats": null
				}`,
			},
		},
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"name": "John",
					"_sum": float64(0),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArrayWithSumAndEmptyArray(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array with no filter, sum of empty float array",
		Request: `query {
					Users {
						name
						_sum(favouriteFloats: {})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "John",
					"favouriteFloats": []
				}`,
			},
		},
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"name": "John",
					"_sum": float64(0),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArrayWithSumAndPopulatedArray(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array with no filter, sum of float array",
		Request: `query {
					Users {
						name
						_sum(favouriteFloats: {})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "John",
					"favouriteFloats": [3.1425, 0.00000000001, 10]
				}`,
			},
		},
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"name": "John",
					"_sum": float64(13.14250000001),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableFloatArrayWithSumAndPopulatedArray(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array with no filter, sum of nillable float array",
		Request: `query {
					Users {
						name
						_sum(pageRatings: {})
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
					"_sum": float64(13.14250000001),
				},
			},
		},
	}

	executeTestCase(t, test)
}
