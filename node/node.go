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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/sourcenetwork/defradb/net"
)

var log = logging.MustNewLogger("node")

// Options contains start configuration values.
type Options struct {
	storeOpts  []StoreOpt
	dbOpts     []db.Option
	netOpts    []net.NodeOpt
	serverOpts []http.ServerOpt
	peers      []peer.AddrInfo
	disableP2P bool
	disableAPI bool
}

// DefaultOptions returns options with default settings.
func DefaultOptions() *Options {
	return &Options{}
}

// Opt is a function for setting configuration values.
type Opt func(*Options)

// WithStoreOpts sets the store options.
func WithStoreOpts(opts ...StoreOpt) Opt {
	return func(o *Options) {
		o.storeOpts = opts
	}
}

// WithDatabaseOpts sets the database options.
func WithDatabaseOpts(opts ...db.Option) Opt {
	return func(o *Options) {
		o.dbOpts = opts
	}
}

// WithNetOpts sets the net / p2p options.
func WithNetOpts(opts ...net.NodeOpt) Opt {
	return func(o *Options) {
		o.netOpts = opts
	}
}

// WithServerOpts sets the api server options.
func WithServerOpts(opts ...http.ServerOpt) Opt {
	return func(o *Options) {
		o.serverOpts = opts
	}
}

// WithDisableP2P sets the disable p2p flag.
func WithDisableP2P(disable bool) Opt {
	return func(o *Options) {
		o.disableP2P = disable
	}
}

// WithDisableAPI sets the disable api flag.
func WithDisableAPI(disable bool) Opt {
	return func(o *Options) {
		o.disableAPI = disable
	}
}

// WithPeers sets the bootstrap peers.
func WithPeers(peers ...peer.AddrInfo) Opt {
	return func(o *Options) {
		o.peers = peers
	}
}

// Node is a DefraDB instance with optional sub-systems.
type Node struct {
	DB     client.DB
	Node   *net.Node
	Server *http.Server
}

// New returns a new node instance configured with the given options.
func New(ctx context.Context, opts ...Opt) (*Node, error) {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	rootstore, err := NewStore(options.storeOpts...)
	if err != nil {
		return nil, err
	}
	db, err := db.NewDB(ctx, rootstore, options.dbOpts...)
	if err != nil {
		return nil, err
	}

	var node *net.Node
	if !options.disableP2P {
		// setup net node
		node, err = net.NewNode(ctx, db, options.netOpts...)
		if err != nil {
			return nil, err
		}
		if len(options.peers) > 0 {
			node.Bootstrap(options.peers)
		}
	}

	var server *http.Server
	if !options.disableAPI {
		// setup http server
		var handler *http.Handler
		if node != nil {
			handler, err = http.NewHandler(node)
		} else {
			handler, err = http.NewHandler(db)
		}
		if err != nil {
			return nil, err
		}
		server, err = http.NewServer(handler, options.serverOpts...)
		if err != nil {
			return nil, err
		}
	}

	return &Node{
		DB:     db,
		Node:   node,
		Server: server,
	}, nil
}

// Start starts the node sub-systems.
func (n *Node) Start(ctx context.Context) error {
	if n.Node != nil {
		if err := n.Node.Start(); err != nil {
			return err
		}
	}
	if n.Server != nil {
		go func() {
			if err := n.Server.ListenAndServe(); err != nil {
				log.FeedbackErrorE(ctx, "HTTP server stopped", err)
			}
		}()
	}
	return nil
}

// Close stops the node sub-systems.
func (n *Node) Close(ctx context.Context) error {
	var err error
	if n.Server != nil {
		err = n.Server.Shutdown(ctx)
	}
	if n.Node != nil {
		n.Node.Close()
	} else {
		n.DB.Close()
	}
	return err
}
