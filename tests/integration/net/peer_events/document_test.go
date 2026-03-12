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

package peer_events

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestPeerEvents_OnSubscribeToDocument_ShouldReceiveJoinEventOnDocumentTopic(t *testing.T) {
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
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.AddDocumentSubscription{
				NodeID: 0,
				DocIDs: []state.ColDocIndex{{Col: 0, Doc: 0}},
			},
			testUtils.AddDocumentSubscription{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{{Col: 0, Doc: 0}},
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.WaitForPeersEvents{
				NodeID: 0,
				ExpectedPeersByDocument: map[state.ColDocIndex][]int{
					{Col: 0, Doc: 0}: {1},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestPeerEvents_OnSubscribeToMultipleDocuments_ShouldReceiveJoinEventsOnAllTopics(t *testing.T) {
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
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Alice",
				},
			},
			testUtils.AddDocumentSubscription{
				NodeID: 0,
				DocIDs: []state.ColDocIndex{
					{Col: 0, Doc: 0},
					{Col: 0, Doc: 1},
				},
			},
			testUtils.AddDocumentSubscription{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{
					{Col: 0, Doc: 0},
					{Col: 0, Doc: 1},
				},
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.WaitForPeersEvents{
				NodeID: 0,
				ExpectedPeersByDocument: map[state.ColDocIndex][]int{
					{Col: 0, Doc: 0}: {1},
					{Col: 0, Doc: 1}: {1},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestPeerEvents_DocumentAndDocSyncTopics_ShouldReceiveJoinEventsOnBoth(t *testing.T) {
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
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.AddDocumentSubscription{
				NodeID: 0,
				DocIDs: []state.ColDocIndex{{Col: 0, Doc: 0}},
			},
			testUtils.AddDocumentSubscription{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{{Col: 0, Doc: 0}},
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.WaitForPeersEvents{
				NodeID: 0,
				ExpectedPeersByTopic: map[string][]int{
					docSyncTopic: {1},
				},
				ExpectedPeersByDocument: map[state.ColDocIndex][]int{
					{Col: 0, Doc: 0}: {1},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestPeerEvents_AllTopicTypes_ShouldReceiveJoinEventsOnAll(t *testing.T) {
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
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.AddCollectionSubscription{
				NodeID:        0,
				CollectionIDs: []int{0},
			},
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.AddDocumentSubscription{
				NodeID: 0,
				DocIDs: []state.ColDocIndex{{Col: 0, Doc: 0}},
			},
			testUtils.AddDocumentSubscription{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{{Col: 0, Doc: 0}},
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.WaitForPeersEvents{
				NodeID: 0,
				ExpectedPeersByTopic: map[string][]int{
					docSyncTopic: {1},
				},
				ExpectedPeersByCollection: map[int][]int{
					0: {1},
				},
				ExpectedPeersByDocument: map[state.ColDocIndex][]int{
					{Col: 0, Doc: 0}: {1},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestPeerEvents_OnUnsubscribeFromDocument_ShouldReceiveLeftEvent(t *testing.T) {
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
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.AddDocumentSubscription{
				NodeID: 0,
				DocIDs: []state.ColDocIndex{{Col: 0, Doc: 0}},
			},
			testUtils.AddDocumentSubscription{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{{Col: 0, Doc: 0}},
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.WaitForPeersEvents{
				NodeID: 0,
				ExpectedPeersByDocument: map[state.ColDocIndex][]int{
					{Col: 0, Doc: 0}: {1},
				},
			},
			testUtils.DeleteDocumentSubscription{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{{Col: 0, Doc: 0}},
			},
			&action.WaitForPeersEvents{
				NodeID:    0,
				EventType: client.PeerEventTypeLeft,
				ExpectedPeersByDocument: map[state.ColDocIndex][]int{
					{Col: 0, Doc: 0}: {1},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestPeerEvents_OnUnsubscribeFromMultipleDocuments_ShouldReceiveLeftEvents(t *testing.T) {
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
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Alice",
				},
			},
			testUtils.AddDocumentSubscription{
				NodeID: 0,
				DocIDs: []state.ColDocIndex{
					{Col: 0, Doc: 0},
					{Col: 0, Doc: 1},
				},
			},
			testUtils.AddDocumentSubscription{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{
					{Col: 0, Doc: 0},
					{Col: 0, Doc: 1},
				},
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.WaitForPeersEvents{
				NodeID: 0,
				ExpectedPeersByDocument: map[state.ColDocIndex][]int{
					{Col: 0, Doc: 0}: {1},
					{Col: 0, Doc: 1}: {1},
				},
			},
			testUtils.DeleteDocumentSubscription{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{
					{Col: 0, Doc: 0},
					{Col: 0, Doc: 1},
				},
			},
			&action.WaitForPeersEvents{
				NodeID:    0,
				EventType: client.PeerEventTypeLeft,
				ExpectedPeersByDocument: map[state.ColDocIndex][]int{
					{Col: 0, Doc: 0}: {1},
					{Col: 0, Doc: 1}: {1},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
