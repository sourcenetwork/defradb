// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package subscribe_test

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestP2PCollectionAddRemoveGetSingle(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
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
			},
			testUtils.UnsubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.GetAllP2PCollections{
				NodeID:                1,
				ExpectedCollectionIDs: []int{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PCollectionAddRemoveGetMultiple(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
					type Giraffes {
						name: String
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0, 1},
			},
			testUtils.UnsubscribeToCollection{
				NodeID: 1,
				// Unsubscribe from Users, but remain subscribed to Giraffes
				CollectionIDs: []int{0},
			},
			testUtils.GetAllP2PCollections{
				NodeID:                1,
				ExpectedCollectionIDs: []int{1},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
