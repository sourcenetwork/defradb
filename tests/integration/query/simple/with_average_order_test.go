// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithAverageWithOrder_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Age": 30,
					"HeightM": 1.8
				}`,
			}, // Average: 15.9

			&action.AddDoc{
				Doc: `{
					"Age": 25,
					"HeightM": 1.6
				}`,
			}, // Sum: 13.3

			// Test descending order by computed total
			&action.Request{
				Request: `query {
					Users(order: {_alias: {total: DESC}}) {
						total: AVG(HeightM: {}, Age: {})
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
			&action.Request{
				Request: `query {
					Users(order: {_alias: {total: ASC}}) {
						total: AVG(HeightM: {}, Age: {})
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
