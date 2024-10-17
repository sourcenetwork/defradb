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
					type User @index(includes: [{field: "name"}, {field: "numbers"}, {field: "age"}]) {
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
					type User @index(includes: [{field: "name"}, {field: "numbers"}, {field: "age"}]) {
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
					type User @index(includes: [{field: "name"}, {field: "numbers"}, {field: "age"}]) {
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

func TestArrayCompositeIndex_With2ConsecutiveArrayFields_Succeed(t *testing.T) {
	req := `query {
		User(filter: {name: {_eq: "Shahzad"}, numbers: {_any: {_eq: 30}}, hobbies: {_any: {_eq: "sports"}} age: {_eq: 30}}) {
			_docID
			numbers
			hobbies
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(includes: [{field: "name"}, {field: "numbers"}, {field: "hobbies"}, {field: "age"}]) {
						name: String 
						numbers: [Int!] 
						hobbies: [String!] 
						age: Int
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 30, 20],
					"hobbies": ["sports", "books"],
					"age": 30
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [30, 40],
					"hobbies": ["sports", "books"],
					"age": 30
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [50],
					"hobbies": ["books", "movies"],
					"age": 60
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [2, 3],
					"hobbies": ["sports", "movies", "books"],
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
							"hobbies": []string{"sports", "books"},
						},
					},
				},
			},
			testUtils.Request{
				Request: makeExplainQuery(req),
				// all "Shahzad" users have in total 5 numbers
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayCompositeIndex_With2SeparateArrayFields_Succeed(t *testing.T) {
	req := `query {
		User(filter: {name: {_eq: "Shahzad"}, numbers: {_any: {_eq: 30}}, hobbies: {_any: {_eq: "sports"}} age: {_eq: 30}}) {
			_docID
			numbers
			hobbies
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(includes: [{field: "numbers"}, {field: "name"}, {field: "age"}, {field: "hobbies"}]) {
						name: String 
						numbers: [Int!] 
						hobbies: [String!] 
						age: Int
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 30, 20],
					"hobbies": ["sports", "books"],
					"age": 30
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [30, 40],
					"hobbies": ["sports", "books"],
					"age": 30
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [50],
					"hobbies": ["books", "movies"],
					"age": 60
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [2, 3],
					"hobbies": ["sports", "movies", "books"],
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
							"hobbies": []string{"sports", "books"},
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

func TestArrayCompositeIndex_WithAnyNoneAll_Succeed(t *testing.T) {
	req := `query {
		User(filter: {
			numbers1: {_all: {_gt: 0}}, 
			numbers2: {_none: {_eq: 40}}, 
			numbers3: {_any: {_le: 200}}
		}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(includes: [{field: "numbers1"}, {field: "numbers2"}, {field: "numbers3"}]) {
						name: String 
						numbers1: [Int!] 
						numbers2: [Int!] 
						numbers3: [Int!] 
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers1": [1, 2, 3],
					"numbers2": [10, 20, 30],
					"numbers3": [100, 200, 300]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers1": [2, 3, 4],
					"numbers2": [20, 30, 40],
					"numbers3": [200, 300, 400]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"numbers1": [0, 1],
					"numbers2": [90],
					"numbers3": [900]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Islam",
					"numbers1": [6, 7, 8],
					"numbers2": [10, 70, 80],
					"numbers3": [100, 700, 800]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"numbers1": [1, 4, 5, 8],
					"numbers2": [60, 80],
					"numbers3": [600, 800]
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
						{"name": "Islam"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayCompositeIndexUpdate_With2ArrayFields_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(includes: [{field: "name"}, {field: "numbers"}, {field: "hobbies"}]) {
						name: String 
						numbers: [Int!] 
						hobbies: [String!] 
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 30, 20, 40],
					"hobbies": ["sports", "books"]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [30, 40, 30],
					"hobbies": ["sports", "books"]
				}`,
			},
			testUtils.UpdateDoc{
				DocID: 1,
				Doc: `{
					"name": "Shahzad",
					"numbers": [40, 50, 50],
					"hobbies": ["books", "movies", "books", "movies"]
				}`,
			},
			testUtils.Request{
				Request: `query {
					User(filter: {numbers: {_any: {_eq: 30}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{{"name": "John"}},
				},
			},
			testUtils.Request{
				Request: `query {
					User(filter: {numbers: {_any: {_eq: 40}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
						{"name": "Shahzad"},
					},
				},
			},
			testUtils.Request{
				Request: `query {
					User(filter: {numbers: {_any: {_eq: 50}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{{"name": "Shahzad"}},
				},
			},
			testUtils.Request{
				Request: `query {
					User(filter: {numbers: {_any: {_gt: 0}}, hobbies: {_any: {_eq: "sports"}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{{"name": "John"}},
				},
			},
			testUtils.Request{
				Request: `query {
					User(filter: {numbers: {_any: {_gt: 0}}, hobbies: {_any: {_eq: "books"}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
						{"name": "Shahzad"},
					},
				},
			},
			testUtils.Request{
				Request: `query {
					User(filter: {numbers: {_any: {_gt: 0}}, hobbies: {_any: {_eq: "movies"}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{{"name": "Shahzad"}},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayCompositeIndexDelete_With2ConsecutiveArrayFields_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(includes: [{field: "name"}, {field: "numbers"}, {field: "hobbies"}]) {
						name: String 
						numbers: [Int!] 
						hobbies: [String!] 
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 30, 20],
					"hobbies": ["sports", "books"]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [30, 40, 30, 50],
					"hobbies": ["sports", "books", "sports", "movies"]
				}`,
			},
			testUtils.DeleteDoc{DocID: 1},
			testUtils.Request{
				Request: `query {
					User(filter: {numbers: {_any: {_eq: 30}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{{"name": "John"}},
				},
			},
			testUtils.Request{
				Request: `query {
					User(filter: {numbers: {_any: {_gt: 0}}, hobbies: {_any: {_eq: "sports"}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{{"name": "John"}},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
