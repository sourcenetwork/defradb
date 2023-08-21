// Copyright 2022 Democratized Data Foundation
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

func sendRequestAndExplain(
	reqBody string,
	results []map[string]any,
	asserter testUtils.ResultAsserter,
) []testUtils.Request {
	return []testUtils.Request{
		{
			Request: "query {" + reqBody + "}",
			Results: results,
		},
		/*{
			Request:  "query @explain(type: execute) {" + reqBody + "}",
			Asserter: asserter,
		},*/
	}
}

func TestQueryWithIndex_WithOnlyIndexedField_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "If there is only one indexed field in the query, it should be fetched",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String @index
				} 
			`),
			sendRequestAndExplain(`
				users(filter: {name: {_eq: "Islam"}}) {
					name
				}`,
				[]map[string]any{
					{"name": "Islam"},
				},
				NewExplainAsserter().WithDocFetches(1).WithFieldFetches(1),
			),
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithNonIndexedFields_ShouldFetchAllOfThem(t *testing.T) {
	test := testUtils.TestCase{
		Description: "If there are non-indexed fields in the query, they should be fetched",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String @index
					age: Int
				} 
			`),
			sendRequestAndExplain(`
				users(filter: {name: {_eq: "Islam"}}) {
					name
					age
				}`,
				[]map[string]any{{
					"name": "Islam",
					"age":  uint64(32),
				}},
				NewExplainAsserter().WithDocFetches(1).WithFieldFetches(2),
			),
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_IfMoreThenOneDoc_ShouldFetchAll(t *testing.T) {
	test := testUtils.TestCase{
		Description: "If there are more than one doc with the same indexed field, they should be fetched",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String @index
					age: Int
				} 
			`),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Islam",
					"age": 18
				}`,
			},
			sendRequestAndExplain(`
				users(filter: {name: {_eq: "Islam"}}) {
					age
				}`,
				[]map[string]any{
					{"age": uint64(32)},
					{"age": uint64(18)},
				},
				NewExplainAsserter().WithDocFetches(2).WithFieldFetches(4),
			),
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithGreaterThanFilter_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _gt filter",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String 
					age: Int @index
				} 
			`),
			sendRequestAndExplain(`
				users(filter: {age: {_gt: 48}}) {
					name
				}`,
				[]map[string]any{
					{"name": "Chris"},
				},
				NewExplainAsserter().WithDocFetches(1).WithFieldFetches(2),
			),
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithGreaterOrEqualFilter_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _ge filter",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String 
					age: Int @index
				} 
			`),
			sendRequestAndExplain(`
				users(filter: {age: {_ge: 48}}) {
					name
				}`,
				[]map[string]any{
					{"name": "Keenan"},
					{"name": "Chris"},
				},
				NewExplainAsserter().WithDocFetches(2).WithFieldFetches(4),
			),
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithLessThanFilter_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _lt filter",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String 
					age: Int @index
				} 
			`),
			sendRequestAndExplain(`
				users(filter: {age: {_lt: 28}}) {
					name
				}`,
				[]map[string]any{
					{"name": "Shahzad"},
				},
				NewExplainAsserter().WithDocFetches(2).WithFieldFetches(4),
			),
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithLessOrEqualFilter_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _le filter",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String 
					age: Int @index
				} 
			`),
			sendRequestAndExplain(`
				users(filter: {age: {_le: 28}}) {
					name
				}`,
				[]map[string]any{
					{"name": "Shahzad"},
					{"name": "Fred"},
				},
				NewExplainAsserter().WithDocFetches(2).WithFieldFetches(4),
			),
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithNotEqualFilter_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _ne filter",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String @index
					age: Int 
				} 
			`),
			sendRequestAndExplain(`
				users(filter: {name: {_ne: "Islam"}}) {
					name
				}`,
				[]map[string]any{
					{"name": "Addo"},
					{"name": "Andy"},
					{"name": "Fred"},
					{"name": "John"},
					{"name": "Chris"},
					{"name": "Keenan"},
					{"name": "Shahzad"},
				},
				NewExplainAsserter().WithDocFetches(2).WithFieldFetches(4),
			),
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithInFilter_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _in filter",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String 
					age: Int @index
				} 
			`),
			sendRequestAndExplain(`
				users(filter: {age: {_in: [20, 33]}}) {
					name
				}`,
				[]map[string]any{
					{"name": "Shahzad"},
					{"name": "Andy"},
				},
				NewExplainAsserter().WithDocFetches(2).WithFieldFetches(4),
			),
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithNotInFilter_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _nin filter",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String 
					age: Int @index
				} 
			`),
			sendRequestAndExplain(`
				users(filter: {age: {_nin: [20, 28, 33, 42, 55]}}) {
					name
				}`,
				[]map[string]any{
					{"name": "John"},
					{"name": "Islam"},
					{"name": "Keenan"},
				},
				NewExplainAsserter().WithDocFetches(2).WithFieldFetches(4),
			),
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithLikeFilter_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _like filter",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String @index
					age: Int 
				} 
			`),
			sendRequestAndExplain(`
				users(filter: {name: {_like: "A%"}}) {
					name
				}`,
				[]map[string]any{
					{"name": "Addo"},
					{"name": "Andy"},
				},
				NewExplainAsserter().WithDocFetches(2).WithFieldFetches(4),
			),
			sendRequestAndExplain(`
				users(filter: {name: {_like: "%d"}}) {
					name
				}`,
				[]map[string]any{
					{"name": "Fred"},
					{"name": "Shahzad"},
				},
				NewExplainAsserter().WithDocFetches(2).WithFieldFetches(4),
			),
			sendRequestAndExplain(`
				users(filter: {name: {_like: "%e%"}}) {
					name
				}`,
				[]map[string]any{
					{"name": "Fred"},
					{"name": "Keenan"},
				},
				NewExplainAsserter().WithDocFetches(2).WithFieldFetches(4),
			),
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithNotLikeFilter_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _nlike filter",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String @index
					age: Int 
				} 
			`),
			sendRequestAndExplain(`
				users(filter: {name: {_nlike: "%h%"}}) {
					name
				}`,
				[]map[string]any{
					{"name": "Addo"},
					{"name": "Andy"},
					{"name": "Fred"},
					{"name": "Islam"},
					{"name": "Keenan"},
				},
				NewExplainAsserter().WithDocFetches(2).WithFieldFetches(4),
			),
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
