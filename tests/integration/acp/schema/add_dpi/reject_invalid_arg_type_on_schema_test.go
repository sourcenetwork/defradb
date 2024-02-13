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

func TestACP_AddDPISchema_InvalidPolicyIDArgTypeWasSpecifiedOnSchema_SchemaRejected(t *testing.T) {
	policyIDOfValidDPI := "4f13c5084c3d0e1e5c5db702fceef84c3b6ab948949ca8e27fcaad3fb8bc39f4"

	test := testUtils.TestCase{
		Description: "Test acp, add dpi schema, but invalid policyID arg type was specified on schema, reject schema",

		Actions: []any{

			testUtils.AddPolicy{
				IsYAML: true,

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

				ExpectedPolicyID: policyIDOfValidDPI,
			},

			testUtils.SchemaUpdate{
				Schema: `
					type Users @policy(id: 123 , resource: "users") {
						name: String
						age: Int
					}
				`,
				ExpectedError: "policy directive with invalid id property",
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

func TestACP_AddDPISchema_InvalidResourceArgTypeWasSpecifiedOnSchema_SchemaRejected(t *testing.T) {
	policyIDOfValidDPI := "4f13c5084c3d0e1e5c5db702fceef84c3b6ab948949ca8e27fcaad3fb8bc39f4"

	test := testUtils.TestCase{
		Description: "Test acp, add dpi schema, but invalid resource arg type was specified on schema, reject schema",

		Actions: []any{

			testUtils.AddPolicy{
				IsYAML: true,

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

				ExpectedPolicyID: policyIDOfValidDPI,
			},

			testUtils.SchemaUpdate{
				Schema: fmt.Sprintf(`
					type Users @policy(id: "%s" , resource: 123) {
						name: String
						age: Int
					}
				`,
					policyIDOfValidDPI,
				),

				ExpectedError: "policy directive with invalid resource property",
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
