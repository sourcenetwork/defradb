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

// This test documents that we don't allow subscribing to a collection that has a policy
// until the following is implemented:
// TODO-ACP: ACP <> P2P https://github.com/sourcenetwork/defradb/issues/2366
func TestACP_P2PSubscribeAddGetSingleWithPermissionedCollection_Error(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, with p2p subscribe with permissioned collection, error",

		Actions: []any{

			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),

			testUtils.AddPolicy{

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

				ExpectedPolicyID: "94eb195c0e459aa79e02a1986c7e731c5015721c18a373f2b2a0ed140a04b454",
			},

			testUtils.SchemaUpdate{
				Schema: `
					type Users @policy(
						id: "94eb195c0e459aa79e02a1986c7e731c5015721c18a373f2b2a0ed140a04b454",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
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
