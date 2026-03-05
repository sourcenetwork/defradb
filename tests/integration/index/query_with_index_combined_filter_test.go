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
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String @index
						age: Int
					}`,
			},
			testUtils.AddPredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Addo"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
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
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String @index
						age: Int @index
						email: String 
					}`,
			},
			testUtils.AddPredefinedDocs{
				Docs: getUserDocs(),
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
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(6),
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
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String @index
						age: Int @index
						email: String 
					}`,
			},
			testUtils.AddPredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Andy"},
						{"name": "Addo"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(6),
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
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String @index
						age: Int
					}`,
			},
			testUtils.AddPredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Roy"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
