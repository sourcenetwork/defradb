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
	policyIDOfInvalidDPI := "4256d2b54767cafd0e0a2b39a6faebf44bc99a7fc74ff5b51894f7accf2ef638"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, with owner missing required write permission, reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Creator: actor1Identity,

				Policy: `
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
	policyIDOfInvalidDPI := "e8be944571cd6b52faa1e8b75fa339a9f60065b65d78ed126d037722e2512593"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, with owner missing required write permission label, reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Creator: actor1Identity,

				Policy: `
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
	policyIDOfInvalidDPI := "34ff30cb9e80993e2b11f86f85c6daa7cd9bf25724e4d5ff0704518d7970d074"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, owner specified incorrectly on write permission expression, reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Creator: actor1Identity,

				Policy: `
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
	policyIDOfInvalidDPI := "2e9fc5805b0442e856e9893fea0f4759d333e442856a230ed741b88670e6426c"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, owner specified incorrectly on write permission expression (no space), reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Creator: actor1Identity,

				Policy: `
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
	policyIDOfInvalidDPI := "3bcd650ac1e69d5efe6c930d05420231a0a69e6018d0f1015e0ecef9869d8dd5"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, malicious owner specified on write permission expression, reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Creator: actor1Identity,

				Policy: `
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
