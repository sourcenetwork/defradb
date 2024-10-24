// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_p2p

import (
	"fmt"
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_P2PUpdatePrivateDocumentsOnDifferentNodes_SourceHubACP(t *testing.T) {
	expectedPolicyID := "fc56b7509c20ac8ce682b3b9b4fdaad868a9c70dda6ec16720298be64f16e9a4"

	test := testUtils.TestCase{

		Description: "Test acp, p2p update private documents on different nodes, with source-hub",

		SupportedACPTypes: immutable.Some(
			[]testUtils.ACPType{
				testUtils.SourceHubACPType,
			},
		),

		Actions: []any{
			testUtils.RandomNetworkingConfig(),

			testUtils.RandomNetworkingConfig(),

			testUtils.AddPolicy{

				Identity: testUtils.UserIdentity(1),

				Policy: `
                    name: Test Policy

                    description: A Policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          read:
                            expr: owner + reader + writer

                          write:
                            expr: owner + writer

                          nothing:
                            expr: dummy

                        relations:
                          owner:
                            types:
                              - actor

                          reader:
                            types:
                              - actor

                          writer:
                            types:
                              - actor

                          admin:
                            manages:
                              - reader
                            types:
                              - actor

                          dummy:
                            types:
                              - actor
                `,

				ExpectedPolicyID: expectedPolicyID,
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
					expectedPolicyID,
				),
			},

			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},

			testUtils.CreateDoc{
				Identity: testUtils.UserIdentity(1),

				NodeID: immutable.Some(0),

				CollectionID: 0,

				DocMap: map[string]any{
					"name": "Shahzad",
				},
			},

			testUtils.CreateDoc{
				Identity: testUtils.UserIdentity(1),

				NodeID: immutable.Some(1),

				CollectionID: 0,

				DocMap: map[string]any{
					"name": "Shahzad Lone",
				},
			},

			testUtils.WaitForSync{},

			testUtils.UpdateDoc{
				Identity: testUtils.UserIdentity(1),

				NodeID: immutable.Some(0),

				CollectionID: 0,

				DocID: 0,

				Doc: `
					{
						"name": "ShahzadLone"
					}
				`,
			},

			testUtils.UpdateDoc{
				Identity: testUtils.UserIdentity(1),

				NodeID: immutable.Some(1),

				CollectionID: 0,

				DocID: 1,

				Doc: `
					{
						"name": "ShahzadLone"
					}
				`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
