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

func TestACP_AddDPISchema_WhereNoPolicyWasAdded_SchemaRejected(t *testing.T) {
	nonExistingPolicyID := "dfe202ffb4f0fe9b46157c313213a3839e08a6f0a7c3aba55e4724cb49ffde8a"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, but no policy was added, reject schema",

		Actions: []any{

			testUtils.SchemaUpdate{
				Schema: fmt.Sprintf(`
					type Users @policy(
						id: "%s",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
					nonExistingPolicyID,
				),

				ExpectedError: "policyID specified does not exist on acp module",
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

func TestACP_AddDPISchema_WhereAPolicyWasAddedButLinkedPolicyWasNotAdded_SchemaRejected(t *testing.T) {
	policyAdded := "4f13c5084c3d0e1e5c5db702fceef84c3b6ab948949ca8e27fcaad3fb8bc39f4"
	incorrectPolicyID := "dfe202ffb4f0fe9b46157c313213a3839e08a6f0a7c3aba55e4724cb49ffde8a"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, but specify incorrect policy ID, reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Creator: actor1Signature,

				Policy: `
                    description: A Valid Defra Policy Interface (DPI)

                    actor:
                      name: actor

                    resources:
                      users:
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
                `,

				ExpectedPolicyID: policyAdded,
			},

			testUtils.SchemaUpdate{
				Schema: fmt.Sprintf(`
					type Users @policy(
						id: "%s",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
					incorrectPolicyID,
				),

				ExpectedError: "policyID specified does not exist on acp module",
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
