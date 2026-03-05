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

func TestQueryInlineIntegerArray_WithMinWithFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteIntegers": [-1, 2, -1, 1, 0]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						MIN(favouriteIntegers: {filter: {_gt: 0}})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							"MIN":  int64(1),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableIntegerArray_WithMinWithFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"testScores": [-1, 2, null, 1, 0]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						MIN(testScores: {filter: {_gt: 0}})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							"MIN":  int64(1),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArray_WithMinWithFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteFloats": [3.1425, 0.00000000001, 10]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						MIN(favouriteFloats: {filter: {_gt: 1}})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							"MIN":  float64(3.1425),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableFloatArray_WithMinWithFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"pageRatings": [3.1425, 0.00000000001, 10, null]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						MIN(pageRatings: {filter: {_gt: 1}})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							"MIN":  float64(3.1425),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
