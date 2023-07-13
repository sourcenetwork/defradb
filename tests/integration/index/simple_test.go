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

func TestIndexWithExplain(t *testing.T) {
	test := testUtils.TestCase{
		Description: "",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type users {
						name: String 
						age: Int
						verified: Boolean
					} 
				`,
			},
			createUserDocsWithAge(),
			testUtils.Request{
				Request: `
					query @explain(type: execute) {
						users(filter: {name: {_eq: "Islam"}}) {
							name
						}
					}`,
				Asserter: newExplainAsserter(2, 2, 1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryWithIndex_WithOnlyIndexedField_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "If there is only one indexed field in the query, it should be fetched",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type users {
						name: String @index
					} 
				`,
			},
			createUserDocs(),
			testUtils.Request{
				Request: `
					query {
						users(filter: {name: {_eq: "Islam"}}) {
							name
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Islam",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryWithIndex_WithNonIndexedFields_ShouldFetchAllOfThem(t *testing.T) {
	test := testUtils.TestCase{
		Description: "If there are non-indexed fields in the query, they should be fetched",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type users {
						name: String @index
						age: Int
					} 
				`,
			},
			createUserDocsWithAge(),
			testUtils.Request{
				Request: `
					query {
						users(filter: {name: {_eq: "Islam"}}) {
							name
							age
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Islam",
						"age":  uint64(32),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryWithIndex_IfMoreThenOneDoc_ShouldFetchAll(t *testing.T) {
	test := testUtils.TestCase{
		Description: "If there are more than one doc with the same indexed field, they should be fetched",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type users {
						name: String @index
						age: Int
					} 
				`,
			},
			createUserDocsWithAge(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Islam",
					"age": 18
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						users(filter: {name: {_eq: "Islam"}}) {
							age
						}
					}`,
				Results: []map[string]any{
					{
						"age": uint64(32),
					},
					{
						"age": uint64(18),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}
