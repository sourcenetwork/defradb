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
	nonExistingPolicyID := "468a5f345b3afec72f025185159e0fe84b02ead3374ec6aa54c390b2e3299444"

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

				ExpectedError: "policyID specified does not exist with acp",
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
	policyAdded := "230bc7230d06a741e395b1ee093b03d0b3f8fd8c9c21727575db10219fb55be1"
	incorrectPolicyID := "468a5f345b3afec72f025185159e0fe84b02ead3374ec6aa54c390b2e3299444"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, but specify incorrect policy ID, reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
                    name: test
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

				ExpectedError: "policyID specified does not exist with acp",
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
