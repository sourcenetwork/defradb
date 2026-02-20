// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package action

import (
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
)

// PeerInfo returns the p2p host list of addresses.
type PeerInfo struct {
	stateful

	// NodeID is the ID (index) of the node to execute the PeerInfo request on.
	NodeID int

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// Expected number of total peers in the list of peers that will be returned, they will all
	// be validated, we just don't assert them individually due to maintainance cost.
	ExpectedNumberOfPeers int

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

var _ Action = (*PeerInfo)(nil)
var _ Stateful = (*PeerInfo)(nil)

func (a *PeerInfo) Execute() {
	node := a.s.Nodes[a.NodeID]

	opts := options.PeerInfo()
	identOption := getIdentityForRequestSpecificToNode(a.s, a.Identity, a.NodeID)
	if identOption.HasValue() {
		opts.SetIdentity(identOption.Value())
	}
	peerInfos, err := node.PeerInfo(a.s.Ctx, opts)

	expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)
	assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)

	if !expectedErrorRaised {
		assert.Equal(a.s.T, a.ExpectedNumberOfPeers, len(peerInfos))

		// Check that all the returned addresses in the list are valid.
		for _, peerInfo := range peerInfos {
			maddr, err := multiaddr.NewMultiaddr(peerInfo)
			require.NoError(a.s.T, err)

			_, err = peer.IDFromP2PAddr(maddr)
			require.NoError(a.s.T, err)
		}
	}
}
