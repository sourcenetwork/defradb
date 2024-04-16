// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_schema_add_dpi

import (
	"fmt"
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_AddDPISchema_PartialValidDPIButUseInValidDPIResource_RejectSchema(t *testing.T) {
	policyIDOfPartiallyValidDPI := "d5d411825b2d8fa5a550f1e34153b88b375ed9c9af19ce6d2ba1769e237a45d0"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, has both valid & invalid resources, but use invalid resource, schema rejected",

		Actions: []any{

			testUtils.AddPolicy{

				Creator: actor1Identity,

				Policy: `
                    description: A Partially Valid Defra Policy Interface (DPI)

                    actor:
                      name: actor

                    resources:
                      usersValid:
                        permissions:
                          read:
                            expr: owner + reader
                          write:
                            expr: owner

                        relations:
                          owner:
                            types:
                              - actor
                          reader:
                            types:
                              - actor

                      usersInvalid:
                        permissions:
                          read:
                            expr: reader - owner
                          write:
                            expr: reader

                        relations:
                          owner:
                            types:
                              - actor
                          reader:
                            types:
                              - actor
                `,

				ExpectedPolicyID: policyIDOfPartiallyValidDPI,
			},

			testUtils.SchemaUpdate{
				Schema: fmt.Sprintf(`
					type Users @policy(
						id: "%s",
						resource: "usersInvalid"
					) {
						name: String
						age: Int
					}
				`,
					policyIDOfPartiallyValidDPI,
				),

				ExpectedError: fmt.Sprintf(
					"expr of required permission must start with required relation. Permission: %s, Relation: %s",
					"read",
					"owner",
				),
			},

			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "Users") {
							name
							fields {
								name
								type {
								name
								kind
								}
							}
						}
					}
				`,
				ExpectedData: map[string]any{
					"__type": nil, // NOTE: No "Users" should exist.
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
