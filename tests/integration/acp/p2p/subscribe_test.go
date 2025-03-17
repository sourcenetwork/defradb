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
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_P2PSubscribeAddGetSingleWithPermissionedCollection_LocalACP(t *testing.T) {
	test := testUtils.TestCase{
		SupportedACPTypes: immutable.Some(
			[]testUtils.ACPType{
				testUtils.LocalACPType,
			},
		),
		Actions: []any{

			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),

			testUtils.AddPolicy{

				Identity: testUtils.ClientIdentity(1),

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
                          update:
                            expr: owner
                          delete:
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

			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},

			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
				ExpectedError: "p2p collection specified has a policy on it",
			},

			testUtils.GetAllP2PCollections{
				NodeID:                1,
				ExpectedCollectionIDs: []int{}, // Note: Empty
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_P2PSubscribeAddGetSingleWithPermissionedCollection_SourceHubACP(t *testing.T) {
	test := testUtils.TestCase{
		SupportedACPTypes: immutable.Some(
			[]testUtils.ACPType{
				testUtils.SourceHubACPType,
			},
		),
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.AddPolicy{
				Identity: testUtils.ClientIdentity(1),
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
                          update:
                            expr: owner
                          delete:
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
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.CreateDoc{
				NodeID:   immutable.Some(0),
				Identity: testUtils.ClientIdentity(1),
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.Request{
				// The document will only be accessible on node 0 since node 1 is not authorized to
				// access the document.
				NodeID:   immutable.Some(0),
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
							"name": "John",
						},
					},
				},
			},
			testUtils.Request{
				// Since node 1 is not authorized to access the document, it won't have to document
				// so even if requesting with an authorized identity, the document won't be returned.
				NodeID:   immutable.Some(1),
				Identity: testUtils.ClientIdentity(1),
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
				// Ensure that the document is hidden on all nodes to unidentified actors
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
				// Ensure that the document is hidden on all nodes to unauthorized actors
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
