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

func TestACP_AddDPISchema_OwnerMissingRequiredUpdatePermissionOnDPI_SchemaRejected(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, with owner missing required update permission, reject schema",

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
                          update:
                            expr: w
                          delete:
                            expr: owner

                        relations:
                          owner:
                            types:
                              - actor
                          w:
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

				ExpectedError: fmt.Sprintf(
					"expr of required permission must start with required relation. Permission: %s, Relation: %s",
					"update",
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

func TestACP_AddDPISchema_OwnerMissingRequiredUpdatePermissionLabelOnDPI_SchemaRejected(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, with owner missing required update permission label, reject schema",

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
                           delete:
                             expr: owner

                         relations:
                           owner:
                             types:
                               - actor
                           reader:
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

				ExpectedError: "resource is missing required permission on policy.",
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

func TestACP_AddDPISchema_OwnerSpecifiedIncorrectlyOnUpdatePermissionExprOnDPI_SchemaRejected(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, owner specified incorrectly on update permission expression, reject schema",

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
                           update:
                             expr: updater + owner
                           delete:
                             expr: owner

                         relations:
                           owner:
                             types:
                               - actor
                           updater:
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

				ExpectedError: fmt.Sprintf(
					"expr of required permission must start with required relation. Permission: %s, Relation: %s",
					"update",
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

func TestACP_AddDPISchema_OwnerSpecifiedIncorrectlyOnUpdatePermissionNoSpaceExprOnDPI_SchemaRejected(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, owner specified incorrectly on update permission expression (no space), reject schema",

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
                           update:
                             expr: updater+owner
                           delete:
                             expr: owner

                         relations:
                           owner:
                             types:
                               - actor
                           updater:
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

				ExpectedError: fmt.Sprintf(
					"expr of required permission must start with required relation. Permission: %s, Relation: %s",
					"update",
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

func TestACP_AddDPISchema_MaliciousOwnerSpecifiedOnUpdatePermissionExprOnDPI_SchemaRejected(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, malicious owner specified on update permission expression, reject schema",

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
                           update:
                             expr: ownerBad
                           delete:
                             expr: owner

                         relations:
                           owner:
                             types:
                               - actor
                           ownerBad:
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

				ExpectedError: fmt.Sprintf(
					"expr of required permission has invalid character after relation. Permission: %s, Relation: %s, Character: %s",
					"update",
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
