// Copyright 2026 Democratized Data Foundation
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
	"fmt"
	"testing"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceManagerLimited(t *testing.T) {
	rcmr, err := buildResourceManager(ResourceProfileLimited)
	assert.NoError(t, err)
	assert.NotNil(t, rcmr)

}

func TestResourceManagerServer(t *testing.T) {
	rcmr, err := buildResourceManager(ResourceProfileServer)
	assert.NoError(t, err)
	assert.NotNil(t, rcmr)
}

func TestResourceManagerUnknown(t *testing.T) {
	rcmr, err := buildResourceManager("unknown")
	assert.Error(t, err)
	assert.Nil(t, rcmr)
}

func TestResourceManagerLimitedProfileTransientConnsInbound(t *testing.T) {
	rm, err := buildResourceManager(ResourceProfileLimited)
	require.NoError(t, err)
	defer rm.Close()

	// limited profile sets Transient.ConnsInbound: 16  — .
	conns := make([]network.ConnManagementScope, 16)
	defer func() {
		for _, c := range conns {
			c.Done()
		}
	}()
	for i := range conns {
		// distinct IPs to avoid per IP limit
		addr := multiaddr.StringCast(fmt.Sprintf("/ip4/1.2.3.%d/tcp/1234", i))
		conns[i], err = rm.OpenConnection(network.DirInbound, false, addr)
		require.NoError(t, err)
	}

	// 17th transient inbound connection should be rejected
	addr := multiaddr.StringCast("/ip4/1.2.4.1/tcp/1234")
	_, err = rm.OpenConnection(network.DirInbound, false, addr)
	assert.Error(t, err, "should reject connection beyond inbound limit")
}

func TestResourceManagerLimitedProfileSystemConnsInbound(t *testing.T) {
	rm, err := buildResourceManager(ResourceProfileLimited)
	require.NoError(t, err)
	defer rm.Close()

	// limited profile sets ConnsInbound: 32
	conns := make([]network.ConnManagementScope, 32)
	defer func() {
		for _, c := range conns {
			c.Done()
		}
	}()

	for i := range conns {
		// distinct IPs to avoid per IP limit
		addr := multiaddr.StringCast(fmt.Sprintf("/ip4/1.2.3.%d/tcp/1234", i))
		conns[i], err = rm.OpenConnection(network.DirInbound, false, addr)
		require.NoError(t, err)
		// SetPeer moves each connection out of transient scope so we can exercise the system peer limit.
		err = conns[i].SetPeer(peer.ID(fmt.Sprintf("peer%d", i)))
		require.NoError(t, err)
	}

	// 33rd inbound connection should be rejected
	addr := multiaddr.StringCast("/ip4/1.2.4.1/tcp/1234")
	_, err = rm.OpenConnection(network.DirInbound, false, addr)
	assert.Error(t, err, "should reject connection beyond inbound limit")
}

func TestResourceManagerServerProfilePeerConnsInbound(t *testing.T) {
	rm, err := buildResourceManager(ResourceProfileServer)
	require.NoError(t, err)
	defer rm.Close()

	peerID := peer.ID("peer0")

	// ServerProfile limits connections per peer to 8
	conns := make([]network.ConnManagementScope, 8)
	defer func() {
		for _, c := range conns {
			c.Done()
		}
	}()

	for i := range conns {
		// distinct IPs to avoid per IP limit
		addr := multiaddr.StringCast(fmt.Sprintf("/ip4/1.2.3.%d/tcp/1234", i))
		conns[i], err = rm.OpenConnection(network.DirInbound, false, addr)
		require.NoError(t, err)
		err = conns[i].SetPeer(peerID)
		require.NoError(t, err)
	}
	addr := multiaddr.StringCast("/ip4/1.2.4.1/tcp/1234")
	c, err := rm.OpenConnection(network.DirInbound, false, addr)
	require.NoError(t, err)
	defer func() {
		c.Done()
	}()

	// Assigning connection to this peer will be rejected
	err = c.SetPeer(peerID)
	assert.Error(t, err, "should reject connection beyond per-peer inbound limit")
}
