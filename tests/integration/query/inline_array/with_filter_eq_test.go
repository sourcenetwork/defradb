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
	"github.com/sourcenetwork/defradb/tests/multiplier"
)

func TestQueryInlineBooleanArray_WithEqFilter_ReturnsResults(t *testing.T) {
	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"likedIndexes": [true, false]
				}`,
			},
			&action.AddDoc{
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
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"likedIndexes": [true, false]
				}`,
			},
			&action.AddDoc{
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
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"indexLikesDislikes": [true, null, false]
				}`,
			},
			&action.AddDoc{
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
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"indexLikesDislikes": [true, null, false]
				}`,
			},
			&action.AddDoc{
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
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteIntegers": [1, 2, 3]
				}`,
			},
			&action.AddDoc{
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
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteIntegers": [1, 2, 3]
				}`,
			},
			&action.AddDoc{
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
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"testScores": [90, null, 85]
				}`,
			},
			&action.AddDoc{
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
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"testScores": [90, null, 85]
				}`,
			},
			&action.AddDoc{
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
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteFloats": [3.14, 2.71, 1.41]
				}`,
			},
			&action.AddDoc{
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
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteFloats": [3.14, 2.71, 1.41]
				}`,
			},
			&action.AddDoc{
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
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"pageRatings": [4.5, null, 3.2]
				}`,
			},
			&action.AddDoc{
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
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"pageRatings": [4.5, null, 3.2]
				}`,
			},
			&action.AddDoc{
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
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"preferredStrings": ["apple", "banana", "cherry"]
				}`,
			},
			&action.AddDoc{
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
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"preferredStrings": ["apple", "banana", "cherry"]
				}`,
			},
			&action.AddDoc{
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
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"pageHeaders": ["intro", null, "conclusion"]
				}`,
			},
			&action.AddDoc{
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
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"pageHeaders": ["intro", null, "conclusion"]
				}`,
			},
			&action.AddDoc{
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
