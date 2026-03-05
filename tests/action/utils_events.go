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
	"strconv"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/tests/state"
)

// eventTimeout is the amount of time to wait
// for an event before timing out
const eventTimeout = 1 * time.Second

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
					evt, _ = msg.Data.(event.Update)

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
		s.Nodes[id].P2P.ExpectedDAGHeads[getUpdateEventKey(evt)] = append(
			s.Nodes[id].P2P.ExpectedDAGHeads[getUpdateEventKey(evt)],
			state.ExpectedHead{CID: evt.Cid, SourceNodeID: nodeID},
		)
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
			s.Nodes[id].P2P.ExpectedDAGHeads[getUpdateEventKey(evt)] = append(
				s.Nodes[id].P2P.ExpectedDAGHeads[getUpdateEventKey(evt)],
				state.ExpectedHead{CID: evt.Cid, SourceNodeID: nodeID},
			)
		}
		// peer document subscribers receive updates from any other subscriber node
		if _, ok := s.Nodes[id].P2P.PeerDocuments[state.NewColDocIndex(collectionID, docIndex)]; ok {
			s.Nodes[id].P2P.ExpectedDAGHeads[getUpdateEventKey(evt)] = append(
				s.Nodes[id].P2P.ExpectedDAGHeads[getUpdateEventKey(evt)],
				state.ExpectedHead{CID: evt.Cid, SourceNodeID: nodeID},
			)
		}

		updateConnectedNodes(s, id, nodesCovered, ident, collectionID, docIndex, evt)
	}
}

// getUpdateEventKey gets the identifier to which this event is scoped to.
//
// For example, if this is scoped to a document, the document ID will be
// returned.  If it is scoped to a collection, the collection root will be returned.
func getUpdateEventKey(evt event.Update) string {
	if evt.DocID == "" {
		return evt.CollectionID
	}

	return evt.DocID
}
