// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package subscribe_test

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestP2PCollectionAddRemoveGetSingle(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Subscribing then unsubscribing from a collection leaves the subscription list empty.",
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.DeleteCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.ListP2PCollections{
				NodeID:                1,
				ExpectedCollectionIDs: []int{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PCollectionAddRemoveGetMultiple(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Unsubscribing from one of two collections leaves only the other collection in the list.",
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
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
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0, 1},
			},
			testUtils.DeleteCollectionSubscription{
				NodeID: 1,
				// Unsubscribe from Users, but remain subscribed to Giraffes
				CollectionIDs: []int{0},
			},
			testUtils.ListP2PCollections{
				NodeID:                1,
				ExpectedCollectionIDs: []int{1},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
