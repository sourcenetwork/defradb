// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package create

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"

	"github.com/sourcenetwork/immutable"
)

func TestMutationCreateWithNonNullVariable(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple create mutation with non null variable input.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.Request{
				Variables: immutable.Some(map[string]any{
					"user": map[string]any{
						"name": "Bob",
					},
				}),
				Request: `mutation($user: [UsersMutationInputArg!]!) {
					create_Users(input: $user) {
						name
					}
				}`,
				Results: map[string]any{
					"create_Users": []map[string]any{
						{
							"name": "Bob",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreateWithDefaultVariable(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple create mutation with default variable input.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.Request{
				Request: `mutation($user: [UsersMutationInputArg!] = {name: "Bob"}) {
					create_Users(input: $user) {
						name
					}
				}`,
				Results: map[string]any{
					"create_Users": []map[string]any{
						{
							"name": "Bob",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
