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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithMinWithOrder_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"Age": 30,
					"HeightM": 1.8
				}`,
			}, // Min: 1.8

			&action.CreateDoc{
				Doc: `{
					"Age": 25,
					"HeightM": 1.6
				}`,
			}, // Min: 1.6

			// Test descending order by computed total
			&action.Request{
				Request: `query {
					Users(order: {_alias: {total: DESC}}) {
						total: MIN(HeightM: {}, Age: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"total": 1.8,
						},
						{
							"total": 1.6,
						},
					},
				},
			},

			// Test ascending order by computed total
			&action.Request{
				Request: `query {
					Users(order: {_alias: {total: ASC}}) {
						total: MIN(HeightM: {}, Age: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"total": 1.6,
						},
						{
							"total": 1.8,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
