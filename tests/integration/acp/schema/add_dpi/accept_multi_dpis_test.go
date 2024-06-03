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

func TestACP_AddDPISchema_AddDuplicateDPIsByOtherCreatorsUseBoth_AcceptSchema(t *testing.T) {
	const sameResourceNameOnBothDPI string = "users"
	const validDPIUsedByBoth string = `
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
    `

	const policyIDOfFirstCreatorsDPI string = "d5b240c738dba7fe7d8ae55acf257d8e4010c9d8b78e0b1f0bd26741b1ec5663"
	const policyIDOfSecondCreatorsDPI string = "6d2ec2fd16ed62a1cad05d8e791abe12cbbf9551080c0ca052336b49e635c291"

	test := testUtils.TestCase{

		Description: "Test acp, add duplicate DPIs by different actors, accept both schemas",

		Actions: []any{
			testUtils.AddPolicy{

				Identity: actor1Identity,

				Policy: validDPIUsedByBoth,

				ExpectedPolicyID: policyIDOfFirstCreatorsDPI,
			},

			testUtils.AddPolicy{

				Identity: actor2Identity,

				Policy: validDPIUsedByBoth,

				ExpectedPolicyID: policyIDOfSecondCreatorsDPI,
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
					policyIDOfFirstCreatorsDPI,
					sameResourceNameOnBothDPI,
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
					policyIDOfSecondCreatorsDPI,
					sameResourceNameOnBothDPI,
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
