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
