// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package json

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryJSON_WithNotLikeFilter_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"custom": "Daenerys Stormborn of House Targaryen, the First of Her Name"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"custom": "Viserys I Targaryen, King of the Andals"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"custom": [1, 2]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"custom": {"one": 1}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"custom": false
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"custom": 32
				}`,
			},
			testUtils.Request{
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
