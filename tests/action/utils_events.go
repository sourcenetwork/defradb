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
	"fmt"
	"strconv"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/event"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
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
		if node.IsExternal {
			// A node in another process emits no events we can read, so there is
			// nothing to wait for here. The write still has to reach the nodes this
			// one replicates to, and normally those nodes learn what to expect from
			// this event. Tell them the document ID directly instead, or they would
			// wait for nothing and the test would read the data too early.
			//
			// A node with no replicators has no one to tell, so this does nothing
			// when networking is not in use.
			MarkDocsExpectedOnTargets(s, i, collectionIndex, docIDs, ident)
			continue
		}

		expect := make(map[string]struct{}, len(docIDs))

		collections := node.Collections
		col := collections[collectionIndex]
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
					updateKey := getUpdateEventKey(evt)
					node.Composites[updateKey] = append(node.Composites[updateKey], evt.Cid)
					if node.FieldCIDs == nil {
						node.FieldCIDs = make(map[string]map[string][]cid.Cid)
					}
					if node.FieldCIDs[updateKey] == nil {
						node.FieldCIDs[updateKey] = make(map[string][]cid.Cid)
					}
					block, err := coreblock.GetFromBytes(evt.Block)
					require.NoError(s.T, err)
					for _, link := range block.Links {
						node.FieldCIDs[updateKey][link.Name] = append(
							node.FieldCIDs[updateKey][link.Name],
							link.Link.Cid,
						)
					}
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

// MarkDocsExpectedOnTargets records that the given documents should reach every
// node the source syncs to.
//
// It is the counterpart of [updateNetworkState] for a source node whose events
// cannot be read. The head CID is read back from the node with a query rather
// than taken from an event, so the nodes downstream wait on the same head they
// would have anyway.
func MarkDocsExpectedOnTargets(
	s *state.State,
	sourceNodeID int,
	collectionIndex int,
	docIDs map[string]struct{},
	ident immutable.Option[state.Identity],
) {
	for docID := range docIDs {
		// The source node wrote this document, so it must be able to report the
		// commit. Skipping instead would record nothing to wait for, and the
		// assertions that follow would run against data that never arrived and
		// still pass.
		head, ok := latestCompositeCID(s, sourceNodeID, docID)
		require.True(s.T, ok, "node %d could not report the head of %s", sourceNodeID, docID)

		// Build the event ourselves, since we cannot read the real one.
		evt := event.Update{
			DocID:        docID,
			Cid:          head,
			CollectionID: collectionIDForIndex(s, sourceNodeID, collectionIndex),
		}

		s.Nodes[sourceNodeID].P2P.ActualDAGHeads[docID] = state.DocHeadState{CID: head}

		for targetID := range s.Nodes[sourceNodeID].P2P.Replicators {
			s.Nodes[targetID].P2P.ExpectedDAGHeads[docID] = append(
				s.Nodes[targetID].P2P.ExpectedDAGHeads[docID],
				state.ExpectedHead{CID: head, SourceNodeID: sourceNodeID},
			)
		}

		// Subscribers are reached over connections rather than replicators, so
		// they need the same walk the native path does.
		updateConnectedNodes(
			s, sourceNodeID, sourceNodeID, map[int]struct{}{}, ident,
			collectionIndex, docIndexForID(s, collectionIndex, docID), evt,
		)
	}
}

// collectionIDForIndex returns the collection ID for a collection index on a node.
func collectionIDForIndex(s *state.State, nodeID int, collectionIndex int) string {
	collections := s.Nodes[nodeID].Collections
	if collectionIndex < 0 || collectionIndex >= len(collections) {
		return ""
	}
	return collections[collectionIndex].Version().CollectionID
}

// docIndexForID returns the index a document was added under, or -1 if unknown.
func docIndexForID(s *state.State, collectionIndex int, docID string) int {
	s.DocIDsLock.RLock()
	defer s.DocIDsLock.RUnlock()

	if collectionIndex < 0 || collectionIndex >= len(s.DocIDs) {
		return -1
	}
	for i, id := range s.DocIDs[collectionIndex] {
		if id.String() == docID {
			return i
		}
	}
	return -1
}

// latestCompositeCID asks the node for the newest composite commit of a
// document.
//
// The composite commit is the one a merge event reports, so this is the same
// CID the native path takes from that event.
func latestCompositeCID(s *state.State, nodeID int, docID string) (cid.Cid, bool) {
	result := s.Nodes[nodeID].ExecRequest(
		s.Ctx,
		fmt.Sprintf(
			`query { _commits(docID: %q, filter: {fieldName: {_eq: "_C"}}, order: {height: DESC}, limit: 1) { cid } }`,
			docID,
		),
	)
	if len(result.GQL.Errors) > 0 {
		return cid.Cid{}, false
	}

	data, ok := result.GQL.Data.(map[string]any)
	if !ok {
		return cid.Cid{}, false
	}

	var cidStr string
	switch commits := data["_commits"].(type) {
	case []any:
		if len(commits) == 0 {
			return cid.Cid{}, false
		}
		commit, ok := commits[0].(map[string]any)
		if !ok {
			return cid.Cid{}, false
		}
		cidStr, _ = commit["cid"].(string)
	case []map[string]any:
		if len(commits) == 0 {
			return cid.Cid{}, false
		}
		cidStr, _ = commits[0]["cid"].(string)
	default:
		return cid.Cid{}, false
	}

	parsed, err := cid.Decode(cidStr)
	if err != nil {
		return cid.Cid{}, false
	}
	return parsed, true
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

	updateConnectedNodes(s, nodeID, nodeID, map[int]struct{}{}, ident, collectionID, docIndex, evt)
}

// updateConnectedNodes updates the expected document heads of connected nodes.
// originNodeID is the node that authored the update and stays constant through recursion.
// nodeID is the current node being visited in the connection graph traversal.
func updateConnectedNodes(
	s *state.State,
	originNodeID int,
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
				state.ExpectedHead{CID: evt.Cid, SourceNodeID: originNodeID},
			)
		}
		// peer document subscribers receive updates from any other subscriber node
		if _, ok := s.Nodes[id].P2P.PeerDocuments[state.NewColDocIndex(collectionID, docIndex)]; ok {
			s.Nodes[id].P2P.ExpectedDAGHeads[getUpdateEventKey(evt)] = append(
				s.Nodes[id].P2P.ExpectedDAGHeads[getUpdateEventKey(evt)],
				state.ExpectedHead{CID: evt.Cid, SourceNodeID: originNodeID},
			)
		}

		updateConnectedNodes(s, originNodeID, id, nodesCovered, ident, collectionID, docIndex, evt)
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
