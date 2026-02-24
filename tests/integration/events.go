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
	"fmt"
	"strconv"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/tests/state"
)

// eventTimeout is the amount of time to wait
// for an event before timing out
const eventTimeout = 1 * time.Second

// waitForReplicatorConfigureEvent waits for a  node to publish a
// replicator completed event on the local event bus.
//
// Expected document heads will be updated for the targeted node.
func waitForReplicatorConfigureEvent(s *state.State, cfg AddReplicator) {
	select {
	case _, ok := <-s.Nodes[cfg.SourceNodeID].Event.Replicator.Message():
		if !ok {
			require.Fail(s.T, "subscription closed waiting for replicator event")
		}

	case <-time.After(eventTimeout):
		require.Fail(s.T, "timeout waiting for replicator event")
	}

	// all previous documents should be merged on the subscriber node
	for key, val := range s.Nodes[cfg.SourceNodeID].P2P.ActualDAGHeads {
		s.Nodes[cfg.TargetNodeID].P2P.ExpectedDAGHeads[key] = val.CID
	}

	// update node connections and replicators
	s.Nodes[cfg.TargetNodeID].P2P.Connections[cfg.SourceNodeID] = struct{}{}
	s.Nodes[cfg.SourceNodeID].P2P.Connections[cfg.TargetNodeID] = struct{}{}
	s.Nodes[cfg.SourceNodeID].P2P.Replicators[cfg.TargetNodeID] = struct{}{}
}

// waitForReplicatorDeleteEvent waits for a node to publish a
// replicator completed event on the local event bus.
func waitForReplicatorDeleteEvent(s *state.State, cfg DeleteReplicator) {
	select {
	case _, ok := <-s.Nodes[cfg.SourceNodeID].Event.Replicator.Message():
		if !ok {
			require.Fail(s.T, "subscription closed waiting for replicator event")
		}

	case <-time.After(eventTimeout):
		require.Fail(s.T, "timeout waiting for replicator event")
	}

	delete(s.Nodes[cfg.TargetNodeID].P2P.Connections, cfg.SourceNodeID)
	delete(s.Nodes[cfg.SourceNodeID].P2P.Connections, cfg.TargetNodeID)
	delete(s.Nodes[cfg.SourceNodeID].P2P.Replicators, cfg.TargetNodeID)
}

// waitForAddCollectionSubscriptionEvent waits for a node to publish a
// p2p topic completed event on the local event bus.
//
// Expected document heads will be updated for the subscriber node.
func waitForAddCollectionSubscriptionEvent(s *state.State, action AddCollectionSubscription) {
	// update peer collections of target node
	for _, collectionIndex := range action.CollectionIDs {
		if collectionIndex == NonExistentCollectionID {
			continue // don't track non existent collections
		}
		s.Nodes[action.NodeID].P2P.PeerCollections[collectionIndex] = struct{}{}
	}
}

// waitForDeleteCollectionSubscriptionEvent waits for a node to publish a
// p2p topic completed event on the local event bus.
func waitForDeleteCollectionSubscriptionEvent(s *state.State, action DeleteCollectionSubscription) {
	for _, collectionIndex := range action.CollectionIDs {
		if collectionIndex == NonExistentCollectionID {
			continue // don't track non existent collections
		}
		delete(s.Nodes[action.NodeID].P2P.PeerCollections, collectionIndex)
	}
}

// waitForAddDocumentSubscriptionEvent waits for a node to publish a
// p2p topic completed event on the local event bus.
//
// Expected document heads will be updated for the subscriber node.
func waitForAddDocumentSubscriptionEvent(s *state.State, action AddDocumentSubscription) {
	// update peer documents of target node
	for _, colDocIndex := range action.DocIDs {
		if colDocIndex.Doc == NonExistentDocID {
			continue // don't track non existent documents
		}
		s.Nodes[action.NodeID].P2P.PeerDocuments[colDocIndex] = struct{}{}
	}
}

// waitForDeleteDocumentSubscriptionEvent waits for a node to publish a
// p2p topic completed event on the local event bus.
func waitForDeleteDocumentSubscriptionEvent(s *state.State, action DeleteDocumentSubscription) {
	for _, colDocIndex := range action.DocIDs {
		if colDocIndex.Doc == NonExistentDocID {
			continue // don't track non existent documents
		}
		delete(s.Nodes[action.NodeID].P2P.PeerDocuments, colDocIndex)
	}
}

// waitForUpdateEvents waits for all selected nodes to publish an
// update event to the local event bus.
//
// Expected document heads will be updated for any connected nodes.
func waitForUpdateEvents(
	s *state.State,
	nodeID immutable.Option[int],
	collectionIndex int,
	docIDs map[string]struct{},
	ident immutable.Option[state.Identity],
) {
	for i := 0; i < len(s.Nodes); i++ {
		if nodeID.HasValue() && nodeID.Value() != i {
			continue // node is not selected
		}

		node := s.Nodes[i]
		if node.Closed {
			continue // node is closed
		}

		expect := make(map[string]struct{}, len(docIDs))

		col := node.Collections[collectionIndex]
		if col.Version().IsBranchable {
			expect[col.CollectionID()] = struct{}{}
		}
		for k := range docIDs {
			expect[k] = struct{}{}
		}

		for len(expect) > 0 {
			var evt event.Update
		relayCheck:
			// We need to ensure the message was not from a previously relayed update.
			// If it is, we try the next one.
			for {
				select {
				case msg, ok := <-node.Event.Update.Message():
					if !ok {
						require.Fail(s.T, "subscription closed waiting for update event", "Node %d", i)
					}
					evt = msg.Data.(event.Update)

					node.CompositesLock.Lock()
					// We keep track of the list of cids for all documents in the test
					// in case we want to use them in subsequent test actions without having
					// to know in advance what the CID will be.
					if node.Composites == nil {
						node.Composites = make(map[string][]cid.Cid)
					}
					node.Composites[evt.DocID] = append(node.Composites[evt.DocID], evt.Cid)
					node.CompositesLock.Unlock()

					if !evt.IsRelay {
						break relayCheck
					}

				case <-time.After(eventTimeout):
					require.Fail(s.T, "timeout waiting for update event", "Node %d", i)
				}
			}

			// make sure the event is expected
			_, ok := expect[getUpdateEventKey(evt)]
			require.True(s.T, ok, "unexpected document update", getUpdateEventKey(evt))
			delete(expect, getUpdateEventKey(evt))

			// we only need to update the network state if the nodes
			// are configured for networking
			if s.IsNetworkEnabled {
				updateNetworkState(s, i, evt, ident)
			}
		}
	}
}

// waitForMergeEvents waits for all expected heads to be merged to all nodes.
//
// Will fail the test if an event is not received within the expected time interval to prevent tests
// from running forever.
func waitForMergeEvents(s *state.State, action WaitForSync) {
	for nodeID := 0; nodeID < len(s.Nodes); nodeID++ {
		node := s.Nodes[nodeID]
		if node.Closed {
			continue // node is closed
		}

		expect := node.P2P.ExpectedDAGHeads

		// remove any heads that are already merged
		// up to the expected head
		for key, val := range node.P2P.ActualDAGHeads {
			if head, ok := expect[key]; ok && head.String() == val.CID.String() {
				delete(expect, key)
			}
		}

		// wait for all expected heads to be merged
		//
		// the order of merges does not matter as we only
		// expect the latest head to eventually be merged
		//
		// unexpected merge events are ignored
		for len(expect) > 0 {
			var evt event.MergeComplete
			select {
			case msg, ok := <-node.Event.Merge.Message():
				if !ok {
					require.Fail(s.T, "subscription closed waiting for merge complete event")
				}
				evt = msg.Data.(event.MergeComplete)

			case <-time.After(30 * eventTimeout):
				require.Fail(s.T, "timeout waiting for merge complete event")
			}

			head, ok := expect[getMergeEventKey(evt.Merge)]
			if ok && head.String() == evt.Merge.Cid.String() {
				delete(expect, getMergeEventKey(evt.Merge))
			}
			node.P2P.ActualDAGHeads[getMergeEventKey(evt.Merge)] = state.DocHeadState{
				CID: evt.Merge.Cid,
			}
		}
	}
}

func waitForSESync(s *state.State, action WaitForSESync) {
	var docIDsToWait []string
	s.DocIDsLock.RLock()
	if len(action.DocIDs) > 0 {
		for _, docIndex := range action.DocIDs {
			if len(s.DocIDs[0]) <= docIndex {
				require.Fail(s.T, "doc index %d out of range", docIndex)
			}
			docIDsToWait = append(docIDsToWait, s.DocIDs[0][docIndex].String())
		}
	} else {
		// Wait for all documents if no specific IDs provided
		for _, docID := range s.DocIDs[0] {
			docIDsToWait = append(docIDsToWait, docID.String())
		}
	}
	s.DocIDsLock.RUnlock()

	// SE sync events are only published on replicator nodes (nodes that receive artifacts)
	// We wait for events from any non-source node with active replicators
	for nodeID := 1; nodeID < len(s.Nodes); nodeID++ {
		node := s.Nodes[nodeID]
		if node.Closed {
			continue // node is closed
		}

		expectedSyncs := make(map[string]struct{}, len(docIDsToWait))
		for _, docID := range docIDsToWait {
			expectedSyncs[docID] = struct{}{}
		}

		for len(expectedSyncs) > 0 {
			select {
			case msg, ok := <-node.Event.SESync.Message():
				if !ok {
					require.Fail(s.T, "subscription closed waiting for SE sync complete event")
				}
				evt := msg.Data.(event.SEArtifactReceived)

				delete(expectedSyncs, evt.DocID)

			case <-time.After(30 * eventTimeout):
				require.Fail(s.T, fmt.Sprintf("timeout waiting for SE sync complete event on node %d. Remaining: %v",
					nodeID, expectedSyncs))
			}
		}
	}
}

// updateNetworkState updates the network state by checking which
// nodes should receive the updated document in the given update event.
func updateNetworkState(s *state.State, nodeID int, evt event.Update, ident immutable.Option[state.Identity]) {
	// find the correct collection index for this update
	collectionID := -1
	for i, c := range s.Nodes[nodeID].Collections {
		if c.Version().CollectionID == evt.CollectionID {
			collectionID = i
		}
	}
	docIndex := -1
	if collectionID != -1 {
		s.DocIDsLock.RLock()
		for i, docID := range s.DocIDs[collectionID] {
			if docID.String() == evt.DocID {
				docIndex = i
			}
		}
		s.DocIDsLock.RUnlock()
	}

	node := s.Nodes[nodeID]

	// update the actual document head on the node that updated it
	// as the node added the document, it is already decrypted
	node.P2P.ActualDAGHeads[getUpdateEventKey(evt)] = state.DocHeadState{CID: evt.Cid}

	// update the expected document heads of replicator targets
	for id := range node.P2P.Replicators {
		// replicator target nodes push updates to source nodes
		s.Nodes[id].P2P.ExpectedDAGHeads[getUpdateEventKey(evt)] = evt.Cid
	}

	updateConnectedNodes(s, nodeID, map[int]struct{}{}, ident, collectionID, docIndex, evt)
}

// updateConnectedNodes updates the expected document heads of connected nodes
func updateConnectedNodes(
	s *state.State,
	nodeID int,
	nodesCovered map[int]struct{},
	ident immutable.Option[state.Identity],
	collectionID int,
	docIndex int,
	evt event.Update,
) {
	if _, ok := nodesCovered[nodeID]; ok {
		return
	}
	nodesCovered[nodeID] = struct{}{}
	for id := range s.Nodes[nodeID].P2P.Connections {
		if _, ok := nodesCovered[id]; ok {
			continue
		}
		if ident.HasValue() && ident.Value().Selector != strconv.Itoa(id) {
			// If the document is created by a specific identity, only the node with the
			// same index as the identity can initially access it.
			// If this network state update comes from the adding of an actor relationship,
			// then the identity reflects that of the target node.
			continue
		}
		// peer collection subscribers receive updates from any other subscriber node
		if _, ok := s.Nodes[id].P2P.PeerCollections[collectionID]; ok {
			s.Nodes[id].P2P.ExpectedDAGHeads[getUpdateEventKey(evt)] = evt.Cid
		}
		// peer document subscribers receive updates from any other subscriber node
		if _, ok := s.Nodes[id].P2P.PeerDocuments[state.NewColDocIndex(collectionID, docIndex)]; ok {
			s.Nodes[id].P2P.ExpectedDAGHeads[getUpdateEventKey(evt)] = evt.Cid
		}

		updateConnectedNodes(s, id, nodesCovered, ident, collectionID, docIndex, evt)
	}
}

// getEventsForUpdateDoc returns a map of docIDs that should be
// published to the local event bus after an UpdateDoc action.
func getEventsForUpdateDoc(s *state.State, action UpdateDoc) map[string]struct{} {
	s.DocIDsLock.RLock()
	docID := s.DocIDs[action.CollectionID][action.DocID]
	s.DocIDsLock.RUnlock()

	docMap := make(map[string]any)
	err := json.Unmarshal([]byte(action.Doc), &docMap)
	require.NoError(s.T, err)

	return map[string]struct{}{
		docID.String(): {},
	}
}

func waitForSync(s *state.State, action WaitForSync) {
	waitForMergeEvents(s, action)
}

// getEventsForUpdateWithFilter returns a map of docIDs that should be
// published to the local event bus after a UpdateWithFilter action.
func getEventsForUpdateWithFilter(
	s *state.State,
	action UpdateWithFilter,
	result *client.UpdateResult,
) map[string]struct{} {
	var docPatch map[string]any
	err := json.Unmarshal([]byte(action.Updater), &docPatch)
	require.NoError(s.T, err)

	expect := make(map[string]struct{}, len(result.DocIDs))

	for _, docID := range result.DocIDs {
		expect[docID] = struct{}{}
	}

	return expect
}

// getUpdateEventKey gets the identifier to which this event is scoped to.
//
// For example, if this is scoped to a document, the document ID will be
// returned.  If it is scoped to a schema, the schema root will be returned.
func getUpdateEventKey(evt event.Update) string {
	if evt.DocID == "" {
		return evt.CollectionID
	}

	return evt.DocID
}

// getMergeEventKey gets the identifier to which this event is scoped to.
//
// For example, if this is scoped to a document, the document ID will be
// returned.  If it is scoped to a schema, the schema root will be returned.
func getMergeEventKey(evt event.Merge) string {
	if evt.DocID == "" {
		return evt.CollectionID
	}

	return evt.DocID
}
