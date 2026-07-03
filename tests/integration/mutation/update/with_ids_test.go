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

package update

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpdate_WithIds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						points: Float
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"points": 42.1
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Bob",
					"points": 66.6
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Fred",
					"points": 33
				}`,
			},
			&action.Request{
				Request: `mutation {
					update_Users(
						docID: ["{{.DocID0_0}}", "{{.DocID0_2}}"],
						input: {points: 59}
					) {
						name
						points
					}
				}`,
				Results: map[string]any{
					"update_Users": []map[string]any{
						{
							"name":   "John",
							"points": float64(59),
						},
						{
							"name":   "Fred",
							"points": float64(59),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
