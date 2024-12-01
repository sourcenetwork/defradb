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

func TestJSONIndex_WithFilterOnNumberField_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {height: {_eq: 168}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
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
					"custom": {"height": 168}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": {"weight": 80, "BMI": 25}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"custom": {"height": 190}
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
						{"name": "Islam"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(2).WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
