// Copyright 2025 Democratized Data Foundation
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

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/tests/state"
)

// ConnectPeers connects two nodes together as peers.
//
// Updates between shared documents should be synced in either direction,
// but new documents will only be synced if explicitly requested (e.g. via
// collection subscription).
type ConnectPeers struct {
	// SourceNodeID is the node ID (index) of the first node to connect.
	//
	// Is completely interchangeable with TargetNodeID and which way round
	// these properties are specified is purely cosmetic.
	SourceNodeID int

	// TargetNodeID is the node ID (index) of the second node to connect.
	//
	// Is completely interchangeable with SourceNodeID and which way round
	// these properties are specified is purely cosmetic.
	TargetNodeID int
}

// WaitForSync is an action that instructs the test framework to wait for all document synchronization
// to complete before progressing.
//
// For example you will likely wish to `WaitForSync` after creating a document in node 0 before querying
// node 1 to see if it has been replicated.
type WaitForSync struct {
	// Decrypted is a list of document indexes that are expected to be merged and synced decrypted.
	Decrypted []int
}

// WaitForSESync waits for SE artifact synchronization to complete.
type WaitForSESync struct {
	// DocIDs is a list of document indexes expected to have SE artifacts synced.
	// If empty, waits for SE sync of all created documents.
	DocIDs []int
}

// connectPeers connects two existing, started, nodes as peers.  It returns a channel
// that will receive an empty struct upon sync completion of all expected peer-sync events.
//
// Any errors generated whilst configuring the peers or waiting on sync will result in a test failure.
func connectPeers(
	s *state.State,
	cfg ConnectPeers,
) {
	sourceNode := s.Nodes[cfg.SourceNodeID]
	targetNode := s.Nodes[cfg.TargetNodeID]

	sourceAddresses, err := sourceNode.PeerInfo()
	require.NoError(s.T, err)
	targetAddresses, err := targetNode.PeerInfo()
	require.NoError(s.T, err)

	log.InfoContext(s.Ctx, "Connect peers",
		corelog.Any("Source", sourceAddresses),
		corelog.Any("Target", targetAddresses))

	err = sourceNode.Connect(s.Ctx, targetAddresses)
	require.NoError(s.T, err)

	s.Nodes[cfg.SourceNodeID].P2P.Connections[cfg.TargetNodeID] = struct{}{}
	s.Nodes[cfg.TargetNodeID].P2P.Connections[cfg.SourceNodeID] = struct{}{}

	// Bootstrap triggers a bunch of async stuff for which we have no good way of waiting on.  It must be
	// allowed to complete before documentation begins or it will not even try and sync it. So for now, we
	// sleep a little.
	time.Sleep(10 * time.Millisecond)
}

// reconnectPeers makes sure that all peers are connected after a node restart action.
func reconnectPeers(s *state.State) {
	for i, n := range s.Nodes {
		for j := range n.P2P.Connections {
			sourceNode := s.Nodes[i]
			targetNode := s.Nodes[j]

			sourceAddresses, err := sourceNode.PeerInfo()
			require.NoError(s.T, err)
			targetAddresses, err := targetNode.PeerInfo()
			require.NoError(s.T, err)

			log.InfoContext(s.Ctx, "Connect peers",
				corelog.Any("Source", sourceAddresses),
				corelog.Any("Target", targetAddresses))

			err = sourceNode.Connect(s.Ctx, targetAddresses)
			require.NoError(s.T, err)
		}
	}
}

// syncDocs requests document sync from peers.
func syncDocs(s *state.State, action SyncDocs) {
	node := s.Nodes[action.NodeID]

	docIDStrings := make([]string, len(action.DocIDs))
	for i, docIndex := range action.DocIDs {
		docIDStrings[i] = s.DocIDs[action.CollectionID][docIndex].String()
	}

	collectionName := s.Nodes[action.NodeID].Collections[action.CollectionID].Name()

	err := withRetryOnNode(
		node,
		func() error {
			return node.SyncDocuments(
				s.Ctx,
				collectionName,
				docIDStrings,
			)
		},
	)

	expectedErrorRaised := AssertError(s.T, err, action.ExpectedError)

	assertExpectedErrorRaised(s.T, action.ExpectedError, expectedErrorRaised)

	if !expectedErrorRaised {
		for i, docInd := range action.DocIDs {
			nodeID := action.SourceNodes[i]
			docID := s.DocIDs[action.CollectionID][docInd].String()
			node.P2P.ExpectedDAGHeads[docID] = s.Nodes[nodeID].P2P.ActualDAGHeads[docID].CID
		}
	}
}
