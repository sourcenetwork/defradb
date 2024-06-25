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

	"github.com/libp2p/go-libp2p/core/peer"
	badger "github.com/sourcenetwork/badger/v4"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/client"
	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v4"
	"github.com/sourcenetwork/defradb/datastore/memory"
	"github.com/sourcenetwork/defradb/internal/db"
	netutils "github.com/sourcenetwork/defradb/net/utils"
)

// Node.Boostrap is not tested because the underlying, *ipfslite.Peer.Bootstrap is a best-effort function.

func FixtureNewMemoryDBWithBroadcaster(t *testing.T) client.DB {
	var database client.DB
	ctx := context.Background()
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	require.NoError(t, err)
	database, err = db.NewDB(ctx, rootstore, acp.NoACP, nil)
	require.NoError(t, err)
	return database
}

func TestNewPeer_WithEnableRelay_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, acp.NoACP, nil)
	require.NoError(t, err)
	defer db.Close()
	n, err := NewPeer(
		context.Background(),
		db.Root(),
		db.Blockstore(),
		db.Events(),
		WithEnableRelay(true),
	)
	require.NoError(t, err)
	n.Close()
}

func TestNewPeer_WithDBClosed_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)

	db, err := db.NewDB(ctx, store, acp.NoACP, nil)
	require.NoError(t, err)
	db.Close()

	_, err = NewPeer(
		context.Background(),
		db.Root(),
		db.Blockstore(),
		db.Events(),
	)
	require.ErrorContains(t, err, "datastore closed")
}

func TestNewPeer_NoPubSub_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, acp.NoACP, nil)
	require.NoError(t, err)
	defer db.Close()

	n, err := NewPeer(
		context.Background(),
		db.Root(),
		db.Blockstore(),
		db.Events(),
		WithEnablePubSub(false),
	)
	require.NoError(t, err)
	require.Nil(t, n.ps)
	n.Close()
}

func TestNewPeer_WithEnablePubSub_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, acp.NoACP, nil)
	require.NoError(t, err)
	defer db.Close()

	n, err := NewPeer(
		ctx,
		db.Root(),
		db.Blockstore(),
		db.Events(),
		WithEnablePubSub(true),
	)

	require.NoError(t, err)
	// overly simple check of validity of pubsub, avoiding the process of creating a PubSub
	require.NotNil(t, n.ps)
	n.Close()
}

func TestNodeClose_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, acp.NoACP, nil)
	require.NoError(t, err)
	defer db.Close()
	n, err := NewPeer(
		context.Background(),
		db.Root(),
		db.Blockstore(),
		db.Events(),
	)
	require.NoError(t, err)
	n.Close()
}

func TestNewPeer_BootstrapWithNoPeer_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, acp.NoACP, nil)
	require.NoError(t, err)
	defer db.Close()

	n1, err := NewPeer(
		ctx,
		db.Root(),
		db.Blockstore(),
		db.Events(),
		WithListenAddresses("/ip4/0.0.0.0/tcp/0"),
	)
	require.NoError(t, err)
	n1.Bootstrap([]peer.AddrInfo{})
	n1.Close()
}

func TestNewPeer_BootstrapWithOnePeer_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, acp.NoACP, nil)
	require.NoError(t, err)
	defer db.Close()
	n1, err := NewPeer(
		ctx,
		db.Root(),
		db.Blockstore(),
		db.Events(),
		WithListenAddresses("/ip4/0.0.0.0/tcp/0"),
	)
	require.NoError(t, err)
	defer n1.Close()
	n2, err := NewPeer(
		ctx,
		db.Root(),
		db.Blockstore(),
		db.Events(),
		WithListenAddresses("/ip4/0.0.0.0/tcp/0"),
	)
	require.NoError(t, err)
	defer n2.Close()
	addrs, err := netutils.ParsePeers([]string{n1.host.Addrs()[0].String() + "/p2p/" + n1.PeerID().String()})
	if err != nil {
		t.Fatal(err)
	}
	n2.Bootstrap(addrs)
}

func TestNewPeer_BootstrapWithOneValidPeerAndManyInvalidPeers_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, acp.NoACP, nil)
	require.NoError(t, err)
	defer db.Close()

	n1, err := NewPeer(
		ctx,
		db.Root(),
		db.Blockstore(),
		db.Events(),
		WithListenAddresses("/ip4/0.0.0.0/tcp/0"),
	)
	require.NoError(t, err)
	defer n1.Close()
	n2, err := NewPeer(
		ctx,
		db.Root(),
		db.Blockstore(),
		db.Events(),
		WithListenAddresses("/ip4/0.0.0.0/tcp/0"),
	)
	require.NoError(t, err)
	defer n2.Close()
	addrs, err := netutils.ParsePeers([]string{
		n1.host.Addrs()[0].String() + "/p2p/" + n1.PeerID().String(),
		"/ip4/0.0.0.0/tcp/1234/p2p/" + "12D3KooWC8YY6Tx3uAeHsdBmoy7PJPwqXAHE4HkCZ5veankKWci6",
		"/ip4/0.0.0.0/tcp/1235/p2p/" + "12D3KooWC8YY6Tx3uAeHsdBmoy7PJPwqXAHE4HkCZ5veankKWci5",
		"/ip4/0.0.0.0/tcp/1236/p2p/" + "12D3KooWC8YY6Tx3uAeHsdBmoy7PJPwqXAHE4HkCZ5veankKWci4",
	})
	require.NoError(t, err)
	n2.Bootstrap(addrs)
}

func TestListenAddrs_WithListenAddresses_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, acp.NoACP, nil)
	require.NoError(t, err)
	defer db.Close()

	n, err := NewPeer(
		context.Background(),
		db.Root(),
		db.Blockstore(),
		db.Events(),
		WithListenAddresses("/ip4/0.0.0.0/tcp/0"),
	)
	require.NoError(t, err)
	require.Contains(t, n.ListenAddrs()[0].String(), "/tcp/")
	n.Close()
}
