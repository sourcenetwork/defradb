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
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client/options"
)

func RandomNetworkingConfig() *NewNode {
	return &NewNode{
		Network: immutable.Some[ConfigureNode](func() options.NodeP2POptions {
			return options.NodeP2POptions{}
		}),
	}
}

// NoNetworkingConfig returns a node configured with P2P disabled entirely (the node's
// internal db.p2p stays nil).
func NoNetworkingConfig() *NewNode {
	return &NewNode{
		// A present but nil config is what tells node setup to leave p2p out.
		Network: immutable.Some[ConfigureNode](nil),
	}
}

func WithPrivateKey(_ *options.NodeP2POptions, _ []byte) {
	// JS builds don't support P2P
}

func WithListenAddresses(_ *options.NodeP2POptions, _ ...string) {
	// JS builds don't support P2P
}
