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
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
)

// AddReplicator configures a directional replicator relationship between
// two nodes.
//
// All document changes made in the source node will be synced to the target node.
// New documents added in the target node will not be synced to the source node,
// however updates in the target node to documents synced from the source node will
// be synced back to the source node.
type AddReplicator struct {
	// SourceNodeID is the node ID (index) of the node from which data should be replicated.
	//
	// Note: The request will use identity (if specified) of the Source Node.
	SourceNodeID int

	// TargetNodeID is the node ID (index) of the node to which data should be replicated.
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

// DeleteReplicator deletes a directional replicator relationship between two nodes.
type DeleteReplicator struct {
	// SourceNodeID is the node ID (index) of the node from which the replicator should be deleted.
	//
	// Note: The request will use identity (if specified) of the Source Node.
	SourceNodeID int

	// TargetNodeID is the node ID (index) of the node to which the replicator should be deleted.
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

// ListReplicators gets the configured replicators for the given node and compares them against the
// expected results.
// TODO: Test ListReplicators with and without NAC.
// https://github.com/sourcenetwork/defradb/issues/4109
type ListReplicators struct {
	// NodeID is the node ID (index) of the node in which to get the replicators for.
	NodeID int

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// ExpectedCollectionIDs are the collection IDs (indexes) of the collections expected.
	ExpectedTargetNodeIDs []int

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// addReplicator configures a replicator relationship between two existing, started, nodes.
// It returns a channel that will receive an empty struct upon sync completion of all expected
// replicator-sync events.
//
// Any errors generated whilst configuring the peers or waiting on sync will result in a test failure.
func addReplicator(
	s *state.State,
	cfg AddReplicator,
) {
	sourceNode := s.Nodes[cfg.SourceNodeID]
	targetNode := s.Nodes[cfg.TargetNodeID]

	// Inject target node's identity into the context to bypass NAC for the gated [PeerInfo] operation,
	// otherwise due to lack of authorization(s) we might not be able to see the peer addresses at all.
	nodeIdentity := NodeIdentity(cfg.TargetNodeID)

	peerInfoOpts := options.PeerInfo()
	identOption := getIdentityForRequestSpecificToNode(s, nodeIdentity, cfg.TargetNodeID)
	if identOption.HasValue() {
		peerInfoOpts.SetIdentity(identOption.Value())
	}
	targetAddresses, err := targetNode.PeerInfo(s.Ctx, peerInfoOpts)
	require.NoError(s.T, err)

	opt := options.WithIdentity(options.AddReplicator(),
		getIdentityForRequestSpecificToNode(s, cfg.Identity, cfg.SourceNodeID))
	err = sourceNode.AddReplicator(s.Ctx, targetAddresses, nil, opt)

	expectedErrorRaised := AssertError(s.T, err, cfg.ExpectedError)
	assertExpectedErrorRaised(s.T, cfg.ExpectedError, expectedErrorRaised)

	if err == nil {
		waitForReplicatorConfigureEvent(s, cfg)
	}
}

func deleteReplicator(
	s *state.State,
	cfg DeleteReplicator,
) {
	sourceNode := s.Nodes[cfg.SourceNodeID]
	targetNode := s.Nodes[cfg.TargetNodeID]

	// Inject target node's identity into the context to bypass NAC for the gated [PeerInfo] operation,
	// otherwise due to lack of authorization(s) we might not be able to see the peer addresses at all.
	nodeIdentity := NodeIdentity(cfg.TargetNodeID)
	peerInfoOpts := options.PeerInfo()
	identOption := getIdentityForRequestSpecificToNode(s, nodeIdentity, cfg.TargetNodeID)
	if identOption.HasValue() {
		peerInfoOpts.SetIdentity(identOption.Value())
	}
	targetAddresses, err := targetNode.PeerInfo(s.Ctx, peerInfoOpts)
	require.NoError(s.T, err)
	require.NotZero(s.T, len(targetAddresses))

	maddr, err := multiaddr.NewMultiaddr(targetAddresses[0])
	require.NoError(s.T, err)
	id, err := maddr.ValueForProtocol(multiaddr.P_P2P)
	require.NoError(s.T, err)

	opt := options.WithIdentity(options.DeleteReplicator(),
		getIdentityForRequestSpecificToNode(s, cfg.Identity, cfg.SourceNodeID))
	err = sourceNode.DeleteReplicator(s.Ctx, id, nil, opt)

	expectedErrorRaised := AssertError(s.T, err, cfg.ExpectedError)
	assertExpectedErrorRaised(s.T, cfg.ExpectedError, expectedErrorRaised)

	if err == nil {
		waitForReplicatorDeleteEvent(s, cfg)
	}
}
