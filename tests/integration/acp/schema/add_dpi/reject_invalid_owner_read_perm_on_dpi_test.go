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
	policyIDOfInvalidDPI := "017d821ad9a540b639769fab6a58a5bf27591428f7acbf0d1ecf92dbcc4c4ba4"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, with owner missing required read permission, reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Identity: actor1Identity,

				Policy: `
                    name: test
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
	policyIDOfInvalidDPI := "d56ef73738a0dd50fe1763c451ed3a9b774b38c268ed0c510058bb60a1df02de"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, with owner missing required read permission label, reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Identity: actor1Identity,

				Policy: `
                    name: test
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
	policyIDOfInvalidDPI := "75f0320e2d949cd201ce40e556c2e7f615ae5c1039197329fc6ed4bca6c7135d"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, owner specified incorrectly on read permission expression, reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Identity: actor1Identity,

				Policy: `
                    name: test
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
	policyIDOfInvalidDPI := "481cc0091eb0c30eb75e8676db564dadf3166fe617dcd576f34b38e4ef4797eb"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, owner specified incorrectly on read permission expression (no space), reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Identity: actor1Identity,

				Policy: `
                    name: test
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
	policyIDOfInvalidDPI := "79095d371699ab7c1aa0d7336645bafea4cd02d15af9c8bacf4d3ef6078488c6"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, malicious owner specified on read permission expression, reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Identity: actor1Identity,

				Policy: `
                    name: test
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
