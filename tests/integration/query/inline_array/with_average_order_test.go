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

func TestQueryInlineIntegerArrayWithAverageAndOrder_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"testScores": [3, 4, 5],
					"pageRatings": [1.0, 2.0, 3.0]
				}`, // Average: 3.0
			},

			&action.AddDoc{
				Doc: `{
					"testScores": [30, 40, 50],
					"pageRatings": [10.0, 20.0, 30.0]
				}`, // Average: 30
			},

			// Test descending order
			&action.Request{
				Request: `query {
					Users(order: {_alias: {total: DESC}}) {
						total: AVG(testScores: {}, pageRatings: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"total": 30,
						},
						{
							"total": 3,
						},
					},
				},
			},

			// Test ascending order
			&action.Request{
				Request: `query {
					Users(order: {_alias: {total: ASC}}) {
						total: AVG(testScores: {}, pageRatings: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"total": 3,
						},
						{
							"total": 30,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArrayWithNullWithAverageAndOrder_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"testScores": [3, 4, 5, null],
					"pageRatings": [1.0, 2.0, 3.0, null]
				}`, // Average: 3.0
			},

			&action.AddDoc{
				Doc: `{
					"testScores": [30, 40, 50, null],
					"pageRatings": [10.0, 20.0, 30.0, null]
				}`, // Average: 30
			},

			// Test descending order
			&action.Request{
				Request: `query {
					Users(order: {_alias: {total: DESC}}) {
						total: AVG(testScores: {}, pageRatings: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"total": 30,
						},
						{
							"total": 3,
						},
					},
				},
			},

			// Test ascending order
			&action.Request{
				Request: `query {
					Users(order: {_alias: {total: ASC}}) {
						total: AVG(testScores: {}, pageRatings: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"total": 3,
						},
						{
							"total": 30,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
