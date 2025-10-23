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
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/stretchr/testify/require"
)

// ConfigureReplicator configures a directional replicator relationship between
// two nodes.
//
// All document changes made in the source node will be synced to the target node.
// New documents created in the target node will not be synced to the source node,
// however updates in the target node to documents synced from the source node will
// be synced back to the source node.
type ConfigureReplicator struct {
	// SourceNodeID is the node ID (index) of the node from which data should be replicated.
	SourceNodeID int

	// TargetNodeID is the node ID (index) of the node to which data should be replicated.
	TargetNodeID int

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// DeleteReplicator deletes a directional replicator relationship between two nodes.
type DeleteReplicator struct {
	// SourceNodeID is the node ID (index) of the node from which the replicator should be deleted.
	SourceNodeID int

	// TargetNodeID is the node ID (index) of the node to which the replicator should be deleted.
	TargetNodeID int
}

// GetAllReplicators gets the configured replicators for the given node and compares them against the
// expected results.
type GetAllReplicators struct {
	// NodeID is the node ID (index) of the node in which to get the subscriptions for.
	NodeID int

	// ExpectedCollectionIDs are the collection IDs (indexes) of the collections expected.
	ExpectedTargetNodeIDs []int
}

// configureReplicator configures a replicator relationship between two existing, started, nodes.
// It returns a channel that will receive an empty struct upon sync completion of all expected
// replicator-sync events.
//
// Any errors generated whilst configuring the peers or waiting on sync will result in a test failure.
func configureReplicator(
	s *state.State,
	cfg ConfigureReplicator,
) {
	sourceNode := s.Nodes[cfg.SourceNodeID]
	targetNode := s.Nodes[cfg.TargetNodeID]

	targetAddresses, err := targetNode.PeerInfo()
	require.NoError(s.T, err)
	err = sourceNode.SetReplicator(s.Ctx, targetAddresses)

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

	targetAddresses, err := targetNode.PeerInfo()
	require.NoError(s.T, err)
	require.NotZero(s.T, len(targetAddresses))
	maddr, err := multiaddr.NewMultiaddr(targetAddresses[0])
	require.NoError(s.T, err)
	id, err := maddr.ValueForProtocol(multiaddr.P_P2P)
	require.NoError(s.T, err)
	err = sourceNode.DeleteReplicator(s.Ctx, id)
	require.NoError(s.T, err)
	waitForReplicatorDeleteEvent(s, cfg)
}
