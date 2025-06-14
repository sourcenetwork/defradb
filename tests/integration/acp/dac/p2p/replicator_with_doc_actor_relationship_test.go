// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_dac_p2p

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_P2PReplicatorWithPermissionedCollectionCreateDocActorRelationship_SourceHubACP(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, p2p replicator with collection that has a policy, create a new doc-actor relationship",

		SupportedDocumentACPTypes: immutable.Some(
			[]testUtils.DocumentACPType{
				testUtils.SourceHubDocumentACPType,
			},
		),

		Actions: []any{
			testUtils.RandomNetworkingConfig(),

			testUtils.RandomNetworkingConfig(),

			testUtils.AddDACPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
                    name: Test Policy

                    description: A Policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          read:
                            expr: owner + reader + updater + deleter

                          update:
                            expr: owner + updater

                          delete:
                            expr: owner + deleter

                          nothing:
                            expr: dummy

                        relations:
                          owner:
                            types:
                              - actor

                          reader:
                            types:
                              - actor

                          updater:
                            types:
                              - actor

                          deleter:
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
			},

			testUtils.ConfigureReplicator{
				SourceNodeID: 0,

				TargetNodeID: 1,
			},

			testUtils.CreateDoc{
				Identity: testUtils.ClientIdentity(1),

				NodeID: immutable.Some(0),

				CollectionID: 0,

				DocMap: map[string]any{
					"name": "Shahzad",
				},
			},

			testUtils.WaitForSync{},

			testUtils.Request{
				// Ensure that the document is hidden on all nodes to an unauthorized actor
				Identity: testUtils.ClientIdentity(2),

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

			testUtils.AddDACActorRelationship{
				NodeID: immutable.Some(0),

				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedExistence: false,
			},

			testUtils.AddDACActorRelationship{
				NodeID: immutable.Some(1), // Note: Different node than the previous

				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedExistence: true, // Making the same relation through any node should be a no-op
			},

			testUtils.Request{
				// Ensure that the document is now accessible on all nodes to the newly authorized actor.
				Identity: testUtils.ClientIdentity(2),

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
				Identity: testUtils.ClientIdentity(1),

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

			testUtils.DeleteDACActorRelationship{
				NodeID: immutable.Some(1),

				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedRecordFound: true,
			},

			testUtils.DeleteDACActorRelationship{
				NodeID: immutable.Some(0), // Note: Different node than the previous

				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedRecordFound: false, // Making the same relation through any node should be a no-op
			},

			testUtils.Request{
				// Ensure that the document is now inaccessible on all nodes to the actor we revoked access from.
				Identity: testUtils.ClientIdentity(2),

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
				Identity: testUtils.ClientIdentity(1),

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
