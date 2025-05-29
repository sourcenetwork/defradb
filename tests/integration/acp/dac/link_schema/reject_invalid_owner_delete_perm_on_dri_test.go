// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_dac_link_schema

import (
	"fmt"
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_LinkSchema_OwnerMissingRequiredDeletePermissionOnDRI_SchemaRejected(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, link schema, with owner missing required delete permission, reject schema",

		Actions: []any{

			testUtils.AddPolicyWithDAC{

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
                            expr: owner
                          delete:
                            expr: w

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
					"delete",
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

func TestACP_LinkSchema_OwnerMissingRequiredDeletePermissionLabelOnDRI_SchemaRejected(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, link schema, with owner missing required delete permission label, reject schema",

		Actions: []any{

			testUtils.AddPolicyWithDAC{

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

func TestACP_LinkSchema_OwnerSpecifiedIncorrectlyOnDeletePermissionExprOnDRI_SchemaRejected(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, link schema, owner specified incorrectly on delete permission expression, reject schema",

		Actions: []any{

			testUtils.AddPolicyWithDAC{

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
                             expr: owner
                           delete:
                             expr: deleter + owner

                         relations:
                           owner:
                             types:
                               - actor
                           deleter:
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
					"delete",
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

func TestACP_LinkSchema_OwnerSpecifiedIncorrectlyOnDeletePermissionNoSpaceExprOnDRI_SchemaRejected(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, link schema, owner specified incorrectly on delete permission expression (no space), reject schema",

		Actions: []any{

			testUtils.AddPolicyWithDAC{

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
                             expr: owner
                           delete:
                             expr: deleter+owner

                         relations:
                           owner:
                             types:
                               - actor
                           deleter:
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
					"delete",
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

func TestACP_LinkSchema_MaliciousOwnerSpecifiedOnDeletePermissionExprOnDRI_SchemaRejected(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, link schema, malicious owner specified on delete permission expression, reject schema",

		Actions: []any{

			testUtils.AddPolicyWithDAC{

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
                             expr: owner
                           delete:
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
					"delete",
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
