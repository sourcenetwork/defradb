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
	schemaUtils "github.com/sourcenetwork/defradb/tests/integration/schema"
)

func TestACP_AddDPISchema_BasicYAML_SchemaAccepted(t *testing.T) {
	policyIDOfValidDPI := "aa664afaf8dff947ba85f4d464662d595af6c1e2466bd11fd6b82ea95b547ea3"

	test := testUtils.TestCase{

		Description: "Test acp, specify basic policy that was added in YAML format, accept schema",

		Actions: []any{
			testUtils.AddPolicy{

				Identity: actor1Identity,

				Policy: `
                    description: a basic policy that satisfies minimum DPI requirements

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          read:
                            expr: owner
                          write:
                            expr: owner

                        relations:
                          owner:
                            types:
                              - actor

                `,

				ExpectedPolicyID: policyIDOfValidDPI,
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
					policyIDOfValidDPI,
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
					"__type": map[string]any{
						"name": "Users", // NOTE: "Users" MUST exist
						"fields": schemaUtils.DefaultFields.Append(
							schemaUtils.Field{
								"name": "name",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "String",
								},
							},
						).Append(
							schemaUtils.Field{
								"name": "age",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "Int",
								},
							},
						).Tidy(),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddDPISchema_BasicJSON_SchemaAccepted(t *testing.T) {
	policyIDOfValidDPI := "aa664afaf8dff947ba85f4d464662d595af6c1e2466bd11fd6b82ea95b547ea3"

	test := testUtils.TestCase{

		Description: "Test acp, specify basic policy that was added in JSON format, accept schema",

		Actions: []any{
			testUtils.AddPolicy{

				Identity: actor1Identity,

				Policy: `
					{
					  "description": "a basic policy that satisfies minimum DPI requirements",
					  "resources": {
					    "users": {
					      "permissions": {
					        "read": {
					          "expr": "owner"
					        },
					        "write": {
					          "expr": "owner"
					        }
					      },
					      "relations": {
					        "owner": {
					          "types": [
					            "actor"
					          ]
					        }
					      }
					    }
					  },
					  "actor": {
					    "name": "actor"
					  }
					}
                `,

				ExpectedPolicyID: policyIDOfValidDPI,
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
					policyIDOfValidDPI,
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
					"__type": map[string]any{
						"name": "Users", // NOTE: "Users" MUST exist
						"fields": schemaUtils.DefaultFields.Append(
							schemaUtils.Field{
								"name": "name",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "String",
								},
							},
						).Append(
							schemaUtils.Field{
								"name": "age",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "Int",
								},
							},
						).Tidy(),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
