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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryInlineIntegerArrayWithSumWithOffsetWithLimitWithOrderAsc(t *testing.T) {
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
						SUM(favouriteIntegers: {offset: 1, limit: 3, order: ASC})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							// 0 + 1 + 2
							"SUM": int64(3),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArrayWithSumWithOffsetWithLimitWithOrderDesc(t *testing.T) {
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
						SUM(favouriteIntegers: {offset: 1, limit: 3, order: DESC})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							// 5 + 2 + 1
							"SUM": int64(8),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableIntegerArrayWithSumWithOffsetWithLimitWithOrderAsc(t *testing.T) {
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
						SUM(testScores: {offset: 1, limit: 3, order: ASC})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							// 0 + 1 + 2
							"SUM": int64(3),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableIntegerArrayWithSumWithOffsetWithLimitWithOrderDesc(t *testing.T) {
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
						SUM(testScores: {offset: 1, limit: 3, order: DESC})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							// 5 + 2 + 1
							"SUM": int64(8),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArrayWithSumWithOffsetWithLimitWithOrderAsc(t *testing.T) {
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
						SUM(favouriteFloats: {offset: 1, limit: 3, order: ASC})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							// 0.577 + 2.718 + 3.1425
							"SUM": float64(6.4375),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArrayWithSumWithOffsetWithLimitWithOrderDesc(t *testing.T) {
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
						SUM(favouriteFloats: {offset: 1, limit: 3, order: DESC})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							// 6.283 + 3.1425 + 2.718
							"SUM": float64(12.1435),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableFloatArrayWithSumWithOffsetWithLimitWithOrderAsc(t *testing.T) {
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
						SUM(pageRatings: {offset: 1, limit: 3, order: ASC})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							// 0.577 + 2.718 + 3.1425
							"SUM": float64(6.4375),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableFloatArrayWithSumWithOffsetWithLimitWithOrderDesc(t *testing.T) {
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
						SUM(pageRatings: {offset: 1, limit: 3, order: DESC})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							// 6.283 + 3.1425 + 2.718
							"SUM": float64(12.1435),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
