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
)

func TestDial_WithConnectedPeer_NoError(t *testing.T) {
	db1 := FixtureNewMemoryDBWithBroadcaster(t)
	db2 := FixtureNewMemoryDBWithBroadcaster(t)
	defer db1.Close()
	defer db2.Close()
	ctx := context.Background()
	n1, err := NewPeer(
		ctx,
		db1.Blockstore(),
		db1.Events(),
		WithListenAddresses("/ip4/127.0.0.1/tcp/0"),
	)
	assert.NoError(t, err)
	defer n1.Close()
	n2, err := NewPeer(
		ctx,
		db2.Blockstore(),
		db2.Events(),
		WithListenAddresses("/ip4/127.0.0.1/tcp/0"),
	)
	assert.NoError(t, err)
	defer n2.Close()

	err = n2.Connect(ctx, n1.PeerInfo())
	require.NoError(t, err)

	_, err = n1.server.dial(n2.PeerID())
	require.NoError(t, err)
}

func TestDial_WithConnectedPeerAndSecondConnection_NoError(t *testing.T) {
	db1 := FixtureNewMemoryDBWithBroadcaster(t)
	db2 := FixtureNewMemoryDBWithBroadcaster(t)
	defer db1.Close()
	defer db2.Close()
	ctx := context.Background()
	n1, err := NewPeer(
		ctx,
		db1.Blockstore(),
		db1.Events(),
		WithListenAddresses("/ip4/127.0.0.1/tcp/0"),
	)
	assert.NoError(t, err)
	defer n1.Close()
	n2, err := NewPeer(
		ctx,
		db2.Blockstore(),
		db2.Events(),
		WithListenAddresses("/ip4/127.0.0.1/tcp/0"),
	)
	assert.NoError(t, err)
	defer n2.Close()

	err = n2.Connect(ctx, n1.PeerInfo())
	require.NoError(t, err)

	_, err = n1.server.dial(n2.PeerID())
	require.NoError(t, err)

	_, err = n1.server.dial(n2.PeerID())
	require.NoError(t, err)
}

func TestDial_WithConnectedPeerAndSecondConnectionWithConnectionShutdown_ClosingConnectionError(t *testing.T) {
	db1 := FixtureNewMemoryDBWithBroadcaster(t)
	db2 := FixtureNewMemoryDBWithBroadcaster(t)
	defer db1.Close()
	defer db2.Close()
	ctx := context.Background()
	n1, err := NewPeer(
		ctx,
		db1.Blockstore(),
		db1.Events(),
		WithListenAddresses("/ip4/127.0.0.1/tcp/0"),
	)
	assert.NoError(t, err)
	defer n1.Close()
	n2, err := NewPeer(
		ctx,
		db2.Blockstore(),
		db2.Events(),
		WithListenAddresses("/ip4/127.0.0.1/tcp/0"),
	)
	assert.NoError(t, err)
	defer n2.Close()

	err = n2.Connect(ctx, n1.PeerInfo())
	require.NoError(t, err)

	_, err = n1.server.dial(n2.PeerID())
	require.NoError(t, err)

	err = n1.server.conns[n2.PeerID()].Close()
	require.NoError(t, err)

	_, err = n1.server.dial(n2.PeerID())
	require.Contains(t, err.Error(), "grpc: the client connection is closing")
}
