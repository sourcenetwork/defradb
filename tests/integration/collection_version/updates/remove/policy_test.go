// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package remove

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColVersionUpdateRemovePolicy_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.AddPolicyWithDAC{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
                    name: test
                    description: a test policy which marks a collection in a database as a resource

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          read:
                            expr: owner + reader
                          update:
                            expr: owner
                          delete:
                            expr: owner

                        relations:
                          owner:
                            types:
                              - actor
                          reader:
                            types:
                              - actor
                          admin:
                            manages:
                              - reader
                            types:
                              - actor
                `,
			},

			testUtils.SchemaUpdate{
				Schema: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,

				Replace: map[string]testUtils.ReplaceType{
					"Policy0": testUtils.NewPolicyIndex(0),
				},
			},

			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "remove",
							"path": "/bafkreihhd6bqrjhl5zidwztgxzeseveplv3cj3fwtn3unjkdx7j2vr2vrq/Policy"
						}
					]
				`,
				ExpectedError: "collection policy cannot be mutated.",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
