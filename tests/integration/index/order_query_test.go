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

func TestOrderQueryWithIndex_WithAscendingOrder_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(order: {age: ASC}) {
			name
			age
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						age: Int @index
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Shahzad",
							"age":  int64(20),
						},
						{
							"name": "Bruno",
							"age":  int64(23),
						},
						{
							"name": "Fred",
							"age":  int64(28),
						},
						{
							"name": "John",
							"age":  int64(30),
						},
						{
							"name": "Islam",
							"age":  int64(32),
						},
						{
							"name": "Andy",
							"age":  int64(33),
						},
						{
							"name": "Addo",
							"age":  int64(42),
						},
						{
							"name": "Roy",
							"age":  int64(44),
						},
						{
							"name": "Keenan",
							"age":  int64(48),
						},
						{
							"name": "Chris",
							"age":  int64(55),
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithOrder().WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOrderQueryWithIndex_WithLimitDescending_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(order: {age: DESC}, limit: 3) {
			name
			age
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						age: Int @index
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Chris",
							"age":  int64(55),
						},
						{
							"name": "Keenan",
							"age":  int64(48),
						},
						{
							"name": "Roy",
							"age":  int64(44),
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithLimit().WithOrder().WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOrderQueryWithIndex_WithLimitAscending_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(order: {age: ASC}, limit: 3) {
			name
			age
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						age: Int @index
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Shahzad",
							"age":  int64(20),
						},
						{
							"name": "Bruno",
							"age":  int64(23),
						},
						{
							"name": "Fred",
							"age":  int64(28),
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithLimit().WithOrder().WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOrderQueryWithIndex_WithFilterOnNonIndexedFieldAscending_ShouldUseIndexForOrdering(t *testing.T) {
	req := `query {
		User(order: {age: ASC}, filter: {name: {_like: "A%"}}) {
			name
			age
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						age: Int @index
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Andy",
							"age":  int64(33),
						},
						{
							"name": "Addo",
							"age":  int64(42),
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				// we fetch all available docs with index
				Asserter: testUtils.NewExplainAsserter().WithOrder().WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOrderQueryWithIndex_WithFilterOnNonIndexedFieldDescending_ShouldUseIndexForOrdering(t *testing.T) {
	req := `query {
		User(order: {age: DESC}, filter: {name: {_like: "A%"}}) {
			name
			age
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						age: Int @index
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Addo",
							"age":  int64(42),
						},
						{
							"name": "Andy",
							"age":  int64(33),
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				// we fetch all available docs with index
				Asserter: testUtils.NewExplainAsserter().WithOrder().WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
