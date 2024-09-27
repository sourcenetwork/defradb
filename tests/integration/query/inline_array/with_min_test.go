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

func TestQueryInlineIntegerArray_WithMinAndNullArray_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, min of nil integer array",
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
						_min(favouriteIntegers: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"_min": int64(math.MaxInt64),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArray_WithMinAndEmptyArray_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, min of empty integer array",
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
						_min(favouriteIntegers: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"_min": int64(math.MaxInt64),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArray_WithMinAndPopulatedArray_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, min of integer array",
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
						_min(favouriteIntegers: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							"_min": int64(-1),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableIntegerArray_WithMinAndPopulatedArray_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, min of nillable integer array",
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
						_min(testScores: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							"_min": int64(-1),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArray_WithMinAndNullArray_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, min of nil float array",
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
						_min(favouriteFloats: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"_min": float64(math.MaxFloat64),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArray_WithMinAndEmptyArray_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, min of empty float array",
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
						_min(favouriteFloats: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"_min": float64(math.MaxFloat64),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArray_WithMinAndPopulatedArray_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, min of float array",
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
						_min(favouriteFloats: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"_min": float64(0.00000000001),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableFloatArray_WithMinAndPopulatedArray_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, min of nillable float array",
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
						_min(pageRatings: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							"_min": float64(0.00000000001),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
