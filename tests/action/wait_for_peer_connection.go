// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package action

import (
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/event"
)

// WaitForPeersConnection waits for peer join events on a specific topic.
// This is used to ensure peers have connected and are ready to communicate
// on a pubsub topic before proceeding with tests.
type WaitForPeersConnection struct {
	stateful

	// NodeID is the node that should receive the peer join event.
	NodeID int

	// PeerNodeIDs are the nodes that should join the topic.
	PeerNodeIDs []int

	// Topic is the pubsub topic on which the peer join events are expected.
	Topic string

	// Timeout is the maximum time to wait for the peer connection.
	// Defaults to 5 seconds if not specified.
	Timeout time.Duration
}

var _ Action = (*WaitForPeersConnection)(nil)
var _ Stateful = (*WaitForPeersConnection)(nil)

func (a *WaitForPeersConnection) Execute() {
	timeout := a.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	sourceNode := a.s.Nodes[a.NodeID]

	expectedPeerIDs := make(map[string]bool)
	for _, peerNodeID := range a.PeerNodeIDs {
		targetNode := a.s.Nodes[peerNodeID]
		targetAddresses, err := targetNode.PeerInfo()
		require.NoError(a.s.T, err)
		require.NotEmpty(a.s.T, targetAddresses, "target node %d has no addresses", peerNodeID)

		peerID := extractPeerID(targetAddresses[0])
		require.NotEmpty(a.s.T, peerID, "could not extract peer ID from address for node %d", peerNodeID)
		expectedPeerIDs[peerID] = true
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for len(expectedPeerIDs) > 0 {
		select {
		case msg := <-sourceNode.Event.TopicPeerEvent.Message():
			peerEvent, ok := msg.Data.(event.TopicPeerEvent)
			if !ok {
				continue
			}
			if peerEvent.Joined && peerEvent.Topic == a.Topic {
				delete(expectedPeerIDs, peerEvent.PeerID)
			}
		case <-timer.C:
			var remaining []string
			for peerID := range expectedPeerIDs {
				remaining = append(remaining, peerID)
			}
			require.Fail(a.s.T, "timeout waiting for peer connections",
				"source node %d did not receive join events from peers: %v",
				a.NodeID, remaining)
			return
		}
	}
}

// extractPeerID extracts the peer ID from a multiaddr string.
// Example: /ip4/127.0.0.1/tcp/4001/p2p/12D3KooWExample -> 12D3KooWExample
func extractPeerID(addr string) string {
	// Find the /p2p/ component and extract what follows
	const p2pPrefix = "/p2p/"
	for i := 0; i < len(addr); i++ {
		if i+len(p2pPrefix) <= len(addr) && addr[i:i+len(p2pPrefix)] == p2pPrefix {
			return addr[i+len(p2pPrefix):]
		}
	}
	return ""
}
