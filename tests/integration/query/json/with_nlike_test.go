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

package json

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryJSON_WithNotLikeFilter_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						custom: JSON
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"custom": "Daenerys Stormborn of House Targaryen, the First of Her Name"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"custom": "Viserys I Targaryen, King of the Andals"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"custom": [1, 2]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"custom": {"one": 1}
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"custom": false
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"custom": 32
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {custom: {_nlike: "%Stormborn%"}}) {
						custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"custom": false,
						},
						{
							"custom": "Viserys I Targaryen, King of the Andals",
						},
						{
							"custom": map[string]any{"one": float64(1)},
						},
						{
							"custom": float64(32),
						},
						{
							"custom": []any{float64(1), float64(2)},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
