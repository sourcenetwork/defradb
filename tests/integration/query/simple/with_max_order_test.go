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

func TestQuerySimpleWithMaxWithOrder_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Age": 30,
					"HeightM": 1.8
				}`,
			}, // Max: 30

			&action.AddDoc{
				Doc: `{
					"Age": 25,
					"HeightM": 1.6
				}`,
			}, // Max: 25

			// Test descending order by computed total
			&action.Request{
				Request: `query {
					Users(order: {_alias: {total: DESC}}) {
						total: MAX(HeightM: {}, Age: {})
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
			&action.Request{
				Request: `query {
					Users(order: {_alias: {total: ASC}}) {
						total: MAX(HeightM: {}, Age: {})
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
