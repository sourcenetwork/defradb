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

func TestQuerySimpleWithMaxWithOrder_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Age": 30,
					"HeightM": 1.8
				}`,
			}, // Max: 30

			testUtils.CreateDoc{
				Doc: `{
					"Age": 25,
					"HeightM": 1.6
				}`,
			}, // Max: 25

			// Test descending order by computed total
			testUtils.Request{
				Request: `query {
					Users(order: {_alias: {total: DESC}}) {
						total: _max(HeightM: {}, Age: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"total": 30,
						},
						{
							"total": 25,
						},
					},
				},
			},

			// Test ascending order by computed total
			testUtils.Request{
				Request: `query {
					Users(order: {_alias: {total: ASC}}) {
						total: _max(HeightM: {}, Age: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"total": 25,
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
