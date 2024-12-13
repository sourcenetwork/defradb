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

func TestACP_AddDPISchema_OwnerMissingRequiredWritePermissionOnDPI_SchemaRejected(t *testing.T) {
	policyIDOfInvalidDPI := "d0f27cd0f75f3f3048bfa06e2a45f8dcfd172767437d48bf584eb3881f90c373"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, with owner missing required write permission, reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
                    name: test
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          write:
                            expr: w
                          read:
                            expr: owner

                        relations:
                          owner:
                            types:
                              - actor
                          w:
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
					"write",
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

func TestACP_AddDPISchema_OwnerMissingRequiredWritePermissionLabelOnDPI_SchemaRejected(t *testing.T) {
	policyIDOfInvalidDPI := "73c85066924b28fafb4103fc764ae35f6891a50292634df5a870f45525095bca"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, with owner missing required write permission label, reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
                     name: test
                     description: a policy

                     actor:
                       name: actor

                     resources:
                       users:
                         permissions:
                           read:
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
					"write",
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

func TestACP_AddDPISchema_OwnerSpecifiedIncorrectlyOnWritePermissionExprOnDPI_SchemaRejected(t *testing.T) {
	policyIDOfInvalidDPI := "1e4cd20ab842fb837ed3e03c69014732593261ffc56142f7c907f32200f6d255"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, owner specified incorrectly on write permission expression, reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
                     name: test
                     description: a policy

                     actor:
                       name: actor

                     resources:
                       users:
                         permissions:
                           read:
                             expr: owner
                           write:
                             expr: writer + owner

                         relations:
                           owner:
                             types:
                               - actor
                           writer:
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
					"write",
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

func TestACP_AddDPISchema_OwnerSpecifiedIncorrectlyOnWritePermissionNoSpaceExprOnDPI_SchemaRejected(t *testing.T) {
	policyIDOfInvalidDPI := "44267005866f1bef0613e442254c90491bd26625be54f6c5bec10a9b83cf6f05"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, owner specified incorrectly on write permission expression (no space), reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
                     name: test
                     description: a policy

                     actor:
                       name: actor

                     resources:
                       users:
                         permissions:
                           read:
                             expr: owner
                           write:
                             expr: writer+owner

                         relations:
                           owner:
                             types:
                               - actor
                           writer:
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
					"write",
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

func TestACP_AddDPISchema_MaliciousOwnerSpecifiedOnWritePermissionExprOnDPI_SchemaRejected(t *testing.T) {
	policyIDOfInvalidDPI := "306fb2351dbe999e4f6f785f0f3ec353a389c82babedca49e375cca2d45b1ca8"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, malicious owner specified on write permission expression, reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
                     name: test
                     description: a policy

                     actor:
                       name: actor

                     resources:
                       users:
                         permissions:
                           read:
                             expr: owner
                           write:
                             expr: ownerBad

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
					"write",
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
