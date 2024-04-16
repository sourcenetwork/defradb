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

func TestACP_AddDPISchema_UseSameResourceOnDifferentSchemas_AcceptSchemas(t *testing.T) {
	policyIDOfValidDPI := "4f13c5084c3d0e1e5c5db702fceef84c3b6ab948949ca8e27fcaad3fb8bc39f4"
	sharedSameResourceName := "users"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, where one resource is specified on different schemas, schemas accepted",

		Actions: []any{

			testUtils.AddPolicy{

				Creator: actor1Identity,

				Policy: `
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
				Schema: fmt.Sprintf(`
					type OldUsers @policy(
						id: "%s",
						resource: "%s"
					) {
						name: String
						age: Int
					}
				`,
					policyIDOfValidDPI,
					sharedSameResourceName,
				),
			},

			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "OldUsers") {
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
						"name": "OldUsers", // NOTE: "OldUsers" MUST exist
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

			testUtils.SchemaUpdate{
				Schema: fmt.Sprintf(`
					type NewUsers @policy(
						id: "%s",
						resource: "%s"
					) {
						name: String
						age: Int
					}
				`,
					policyIDOfValidDPI,
					sharedSameResourceName,
				),
			},

			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "NewUsers") {
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
						"name": "NewUsers", // NOTE: "NewUsers" MUST exist
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
