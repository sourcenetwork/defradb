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

func TestQueryInlineIntegerArray_WithMaxAndOrder_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"testScores": [3, 4, 5],
					"pageRatings": [1.0, 2.0, 3.0]
				}`, // Maximum: 5
			},

			&action.CreateDoc{
				Doc: `{
					"testScores": [30, 40, 50],
					"pageRatings": [10.0, 20.0, 30.0]
				}`, // Maximum: 50
			},

			// Test descending order
			&action.Request{
				Request: `query {
					Users(order: {_alias: {total: DESC}}) {
						total: MAX(testScores: {}, pageRatings: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"total": 50,
						},
						{
							"total": 5,
						},
					},
				},
			},

			// Test ascending order
			&action.Request{
				Request: `query {
					Users(order: {_alias: {total: ASC}}) {
						total: MAX(testScores: {}, pageRatings: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"total": 5,
						},
						{
							"total": 50,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArray_WithNullAndMaxAndOrder_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"testScores": [3, 4, 5, null],
					"pageRatings": [1.0, 2.0, 3.0, null]
				}`, // Maximum: 5
			},

			&action.CreateDoc{
				Doc: `{
					"testScores": [30, 40, 50, null],
					"pageRatings": [10.0, 20.0, 30.0, null]
				}`, // Maximum: 50
			},

			// Test descending order
			&action.Request{
				Request: `query {
					Users(order: {_alias: {total: DESC}}) {
						total: MAX(testScores: {}, pageRatings: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"total": 50,
						},
						{
							"total": 5,
						},
					},
				},
			},

			// Test ascending order
			&action.Request{
				Request: `query {
					Users(order: {_alias: {total: ASC}}) {
						total: MAX(testScores: {}, pageRatings: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"total": 5,
						},
						{
							"total": 50,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
