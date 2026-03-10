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

package index

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestJSONArrayIndex_WithDifferentElementValuesAndTypes_ShouldFetchCorrectlyUsingIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {numbers: {_any: {_eq: 4}}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
					"custom": map[string]any{
						"numbers": []int{3, 5, 7},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Islam",
					"custom": map[string]any{
						"numbers": []int{3},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Shahzad",
					"custom": map[string]any{
						"numbers": []int{4, 8, 4, 4, 5, 4},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Keenan",
					"custom": map[string]any{
						"numbers": []any{8, nil},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Bruno",
					"custom": map[string]any{
						"numbers": []any{10, "str", true},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"custom": map[string]any{
						"numbers": 4,
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Chris",
					"custom": map[string]any{
						"numbers": nil,
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Fred",
					"custom": map[string]any{
						"height": 198,
					},
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONArrayIndex_WithAnyEqFilter_ShouldNotConsiderThem(t *testing.T) {
	req := `query {
		User(filter: {custom: {numbers: {_any: {_eq: 4}}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
					"custom": map[string]any{
						"numbers": []any{3, 5, []int{9, 4}, 7},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Islam",
					"custom": map[string]any{
						"numbers": []any{0, []int{2}, 4},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Fred",
					"custom": map[string]any{
						"numbers": []any{3, 5, []any{1, 0, []int{9, 4, 6}}, 7},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"custom": map[string]any{
						"numbers": []any{1, 2, []int{8, 6}, 10},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Shahzad",
					"custom": map[string]any{
						"numbers": 4,
					},
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Islam"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONArrayIndex_WithAnyAndComparisonFilter_ShouldNotConsiderThem(t *testing.T) {
	req := `query {
		User(filter: {custom: {numbers: {_any: {_gt: 4}}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
					"custom": map[string]any{
						"numbers": []any{3, 5, 7},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Islam",
					"custom": map[string]any{
						"numbers": []any{0, []int{6}, 4},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"custom": map[string]any{
						"numbers": 5,
					},
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONArrayIndex_WithNoneEqFilter_ShouldFetchCorrectlyUsingIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {numbers: {_none: {_eq: 4}}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
					"custom": map[string]any{
						"numbers": []int{3},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Shahzad",
					"custom": map[string]any{
						"numbers": []int{4, 8},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Islam",
					"custom": map[string]any{
						"numbers": []any{8, nil},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Fred",
					"custom": map[string]any{
						"numbers": []any{1, []int{4}},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"custom": map[string]any{
						"numbers": 4,
					},
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
						{"name": "Fred"},
						{"name": "Islam"},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				Request: makeExplainQuery(req),
				// We don't use index for _none operator
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONArrayIndex_WithNoneEqAndComparisonFilter_ShouldFetchCorrectlyUsingIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {numbers: {_none: {_gt: 4}}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
					"custom": map[string]any{
						"numbers": []int{3},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Shahzad",
					"custom": map[string]any{
						"numbers": []int{3, 8},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Islam",
					"custom": map[string]any{
						"numbers": []any{2, nil},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Fred",
					"custom": map[string]any{
						"numbers": []any{1, []int{5}},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"custom": map[string]any{
						"numbers": 5,
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Bruno",
					"custom": map[string]any{
						"numbers": nil,
					},
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
						{"name": "Fred"},
						{"name": "Islam"},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				Request: makeExplainQuery(req),
				// We don't use index for _none operator
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONArrayIndex_WithAllEqFilter_ShouldFetchCorrectlyUsingIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {numbers: {_all: {_eq: 4}}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
					"custom": map[string]any{
						"numbers": []int{3, 4},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Shahzad",
					"custom": map[string]any{
						"numbers": []any{4, []int{4, 8}},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Islam",
					"custom": map[string]any{
						"numbers": 4,
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Fred",
					"custom": map[string]any{
						"numbers": []any{4, []any{4, []int{4}}},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Bruno",
					"custom": map[string]any{
						"numbers": []any{4, 4, 4},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"custom": map[string]any{
						"numbers": 3,
					},
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Bruno"},
					},
				},
			},
			&action.Request{
				Request: makeExplainQuery(req),
				// 4 docs have the value 4 in the numbers array
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(4),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONArrayIndex_WithAllEqAndComparisonFilter_ShouldFetchCorrectlyUsingIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {numbers: {_all: {_gt: 4}}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
					"custom": map[string]any{
						"numbers": []int{3, 7},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Shahzad",
					"custom": map[string]any{
						"numbers": []any{5, []int{6}},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"custom": map[string]any{
						"numbers": []any{7, 8},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Islam",
					"custom": map[string]any{
						"numbers": 8,
					},
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Andy"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(4),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
