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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryInlineIntegerArray_WithCountAndOrder_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"testScores": [3, 4, 5],
					"pageRatings": [1.0, 2.0, 3.0]
				}`, // Count: 6
			},

			&action.AddDoc{
				Doc: `{
					"testScores": [3, 4, 5, 6],
					"pageRatings": [1.0, 2.0, 3.0, 4.0]
				}`, // Count: 8
			},

			// Test descending order
			&action.Request{
				Request: `query {
					Users(order: {_alias: {total: DESC}}) {
						total: COUNT(testScores: {}, pageRatings: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"total": 8,
						},
						{
							"total": 6,
						},
					},
				},
			},

			// Test ascending order
			&action.Request{
				Request: `query {
					Users(order: {_alias: {total: ASC}}) {
						total: COUNT(testScores: {}, pageRatings: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"total": 6,
						},
						{
							"total": 8,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArray_WithNullAndCountAndOrder_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"testScores": [3, 4, 5, null],
					"pageRatings": [1.0, 2.0, 3.0, null]
				}`, // Count: 8
			},

			&action.AddDoc{
				Doc: `{
					"testScores": [3, 4, 5, 6, null],
					"pageRatings": [1.0, 2.0, 3.0, 4.0, null]
				}`, // Count: 10
			},

			// Test descending order
			&action.Request{
				Request: `query {
					Users(order: {_alias: {total: DESC}}) {
						total: COUNT(testScores: {}, pageRatings: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"total": 10,
						},
						{
							"total": 8,
						},
					},
				},
			},

			// Test ascending order
			&action.Request{
				Request: `query {
					Users(order: {_alias: {total: ASC}}) {
						total: COUNT(testScores: {}, pageRatings: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"total": 8,
						},
						{
							"total": 10,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
