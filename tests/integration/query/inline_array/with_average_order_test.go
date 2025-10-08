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

func TestQueryInlineIntegerArray_WithAverageAndOrder_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"books": [3, 4, 5],
					"movies": [1, 2, 3],
					"games": [3, 4, 2]
				}`, // Average: 3
			},

			testUtils.CreateDoc{
				Doc: `{
					"books": [30, 40, 50],
					"movies": [10, 20, 30],
					"games": [30, 40, 20]
				}`, // Average: 30
			},

			// Test descending order
			testUtils.Request{
				Request: `query {
					Users(order: {_alias: {total: DESC}}) {
						total: _avg(books: {}, games: {}, movies: {})
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
			testUtils.Request{
				Request: `query {
					Users(order: {_alias: {total: ASC}}) {
						total: _avg(books: {}, games: {}, movies: {})
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
