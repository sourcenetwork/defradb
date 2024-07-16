// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tests

import (
	"time"

	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/event"
)

// eventTimeout is the default amount of time
// to wait for an event before timing out
const eventTimeout = 5 * time.Second

// waitForNetworkSetupEvents waits for p2p topic completed and
// replicator completed events to be published on the local node event bus.
func waitForNetworkSetupEvents(s *state, nodeID int) {
	cols, err := s.nodes[nodeID].GetAllP2PCollections(s.ctx)
	require.NoError(s.t, err)

	reps, err := s.nodes[nodeID].GetAllReplicators(s.ctx)
	require.NoError(s.t, err)

	p2pTopicEvents := 0
	replicatorEvents := len(reps)

	// there is only one message for loading of P2P collections
	if len(cols) > 0 {
		p2pTopicEvents = 1
	}

	for p2pTopicEvents > 0 && replicatorEvents > 0 {
		select {
		case <-s.nodeEvents[nodeID].replicator.Message():
			replicatorEvents--

		case <-s.nodeEvents[nodeID].p2pTopic.Message():
			p2pTopicEvents--

		case <-time.After(eventTimeout):
			s.t.Fatalf("timeout waiting for network setup events")
		}
	}
}

// waitForReplicatorConfigureEvent waits for a  node to publish a
// replicator completed event on the local event bus.
//
// Expected document heads will be updated for the targeted node.
func waitForReplicatorConfigureEvent(s *state, cfg ConfigureReplicator) {
	select {
	case <-s.nodeEvents[cfg.SourceNodeID].replicator.Message():
		// event recieved

	case <-time.After(eventTimeout):
		require.Fail(s.t, "timeout waiting for replicator event")
	}

	// all previous documents should be merged on the subscriber node
	for key, val := range s.nodeP2P[cfg.SourceNodeID].actualDocHeads {
		s.nodeP2P[cfg.TargetNodeID].expectedDocHeads[key] = val
	}
	s.nodeP2P[cfg.TargetNodeID].replicatorTargets[cfg.SourceNodeID] = struct{}{}
	s.nodeP2P[cfg.SourceNodeID].replicatorSources[cfg.TargetNodeID] = struct{}{}
}

// waitForReplicatorConfigureEvent waits for a node to publish a
// replicator completed event on the local event bus.
func waitForReplicatorDeleteEvent(s *state, cfg DeleteReplicator) {
	select {
	case <-s.nodeEvents[cfg.SourceNodeID].replicator.Message():
		// event recieved

	case <-time.After(eventTimeout):
		require.Fail(s.t, "timeout waiting for replicator event")
	}

	delete(s.nodeP2P[cfg.TargetNodeID].replicatorTargets, cfg.SourceNodeID)
	delete(s.nodeP2P[cfg.SourceNodeID].replicatorSources, cfg.TargetNodeID)
}

// waitForSubscribeToCollectionEvent waits for a node to publish a
// p2p topic completed event on the local event bus.
//
// Expected document heads will be updated for the subscriber node.
func waitForSubscribeToCollectionEvent(s *state, action SubscribeToCollection) {
	select {
	case <-s.nodeEvents[action.NodeID].p2pTopic.Message():
		// event recieved

	case <-time.After(eventTimeout):
		require.Fail(s.t, "timeout waiting for p2p topic event")
	}

	// update peer collections and expected documents of subscribed node
	for _, collectionIndex := range action.CollectionIDs {
		if collectionIndex == NonExistentCollectionID {
			continue // don't track non existent collections
		}
		s.nodeP2P[action.NodeID].peerCollections[collectionIndex] = struct{}{}
		if collectionIndex >= len(s.documents) {
			continue // no documents to track
		}
		// all previous documents should be merged on the subscriber node
		for _, doc := range s.documents[collectionIndex] {
			s.nodeP2P[action.NodeID].expectedDocHeads[doc.ID().String()] = doc.Head()
		}
	}
}

// waitForSubscribeToCollectionEvent waits for a node to publish a
// p2p topic completed event on the local event bus.
func waitForUnsubscribeToCollectionEvent(s *state, action UnsubscribeToCollection) {
	select {
	case <-s.nodeEvents[action.NodeID].p2pTopic.Message():
		// event recieved

	case <-time.After(eventTimeout):
		require.Fail(s.t, "timeout waiting for p2p topic event")
	}

	for _, collectionIndex := range action.CollectionIDs {
		if collectionIndex == NonExistentCollectionID {
			continue // don't track non existent collections
		}
		delete(s.nodeP2P[action.NodeID].peerCollections, collectionIndex)
	}
}

// waitForUpdateEvents waits for all selected nodes to publish an
// update event to the local event bus.
//
// Expected document heads will be updated for any connected nodes.
func waitForUpdateEvents(s *state, nodeID immutable.Option[int], collectionID int) {
	for i := 0; i < len(s.nodes); i++ {
		if nodeID.HasValue() && nodeID.Value() != i {
			continue // node is not selected
		}

		var evt event.Update
		select {
		case msg := <-s.nodeEvents[i].update.Message():
			evt = msg.Data.(event.Update)

		case <-time.After(eventTimeout):
			require.Fail(s.t, "timeout waiting for update event")
		}

		// update the actual document head on the node that updated it
		s.nodeP2P[i].actualDocHeads[evt.DocID] = evt.Cid

		// update the expected document heads of connected nodes
		//
		// connected nodes share updates of documents they have in common
		for id := range s.nodeP2P[i].connections {
			if _, ok := s.nodeP2P[id].actualDocHeads[evt.DocID]; ok {
				s.nodeP2P[id].expectedDocHeads[evt.DocID] = evt.Cid
			}
		}

		// update the expected document heads of replicator sources
		//
		// replicator source nodes receive updates from target nodes
		for id := range s.nodeP2P[i].replicatorTargets {
			if _, ok := s.nodeP2P[id].actualDocHeads[evt.DocID]; ok {
				s.nodeP2P[id].expectedDocHeads[evt.DocID] = evt.Cid
			}
		}

		// update the expected document heads of replicator targets
		//
		// replicator target nodes push updates to source nodes
		for id := range s.nodeP2P[i].replicatorSources {
			s.nodeP2P[id].expectedDocHeads[evt.DocID] = evt.Cid
		}

		// update the expected document heads of peer collection subs
		//
		// peer collection subscribers receive updates from any other subscriber node
		for id := range s.nodes {
			if _, ok := s.nodeP2P[id].peerCollections[collectionID]; ok {
				s.nodeP2P[id].expectedDocHeads[evt.DocID] = evt.Cid
			}
		}
	}
}

// waitForMergeEvents waits for all expected document heads to be merged to all nodes.
//
// Will fail the test if an event is not received within the expected time interval to prevent tests
// from running forever.
func waitForMergeEvents(s *state, action WaitForSync) {
	var timeout time.Duration
	if action.ExpectedTimeout != 0 {
		timeout = action.ExpectedTimeout
	} else {
		timeout = eventTimeout
	}

	for nodeID := 0; nodeID < len(s.nodes); nodeID++ {
		expect := s.nodeP2P[nodeID].expectedDocHeads

		// remove any docs that are already merged
		// up to the expected document head
		for key, val := range s.nodeP2P[nodeID].actualDocHeads {
			if head, ok := expect[key]; ok && head.String() == val.String() {
				delete(expect, key)
			}
		}
		// wait for all expected doc heads to be merged
		//
		// the order of merges does not matter as we only
		// expect the latest head to eventually be merged
		//
		// unexpected merge events are ignored
		for len(expect) > 0 {
			var evt event.Merge
			select {
			case msg := <-s.nodeEvents[nodeID].merge.Message():
				evt = msg.Data.(event.Merge)

			case <-time.After(timeout):
				require.Fail(s.t, "timeout waiting for merge complete event")
			}

			head, ok := expect[evt.DocID]
			if ok && head.String() == evt.Cid.String() {
				delete(expect, evt.DocID)
			}
		}
	}
}
