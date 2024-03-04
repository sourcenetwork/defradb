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

func TestQueryWithIndex_IfIndexFilterWithRegular_ShouldFilter(t *testing.T) {
	req := `query {
		User(filter: {
			name: {_in: ["Fred", "Islam", "Addo"]}, 
			age:  {_gt: 40}
		}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Combination of a filter on regular and of an indexed field",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String @index
						age: Int
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: []map[string]any{
					{"name": "Addo"},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(3).WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_IfMultipleIndexFiltersWithRegular_ShouldFilter(t *testing.T) {
	req := `query {
		User(filter: {
			name: {_like: "%a%"}, 
			age:  {_gt: 30},
			email: {_like: "%m@gmail.com"}
		}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Combination of a filter on regular and of 2 indexed fields",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String @index
						age: Int @index
						email: String 
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: []map[string]any{
					{"name": "Islam"},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(12),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_IfMultipleIndexFiltersWithRegularCaseInsensitive_ShouldFilter(t *testing.T) {
	req := `query {
		User(filter: {
			name: {_ilike: "a%"}, 
			age:  {_gt: 30},
		}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Combination of a filter on regular and of 2 indexed fields and case insensitive operator",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String @index
						age: Int @index
						email: String 
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: []map[string]any{
					{"name": "Andy"},
					{"name": "Addo"},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(6),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_FilterOnNonIndexedField_ShouldIgnoreIndex(t *testing.T) {
	req := `query {
		User(filter: {
			age:  {_eq: 44}
		}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "If filter does not contain indexed field, index should be ignored",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String @index
						age: Int
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: []map[string]any{
					{"name": "Roy"},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
