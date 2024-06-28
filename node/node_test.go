// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package node

import (
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithDisableP2P(t *testing.T) {
	options := &Options{}
	WithDisableP2P(true)(options)
	assert.Equal(t, true, options.disableP2P)
}

func TestWithDisableAPI(t *testing.T) {
	options := &Options{}
	WithDisableAPI(true)(options)
	assert.Equal(t, true, options.disableAPI)
}

func TestWithPeers(t *testing.T) {
	peer, err := peer.AddrInfoFromString("/ip4/127.0.0.1/tcp/9000/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	require.NoError(t, err)

	options := &Options{}
	WithPeers(*peer)(options)

	require.Len(t, options.peers, 1)
	assert.Equal(t, *peer, options.peers[0])
}
