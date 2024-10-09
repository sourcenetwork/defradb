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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryJSON_WithLikeFilter_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"custom": "Daenerys Stormborn of House Targaryen, the First of Her Name",
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"custom": "Viserys I Targaryen, King of the Andals",
				},
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
					Users(filter: {custom: {_like: "Daenerys%Name"}}) {
						custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"custom": "Daenerys Stormborn of House Targaryen, the First of Her Name",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
