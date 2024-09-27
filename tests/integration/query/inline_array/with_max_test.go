// Copyright 2024 Democratized Data Foundation
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
	"math"
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryInlineIntegerArray_WithMaxAndNullArray_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, max of nil integer array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": null
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_max(favouriteIntegers: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"_max": int64(math.MinInt64),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArray_WithMaxAndEmptyArray_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, max of empty integer array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": []
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_max(favouriteIntegers: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"_max": int64(math.MinInt64),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArray_WithMaxAndPopulatedArray_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, max of integer array",
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
						_max(favouriteIntegers: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							"_max": int64(2),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableIntegerArray_WithMaxAndPopulatedArray_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, max of nillable integer array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"testScores": [-1, 2, null, 1, 0]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_max(testScores: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							"_max": int64(2),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArray_WithMaxAndNullArray_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, max of nil float array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteFloats": null
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_max(favouriteFloats: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"_max": float64(-math.MaxFloat64),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArray_WithMaxAndEmptyArray_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, max of empty float array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteFloats": []
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_max(favouriteFloats: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"_max": float64(-math.MaxFloat64),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArray_WithMaxAndPopulatedArray_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, max of float array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteFloats": [3.1425, 0.00000000001, 10]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_max(favouriteFloats: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"_max": float64(10),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableFloatArray_WithMaxAndPopulatedArray_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, max of nillable float array",
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
						_max(pageRatings: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							"_max": float64(10),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
