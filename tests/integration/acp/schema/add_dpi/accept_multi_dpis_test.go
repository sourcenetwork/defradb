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

	const policyIDOfFirstCreatorsDPI string = "4f13c5084c3d0e1e5c5db702fceef84c3b6ab948949ca8e27fcaad3fb8bc39f4"
	const policyIDOfSecondCreatorsDPI string = "d33aa07a28ea19ed07a5256eb7e7f5600b0e0af13254889a7fce60202c4f6c7e"

	test := testUtils.TestCase{
		Description: "Test acp, add duplicate DPIs by different actors, accept both schemas",

		Actions: []any{
			testUtils.AddPolicy{
				IsYAML: true,

				Creator: actor1Signature,

				Policy: validDPIUsedByBoth,

				ExpectedPolicyID: policyIDOfFirstCreatorsDPI,
			},

			testUtils.AddPolicy{
				IsYAML: true,

				Creator: actor2Signature,

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
