// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package peer_events

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestPeerEvents_OnSubscribeToCollection_ShouldReceiveJoinEventOnCollectionTopic(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			testUtils.AddCollectionSubscription{
				NodeID:        0,
				CollectionIDs: []int{0},
			},
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.WaitForPeersEvents{
				NodeID: 0,
				ExpectedPeersByCollection: map[int][]int{
					0: {1},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestPeerEvents_OnSubscribeToMultipleCollections_ShouldReceiveJoinEventsOnAllTopics(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
					type Product {
						title: String
					}
				`,
			},
			testUtils.AddCollectionSubscription{
				NodeID:        0,
				CollectionIDs: []int{0, 1},
			},
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0, 1},
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.WaitForPeersEvents{
				NodeID: 0,
				ExpectedPeersByCollection: map[int][]int{
					0: {1},
					1: {1},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestPeerEvents_MultipleNodesSubscribedToCollection_ShouldReceiveAllJoinEvents(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			testUtils.AddCollectionSubscription{
				NodeID:        0,
				CollectionIDs: []int{0},
			},
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.AddCollectionSubscription{
				NodeID:        2,
				CollectionIDs: []int{0},
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 2,
			},
			&action.WaitForPeersEvents{
				NodeID: 0,
				ExpectedPeersByCollection: map[int][]int{
					0: {1, 2},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestPeerEvents_OnUnsubscribeFromCollection_ShouldReceiveLeftEvent(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			testUtils.AddCollectionSubscription{
				NodeID:        0,
				CollectionIDs: []int{0},
			},
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.WaitForPeersEvents{
				NodeID: 0,
				ExpectedPeersByCollection: map[int][]int{
					0: {1},
				},
			},
			testUtils.DeleteCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			&action.WaitForPeersEvents{
				NodeID:    0,
				EventType: client.PeerEventTypeLeft,
				ExpectedPeersByCollection: map[int][]int{
					0: {1},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestPeerEvents_OnUnsubscribeFromMultipleCollections_ShouldReceiveLeftEvents(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
					type Product {
						title: String
					}
				`,
			},
			testUtils.AddCollectionSubscription{
				NodeID:        0,
				CollectionIDs: []int{0, 1},
			},
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0, 1},
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.WaitForPeersEvents{
				NodeID: 0,
				ExpectedPeersByCollection: map[int][]int{
					0: {1},
					1: {1},
				},
			},
			testUtils.DeleteCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0, 1},
			},
			&action.WaitForPeersEvents{
				NodeID:    0,
				EventType: client.PeerEventTypeLeft,
				ExpectedPeersByCollection: map[int][]int{
					0: {1},
					1: {1},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
