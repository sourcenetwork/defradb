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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/tests/state"
)

// WaitForPeersEvents waits for peer join events on pubsub topics.
type WaitForPeersEvents struct {
	stateful

	// NodeID is the node that should receive the peer join events.
	NodeID int

	// ExpectedPeersByTopic maps named topics (like "doc-sync") to expected peer node IDs.
	ExpectedPeersByTopic map[string][]int

	// ExpectedPeersByCollection maps collection indexes to expected peer node IDs.
	ExpectedPeersByCollection map[int][]int

	// ExpectedPeersByDocument maps document indexes to expected peer node IDs.
	ExpectedPeersByDocument map[state.ColDocIndex][]int

	// Timeout is the maximum time to wait for the peer connection.
	// Defaults to 5 seconds if not specified.
	Timeout time.Duration
}

var _ Action = (*WaitForPeersEvents)(nil)
var _ Stateful = (*WaitForPeersEvents)(nil)

func (a *WaitForPeersEvents) Execute() {
	timeout := a.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	sourceNode := a.s.Nodes[a.NodeID]
	expectedPeers := make(map[string]map[string]bool)

	addExpectedPeers := func(topic string, peerNodeIDs []int) {
		if _, exists := expectedPeers[topic]; !exists {
			expectedPeers[topic] = make(map[string]bool)
		}
		for _, peerNodeID := range peerNodeIDs {
			targetNode := a.s.Nodes[peerNodeID]
			targetAddresses, err := targetNode.PeerInfo()
			require.NoError(a.s.T, err)
			require.NotEmpty(a.s.T, targetAddresses, "target node %d has no addresses", peerNodeID)

			peerID := extractPeerID(targetAddresses[0])
			require.NotEmpty(a.s.T, peerID, "could not extract peer ID from address for node %d", peerNodeID)
			expectedPeers[topic][peerID] = true
		}
	}

	for topic, peerNodeIDs := range a.ExpectedPeersByTopic {
		addExpectedPeers(topic, peerNodeIDs)
	}

	for colIndex, peerNodeIDs := range a.ExpectedPeersByCollection {
		col := a.s.Nodes[a.NodeID].Collections[colIndex]
		topic := col.CollectionID()
		addExpectedPeers(topic, peerNodeIDs)
	}

	for colDocIndex, peerNodeIDs := range a.ExpectedPeersByDocument {
		docID := a.s.DocIDs[colDocIndex.Col][colDocIndex.Doc]
		topic := docID.String()
		addExpectedPeers(topic, peerNodeIDs)
	}

	totalExpected := 0
	for _, peers := range expectedPeers {
		totalExpected += len(peers)
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for totalExpected > 0 {
		select {
		case msg := <-sourceNode.Event.TopicPeerEvent.Message():
			peerEvent, ok := msg.Data.(event.TopicPeerEvent)
			if !ok {
				continue
			}
			if peerEvent.EventType != client.PeerEventTypeJoined {
				continue
			}
			if topicPeers, topicExists := expectedPeers[peerEvent.Topic]; topicExists {
				if topicPeers[peerEvent.PeerID] {
					delete(topicPeers, peerEvent.PeerID)
					totalExpected--
				}
			}
		case <-timer.C:
			var remaining []string
			for topic, peers := range expectedPeers {
				for peerID := range peers {
					remaining = append(remaining, topic+":"+peerID)
				}
			}
			require.Fail(a.s.T, "timeout waiting for peer connections",
				"source node %d did not receive join events for: %v",
				a.NodeID, remaining)
			return
		}
	}
}

// extractPeerID extracts the peer ID from a multiaddr string.
// Example: /ip4/127.0.0.1/tcp/4001/p2p/12D3KooWExample -> 12D3KooWExample
func extractPeerID(addr string) string {
	const p2pPrefix = "/p2p/"
	for i := 0; i < len(addr); i++ {
		if i+len(p2pPrefix) <= len(addr) && addr[i:i+len(p2pPrefix)] == p2pPrefix {
			return addr[i+len(p2pPrefix):]
		}
	}
	return ""
}
