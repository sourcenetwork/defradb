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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryInlineIntegerArray_WithMinAndNullArray_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": null
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						MIN(favouriteIntegers: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"MIN":  nil,
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
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": []
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						MIN(favouriteIntegers: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"MIN":  nil,
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
						MIN(favouriteIntegers: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							"MIN":  int64(-1),
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
						MIN(testScores: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							"MIN":  int64(-1),
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
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteFloats": null
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						MIN(favouriteFloats: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"MIN":  nil,
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
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteFloats": []
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						MIN(favouriteFloats: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"MIN":  nil,
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
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteFloats": [3.1425, 0.00000000001, 10]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						MIN(favouriteFloats: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"MIN":  float64(0.00000000001),
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
						MIN(pageRatings: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							"MIN":  float64(0.00000000001),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
