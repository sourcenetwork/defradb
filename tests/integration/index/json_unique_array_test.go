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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestJSONArrayUniqueIndex_ShouldAllowOnlyUniqueValuesAndUseThemForFetching(t *testing.T) {
	req := `query {
		User(filter: {custom: {numbers: {_any: {_eq: 4}}}}) {
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
				DocMap: map[string]any{
					"name": "John",
					"custom": map[string]any{
						"numbers": []any{3, 4, nil},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Bruno",
					"custom": map[string]any{
						// use existing value of a different type
						"numbers": []any{"3", "str", true},
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"custom": map[string]any{
						// existing non-array-element value
						"numbers": 3,
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Islam",
					"custom": map[string]any{
						// use existing value
						"numbers": []int{4},
					},
				},
				ExpectedError: "can not index a doc's field(s) that violates unique index.",
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Shahzad",
					"custom": map[string]any{
						// array with duplicate values
						"numbers": []int{5, 8, 5},
					},
				},
				ExpectedError: "can not index a doc's field(s) that violates unique index.",
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Keenan",
					"custom": map[string]any{
						// use existing nil value
						"numbers": []any{6, nil},
					},
				},
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
