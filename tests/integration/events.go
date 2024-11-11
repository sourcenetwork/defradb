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
	"encoding/json"
	"time"

	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
)

// eventTimeout is the amount of time to wait
// for an event before timing out
const eventTimeout = 1 * time.Second

// waitForNetworkSetupEvents waits for p2p topic completed and
// replicator completed events to be published on the local node event bus.
func waitForNetworkSetupEvents(s *state, nodeID int) {
	cols, err := s.nodes[nodeID].client.GetAllP2PCollections(s.ctx)
	require.NoError(s.t, err)

	reps, err := s.nodes[nodeID].client.GetAllReplicators(s.ctx)
	require.NoError(s.t, err)

	replicatorEvents := len(reps)
	p2pTopicEvent := len(cols) > 0

	for p2pTopicEvent && replicatorEvents > 0 {
		select {
		case _, ok := <-s.nodes[nodeID].event.replicator.Message():
			if !ok {
				require.Fail(s.t, "subscription closed waiting for network setup events")
			}
			replicatorEvents--

		case _, ok := <-s.nodes[nodeID].event.p2pTopic.Message():
			if !ok {
				require.Fail(s.t, "subscription closed waiting for network setup events")
			}
			p2pTopicEvent = false

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
	case _, ok := <-s.nodes[cfg.SourceNodeID].event.replicator.Message():
		if !ok {
			require.Fail(s.t, "subscription closed waiting for replicator event")
		}

	case <-time.After(eventTimeout):
		require.Fail(s.t, "timeout waiting for replicator event")
	}

	// all previous documents should be merged on the subscriber node
	for key, val := range s.nodes[cfg.SourceNodeID].p2p.actualDocHeads {
		s.nodes[cfg.TargetNodeID].p2p.expectedDocHeads[key] = val.cid
	}

	// update node connections and replicators
	s.nodes[cfg.TargetNodeID].p2p.connections[cfg.SourceNodeID] = struct{}{}
	s.nodes[cfg.SourceNodeID].p2p.connections[cfg.TargetNodeID] = struct{}{}
	s.nodes[cfg.SourceNodeID].p2p.replicators[cfg.TargetNodeID] = struct{}{}
}

// waitForReplicatorConfigureEvent waits for a node to publish a
// replicator completed event on the local event bus.
func waitForReplicatorDeleteEvent(s *state, cfg DeleteReplicator) {
	select {
	case _, ok := <-s.nodes[cfg.SourceNodeID].event.replicator.Message():
		if !ok {
			require.Fail(s.t, "subscription closed waiting for replicator event")
		}

	case <-time.After(eventTimeout):
		require.Fail(s.t, "timeout waiting for replicator event")
	}

	delete(s.nodes[cfg.TargetNodeID].p2p.connections, cfg.SourceNodeID)
	delete(s.nodes[cfg.SourceNodeID].p2p.connections, cfg.TargetNodeID)
	delete(s.nodes[cfg.SourceNodeID].p2p.replicators, cfg.TargetNodeID)
}

// waitForSubscribeToCollectionEvent waits for a node to publish a
// p2p topic completed event on the local event bus.
//
// Expected document heads will be updated for the subscriber node.
func waitForSubscribeToCollectionEvent(s *state, action SubscribeToCollection) {
	select {
	case _, ok := <-s.nodes[action.NodeID].event.p2pTopic.Message():
		if !ok {
			require.Fail(s.t, "subscription closed waiting for p2p topic event")
		}

	case <-time.After(eventTimeout):
		require.Fail(s.t, "timeout waiting for p2p topic event")
	}

	// update peer collections of target node
	for _, collectionIndex := range action.CollectionIDs {
		if collectionIndex == NonExistentCollectionID {
			continue // don't track non existent collections
		}
		s.nodes[action.NodeID].p2p.peerCollections[collectionIndex] = struct{}{}
	}
}

// waitForSubscribeToCollectionEvent waits for a node to publish a
// p2p topic completed event on the local event bus.
func waitForUnsubscribeToCollectionEvent(s *state, action UnsubscribeToCollection) {
	select {
	case _, ok := <-s.nodes[action.NodeID].event.p2pTopic.Message():
		if !ok {
			require.Fail(s.t, "subscription closed waiting for p2p topic event")
		}

	case <-time.After(eventTimeout):
		require.Fail(s.t, "timeout waiting for p2p topic event")
	}

	for _, collectionIndex := range action.CollectionIDs {
		if collectionIndex == NonExistentCollectionID {
			continue // don't track non existent collections
		}
		delete(s.nodes[action.NodeID].p2p.peerCollections, collectionIndex)
	}
}

// waitForUpdateEvents waits for all selected nodes to publish an
// update event to the local event bus.
//
// Expected document heads will be updated for any connected nodes.
func waitForUpdateEvents(
	s *state,
	nodeID immutable.Option[int],
	docIDs map[string]struct{},
) {
	for i := 0; i < len(s.nodes); i++ {
		if nodeID.HasValue() && nodeID.Value() != i {
			continue // node is not selected
		}

		node := s.nodes[i]
		if node.closed {
			continue // node is closed
		}

		expect := make(map[string]struct{}, len(docIDs))
		for k := range docIDs {
			expect[k] = struct{}{}
		}

		for len(expect) > 0 {
			var evt event.Update
			select {
			case msg, ok := <-node.event.update.Message():
				if !ok {
					require.Fail(s.t, "subscription closed waiting for update event", "Node %d", i)
				}
				evt = msg.Data.(event.Update)

			case <-time.After(eventTimeout):
				require.Fail(s.t, "timeout waiting for update event", "Node %d", i)
			}

			if evt.DocID == "" {
				// Todo: This will almost certainly need to change once P2P for collection-level commits
				// is enabled. See: https://github.com/sourcenetwork/defradb/issues/3212
				continue
			}

			// make sure the event is expected
			_, ok := expect[evt.DocID]
			require.True(s.t, ok, "unexpected document update", "Node %d", i)
			delete(expect, evt.DocID)

			// we only need to update the network state if the nodes
			// are configured for networking
			if s.isNetworkEnabled {
				updateNetworkState(s, i, evt)
			}
		}
	}
}

// waitForMergeEvents waits for all expected document heads to be merged to all nodes.
//
// Will fail the test if an event is not received within the expected time interval to prevent tests
// from running forever.
func waitForMergeEvents(s *state, action WaitForSync) {
	for nodeID := 0; nodeID < len(s.nodes); nodeID++ {
		node := s.nodes[nodeID]
		if node.closed {
			continue // node is closed
		}

		expect := node.p2p.expectedDocHeads

		// remove any docs that are already merged
		// up to the expected document head
		for key, val := range node.p2p.actualDocHeads {
			if head, ok := expect[key]; ok && head.String() == val.cid.String() {
				delete(expect, key)
			}
		}

		expectDecrypted := make(map[string]struct{}, len(action.Decrypted))
		for _, docIndex := range action.Decrypted {
			if len(s.docIDs[0]) <= docIndex {
				require.Fail(s.t, "doc index %d out of range", docIndex)
			}
			docID := s.docIDs[0][docIndex].String()
			actual, hasActual := node.p2p.actualDocHeads[docID]
			if !hasActual || !actual.decrypted {
				expectDecrypted[docID] = struct{}{}
			}
		}

		// wait for all expected doc heads to be merged
		//
		// the order of merges does not matter as we only
		// expect the latest head to eventually be merged
		//
		// unexpected merge events are ignored
		for len(expect) > 0 || len(expectDecrypted) > 0 {
			var evt event.MergeComplete
			select {
			case msg, ok := <-node.event.merge.Message():
				if !ok {
					require.Fail(s.t, "subscription closed waiting for merge complete event")
				}
				evt = msg.Data.(event.MergeComplete)

			case <-time.After(30 * eventTimeout):
				require.Fail(s.t, "timeout waiting for merge complete event")
			}

			_, ok := expectDecrypted[evt.Merge.DocID]
			if ok && evt.Decrypted {
				delete(expectDecrypted, evt.Merge.DocID)
			}

			head, ok := expect[evt.Merge.DocID]
			if ok && head.String() == evt.Merge.Cid.String() {
				delete(expect, evt.Merge.DocID)
			}
			node.p2p.actualDocHeads[evt.Merge.DocID] = docHeadState{cid: evt.Merge.Cid, decrypted: evt.Decrypted}
		}
	}
}

// updateNetworkState updates the network state by checking which
// nodes should receive the updated document in the given update event.
func updateNetworkState(s *state, nodeID int, evt event.Update) {
	// find the correct collection index for this update
	collectionID := -1
	for i, c := range s.nodes[nodeID].collections {
		if c.SchemaRoot() == evt.SchemaRoot {
			collectionID = i
		}
	}

	node := s.nodes[nodeID]

	// update the actual document head on the node that updated it
	// as the node created the document, it is already decrypted
	node.p2p.actualDocHeads[evt.DocID] = docHeadState{cid: evt.Cid, decrypted: true}

	// update the expected document heads of replicator targets
	for id := range node.p2p.replicators {
		// replicator target nodes push updates to source nodes
		s.nodes[id].p2p.expectedDocHeads[evt.DocID] = evt.Cid
	}

	// update the expected document heads of connected nodes
	for id := range node.p2p.connections {
		// connected nodes share updates of documents they have in common
		if _, ok := s.nodes[id].p2p.actualDocHeads[evt.DocID]; ok {
			s.nodes[id].p2p.expectedDocHeads[evt.DocID] = evt.Cid
		}
		// peer collection subscribers receive updates from any other subscriber node
		if _, ok := s.nodes[id].p2p.peerCollections[collectionID]; ok {
			s.nodes[id].p2p.expectedDocHeads[evt.DocID] = evt.Cid
		}
	}

	// make sure the event is published on the network before proceeding
	// this prevents nodes from missing messages that are sent before
	// subscriptions are setup
	time.Sleep(100 * time.Millisecond)
}

// getEventsForUpdateDoc returns a map of docIDs that should be
// published to the local event bus after an UpdateDoc action.
func getEventsForUpdateDoc(s *state, action UpdateDoc) map[string]struct{} {
	docID := s.docIDs[action.CollectionID][action.DocID]

	docMap := make(map[string]any)
	err := json.Unmarshal([]byte(action.Doc), &docMap)
	require.NoError(s.t, err)

	return map[string]struct{}{
		docID.String(): {},
	}
}

// getEventsForCreateDoc returns a map of docIDs that should be
// published to the local event bus after a CreateDoc action.
func getEventsForCreateDoc(s *state, action CreateDoc) map[string]struct{} {
	var collection client.Collection
	if action.NodeID.HasValue() {
		collection = s.nodes[action.NodeID.Value()].collections[action.CollectionID]
	} else {
		collection = s.nodes[0].collections[action.CollectionID]
	}

	docs, err := parseCreateDocs(action, collection)
	require.NoError(s.t, err)

	expect := make(map[string]struct{})

	for _, doc := range docs {
		expect[doc.ID().String()] = struct{}{}
	}

	return expect
}

func waitForSync(s *state, action WaitForSync) {
	waitForMergeEvents(s, action)
}

// getEventsForUpdateWithFilter returns a map of docIDs that should be
// published to the local event bus after a UpdateWithFilter action.
func getEventsForUpdateWithFilter(
	s *state,
	action UpdateWithFilter,
	result *client.UpdateResult,
) map[string]struct{} {
	var docPatch map[string]any
	err := json.Unmarshal([]byte(action.Updater), &docPatch)
	require.NoError(s.t, err)

	expect := make(map[string]struct{})

	for _, docID := range result.DocIDs {
		expect[docID] = struct{}{}
	}

	return expect
}
