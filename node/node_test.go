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
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/net"
)

func TestWithStoreOpts(t *testing.T) {
	storeOpts := []StoreOpt{WithPath("test")}

	options := &Options{}
	WithStoreOpts(storeOpts...)(options)
	assert.Equal(t, storeOpts, options.storeOpts)
}

func TestWithDatabaseOpts(t *testing.T) {
	dbOpts := []db.Option{db.WithMaxRetries(10)}

	options := &Options{}
	WithDatabaseOpts(dbOpts...)(options)
	assert.Equal(t, dbOpts, options.dbOpts)
}

func TestWithNetOpts(t *testing.T) {
	netOpts := []net.NodeOpt{net.WithEnablePubSub(true)}

	options := &Options{}
	WithNetOpts(netOpts...)(options)
	assert.Equal(t, netOpts, options.netOpts)
}

func TestWithServerOpts(t *testing.T) {
	serverOpts := []http.ServerOpt{http.WithAddress("127.0.0.1:8080")}

	options := &Options{}
	WithServerOpts(serverOpts...)(options)
	assert.Equal(t, serverOpts, options.serverOpts)
}

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

func TestNodeStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := []NodeOpt{
		WithStoreOpts(WithPath(t.TempDir())),
		WithDatabaseOpts(db.WithUpdateEvents()),
	}

	node, err := NewNode(ctx, opts...)
	require.NoError(t, err)

	err = node.Start(ctx)
	require.NoError(t, err)

	<-time.After(5 * time.Second)

	err = node.Close(ctx)
	require.NoError(t, err)
}
