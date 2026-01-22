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
	"github.com/sourcenetwork/defradb/tests/multiplier"
)

func TestQueryInlineBooleanArray_WithEqFilter_ReturnsResults(t *testing.T) {
	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"likedIndexes": [true, false]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"likedIndexes": [true, true]
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {likedIndexes: {_eq: [true, false]}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineBooleanArray_WithNeqFilter_ReturnsResults(t *testing.T) {
	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"likedIndexes": [true, false]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"likedIndexes": [true, true]
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {likedIndexes: {_neq: [true, false]}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNullableBooleanArray_WithEqFilter_ReturnsResults(t *testing.T) {
	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"indexLikesDislikes": [true, null, false]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"indexLikesDislikes": [true, true]
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {indexLikesDislikes: {_eq: [true, null, false]}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNullableBooleanArray_WithNeqFilter_ReturnsResults(t *testing.T) {
	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"indexLikesDislikes": [true, null, false]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"indexLikesDislikes": [true, true]
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {indexLikesDislikes: {_neq: [true, null, false]}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArray_WithEqFilter_ReturnsResults(t *testing.T) {
	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteIntegers": [1, 2, 3]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"favouriteIntegers": [4, 5, 6]
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {favouriteIntegers: {_eq: [1, 2, 3]}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArray_WithNeqFilter_ReturnsResults(t *testing.T) {
	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteIntegers": [1, 2, 3]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"favouriteIntegers": [4, 5, 6]
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {favouriteIntegers: {_neq: [1, 2, 3]}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNullableIntegerArray_WithEqFilter_ReturnsResults(t *testing.T) {
	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"testScores": [90, null, 85]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"testScores": [100, 95]
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {testScores: {_eq: [90, null, 85]}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNullableIntegerArray_WithNeqFilter_ReturnsResults(t *testing.T) {
	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"testScores": [90, null, 85]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"testScores": [100, 95]
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {testScores: {_neq: [90, null, 85]}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArray_WithEqFilter_ReturnsResults(t *testing.T) {
	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteFloats": [3.14, 2.71, 1.41]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"favouriteFloats": [1.61, 0.57]
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {favouriteFloats: {_eq: [3.14, 2.71, 1.41]}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArray_WithNeqFilter_ReturnsResults(t *testing.T) {
	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteFloats": [3.14, 2.71, 1.41]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"favouriteFloats": [1.61, 0.57]
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {favouriteFloats: {_neq: [3.14, 2.71, 1.41]}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
func TestQueryInlineNullableFloatArray_WithEqFilter_ReturnsResults(t *testing.T) {
	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"pageRatings": [4.5, null, 3.2]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"pageRatings": [5.0, 4.8]
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {pageRatings: {_eq: [4.5, null, 3.2]}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNullableFloatArray_WithNeqFilter_ReturnsResults(t *testing.T) {
	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"pageRatings": [4.5, null, 3.2]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"pageRatings": [5.0, 4.8]
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {pageRatings: {_neq: [4.5, null, 3.2]}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineStringArray_WithEqFilter_ReturnsResults(t *testing.T) {
	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"preferredStrings": ["apple", "banana", "cherry"]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"preferredStrings": ["dog", "elephant"]
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {preferredStrings: {_eq: ["apple", "banana", "cherry"]}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineStringArray_WithNeqFilter_ReturnsResults(t *testing.T) {
	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"preferredStrings": ["apple", "banana", "cherry"]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"preferredStrings": ["dog", "elephant"]
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {preferredStrings: {_neq: ["apple", "banana", "cherry"]}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNullableStringArray_WithEqFilter_ReturnsResults(t *testing.T) {
	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"pageHeaders": ["intro", null, "conclusion"]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"pageHeaders": ["summary", "details"]
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {pageHeaders: {_eq: ["intro", null, "conclusion"]}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNullableStringArray_WithNeqFilter_ReturnsResults(t *testing.T) {
	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"pageHeaders": ["intro", null, "conclusion"]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"pageHeaders": ["summary", "details"]
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {pageHeaders: {_neq: ["intro", null, "conclusion"]}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
