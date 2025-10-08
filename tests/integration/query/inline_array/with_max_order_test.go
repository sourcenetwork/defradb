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

func TestQueryInlineIntegerArray_WithMaxAndOrder_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"books": [3, 4, 5],
					"movies": [1, 2, 3],
					"games": [3, 4, 2]
				}`, // Maximum: 5
			},

			testUtils.CreateDoc{
				Doc: `{
					"books": [30, 40, 50],
					"movies": [10, 20, 30],
					"games": [30, 40, 20]
				}`, // Maximum: 50
			},

			// Test descending order
			testUtils.Request{
				Request: `query {
					Users(order: {_alias: {total: DESC}}) {
						total: _max(books: {}, games: {}, movies: {})
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
			testUtils.Request{
				Request: `query {
					Users(order: {_alias: {total: ASC}}) {
						total: _max(books: {}, games: {}, movies: {})
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
