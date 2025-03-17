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
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						custom: JSON @index(unique: true)
					}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"custom": map[string]any{
						"numbers": []any{3, 4, nil},
					},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Bruno",
					"custom": map[string]any{
						// use existing value of a different type
						"numbers": []any{"3", "str", true},
					},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"custom": map[string]any{
						// existing non-array-element value
						"numbers": 3,
					},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Islam",
					"custom": map[string]any{
						// use existing value
						"numbers": []int{4},
					},
				},
				ExpectedError: db.NewErrCanNotIndexNonUniqueFields(
					"bae-8ba4aee7-0f15-5bfd-b1c8-7ae19782982b",
					errors.NewKV("custom", map[string]any{"numbers": []int{4}})).Error(),
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Shahzad",
					"custom": map[string]any{
						// array with duplicate values
						"numbers": []int{5, 8, 5},
					},
				},
				ExpectedError: db.NewErrCanNotIndexNonUniqueFields(
					"bae-d7cd78f3-d14e-55a7-bfbc-8c0deb2220b4",
					errors.NewKV("custom", map[string]any{"numbers": []int{5, 8, 5}})).Error(),
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Keenan",
					"custom": map[string]any{
						// use existing nil value
						"numbers": []any{6, nil},
					},
				},
				ExpectedError: db.NewErrCanNotIndexNonUniqueFields(
					"bae-bde18215-f623-568e-868d-1156c30e45d3",
					errors.NewKV("custom", map[string]any{"numbers": []any{6, nil}})).Error(),
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
