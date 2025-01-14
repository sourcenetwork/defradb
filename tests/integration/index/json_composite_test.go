// Copyright 2025 Democratized Data Foundation
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

func TestJSONCompositeIndex_JSONWithScalarWithEqFilter_ShouldFetchUsingIndex(t *testing.T) {
	type testCase struct {
		name         string
		req          string
		result       map[string]any
		indexFetches int
	}

	testCases := []testCase{
		{
			name: "Unique combination. Non-unique custom.val",
			req: `query {
				User(filter: {
					custom: {val: {_eq: 3}}, 
					age: {_eq: 25}
				}) {
					name
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "Islam"},
				},
			},
			indexFetches: 1,
		},
		{
			name: "Unique combination. Non-unique age",
			req: `query {
				User(filter: {
					custom: {val: {_eq: 3}}, 
					age: {_eq: 30}
				}) {
					name
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "John"},
				},
			},
			indexFetches: 1,
		},
		{
			name: "Match first part of the composite index",
			req: `query {
				User(filter: {custom: {val: {_eq: 3}}}) {
					name
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "Islam"},
					{"name": "John"},
				},
			},
			indexFetches: 2,
		},
		{
			name: "Non-unique combination",
			req: `query {
				User(filter: {
					custom: {val: {_eq: 5}},
					age: {_eq: 35},
				}) {
					name
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "Addo"},
					{"name": "Keenan"},
				},
			},
			indexFetches: 2,
		},
		{
			name: "Match second part of the composite index",
			req: `query {
				User(filter: {age: {_eq: 40}}) {
					name
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "Bruno"},
				},
			},
			indexFetches: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			test := testUtils.TestCase{
				Actions: []any{
					testUtils.SchemaUpdate{
						Schema: `
							type User @index(includes: [{field: "custom"}, {field: "age"}]) {
								name: String 
								custom: JSON 
								age: Int
							}`,
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "John",
							"custom": map[string]any{
								"val": 3,
							},
							"age": 30,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Islam",
							"custom": map[string]any{
								"val": 3,
							},
							"age": 25,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Shahzad",
							"custom": map[string]any{
								"val": 4,
							},
							"age": 25,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Keenan",
							"custom": map[string]any{
								"val": 5,
							},
							"age": 35,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Addo",
							"custom": map[string]any{
								"val": 5,
							},
							"age": 35,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Bruno",
							"custom": map[string]any{
								"val": 6,
							},
							"age": 40,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Andy",
							"custom": map[string]any{
								"val": nil,
							},
							"age": 50,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Chris",
							"custom": map[string]any{
								"val": 7,
							},
							"age": nil,
						},
					},
					testUtils.Request{
						Request: tc.req,
						Results: tc.result,
					},
					testUtils.Request{
						Request:  makeExplainQuery(tc.req),
						Asserter: testUtils.NewExplainAsserter().WithIndexFetches(tc.indexFetches),
					},
				},
			}

			testUtils.ExecuteTestCase(t, test)
		})
	}
}

func TestJSONCompositeIndex_JSONWithScalarWithOtherFilters_ShouldFetchUsingIndex2(t *testing.T) {
	type testCase struct {
		name         string
		req          string
		result       map[string]any
		indexFetches int
	}

	testCases := []testCase{
		{
			name: "With _le and _gt filters",
			req: `query {
				User(filter: {
					age: {_le: 35},
					custom: {val: {_gt: 4}}
				}) {
					name
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "Keenan"},
					{"name": "Addo"},
				},
			},
			indexFetches: 8,
		},
		{
			name: "With _lt and _eq filters",
			req: `query {
				User(filter: {
					age: {_lt: 100},
					custom: {val: {_eq: null}}
				}) {
					name
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "Andy"},
				},
			},
			indexFetches: 8,
		},
		{
			name: "With _ne and _ge filters",
			req: `query {
				User(filter: {
					_and: [{ age: {_ne: 35} }, { age: {_ne: 40} }],
					custom: {val: {_ge: 5}} 
				}) {
					name
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "Chris"},
				},
			},
			// we shouldn't use index as _ne filter is present
			indexFetches: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			test := testUtils.TestCase{
				Actions: []any{
					testUtils.SchemaUpdate{
						Schema: `
							type User @index(includes: [{field: "age"}, {field: "custom"}]) {
								name: String 
								custom: JSON 
								age: Int
							}`,
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "John",
							"custom": map[string]any{
								"val": 3,
							},
							"age": 30,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Islam",
							"custom": map[string]any{
								"val": 3,
							},
							"age": 25,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Shahzad",
							"custom": map[string]any{
								"val": 4,
							},
							"age": 25,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Keenan",
							"custom": map[string]any{
								"val": 5,
							},
							"age": 35,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Addo",
							"custom": map[string]any{
								"val": 6,
							},
							"age": 35,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Bruno",
							"custom": map[string]any{
								"val": 6,
							},
							"age": 40,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Andy",
							"custom": map[string]any{
								"val": nil,
							},
							"age": 50,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Chris",
							"custom": map[string]any{
								"val": 7,
							},
							"age": nil,
						},
					},
					testUtils.Request{
						Request: tc.req,
						Results: tc.result,
					},
					testUtils.Request{
						Request:  makeExplainQuery(tc.req),
						Asserter: testUtils.NewExplainAsserter().WithIndexFetches(tc.indexFetches),
					},
				},
			}

			testUtils.ExecuteTestCase(t, test)
		})
	}
}

func TestJSONCompositeIndex_ScalarWithJSON_ShouldFetchUsingIndex(t *testing.T) {
	type testCase struct {
		name         string
		req          string
		result       map[string]any
		indexFetches int
	}

	testCases := []testCase{
		{
			name: "Unique combination. Non-unique custom.val",
			req: `query {
				User(filter: {
					age: {_eq: 25}, 
					custom: {val: {_eq: 3}}
				}) {
					name
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "Islam"},
				},
			},
			indexFetches: 1,
		},
		{
			name: "Unique combination. Non-unique age",
			req: `query {
				User(filter: {
					age: {_eq: 30}, 
					custom: {val: {_eq: 3}}
				}) {
					name
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "John"},
				},
			},
			indexFetches: 1,
		},
		{
			name: "Match first part of the composite index",
			req: `query {
				User(filter: {age: {_eq: 25}}) {
					name
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "Islam"},
					{"name": "Shahzad"},
				},
			},
			indexFetches: 2,
		},
		{
			name: "Non-unique combination",
			req: `query {
				User(filter: {
					age: {_eq: 35}, 
					custom: {val: {_eq: 5}}
				}) {
					name
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "Addo"},
					{"name": "Keenan"},
				},
			},
			indexFetches: 2,
		},
		{
			name: "Match second part of the composite index",
			req: `query {
				User(filter: {custom: {val: {_eq: 6}}}) {
					name
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "Bruno"},
				},
			},
			indexFetches: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			test := testUtils.TestCase{
				Actions: []any{
					testUtils.SchemaUpdate{
						Schema: `
							type User @index(includes: [{field: "age"}, {field: "custom"}]) {
								name: String 
								custom: JSON 
								age: Int
							}`,
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "John",
							"custom": map[string]any{
								"val": 3,
							},
							"age": 30,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Islam",
							"custom": map[string]any{
								"val": 3,
							},
							"age": 25,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Shahzad",
							"custom": map[string]any{
								"val": 4,
							},
							"age": 25,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Keenan",
							"custom": map[string]any{
								"val": 5,
							},
							"age": 35,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Addo",
							"custom": map[string]any{
								"val": 5,
							},
							"age": 35,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Bruno",
							"custom": map[string]any{
								"val": 6,
							},
							"age": 40,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Andy",
							"custom": map[string]any{
								"val": nil,
							},
							"age": 50,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Chris",
							"custom": map[string]any{
								"val": 7,
							},
							"age": nil,
						},
					},
					testUtils.Request{
						Request: tc.req,
						Results: tc.result,
					},
					testUtils.Request{
						Request:  makeExplainQuery(tc.req),
						Asserter: testUtils.NewExplainAsserter().WithIndexFetches(tc.indexFetches),
					},
				},
			}

			testUtils.ExecuteTestCase(t, test)
		})
	}
}

func TestJSONArrayCompositeIndex_JSONArrayWithScalar_ShouldFetchUsingIndex(t *testing.T) {
	type testCase struct {
		name         string
		req          string
		result       map[string]any
		indexFetches int
	}

	testCases := []testCase{
		{
			name: "Unique combination. Non-unique custom.numbers element",
			req: `query {
				User(filter: {
					custom: {numbers: {_any: {_eq: 3}}}, 
					age: {_eq: 25}
				}) {
					name
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "Islam"},
				},
			},
			indexFetches: 1,
		},
		{
			name: "Unique combination. Non-unique age",
			req: `query {
				User(filter: {
					custom: {numbers: {_any: {_eq: 3}}}, 
					age: {_eq: 30}
				}) {
					name
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "John"},
				},
			},
			indexFetches: 1,
		},
		{
			name: "Match first part of the composite index",
			req: `query {
				User(filter: {custom: {numbers: {_any: {_eq: 3}}}}) {
					name
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "Islam"},
					{"name": "John"},
				},
			},
			indexFetches: 2,
		},
		{
			name: "Non-unique combination",
			req: `query {
				User(filter: {
					custom: {numbers: {_any: {_eq: 5}}}, 
					age: {_eq: 35}
				}) {
					name
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "Keenan"},
					{"name": "Addo"},
				},
			},
			indexFetches: 2,
		},
		{
			name: "Match second part of the composite index",
			req: `query {
				User(filter: {age: {_eq: 40}}) {
					name
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "Bruno"},
				},
			},
			indexFetches: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			test := testUtils.TestCase{
				Actions: []any{
					testUtils.SchemaUpdate{
						Schema: `
							type User @index(includes: [{field: "custom"}, {field: "age"}]) {
								name: String 
								custom: JSON 
								age: Int
							}`,
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "John",
							"custom": map[string]any{
								"numbers": []int{3, 4},
							},
							"age": 30,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Islam",
							"custom": map[string]any{
								"numbers": []int{3, 5},
							},
							"age": 25,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Shahzad",
							"custom": map[string]any{
								"numbers": []int{4, 6},
							},
							"age": 30,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Keenan",
							"custom": map[string]any{
								"numbers": []int{5, 7},
							},
							"age": 35,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Addo",
							"custom": map[string]any{
								"numbers": []int{1, 5, 8},
							},
							"age": 35,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Bruno",
							"custom": map[string]any{
								"numbers": []int{6, 9},
							},
							"age": 40,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Andy",
							"custom": map[string]any{
								"numbers": []int{},
							},
							"age": 35,
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Chris",
							"custom": map[string]any{
								"numbers": []int{7, 10},
							},
							"age": nil,
						},
					},
					testUtils.Request{
						Request: tc.req,
						Results: tc.result,
					},
					testUtils.Request{
						Request:  makeExplainQuery(tc.req),
						Asserter: testUtils.NewExplainAsserter().WithIndexFetches(tc.indexFetches),
					},
				},
			}

			testUtils.ExecuteTestCase(t, test)
		})
	}
}

func TestJSONArrayCompositeIndex_JSONArrayWithArrayField_ShouldFetchUsingIndex(t *testing.T) {
	type testCase struct {
		name         string
		req          string
		result       map[string]any
		indexFetches int
	}

	testCases := []testCase{
		{
			name: "Unique combination. Non-unique custom.numbers element",
			req: `query {
				User(filter: {
					custom: {numbers: {_any: {_eq: 3}}},
					tags: {_any: {_eq: "unique"}}
				}) {
					name
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "John"},
				},
			},
			indexFetches: 1,
		},
		{
			name: "Unique combination. Non-unique tags",
			req: `query {
				User(filter: {
					custom: {numbers: {_any: {_eq: 15}}},
					tags: {_any: {_eq: "mentor"}}
				}) {
					name
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "Islam"},
				},
			},
			indexFetches: 1,
		},
		{
			name: "Match first part of the composite index",
			req: `query {
				User(filter: {custom: {numbers: {_any: {_eq: 5}}}}) {
					name
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "Addo"},
					{"name": "Keenan"},
				},
			},
			indexFetches: 2,
		},
		{
			name: "Non-unique combination",
			req: `query {
				User(filter: {
					custom: {numbers: {_any: {_eq: 5}}},
					tags: {_any: {_eq: "family"}}
				}) {
					name
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "Addo"},
					{"name": "Keenan"},
				},
			},
			indexFetches: 2,
		},
		{
			name: "Match second part of the composite index",
			req: `query {
				User(filter: {tags: {_any: {_eq: "dude"}}}) {
					name
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "Bruno"},
				},
			},
			indexFetches: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			test := testUtils.TestCase{
				Actions: []any{
					testUtils.SchemaUpdate{
						Schema: `
							type User @index(includes: [{field: "custom"}, {field: "tags"}]) {
								name: String 
								custom: JSON 
								tags: [String]
							}`,
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "John",
							"custom": map[string]any{
								"numbers": []int{3, 4},
							},
							"tags": []any{"colleague", "mentor", "unique"},
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Islam",
							"custom": map[string]any{
								"numbers": []int{3, 15},
							},
							"tags": []any{"friend", "mentor"},
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Shahzad",
							"custom": map[string]any{
								"numbers": []int{4, 6},
							},
							"tags": []any{"colleague"},
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Keenan",
							"custom": map[string]any{
								"numbers": []int{5, 7},
							},
							"tags": []any{"family"},
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Addo",
							"custom": map[string]any{
								"numbers": []int{1, 5, 8},
							},
							"tags": []any{"family"},
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Bruno",
							"custom": map[string]any{
								"numbers": []int{6, 9},
							},
							"tags": []any{"dude"},
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Andy",
							"custom": map[string]any{
								"numbers": []int{},
							},
							"tags": []any{"friend"},
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Chris",
							"custom": map[string]any{
								"numbers": []int{7, 10},
							},
							"tags": []any{"colleague"},
						},
					},
					testUtils.Request{
						Request: tc.req,
						Results: tc.result,
					},
					testUtils.Request{
						Request:  makeExplainQuery(tc.req),
						Asserter: testUtils.NewExplainAsserter().WithIndexFetches(tc.indexFetches),
					},
				},
			}

			testUtils.ExecuteTestCase(t, test)
		})
	}
}
