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

func TestJSONUniqueIndex_WithRandomValues_ShouldGuaranteeUniquenessAndBeAbelToUseIndexForFetching(t *testing.T) {
	req := `query {
		User(filter: {custom: {height: {_eq: 168}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						custom: JSON @index(unique: true)
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"custom": {"height": 168, "weight": 70}
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Andy",
					"custom": {"height": 190}
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Islam",
					"custom": {"height": 168}
				}`,
				ExpectedError: "can not index a doc's field(s) that violates unique index.",
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": 30
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Bruno",
					"custom": 20
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Keenan",
					"custom": 30
				}`,
				ExpectedError: "can not index a doc's field(s) that violates unique index.",
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONUniqueIndex_UponUpdate_ShouldUseNewIndexValues(t *testing.T) {
	req1 := `query {
		User(filter: {custom: {height: {_eq: 172}}}) {
			name
		}
	}`
	req2 := `query {
		User(filter: {custom: {BMI: {_eq: 22}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						custom: JSON @index(unique: true)
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"custom": {"height": 168, "weight": 70}
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Islam",
					"custom": {"height": 180, "BMI": 25}
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"name": "John",
					"custom": {"height": 172, "BMI": 22}
				}`,
			},
			&action.Request{
				Request: req1,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
			&action.Request{
				Request: req2,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
