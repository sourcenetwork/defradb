// Copyright 2023 Democratized Data Foundation
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

func TestQueryWithIndex_WithNonIndexedFields_ShouldFetchAllOfThem(t *testing.T) {
	test := testUtils.TestCase{
		Description: "If there are non-indexed fields in the query, they should be fetched",
		Actions: []any{
			createSchemaWithDocs(`
				type User {
					name: String @index
					age: Int
				} 
			`),
			sendRequestAndExplain(`
				User(filter: {name: {_eq: "Islam"}}) {
					name
					age
				}`,
				[]map[string]any{{
					"name": "Islam",
					"age":  uint64(32),
				}},
				testUtils.NewExplainAsserter().WithDocFetches(1).WithFieldFetches(2).WithIndexFetches(1),
			),
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithEqualFilter_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _eq filter",
		Actions: []any{
			createSchemaWithDocs(`
				type User {
					name: String @index
				} 
			`),
			sendRequestAndExplain(`
				User(filter: {name: {_eq: "Islam"}}) {
					name
				}`,
				[]map[string]any{
					{"name": "Islam"},
				},
				testUtils.NewExplainAsserter().WithDocFetches(1).WithFieldFetches(1).WithIndexFetches(1),
			),
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_IfSeveralDocsWithEqFilter_ShouldFetchAll(t *testing.T) {
	test := testUtils.TestCase{
		Description: "If there are several docs matching _eq filter, they should be fetched",
		Actions: []any{
			createSchemaWithDocs(`
				type User {
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
				User(filter: {name: {_eq: "Islam"}}) {
					age
				}`,
				[]map[string]any{
					{"age": uint64(32)},
					{"age": uint64(18)},
				},
				testUtils.NewExplainAsserter().WithDocFetches(2).WithFieldFetches(4).WithIndexFetches(2),
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
				type User {
					name: String 
					age: Int @index
				} 
			`),
			sendRequestAndExplain(`
				User(filter: {age: {_gt: 48}}) {
					name
				}`,
				[]map[string]any{
					{"name": "Chris"},
				},
				testUtils.NewExplainAsserter().WithDocFetches(1).WithFieldFetches(2).WithIndexFetches(8),
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
				type User {
					name: String 
					age: Int @index
				} 
			`),
			sendRequestAndExplain(`
				User(filter: {age: {_ge: 48}}) {
					name
				}`,
				[]map[string]any{
					{"name": "Keenan"},
					{"name": "Chris"},
				},
				testUtils.NewExplainAsserter().WithDocFetches(2).WithFieldFetches(4).WithIndexFetches(8),
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
				type User {
					name: String 
					age: Int @index
				} 
			`),
			sendRequestAndExplain(`
				User(filter: {age: {_lt: 28}}) {
					name
				}`,
				[]map[string]any{
					{"name": "Shahzad"},
				},
				testUtils.NewExplainAsserter().WithDocFetches(1).WithFieldFetches(2).WithIndexFetches(8),
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
				type User {
					name: String 
					age: Int @index
				} 
			`),
			sendRequestAndExplain(`
				User(filter: {age: {_le: 28}}) {
					name
				}`,
				[]map[string]any{
					{"name": "Shahzad"},
					{"name": "Fred"},
				},
				testUtils.NewExplainAsserter().WithDocFetches(2).WithFieldFetches(4).WithIndexFetches(8),
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
				type User {
					name: String @index
					age: Int 
				} 
			`),
			sendRequestAndExplain(`
				User(filter: {name: {_ne: "Islam"}}) {
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
				testUtils.NewExplainAsserter().WithDocFetches(7).WithFieldFetches(7).WithIndexFetches(8),
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
				type User {
					name: String 
					age: Int @index
				} 
			`),
			sendRequestAndExplain(`
				User(filter: {age: {_in: [20, 33]}}) {
					name
				}`,
				[]map[string]any{
					{"name": "Shahzad"},
					{"name": "Andy"},
				},
				testUtils.NewExplainAsserter().WithDocFetches(2).WithFieldFetches(4).WithIndexFetches(2),
			),
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_IfSeveralDocsWithInFilter_ShouldFetchAll(t *testing.T) {
	test := testUtils.TestCase{
		Description: "If there are several docs matching _in filter, they should be fetched",
		Actions: []any{
			createSchemaWithDocs(`
				type User {
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
				User(filter: {name: {_in: ["Islam"]}}) {
					age
				}`,
				[]map[string]any{
					{"age": uint64(32)},
					{"age": uint64(18)},
				},
				testUtils.NewExplainAsserter().WithDocFetches(2).WithFieldFetches(4).WithIndexFetches(2),
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
				type User {
					name: String 
					age: Int @index
				} 
			`),
			sendRequestAndExplain(`
				User(filter: {age: {_nin: [20, 28, 33, 42, 55]}}) {
					name
				}`,
				[]map[string]any{
					{"name": "John"},
					{"name": "Islam"},
					{"name": "Keenan"},
				},
				testUtils.NewExplainAsserter().WithDocFetches(3).WithFieldFetches(6).WithIndexFetches(8),
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
				type User {
					name: String @index
					age: Int 
				} 
			`),
			sendRequestAndExplain(`
				User(filter: {name: {_like: "A%"}}) {
					name
				}`,
				[]map[string]any{
					{"name": "Addo"},
					{"name": "Andy"},
				},
				testUtils.NewExplainAsserter().WithDocFetches(2).WithFieldFetches(2).WithIndexFetches(8),
			),
			sendRequestAndExplain(`
				User(filter: {name: {_like: "%d"}}) {
					name
				}`,
				[]map[string]any{
					{"name": "Fred"},
					{"name": "Shahzad"},
				},
				testUtils.NewExplainAsserter().WithDocFetches(2).WithFieldFetches(2).WithIndexFetches(8),
			),
			sendRequestAndExplain(`
				User(filter: {name: {_like: "%e%"}}) {
					name
				}`,
				[]map[string]any{
					{"name": "Fred"},
					{"name": "Keenan"},
				},
				testUtils.NewExplainAsserter().WithDocFetches(2).WithFieldFetches(2).WithIndexFetches(8),
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
				type User {
					name: String @index
					age: Int 
				} 
			`),
			sendRequestAndExplain(`
				User(filter: {name: {_nlike: "%h%"}}) {
					name
				}`,
				[]map[string]any{
					{"name": "Addo"},
					{"name": "Andy"},
					{"name": "Fred"},
					{"name": "Islam"},
					{"name": "Keenan"},
				},
				testUtils.NewExplainAsserter().WithDocFetches(5).WithFieldFetches(5).WithIndexFetches(8),
			),
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
