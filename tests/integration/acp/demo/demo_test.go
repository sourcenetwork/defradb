// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_demo

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_DEMO(t *testing.T) {
	test := testUtils.TestCase{

		Description: "DEMO",

		Actions: []any{
			testUtils.AddPolicy{ // Specified and subbed in for User1 and User2 schema below

				Identity: immutable.Some(1),

				Policy: `
                    name: test
                    description: a test policy which marks a collection in a database as a resource

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
                          admin:
                            manages:
                              - reader
                            types:
                              - actor
                `,

				// TODO: Remove after draft is approved
				// Note: This is also valid now that it is optional, but we don't need to provide it at all.
				ExpectedPolicyID: immutable.Some(
					"94eb195c0e459aa79e02a1986c7e731c5015721c18a373f2b2a0ed140a04b454",
				),
			},

			testUtils.AddPolicy{ // Specified and subbed in for User3 schema below

				Identity: immutable.Some(1),

				Policy: `
                    name: test
                    description: another policy

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
                          admin:
                            manages:
                              - reader
                            types:
                              - actor
                `,

				// TODO: Remove after draft is approved
				// Note: No policyID assertion needed
			},

			testUtils.SchemaUpdate{
				Schema: `
					type Users1 @policy(
						id: "%policyID%",
						resource: "users"
					) {
						name: String
						age: Int
					}

					type User2 @policy(
						id: "%policyID%",
						resource: "users"
					) {
						name: String
						age: Int
					}

					type User3 @policy(
						id: "%policyID%",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,

				PolicyIDs: immutable.Some([]int{0, 0, 1}),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
