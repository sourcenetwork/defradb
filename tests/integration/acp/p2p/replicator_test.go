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

func TestACP_P2POneToOneReplicatorWithPermissionedCollection_LocalACP(t *testing.T) {
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
				ExpectedPolicyID: "abe378ae8dac56f43238b56126a5a5ff1d1021e6bf8027d477b5a366e6238fc2",
			},
			testUtils.SchemaUpdate{
				Schema: `
					type Users @policy(
						id: "abe378ae8dac56f43238b56126a5a5ff1d1021e6bf8027d477b5a366e6238fc2",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},
			testUtils.ConfigureReplicator{
				SourceNodeID:  0,
				TargetNodeID:  1,
				ExpectedError: "replicator collection specified has a policy on it",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_P2POneToOneReplicatorWithPermissionedCollection_SourceHubACP(t *testing.T) {
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
				ExpectedPolicyID: "abe378ae8dac56f43238b56126a5a5ff1d1021e6bf8027d477b5a366e6238fc2",
			},
			testUtils.SchemaUpdate{
				Schema: `
					type Users @policy(
						id: "abe378ae8dac56f43238b56126a5a5ff1d1021e6bf8027d477b5a366e6238fc2",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.CreateDoc{
				NodeID:   immutable.Some(0),
				Identity: testUtils.ClientIdentity(1),
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				// Ensure that the document is accessible on all nodes to authorized actors
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
