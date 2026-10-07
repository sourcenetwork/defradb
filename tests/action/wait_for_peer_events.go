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

package action

import (
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/tests/state"
)

// WaitForPeersEvents waits for peer events on pubsub topics.
type WaitForPeersEvents struct {
	stateful

	// NodeID is the node that should receive the peer events.
	NodeID int

	// EventType is the type of event to wait for.
	// Defaults to client.PeerEventTypeJoined if not specified.
	EventType string

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

	eventType := a.EventType
	if eventType == "" {
		eventType = client.PeerEventTypeJoined
	}

	sourceNode := a.s.Nodes[a.NodeID]

	expectedPeers := make(map[string]map[string]bool)

	addExpectedPeers := func(topic string, peerNodeIDs []int) {
		if _, exists := expectedPeers[topic]; !exists {
			expectedPeers[topic] = make(map[string]bool)
		}
		for _, peerNodeID := range peerNodeIDs {
			targetNode := a.s.Nodes[peerNodeID]
			targetAddresses, err := targetNode.PeerInfo(a.s.Ctx)
			require.NoError(a.s.T, err)
			require.NotEmpty(a.s.T, targetAddresses, "target node %d has no addresses", peerNodeID)

			peerID, err := extractPeerID(targetAddresses[0])
			require.NoError(a.s.T, err, "could not extract peer ID from address for node %d", peerNodeID)
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
		a.s.DocIDsLock.RLock()
		docID := a.s.DocIDs[colDocIndex.Col][colDocIndex.Doc]
		a.s.DocIDsLock.RUnlock()

		topic := docID.String()
		addExpectedPeers(topic, peerNodeIDs)
	}

	if sourceNode.IsExternal {
		// Leaving a topic does not end the connection, so there is nothing
		// outside the node that shows it happened.
		if eventType == client.PeerEventTypeJoined {
			a.waitOnExternalNode(sourceNode, expectedPeers, timeout)
		}
		return
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
			if peerEvent.EventType != eventType {
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
			require.Fail(a.s.T, "timeout waiting for peer events",
				"source node %d did not receive %s events for: %v",
				a.NodeID, eventType, remaining)
			return
		}
	}
}

// waitOnExternalNode waits until a node in another process is connected to
// every expected peer, which it has to be before it can see them join a topic.
// Its events cannot be read from here, so it is asked over its API instead.
// Without this, the actions that follow run before the node has found its
// peers, and a sync it starts misses the ones it has not found yet.
func (a *WaitForPeersEvents) waitOnExternalNode(
	node *state.NodeState,
	expectedPeers map[string]map[string]bool,
	timeout time.Duration,
) {
	want := make(map[string]struct{})
	for _, peers := range expectedPeers {
		for id := range peers {
			want[id] = struct{}{}
		}
	}

	opts := options.ActivePeers()
	ident := getIdentityForRequestSpecificToNode(a.s, NodeIdentity(a.NodeID), a.NodeID)
	if ident.HasValue() {
		opts.SetIdentity(ident.Value())
	}

	deadline := time.Now().Add(timeout)
	for {
		addrs, err := node.ActivePeers(a.s.Ctx, opts)
		require.NoError(a.s.T, err)
		connected := make(map[string]struct{}, len(addrs))
		for _, addr := range addrs {
			if id, err := extractPeerID(addr); err == nil {
				connected[id] = struct{}{}
			}
		}

		var waitingOn []string
		for id := range want {
			if _, ok := connected[id]; !ok {
				waitingOn = append(waitingOn, id)
			}
		}
		if len(waitingOn) == 0 {
			return
		}
		if time.Now().After(deadline) {
			require.Fail(a.s.T, "timeout waiting for external node peers",
				"node %d did not connect to: %v", a.NodeID, waitingOn)
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// extractPeerID extracts the peer ID from a multiaddr string.
func extractPeerID(addr string) (string, error) {
	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return "", err
	}
	id, err := peer.IDFromP2PAddr(maddr)
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
