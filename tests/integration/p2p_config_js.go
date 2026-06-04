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
	"github.com/sourcenetwork/defradb/client/options"
)

func RandomNetworkingConfig() ConfigureNode {
	return func() options.NodeP2POptions {
		return options.NodeP2POptions{}
	}
}

func withPrivateKey(_ *options.NodeP2POptions, _ []byte) {
	// JS builds don't support P2P
}

func withListenAddresses(_ *options.NodeP2POptions, _ ...string) {
	// JS builds don't support P2P
}
