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
						"numbers": []any{8, nil},
					},
				},
				ExpectedError: db.NewErrCanNotIndexNonUniqueFields(
					"bae-f87bacb3-4741-5208-a432-cbfec654080d",
					errors.NewKV("custom", map[string]any{"numbers": []any{8, nil}})).Error(),
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"custom": map[string]any{
						// existing non-array-element value
						"numbers": 3,
					},
				},
				ExpectedError: db.NewErrCanNotIndexNonUniqueFields(
					"bae-54e76159-66c6-56be-ad65-7ff83edda058",
					errors.NewKV("custom", map[string]any{"numbers": 3})).Error(),
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Chris",
					"custom": map[string]any{
						// existing nested value
						"numbers": []any{9, []int{3}},
					},
				},
				ExpectedError: db.NewErrCanNotIndexNonUniqueFields(
					"bae-8dba1343-148c-590c-a942-dd6c80f204fb",
					errors.NewKV("custom", map[string]any{"numbers": []any{9, []int{3}}})).Error(),
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
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(1).WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
