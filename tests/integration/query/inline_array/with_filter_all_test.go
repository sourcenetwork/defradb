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

func TestQueryInlineStringArray_WithAllFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"pageHeaders": ["first", "second"]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"pageHeaders": [null, "second"]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {pageHeaders: {_all: {_ne: null}}}) {
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

func TestQueryInlineNotNullStringArray_WithAllFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"preferredStrings": ["first", "second"]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"preferredStrings": ["", "second"]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {preferredStrings: {_all: {_ne: ""}}}) {
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

func TestQueryInlineIntArray_WithAllFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"testScores": [50, 80]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"testScores": [null, 60]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {testScores: {_all: {_ne: null}}}) {
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

func TestQueryInlineNotNullIntArray_WithAllFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"testScores": [50, 80]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"testScores": [0, 60]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {testScores: {_all: {_lt: 70}}}) {
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

func TestQueryInlineFloatArray_WithAllFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"pageRatings": [50, 80]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"pageRatings": [null, 60]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {pageRatings: {_all: {_ne: null}}}) {
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

func TestQueryInlineNotNullFloatArray_WithAllFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"pageRatings": [50, 80]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"pageRatings": [0, 60]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {pageRatings: {_all: {_lt: 70}}}) {
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

func TestQueryInlineBooleanArray_WithAllFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"indexLikesDislikes": [false, false]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"indexLikesDislikes": [null, true]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {indexLikesDislikes: {_all: {_ne: null}}}) {
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

func TestQueryInlineNotNullBooleanArray_WithAllFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"likedIndexes": [false, false]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"likedIndexes": [true, true]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {likedIndexes: {_all: {_eq: true}}}) {
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

func TestQueryInlineStringArray_WithAllFilterAndNullValue_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Islam",
					"pageHeaders": null
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {pageHeaders: {_all: {_eq: null}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	executeTestCase(t, test)
}
