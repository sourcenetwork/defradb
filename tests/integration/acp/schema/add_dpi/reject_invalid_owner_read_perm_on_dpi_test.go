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

func TestACP_AddDPISchema_OwnerMissingRequiredReadPermissionOnDPI_SchemaRejected(t *testing.T) {
	policyIDOfInvalidDPI := "d3b58e284fb1117eb9c0111105195c910a6873d70695d8d3af3a967632bd5a5c"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, with owner missing required read permission, reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Identity: actor1Identity,

				Policy: `
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          write:
                            expr: owner
                          read:
                            expr: r

                        relations:
                          owner:
                            types:
                              - actor
                          r:
                            types:
                              - actor
                `,

				ExpectedPolicyID: policyIDOfInvalidDPI,
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
					policyIDOfInvalidDPI,
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

func TestACP_AddDPISchema_OwnerMissingRequiredReadPermissionLabelOnDPI_SchemaRejected(t *testing.T) {
	policyIDOfInvalidDPI := "6f81aa1e33a6888e041ed7686f21007d09b30ae218fb2ce19071ea2155482b23"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, with owner missing required read permission label, reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Identity: actor1Identity,

				Policy: `
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
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

				ExpectedPolicyID: policyIDOfInvalidDPI,
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
					policyIDOfInvalidDPI,
				),

				ExpectedError: fmt.Sprintf(
					"resource is missing required permission on policy. PolicyID: %s, ResourceName: %s, Permission: %s",
					policyIDOfInvalidDPI,
					"users",
					"read",
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

func TestACP_AddDPISchema_OwnerSpecifiedIncorrectlyOnReadPermissionExprOnDPI_SchemaRejected(t *testing.T) {
	policyIDOfInvalidDPI := "7f291632f6772e108830d41a5388391fc04cd4f2473d2a759d2a3326ee973848"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, owner specified incorrectly on read permission expression, reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Identity: actor1Identity,

				Policy: `
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          read:
                            expr: reader + owner
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

				ExpectedPolicyID: policyIDOfInvalidDPI,
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
					policyIDOfInvalidDPI,
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

func TestACP_AddDPISchema_OwnerSpecifiedIncorrectlyOnReadPermissionNoSpaceExprOnDPI_SchemaRejected(t *testing.T) {
	policyIDOfInvalidDPI := "aa4c2b7bc9177a7b1cb808289a7f03564c489cc5c829ef756997cbe47b689a3f"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, owner specified incorrectly on read permission expression (no space), reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Identity: actor1Identity,

				Policy: `
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          read:
                            expr: reader+owner
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

				ExpectedPolicyID: policyIDOfInvalidDPI,
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
					policyIDOfInvalidDPI,
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

func TestACP_AddDPISchema_MaliciousOwnerSpecifiedOnReadPermissionExprOnDPI_SchemaRejected(t *testing.T) {
	policyIDOfInvalidDPI := "f52f9a23643c6d52f4aaf2569885c152f06edc7d95647f88f6c4e8ead757e792"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, malicious owner specified on read permission expression, reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Identity: actor1Identity,

				Policy: `
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          read:
                            expr: ownerBad
                          write:
                            expr: owner

                        relations:
                          owner:
                            types:
                              - actor
                          ownerBad:
                            types:
                              - actor
                `,

				ExpectedPolicyID: policyIDOfInvalidDPI,
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
					policyIDOfInvalidDPI,
				),

				ExpectedError: fmt.Sprintf(
					"expr of required permission has invalid character after relation. Permission: %s, Relation: %s, Character: %s",
					"read",
					"owner",
					"B",
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
