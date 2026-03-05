// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package inline_array

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryInlineIntegerArray_WithMaxWithOffsetWithLimitWithOrderAsc_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteIntegers": [-1, 2, 5, 1, 0, 7]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						MAX(favouriteIntegers: {offset: 1, limit: 3, order: ASC})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							// 0, 1, 2
							"MAX": int64(2),
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
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteIntegers": [-1, 2, 5, 1, 0, 7]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						MAX(favouriteIntegers: {offset: 1, limit: 3, order: DESC})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							// 5, 2, 1
							"MAX": int64(5),
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
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"testScores": [2, null, 5, 1, 0, 7]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						MAX(testScores: {offset: 1, limit: 3, order: ASC})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							// 0, 1, 2
							"MAX": int64(2),
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
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"testScores": [null, 2, 5, 1, 0, 7]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						MAX(testScores: {offset: 1, limit: 3, order: DESC})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							// 5, 2, 1
							"MAX": int64(5),
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
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteFloats": [3.1425, 0.00000000001, 10, 2.718, 0.577, 6.283]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						MAX(favouriteFloats: {offset: 1, limit: 3, order: ASC})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							// 0.577, 2.718, 3.1425
							"MAX": float64(3.1425),
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
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteFloats": [3.1425, 0.00000000001, 10, 2.718, 0.577, 6.283]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						MAX(favouriteFloats: {offset: 1, limit: 3, order: DESC})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							// 6.283, 3.1425, 2.718
							"MAX": float64(6.283),
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
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"pageRatings": [3.1425, null, 10, 2.718, 0.577, 6.283]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						MAX(pageRatings: {offset: 1, limit: 3, order: ASC})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							// 0.577, 2.718, 3.1425
							"MAX": float64(3.1425),
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
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"pageRatings": [3.1425, null, 10, 2.718, 0.577, 6.283]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						MAX(pageRatings: {offset: 1, limit: 3, order: DESC})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							// 6.283, 3.1425, 2.718
							"MAX": float64(6.283),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
