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
	"github.com/sourcenetwork/defradb/node"
)

func RandomNetworkingConfig() ConfigureNode {
	return func() []node.Option {
		return []node.Option{}
	}
}

func getP2POptions(_ []node.Option) []node.Option {
	return []node.Option{}

}

func withPrivateKey(opts []node.Option, _ []byte) []node.Option {
	return opts
}

func withWithListenAddresses(opts []node.Option, _ ...string) []node.Option {
	return opts
}
