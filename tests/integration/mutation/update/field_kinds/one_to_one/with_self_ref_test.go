// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_one

import (
	"fmt"
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpdateOneToOne_SelfReferencingFromPrimary(t *testing.T) {
	user1ID := "bae-93b58e20-b3e1-55b9-b5b8-0617fabe710e"

	test := testUtils.TestCase{
		Description: "One to one update mutation, self referencing from primary",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						boss: User @primary
						underling: User
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred"
				}`,
			},
			testUtils.UpdateDoc{
				DocID: 1,
				Doc: fmt.Sprintf(
					`{
						"boss_id": "%s"
					}`,
					user1ID,
				),
			},
			testUtils.Request{
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
			},
			testUtils.Request{
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
