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
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestPeerEvents_OnSubscribeToDocument_ShouldReceiveJoinEventOnDocumentTopic(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddSchema{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.SubscribeToDocument{
				NodeID: 0,
				DocIDs: []state.ColDocIndex{{Col: 0, Doc: 0}},
			},
			testUtils.SubscribeToDocument{
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
			&action.AddSchema{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Alice",
				},
			},
			testUtils.SubscribeToDocument{
				NodeID: 0,
				DocIDs: []state.ColDocIndex{
					{Col: 0, Doc: 0},
					{Col: 0, Doc: 1},
				},
			},
			testUtils.SubscribeToDocument{
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
			&action.AddSchema{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.SubscribeToDocument{
				NodeID: 0,
				DocIDs: []state.ColDocIndex{{Col: 0, Doc: 0}},
			},
			testUtils.SubscribeToDocument{
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
			&action.AddSchema{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.SubscribeToCollection{
				NodeID:        0,
				CollectionIDs: []int{0},
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.SubscribeToDocument{
				NodeID: 0,
				DocIDs: []state.ColDocIndex{{Col: 0, Doc: 0}},
			},
			testUtils.SubscribeToDocument{
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
			&action.AddSchema{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.SubscribeToDocument{
				NodeID: 0,
				DocIDs: []state.ColDocIndex{{Col: 0, Doc: 0}},
			},
			testUtils.SubscribeToDocument{
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
			testUtils.UnsubscribeToDocument{
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
			&action.AddSchema{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Alice",
				},
			},
			testUtils.SubscribeToDocument{
				NodeID: 0,
				DocIDs: []state.ColDocIndex{
					{Col: 0, Doc: 0},
					{Col: 0, Doc: 1},
				},
			},
			testUtils.SubscribeToDocument{
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
			testUtils.UnsubscribeToDocument{
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
