// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package net

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	netutils "github.com/sourcenetwork/defradb/net/utils"
)

func TestDial_WithConnectedPeer_NoError(t *testing.T) {
	db := FixtureNewMemoryDBWithBroadcaster(t)
	ctx := context.Background()
	n1, err := NewNode(
		ctx,
		db,
		WithListenP2PAddrStrings("/ip4/0.0.0.0/tcp/0"),
		// WithDataPath() is a required option with the current implementation of key management
		WithDataPath(t.TempDir()),
	)
	assert.NoError(t, err)
	n2, err := NewNode(
		ctx,
		db,
		WithListenP2PAddrStrings("/ip4/0.0.0.0/tcp/0"),
		// WithDataPath() is a required option with the current implementation of key management
		WithDataPath(t.TempDir()),
	)
	assert.NoError(t, err)
	addrs, err := netutils.ParsePeers([]string{n1.host.Addrs()[0].String() + "/p2p/" + n1.PeerID().String()})
	if err != nil {
		t.Fatal(err)
	}
	n2.Bootstrap(addrs)
	_, err = n1.server.dial(n2.PeerID())
	require.NoError(t, err)
}

func TestDial_WithConnectedPeerAndSecondConnection_NoError(t *testing.T) {
	db := FixtureNewMemoryDBWithBroadcaster(t)
	ctx := context.Background()
	n1, err := NewNode(
		ctx,
		db,
		WithListenP2PAddrStrings("/ip4/0.0.0.0/tcp/0"),
		// WithDataPath() is a required option with the current implementation of key management
		WithDataPath(t.TempDir()),
	)
	assert.NoError(t, err)
	n2, err := NewNode(
		ctx,
		db,
		WithListenP2PAddrStrings("/ip4/0.0.0.0/tcp/0"),
		// WithDataPath() is a required option with the current implementation of key management
		WithDataPath(t.TempDir()),
	)
	assert.NoError(t, err)
	addrs, err := netutils.ParsePeers([]string{n1.host.Addrs()[0].String() + "/p2p/" + n1.PeerID().String()})
	if err != nil {
		t.Fatal(err)
	}
	n2.Bootstrap(addrs)
	_, err = n1.server.dial(n2.PeerID())
	require.NoError(t, err)

	_, err = n1.server.dial(n2.PeerID())
	require.NoError(t, err)
}

func TestDial_WithConnectedPeerAndSecondConnectionWithConnectionShutdown_ClosingConnectionError(t *testing.T) {
	db := FixtureNewMemoryDBWithBroadcaster(t)
	ctx := context.Background()
	n1, err := NewNode(
		ctx,
		db,
		WithListenP2PAddrStrings("/ip4/0.0.0.0/tcp/0"),
		// WithDataPath() is a required option with the current implementation of key management
		WithDataPath(t.TempDir()),
	)
	assert.NoError(t, err)
	n2, err := NewNode(
		ctx,
		db,
		WithListenP2PAddrStrings("/ip4/0.0.0.0/tcp/0"),
		// WithDataPath() is a required option with the current implementation of key management
		WithDataPath(t.TempDir()),
	)
	assert.NoError(t, err)
	addrs, err := netutils.ParsePeers([]string{n1.host.Addrs()[0].String() + "/p2p/" + n1.PeerID().String()})
	if err != nil {
		t.Fatal(err)
	}
	n2.Bootstrap(addrs)
	_, err = n1.server.dial(n2.PeerID())
	require.NoError(t, err)

	err = n1.server.conns[n2.PeerID()].Close()
	require.NoError(t, err)

	_, err = n1.server.dial(n2.PeerID())
	require.Contains(t, err.Error(), "grpc: the client connection is closing")
}
