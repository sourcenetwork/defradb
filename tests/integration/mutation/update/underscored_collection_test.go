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

func TestMutationUpdateUnderscoredCollection(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			&action.AddCollection{
				SDL: `
					type My_User {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"name": "Fred"
				}`,
			},
			&action.Request{
				Request: `
					query {
						My_User {
							name
						}
					}
				`,
				Results: map[string]any{
					"My_User": []map[string]any{
						{
							"name": "Fred",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
