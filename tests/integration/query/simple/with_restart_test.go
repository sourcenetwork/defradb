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

func TestQuerySimpleWithRestart(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			testUtils.Restart{},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"age": 30
				}`,
			},
			testUtils.Restart{},
			&action.Request{
				Request: ` query {
					Users {
						name
						age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							"age":  int64(30),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
