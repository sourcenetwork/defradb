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

func TestACP_LinkSchema_OwnerRelationWithDifferenceSetOpOnReadPermissionExprOnDRI_SchemaRejected(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, link schema, owner relation with difference (-) set operation on read permission expression, reject schema",

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
                            expr: owner - reader
                          update:
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

				ExpectedError: fmt.Sprintf(
					"expr of required permission has invalid character after relation. Permission: %s, Relation: %s, Character: %s",
					"read",
					"owner",
					"-",
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

func TestACP_LinkSchema_OwnerRelationWithIntersectionSetOpOnReadPermissionExprOnDRI_SchemaRejected(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, link schema, owner relation with intersection (&) set operation on read permission expression, reject schema",

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
                            expr: owner & reader
                          update:
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

				ExpectedError: fmt.Sprintf(
					"expr of required permission has invalid character after relation. Permission: %s, Relation: %s, Character: %s",
					"read",
					"owner",
					"&",
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

func TestACP_LinkSchema_OwnerRelationWithInvalidSetOpOnReadPermissionExprOnDRI_SchemaRejected(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, link schema, owner relation with invalid set operation on read permission expression, reject schema",

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
                            expr: owner - owner
                          update:
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

				ExpectedError: fmt.Sprintf(
					"expr of required permission has invalid character after relation. Permission: %s, Relation: %s, Character: %s",
					"read",
					"owner",
					"-",
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
