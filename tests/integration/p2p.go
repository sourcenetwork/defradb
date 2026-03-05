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

package tests

import (
	"context"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client/options"
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
	//
	// Note: The request will use identity (if specified) of the Source Node.
	SourceNodeID int

	// TargetNodeID is the node ID (index) of the second node to connect.
	//
	// Is completely interchangeable with SourceNodeID and which way round
	// these properties are specified is purely cosmetic.
	//
	// Note: The request will use identity (if specified) of the Source Node.
	TargetNodeID int

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// WaitForSync is an action that instructs the test framework to wait for all document synchronization
// to complete before progressing.
//
// For example you will likely wish to `WaitForSync` after adding a document in node 0 before querying
// node 1 to see if it has been replicated.
type WaitForSync struct{}

// WaitForSESync waits for SE artifact synchronization to complete.
type WaitForSESync struct {
	// DocIDs is a list of document indexes expected to have SE artifacts synced.
	// If empty, waits for SE sync of all added documents.
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

	// Inject source/target node's identity into the context to bypass NAC for the gated [PeerInfo] operation,
	// otherwise due to lack of authorization(s) we might not be able to see the peer addresses at all.
	sourceOpts := options.PeerInfo()
	sourceIdent := getIdentityForRequestSpecificToNode(s, NodeIdentity(cfg.SourceNodeID), cfg.SourceNodeID)
	if sourceIdent.HasValue() {
		sourceOpts.SetIdentity(sourceIdent.Value())
	}

	targetOpts := options.PeerInfo()
	targetIdent := getIdentityForRequestSpecificToNode(s, NodeIdentity(cfg.TargetNodeID), cfg.TargetNodeID)
	if targetIdent.HasValue() {
		targetOpts.SetIdentity(targetIdent.Value())
	}

	sourceAddresses, err := sourceNode.PeerInfo(s.Ctx, sourceOpts)
	require.NoError(s.T, err)
	targetAddresses, err := targetNode.PeerInfo(s.Ctx, targetOpts)
	require.NoError(s.T, err)

	log.InfoContext(s.Ctx, "Connect peers",
		corelog.Any("Source", sourceAddresses),
		corelog.Any("Target", targetAddresses),
	)

	opt := options.WithIdentity(options.Connect(),
		getIdentityForRequestSpecificToNode(s, cfg.Identity, cfg.SourceNodeID))

	err = connectWithRetry(s.Ctx, sourceNode, targetAddresses, opt)

	expectedErrorRaised := AssertError(s.T, err, cfg.ExpectedError)
	assertExpectedErrorRaised(s.T, cfg.ExpectedError, expectedErrorRaised)

	s.Nodes[cfg.SourceNodeID].P2P.Connections[cfg.TargetNodeID] = struct{}{}
	s.Nodes[cfg.TargetNodeID].P2P.Connections[cfg.SourceNodeID] = struct{}{}

	// Bootstrap triggers a bunch of async stuff for which we have no good way of waiting on.  It must be
	// allowed to complete before documentation begins or it will not even try and sync it. So for now, we
	// sleep a little.
	time.Sleep(10 * time.Millisecond)
}

// reconnectPeers makes sure that all peers are connected after a node restart action.
func reconnectPeers(s *state.State) {
	nodeIDs, nodes := getNodesWithIDs(immutable.None[int](), s.Nodes)
	for sourceIndex, sourceNode := range nodes {
		sourceNodeID := nodeIDs[sourceIndex]
		// Inject every source node's identity into the context while refreshing so the [Connect] & [PeerInfo]
		// call doesn't fail due to lack of authorization(s) if NAC is enabled.
		nodeIdentity := NodeIdentity(sourceNodeID)
		sourceOpts := options.PeerInfo()
		sourceIdent := getIdentityForRequestSpecificToNode(s, nodeIdentity, sourceNodeID)
		if sourceIdent.HasValue() {
			sourceOpts.SetIdentity(sourceIdent.Value())
		}

		for targetIndex := range sourceNode.P2P.Connections {
			targetNode := nodes[targetIndex]
			targetNodeID := nodeIDs[targetIndex]
			// Inject target node's identity into the context to bypass NAC for the gated [PeerInfo] operation,
			// otherwise due to lack of authorization(s) we might not be able to see the peer addresses at all.
			targetOpts := options.PeerInfo()
			targetIdent := getIdentityForRequestSpecificToNode(s, NodeIdentity(targetNodeID), targetNodeID)
			if targetIdent.HasValue() {
				targetOpts.SetIdentity(targetIdent.Value())
			}
			sourceAddresses, err := sourceNode.PeerInfo(s.Ctx, sourceOpts)
			require.NoError(s.T, err)
			targetAddresses, err := targetNode.PeerInfo(s.Ctx, targetOpts)
			require.NoError(s.T, err)

			log.InfoContext(s.Ctx, "Connect peers",
				corelog.Any("Source", sourceAddresses),
				corelog.Any("Target", targetAddresses),
			)

			opt := options.WithIdentity(options.Connect(),
				getIdentityForRequestSpecificToNode(s, nodeIdentity, sourceNodeID))
			err = connectWithRetry(s.Ctx, sourceNode, targetAddresses, opt)
			require.NoError(s.T, err)
		}
	}
}

// connectWithRetry attempts to connect to target addresses with retry logic
// to handle transient connection failures.
func connectWithRetry(
	ctx context.Context,
	node *state.NodeState,
	targetAddresses []string,
	opt options.Enumerable[options.ConnectOptions],
) error {
	const maxRetries = 5
	const retryDelay = 50 * time.Millisecond

	var lastErr error
	for attempt := range maxRetries {
		lastErr = node.Connect(ctx, targetAddresses, opt)
		if lastErr == nil {
			return nil
		}
		if attempt < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}
	return lastErr
}

// syncDocs requests document sync from peers.
func syncDocs(s *state.State, action SyncDocs) {
	node := s.Nodes[action.NodeID]

	docIDStrings := make([]string, len(action.DocIDs))
	for i, docIndex := range action.DocIDs {
		s.DocIDsLock.RLock()
		docIDStrings[i] = s.DocIDs[action.CollectionID][docIndex].String()
		s.DocIDsLock.RUnlock()
	}

	collectionName := s.Nodes[action.NodeID].Collections[action.CollectionID].Name()

	syncOpts := options.SyncDocuments()
	identOption := getIdentityForRequestSpecificToNode(s, action.Identity, action.NodeID)
	if identOption.HasValue() {
		syncOpts.SetIdentity(identOption.Value())
	}

	err := withRetryOnNode(
		node,
		func() error {
			return node.SyncDocuments(
				s.Ctx,
				collectionName,
				docIDStrings,
				syncOpts,
			)
		},
	)

	expectedErrorRaised := AssertError(s.T, err, action.ExpectedError)

	assertExpectedErrorRaised(s.T, action.ExpectedError, expectedErrorRaised)

	if !expectedErrorRaised {
		s.DocIDsLock.RLock()
		for i, docInd := range action.DocIDs {
			nodeID := action.SourceNodes[i]
			docID := s.DocIDs[action.CollectionID][docInd].String()
			node.P2P.ExpectedDAGHeads[docID] = s.Nodes[nodeID].P2P.ActualDAGHeads[docID].CID
		}
		s.DocIDsLock.RUnlock()
	}
}
