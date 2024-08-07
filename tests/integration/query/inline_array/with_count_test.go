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

func TestQueryInlineIntegerArrayWithCountAndNullArray(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, count of nil integer array",
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
						_count(favouriteIntegers: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "John",
							"_count": 0,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArrayWithCountAndEmptyArray(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, count of empty integer array",
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
						_count(favouriteIntegers: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "John",
							"_count": 0,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArrayWithCountAndPopulatedArray(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, count of integer array",
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
						_count(favouriteIntegers: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "Shahzad",
							"_count": 5,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableBoolArrayWithCountAndPopulatedArray(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, count of nillable bool array",
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
						_count(indexLikesDislikes: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "John",
							"_count": 4,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
