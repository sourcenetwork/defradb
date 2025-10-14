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

func TestQuerySimpleWithSumWithOrder_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Age": 30,
					"HeightM": 1.8
				}`,
			},

			testUtils.CreateDoc{
				Doc: `{
					"Age": 25,
					"HeightM": 1.6
				}`,
			},

			// Test descending order by computed total
			testUtils.Request{
				Request: `query {
					Users(order: {_alias: {total: DESC}}) {
						total: _sum(Age: {}, HeightM: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"total": 11,
						},
						{
							"total": 7,
						},
					},
				},
			},

			// Test ascending order by computed total
			testUtils.Request{
				Request: `query {
					Users(order: {_alias: {total: ASC}}) {
						total: _sum(Age: {}, HeightM: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"total": 7,
						},
						{
							"total": 11,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
