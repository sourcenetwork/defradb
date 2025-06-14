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

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/internal/kms"
)

var log = corelog.NewLogger("node")

// Peer defines the minimal p2p network interface.
type Peer interface {
	Close()
	PeerID() peer.ID
	PeerInfo() peer.AddrInfo
	Connect(context.Context, peer.AddrInfo) error
}

// Node is a DefraDB instance with optional sub-systems.
type Node struct {
	// DB is the database instance
	DB client.DB
	// Peer is the p2p networking subsystem instance
	Peer Peer
	// api http server instance
	server *http.Server
	// kms subsystem instance
	kmsService kms.Service
	// config values after applying options
	config *Config
	// options the node was created with
	options []Option
	// the URL the API is served at.
	APIURL string
}

// New returns a new node instance configured with the given options.
func New(ctx context.Context, options ...Option) (*Node, error) {
	n := Node{
		config:  DefaultConfig(),
		options: options,
	}
	for _, opt := range filterOptions[NodeOpt](options) {
		opt(n.config)
	}
	return &n, nil
}

// Start starts the node sub-systems.
func (n *Node) Start(ctx context.Context) error {
	rootstore, err := NewStore(ctx, filterOptions[StoreOpt](n.options)...)
	if err != nil {
		return err
	}
	lens, err := NewLens(ctx, filterOptions[LenOpt](n.options)...)
	if err != nil {
		return err
	}
	documentACP, err := NewDocumentACP(ctx, filterOptions[DocumentACPOpt](n.options)...)
	if err != nil {
		return err
	}
	n.DB, err = db.NewDB(ctx, rootstore, documentACP, lens, filterOptions[db.Option](n.options)...)
	if err != nil {
		return err
	}
	err = n.startP2P(ctx)
	if err != nil {
		return err
	}
	return n.startAPI(ctx)
}

// Close stops the node sub-systems.
func (n *Node) Close(ctx context.Context) error {
	var err error
	if n.server != nil {
		err = n.server.Shutdown(ctx)
	}
	if n.Peer != nil {
		n.Peer.Close()
	}
	if n.DB != nil {
		n.DB.Close()
	}
	return err
}

// PurgeAndRestart causes the node to shutdown, purge all data from
// its datastore, and restart.
func (n *Node) PurgeAndRestart(ctx context.Context) error {
	if !n.config.enableDevelopment {
		return ErrPurgeWithDevModeDisabled
	}
	err := n.Close(ctx)
	if err != nil {
		return err
	}
	err = purgeStore(ctx, filterOptions[StoreOpt](n.options)...)
	if err != nil {
		return err
	}

	coreDB, _ := n.DB.(*db.DB)

	// This will purge state.
	// They will be restarted when node is started again.
	err = coreDB.PurgeACPState(ctx)
	if err != nil {
		return err
	}

	return n.Start(ctx)
}
