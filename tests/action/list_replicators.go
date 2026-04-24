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
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/state"
)

// ListReplicators gets the configured replicators for the given node and compares
// them against the expected results.
type ListReplicators struct {
	stateful

	// NodeID is the node ID (index) of the node in which to get the replicators for.
	NodeID int

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// ExpectedTargetNodeIDs are the node IDs (indexes) of the expected replicator targets.
	ExpectedTargetNodeIDs []int

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

var _ Action = (*ListReplicators)(nil)
var _ Stateful = (*ListReplicators)(nil)

func (a *ListReplicators) Execute() {
	node := a.s.Nodes[a.NodeID]

	opts := options.ListReplicators()
	identOption := getIdentityForRequestSpecificToNode(a.s, a.Identity, a.NodeID)
	if identOption.HasValue() {
		opts.SetIdentity(identOption.Value())
	}

	reps, err := node.ListReplicators(a.s.Ctx, opts)

	expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)
	assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)

	if !expectedErrorRaised {
		expectedIDs := []string{}
		for _, targetNodeID := range a.ExpectedTargetNodeIDs {
			// Inject the target node's identity to bypass NAC for the gated [PeerInfo] operation,
			// otherwise due to lack of authorization(s) we might not be able to see the peer
			// addresses at all.
			nodeIdentity := NodeIdentity(targetNodeID)
			peerInfoOpts := options.PeerInfo()
			identOpt := getIdentityForRequestSpecificToNode(a.s, nodeIdentity, targetNodeID)
			if identOpt.HasValue() {
				peerInfoOpts.SetIdentity(identOpt.Value())
			}
			targetAddresses, err := a.s.Nodes[targetNodeID].PeerInfo(a.s.Ctx, peerInfoOpts)
			require.NoError(a.s.T, err)
			require.NotZero(a.s.T, len(targetAddresses))

			maddr, err := multiaddr.NewMultiaddr(targetAddresses[0])
			require.NoError(a.s.T, err)
			id, err := maddr.ValueForProtocol(multiaddr.P_P2P)
			require.NoError(a.s.T, err)
			expectedIDs = append(expectedIDs, id)
		}

		actualIDs := []string{}
		for _, rep := range reps {
			actualIDs = append(actualIDs, rep.ID)
		}

		assert.ElementsMatch(a.s.T, expectedIDs, actualIDs)
	}
}
