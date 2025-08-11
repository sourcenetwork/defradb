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

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/db"
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
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index(unique: true)
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {"height": 168, "weight": 70}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"custom": {"height": 190}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": {"height": 168}
				}`,
				ExpectedError: db.NewErrCanNotIndexNonUniqueFields(
					"bae-0b423e0b-2c5d-566f-8266-91211353ab66",
					errors.NewKV("custom", map[string]any{"height": float64(168)})).Error(),
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": 30
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Bruno",
					"custom": 20
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"custom": 30
				}`,
				ExpectedError: db.NewErrCanNotIndexNonUniqueFields(
					"bae-67dd014b-4a26-55ab-a71d-fbd14a3fcecc",
					errors.NewKV("custom", 30)).Error(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
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
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index(unique: true)
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {"height": 168, "weight": 70}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": {"height": 180, "BMI": 25}
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"name": "John",
					"custom": {"height": 172, "BMI": 22}
				}`,
			},
			testUtils.Request{
				Request: req1,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
			testUtils.Request{
				Request: req2,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
