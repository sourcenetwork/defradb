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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpdateOneToOne_SelfReferencingFromPrimary(t *testing.T) {
	user1ID := "bae-1d57efc8-a1f3-5b0e-9d08-51e03359285e"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
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
			testUtils.UpdateDoc{
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
