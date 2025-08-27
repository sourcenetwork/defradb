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

	"github.com/sourcenetwork/corekv/memory"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/net/config"
)

func TestNewPeer_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	p, err := NewPeer(ctx, datastore.BlockstoreFrom(store))
	require.NoError(t, err)
	p.Close()
}

func TestStart_WithKnownPeer_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	n1, err := NewPeer(
		ctx,
		datastore.BlockstoreFrom(store),
		config.WithListenAddresses("/ip4/127.0.0.1/tcp/0"),
	)
	require.NoError(t, err)
	defer n1.Close()
	n2, err := NewPeer(
		ctx,
		datastore.BlockstoreFrom(store),
		config.WithListenAddresses("/ip4/127.0.0.1/tcp/0"),
	)
	require.NoError(t, err)
	defer n2.Close()

	err = n2.Connect(ctx, n1.PeerInfo())
	require.NoError(t, err)
}

func TestNewPeer_WithEnableRelay_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	n, err := NewPeer(
		context.Background(),
		datastore.BlockstoreFrom(store),
		config.WithEnableRelay(true),
	)
	require.NoError(t, err)
	n.Close()
}

func TestNewPeer_NoPubSub_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	n, err := NewPeer(
		context.Background(),
		datastore.BlockstoreFrom(store),
		config.WithEnablePubSub(false),
	)
	require.NoError(t, err)
	require.Nil(t, n.ps)
	n.Close()
}

func TestNewPeer_WithEnablePubSub_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	n, err := NewPeer(
		ctx,
		datastore.BlockstoreFrom(store),
		config.WithEnablePubSub(true),
	)
	require.NoError(t, err)
	// overly simple check of validity of pubsub, avoiding the process of creating a PubSub
	require.NotNil(t, n.ps)
	n.Close()
}

func TestNodeClose_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	n, err := NewPeer(
		context.Background(),
		datastore.BlockstoreFrom(store),
	)
	require.NoError(t, err)
	n.Close()
}

func TestListenAddrs_WithListenAddresses_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	n, err := NewPeer(
		context.Background(),
		datastore.BlockstoreFrom(store),
		config.WithListenAddresses("/ip4/127.0.0.1/tcp/0"),
	)
	require.NoError(t, err)
	require.Contains(t, n.Addrs()[0], "/tcp/")
	n.Close()
}

func TestPeer_WithBootstrapPeers_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	n, err := NewPeer(
		context.Background(),
		datastore.BlockstoreFrom(store),
		config.WithBootstrapPeers("/ip4/127.0.0.1/tcp/6666/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ"),
	)
	require.NoError(t, err)

	n.Close()
}
