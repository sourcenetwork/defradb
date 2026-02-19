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

func TestQueryInlineIntegerArrayWithCountAndNullArray(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": null
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						COUNT(favouriteIntegers: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":  "John",
							"COUNT": 0,
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
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": []
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						COUNT(favouriteIntegers: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":  "John",
							"COUNT": 0,
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
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteIntegers": [-1, 2, -1, 1, 0]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						COUNT(favouriteIntegers: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":  "Shahzad",
							"COUNT": 5,
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
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"name": "John",
					"indexLikesDislikes": [true, true, false, null]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						COUNT(indexLikesDislikes: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":  "John",
							"COUNT": 4,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
