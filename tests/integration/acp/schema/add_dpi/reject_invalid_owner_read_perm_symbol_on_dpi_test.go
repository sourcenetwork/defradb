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

func TestACP_AddDPISchema_OwnerRelationWithDifferenceSetOpOnReadPermissionExprOnDPI_SchemaRejected(t *testing.T) {
	policyIDOfInvalidDPI := "c9bb1811862ded3a4a8a931054bd99ecabde3f41231c6aa2c50e1f1f5af2b5e8"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, owner relation with difference (-) set operation on read permission expression, reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Creator: actor1Signature,

				Policy: `
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          read:
                            expr: owner - reader
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

func TestACP_AddDPISchema_OwnerRelationWithIntersectionSetOpOnReadPermissionExprOnDPI_SchemaRejected(t *testing.T) {
	policyIDOfInvalidDPI := "7bff1d8a967df4de99f8daaa2567c660eb6e7b2c554c9a49bf831230e5d9eba6"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, owner relation with intersection (&) set operation on read permission expression, reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Creator: actor1Signature,

				Policy: `
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          read:
                            expr: owner & reader
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

func TestACP_AddDPISchema_OwnerRelationWithInvalidSetOpOnReadPermissionExprOnDPI_SchemaRejected(t *testing.T) {
	policyIDOfInvalidDPI := "cc2fab7c299e94e2bd9370708d26ca1262ff3b0d75f9a58d1086658cfec26c65"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, owner relation with invalid set operation on read permission expression, reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Creator: actor1Signature,

				Policy: `
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          read:
                            expr: owner - owner
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
