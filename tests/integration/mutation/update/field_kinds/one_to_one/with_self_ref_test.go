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

package one_to_one

import (
	"fmt"
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpdateOneToOne_SelfReferencingFromPrimary(t *testing.T) {
	user1ID := "bae-1d57efc8-a1f3-5b0e-9d08-51e03359285e"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						boss: User @primary
						underling: User
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Fred"
				}`,
			},
			&action.UpdateDoc{
				DocID: 1,
				Doc: fmt.Sprintf(
					`{
						"_bossID": "%s"
					}`,
					user1ID,
				),
			},
			&action.Request{
				Request: `
					query {
						User {
							name
							boss {
								name
							}
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
							"boss": nil,
						},
						{
							"name": "Fred",
							"boss": map[string]any{
								"name": "John",
							},
						},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				Request: `
					query {
						User {
							name
							underling {
								name
							}
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
							"underling": map[string]any{
								"name": "Fred",
							},
						},
						{
							"name":      "Fred",
							"underling": nil,
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
