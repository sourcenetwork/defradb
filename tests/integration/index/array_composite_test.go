// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package index

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestArrayCompositeIndex_WithFilterOnIndexedArrayUsingAny_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {name: {_eq: "Shahzad"}, numbers: {_any: {_eq: 30}}, age: {_eq: 30}}) {
			_docID
			numbers
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(includes: [{name: "name"}, {name: "numbers"}, {name: "age"}]) {
						name: String 
						numbers: [Int!] 
						age: Int
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 30, 20],
					"age": 30
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [30, 40, 50, 30],
					"age": 30
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [40, 50, 30],
					"age": 60
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [1, 2, 3],
					"age": 30
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"_docID":  testUtils.NewDocIndex(0, 1),
							"numbers": []int64{30, 40, 50, 30},
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayCompositeIndex_WithFilterOnIndexedArrayUsingAll_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {name: {_eq: "Shahzad"}, numbers: {_all: {_gt: 1}}, age: {_eq: 30}}) {
			_docID
			numbers
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(includes: [{name: "name"}, {name: "numbers"}, {name: "age"}]) {
						name: String 
						numbers: [Int!] 
						age: Int
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 30, 20],
					"age": 30
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [30, 40],
					"age": 30
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [50],
					"age": 60
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [1, 2],
					"age": 30
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"_docID":  testUtils.NewDocIndex(0, 1),
							"numbers": []int64{30, 40},
						},
					},
				},
			},
			testUtils.Request{
				Request: makeExplainQuery(req),
				// all "Shahzad" users have in total 5 numbers
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(5),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayCompositeIndex_WithFilterOnIndexedArrayUsingNone_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {name: {_eq: "Shahzad"}, numbers: {_none: {_eq: 3}}, age: {_eq: 30}}) {
			_docID
			numbers
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(includes: [{name: "name"}, {name: "numbers"}, {name: "age"}]) {
						name: String 
						numbers: [Int!] 
						age: Int
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 30, 20],
					"age": 30
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [30, 40],
					"age": 30
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [50],
					"age": 60
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [2, 3],
					"age": 30
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"_docID":  testUtils.NewDocIndex(0, 1),
							"numbers": []int64{30, 40},
						},
					},
				},
			},
			testUtils.Request{
				Request: makeExplainQuery(req),
				// all "Shahzad" users have in total 5 numbers
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(5),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
