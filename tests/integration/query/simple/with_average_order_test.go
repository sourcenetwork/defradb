// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package simple

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithAverageWithOrder_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Age": 30,
					"HeightM": 1.8
				}`,
			}, // Average: 15.9

			testUtils.CreateDoc{
				Doc: `{
					"Age": 25,
					"HeightM": 1.6
				}`,
			}, // Sum: 13.3

			// Test descending order by computed total
			testUtils.Request{
				Request: `query {
					Users(order: {_alias: {total: DESC}}) {
						total: _avg(HeightM: {}, Age: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"total": 15.9,
						},
						{
							"total": 13.3,
						},
					},
				},
			},

			// Test ascending order by computed total
			testUtils.Request{
				Request: `query {
					Users(order: {_alias: {total: ASC}}) {
						total: _avg(HeightM: {}, Age: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"total": 13.3,
						},
						{
							"total": 15.9,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
