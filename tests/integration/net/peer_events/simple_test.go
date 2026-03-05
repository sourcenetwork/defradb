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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

const docSyncTopic = "doc-sync"

func TestPeerEvents_OnConnect_ShouldReceiveJoinEventOnDocSyncTopic(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.WaitForPeersEvents{
				NodeID: 0,
				ExpectedPeersByTopic: map[string][]int{
					docSyncTopic: {1},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestPeerEvents_OnConnectMultiplePeers_ShouldReceiveAllJoinEvents(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
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
				ExpectedPeersByTopic: map[string][]int{
					docSyncTopic: {1, 2},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestPeerEvents_OnConnectBidirectional_BothNodesShouldReceiveJoinEvents(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.WaitForPeersEvents{
				NodeID: 0,
				ExpectedPeersByTopic: map[string][]int{
					docSyncTopic: {1},
				},
			},
			&action.WaitForPeersEvents{
				NodeID: 1,
				ExpectedPeersByTopic: map[string][]int{
					docSyncTopic: {0},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
