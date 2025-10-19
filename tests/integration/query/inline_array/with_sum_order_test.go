// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package inline_array

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryInlineIntegerArray_WithSumAndOrder_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{

			testUtils.CreateDoc{
				Doc: `{
					"testScores": [3, 4, 5],
					"pageRatings": [1.0, 2.0, 3.0]
				}`, // Sum: 18
			},

			testUtils.CreateDoc{
				Doc: `{
					"testScores": [30, 40, 50],
					"pageRatings": [10.0, 20.0, 30.0]
				}`, // Sum: 180
			},

			// Test descending order
			testUtils.Request{
				Request: `query {
					Users(order: {_alias: {total: DESC}}) {
						total: _sum(testScores: {}, pageRatings: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"total": 180,
						},
						{
							"total": 18,
						},
					},
				},
			},

			// Test ascending order
			testUtils.Request{
				Request: `query {
					Users(order: {_alias: {total: ASC}}) {
						total: _sum(testScores: {}, pageRatings: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"total": 18,
						},
						{
							"total": 180,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArray_WithNullAndSumAndOrder_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{

			testUtils.CreateDoc{
				Doc: `{
					"testScores": [3, 4, 5, null],
					"pageRatings": [1.0, 2.0, 3.0, null]
				}`, // Sum: 18
			},

			testUtils.CreateDoc{
				Doc: `{
					"testScores": [30, 40, 50, null],
					"pageRatings": [10.0, 20.0, 30.0, null]
				}`, // Sum: 180
			},

			// Test descending order
			testUtils.Request{
				Request: `query {
					Users(order: {_alias: {total: DESC}}) {
						total: _sum(testScores: {}, pageRatings: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"total": 180,
						},
						{
							"total": 18,
						},
					},
				},
			},

			// Test ascending order
			testUtils.Request{
				Request: `query {
					Users(order: {_alias: {total: ASC}}) {
						total: _sum(testScores: {}, pageRatings: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"total": 18,
						},
						{
							"total": 180,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
