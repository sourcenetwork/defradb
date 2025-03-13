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

func TestACP_AddDPISchema_OwnerSpecifiedIncorrectlyOnWritePermissionExprOnDPI_SchemaRejected(t *testing.T) {
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
