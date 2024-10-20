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

func TestACP_P2PReplicatorWithPermissionedCollectionCreateDocActorRelationship_SourceHubACP(t *testing.T) {
	expectedPolicyID := "fc56b7509c20ac8ce682b3b9b4fdaad868a9c70dda6ec16720298be64f16e9a4"

	test := testUtils.TestCase{

		Description: "Test acp, p2p replicator with collection that has a policy, create a new doc-actor relationship",

		SupportedACPTypes: immutable.Some(
			[]testUtils.ACPType{
				testUtils.SourceHubACPType,
			},
		),

		Actions: []any{
			testUtils.RandomNetworkingConfig(),

			testUtils.RandomNetworkingConfig(),

			testUtils.AddPolicy{

				Identity: immutable.Some(1),

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
				Identity: immutable.Some(1),

				NodeID: immutable.Some(0),

				CollectionID: 0,

				DocMap: map[string]any{
					"name": "Shahzad",
				},
			},

			testUtils.WaitForSync{},

			testUtils.Request{
				// Ensure that the document is hidden on all nodes to an unauthorized actor
				Identity: immutable.Some(2),

				Request: `
					query {
						Users {
							name
						}
					}
				`,

				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},

			testUtils.AddDocActorRelationship{
				NodeID: immutable.Some(0),

				RequestorIdentity: 1,

				TargetIdentity: 2,

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedExistence: false,
			},

			testUtils.AddDocActorRelationship{
				NodeID: immutable.Some(1), // Note: Different node than the previous

				RequestorIdentity: 1,

				TargetIdentity: 2,

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedExistence: true, // Making the same relation through any node should be a no-op
			},

			testUtils.Request{
				// Ensure that the document is now accessible on all nodes to the newly authorized actor.
				Identity: immutable.Some(2),

				Request: `
					query {
						Users {
							name
						}
					}
				`,

				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},

			testUtils.Request{
				// Ensure that the document is still accessible on all nodes to the owner.
				Identity: immutable.Some(1),

				Request: `
					query {
						Users {
							name
						}
					}
				`,

				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},

			testUtils.DeleteDocActorRelationship{
				NodeID: immutable.Some(1),

				RequestorIdentity: 1,

				TargetIdentity: 2,

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedRecordFound: true,
			},

			testUtils.DeleteDocActorRelationship{
				NodeID: immutable.Some(0), // Note: Different node than the previous

				RequestorIdentity: 1,

				TargetIdentity: 2,

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedRecordFound: false, // Making the same relation through any node should be a no-op
			},

			testUtils.Request{
				// Ensure that the document is now inaccessible on all nodes to the actor we revoked access from.
				Identity: immutable.Some(2),

				Request: `
					query {
						Users {
							name
						}
					}
				`,

				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},

			testUtils.Request{
				// Ensure that the document is still accessible on all nodes to the owner.
				Identity: immutable.Some(1),

				Request: `
					query {
						Users {
							name
						}
					}
				`,

				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
