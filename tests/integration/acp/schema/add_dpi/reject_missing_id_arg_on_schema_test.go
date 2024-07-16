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
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_AddDPISchema_NoPolicyIDWasSpecifiedOnSchema_SchemaRejected(t *testing.T) {
	policyIDOfValidDPI := "d59f91ba65fe142d35fc7df34482eafc7e99fed7c144961ba32c4664634e61b7"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, but no policyID was specified on schema, reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Identity: immutable.Some(1),

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

				ExpectedPolicyID: policyIDOfValidDPI,
			},

			testUtils.SchemaUpdate{
				Schema: `
					type Users @policy(resource: "users") {
						name: String
						age: Int
					}
				`,
				ExpectedError: "policyID must not be empty",
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

func TestACP_AddDPISchema_SpecifiedPolicyIDArgIsEmptyOnSchema_SchemaRejected(t *testing.T) {
	policyIDOfValidDPI := "d59f91ba65fe142d35fc7df34482eafc7e99fed7c144961ba32c4664634e61b7"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, specified policyID arg on schema is empty, reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Identity: immutable.Some(1),

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

				ExpectedPolicyID: policyIDOfValidDPI,
			},

			testUtils.SchemaUpdate{
				Schema: `
					type Users @policy(resource: "users", id: "") {
						name: String
						age: Int
					}
				`,
				ExpectedError: "policyID must not be empty",
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
