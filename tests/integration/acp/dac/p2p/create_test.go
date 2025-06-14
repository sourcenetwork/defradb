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

func TestACP_P2PCreatePrivateDocumentsOnDifferentNodes_SourceHubACP(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, p2p create private documents on different nodes, with source-hub",

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

			testUtils.CreateDoc{
				Identity: testUtils.ClientIdentity(1),

				NodeID: immutable.Some(0),

				CollectionID: 0,

				DocMap: map[string]any{
					"name": "Shahzad",
				},
			},

			testUtils.CreateDoc{
				Identity: testUtils.ClientIdentity(1),

				NodeID: immutable.Some(1),

				CollectionID: 0,

				DocMap: map[string]any{
					"name": "Shahzad Lone",
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_P2PCreatePrivateDocumentAndSyncAfterAddingRelationship_SourceHubACP(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, p2p create a private documents and sync after adding actor relationship, with source-hub",
		SupportedDocumentACPTypes: immutable.Some(
			[]testUtils.DocumentACPType{
				testUtils.SourceHubDocumentACPType,
			},
		),
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),

			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},

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

			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},

			testUtils.CreateDoc{
				Identity:     testUtils.ClientIdentity(1),
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "Shahzad",
				},
			},

			// At this point the document is only accessible to the owner so node 1
			// should not have been able to sync the document.
			testUtils.Request{
				Identity: testUtils.ClientIdentity(1),
				NodeID:   immutable.Some(1),
				Request: `query {
					Users{
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},

			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.NodeIdentity(1),
				CollectionID:      0,
				DocID:             0,
				Relation:          "reader",
				ExpectedExistence: false,
			},

			testUtils.WaitForSync{},

			testUtils.Request{
				Identity: testUtils.ClientIdentity(1),
				NodeID:   immutable.Some(1),
				Request: `query {
					Users {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "Shahzad"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
