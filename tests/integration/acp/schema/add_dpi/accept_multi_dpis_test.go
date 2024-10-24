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
        name: test
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

	const policyIDOfFirstCreatorsDPI string = "d59f91ba65fe142d35fc7df34482eafc7e99fed7c144961ba32c4664634e61b7"
	const policyIDOfSecondCreatorsDPI string = "4b9291094984289a8f5557d142db453943549626067eedd8cbd5b64c3bc8a4f3"

	test := testUtils.TestCase{

		Description: "Test acp, add duplicate DPIs by different actors, accept both schemas",

		Actions: []any{
			testUtils.AddPolicy{

				Identity: testUtils.UserIdentity(1),

				Policy: validDPIUsedByBoth,

				ExpectedPolicyID: policyIDOfFirstCreatorsDPI,
			},

			testUtils.AddPolicy{

				Identity: testUtils.UserIdentity(2),

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
