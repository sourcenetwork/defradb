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

func TestQueryJSON_WithInFilter_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"custom": {
						"tree": "maple",
						"age": 250
					}
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Andy",
					"custom": {
						"tree": "oak",
						"age": 450
					}
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {custom: {_in: [{tree:"oak",age:450}]}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "Andy"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
