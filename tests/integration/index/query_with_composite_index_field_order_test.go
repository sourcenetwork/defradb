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

func TestQueryWithCompositeIndex_WithDefaultOrder_ShouldFetchInDefaultOrder(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test composite index with reverted order",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(fields: ["name",  "age"]) {
						name: String
						age: Int
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alan",
						"age":	29
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						User(filter: {name: {_like: "Al%"}}) {
							name
							age
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Alan",
						"age":  29,
					},
					{
						"name": "Alice",
						"age":  22,
					},
					{
						"name": "Alice",
						"age":  24,
					},
					{
						"name": "Alice",
						"age":  38,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithRevertedOrderOnFirstField_ShouldFetchInRevertedOrder(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test composite index with reverted order",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(fields: ["name",  "age"], directions: [DESC, ASC]) {
						name: String
						age: Int
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alan",
						"age":	29
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Andy",
						"age":	24
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						User(filter: {name: {_like: "A%"}}) {
							name
							age
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Andy",
						"age":  24,
					},
					{
						"name": "Alice",
						"age":  22,
					},
					{
						"name": "Alice",
						"age":  24,
					},
					{
						"name": "Alice",
						"age":  38,
					},
					{
						"name": "Alan",
						"age":  29,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithRevertedOrderOnSecondField_ShouldFetchInRevertedOrder(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test composite index with reverted order",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(fields: ["name",  "age"], directions: [ASC, DESC]) {
						name: String
						age: Int
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alan",
						"age":	29
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						User(filter: {name: {_like: "Al%"}}) {
							name
							age
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Alan",
						"age":  29,
					},
					{
						"name": "Alice",
						"age":  38,
					},
					{
						"name": "Alice",
						"age":  24,
					},
					{
						"name": "Alice",
						"age":  22,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_IfExactMatchWithRevertedOrderOnFirstField_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test composite index with reverted order",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(fields: ["name",  "age"], directions: [DESC, ASC]) {
						name: String
						age: Int
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alan",
						"age":	29
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						User(filter: {name: {_eq: "Alice"}, age: {_eq: 22}}) {
							name
							age
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Alice",
						"age":  22,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_IfExactMatchWithRevertedOrderOnSecondField_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test composite index with reverted order",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(fields: ["name",  "age"], directions: [ASC, DESC]) {
						name: String
						age: Int
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alan",
						"age":	29
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						User(filter: {name: {_eq: "Alice"}, age: {_eq: 22}}) {
							name
							age
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Alice",
						"age":  22,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithInFilterOnFirstFieldWithRevertedOrder_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _in filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(fields: ["name",  "age"], directions: [DESC, ASC]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: `query {
						User(filter: {name: {_in: ["Addo", "Andy", "Fred"]}}) {
							name
						}
					}`,
				Results: []map[string]any{
					{"name": "Addo"},
					{"name": "Andy"},
					{"name": "Fred"},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithInFilterOnSecondFieldWithRevertedOrder_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _in filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(fields: ["name",  "age"], directions: [ASC, DESC]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: `query {
						User(filter: {age: {_in: [20, 28, 33]}}) {
							name
						}
					}`,
				Results: []map[string]any{
					{"name": "Shahzad"},
					{"name": "Andy"},
					{"name": "Fred"},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
