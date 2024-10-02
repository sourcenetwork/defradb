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
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryInlineIntegerArray_WithMaxWithOffsetWithLimitWithOrderAsc_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array, ordered offsetted limited max of integer array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteIntegers": [-1, 2, 5, 1, 0, 7]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_max(favouriteIntegers: {offset: 1, limit: 3, order: ASC})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							// 0, 1, 2
							"_max": int64(2),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArray_WithMaxWithOffsetWithLimitWithOrderDesc_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array, ordered offsetted limited max of integer array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteIntegers": [-1, 2, 5, 1, 0, 7]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_max(favouriteIntegers: {offset: 1, limit: 3, order: DESC})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							// 5, 2, 1
							"_max": int64(5),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableIntegerArray_WithMaxWithOffsetWithLimitWithOrderAsc_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array, ordered offsetted limited max of integer array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"testScores": [2, null, 5, 1, 0, 7]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_max(testScores: {offset: 1, limit: 3, order: ASC})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							// 0, 1, 2
							"_max": int64(2),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableIntegerArray_WithMaxWithOffsetWithLimitWithOrderDesc_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array, ordered offsetted limited max of integer array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"testScores": [null, 2, 5, 1, 0, 7]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_max(testScores: {offset: 1, limit: 3, order: DESC})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							// 5, 2, 1
							"_max": int64(5),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArray_WithMaxWithOffsetWithLimitWithOrderAsc_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array, ordered offsetted limited max of integer array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteFloats": [3.1425, 0.00000000001, 10, 2.718, 0.577, 6.283]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_max(favouriteFloats: {offset: 1, limit: 3, order: ASC})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							// 0.577, 2.718, 3.1425
							"_max": float64(3.1425),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArray_WithMaxWithOffsetWithLimitWithOrderDesc_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array, ordered offsetted limited max of integer array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteFloats": [3.1425, 0.00000000001, 10, 2.718, 0.577, 6.283]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_max(favouriteFloats: {offset: 1, limit: 3, order: DESC})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							// 6.283, 3.1425, 2.718
							"_max": float64(6.283),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableFloatArray_WithMaxWithOffsetWithLimitWithOrderAsc_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array, ordered offsetted limited max of integer array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"pageRatings": [3.1425, null, 10, 2.718, 0.577, 6.283]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_max(pageRatings: {offset: 1, limit: 3, order: ASC})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							// 0.577, 2.718, 3.1425
							"_max": float64(3.1425),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableFloatArray_WithMaxWithOffsetWithLimitWithOrderDesc_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array, ordered offsetted limited max of integer array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"pageRatings": [3.1425, null, 10, 2.718, 0.577, 6.283]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_max(pageRatings: {offset: 1, limit: 3, order: DESC})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							// 6.283, 3.1425, 2.718
							"_max": float64(6.283),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
