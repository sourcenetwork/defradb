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
	"errors"
	"fmt"
	gohttp "net/http"

	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/net"
)

var log = corelog.NewLogger("node")

// Option is a generic option that applies to any subsystem.
//
// Invalid option types will be silently ignored. Valid option types are:
// - `ACPOpt`
// - `NodeOpt`
// - `StoreOpt`
// - `db.Option`
// - `http.ServerOpt`
// - `net.NodeOpt`
type Option any

// Options contains start configuration values.
type Options struct {
	disableP2P bool
	disableAPI bool
}

// DefaultOptions returns options with default settings.
func DefaultOptions() *Options {
	return &Options{}
}

// NodeOpt is a function for setting configuration values.
type NodeOpt func(*Options)

// WithDisableP2P sets the disable p2p flag.
func WithDisableP2P(disable bool) NodeOpt {
	return func(o *Options) {
		o.disableP2P = disable
	}
}

// WithDisableAPI sets the disable api flag.
func WithDisableAPI(disable bool) NodeOpt {
	return func(o *Options) {
		o.disableAPI = disable
	}
}

// Node is a DefraDB instance with optional sub-systems.
type Node struct {
	DB     client.DB
	Peer   *net.Peer
	Server *http.Server
}

// NewNode returns a new node instance configured with the given options.
func NewNode(ctx context.Context, opts ...Option) (*Node, error) {
	var (
		dbOpts     []db.Option
		acpOpts    []ACPOpt
		netOpts    []net.NodeOpt
		storeOpts  []StoreOpt
		serverOpts []http.ServerOpt
		lensOpts   []LenOpt
	)

	options := DefaultOptions()
	for _, opt := range opts {
		switch t := opt.(type) {
		case ACPOpt:
			acpOpts = append(acpOpts, t)

		case NodeOpt:
			t(options)

		case StoreOpt:
			storeOpts = append(storeOpts, t)

		case db.Option:
			dbOpts = append(dbOpts, t)

		case http.ServerOpt:
			serverOpts = append(serverOpts, t)

		case net.NodeOpt:
			netOpts = append(netOpts, t)

		case LenOpt:
			lensOpts = append(lensOpts, t)
		}
	}

	rootstore, err := NewStore(ctx, storeOpts...)
	if err != nil {
		return nil, err
	}

	acp, err := NewACP(ctx, acpOpts...)
	if err != nil {
		return nil, err
	}

	lens, err := NewLens(ctx, lensOpts...)
	if err != nil {
		return nil, err
	}

	db, err := db.NewDB(ctx, rootstore, acp, lens, dbOpts...)
	if err != nil {
		return nil, err
	}

	var peer *net.Peer
	if !options.disableP2P {
		// setup net node
		peer, err = net.NewPeer(ctx, db.Blockstore(), db.Events(), netOpts...)
		if err != nil {
			return nil, err
		}
	}

	var server *http.Server
	if !options.disableAPI {
		// setup http server
		handler, err := http.NewHandler(db)
		if err != nil {
			return nil, err
		}
		server, err = http.NewServer(handler, serverOpts...)
		if err != nil {
			return nil, err
		}
	}

	return &Node{
		DB:     db,
		Peer:   peer,
		Server: server,
	}, nil
}

// Start starts the node sub-systems.
func (n *Node) Start(ctx context.Context) error {
	if n.Server != nil {
		err := n.Server.SetListener()
		if err != nil {
			return err
		}
		log.InfoContext(ctx,
			fmt.Sprintf("Providing HTTP API at %s PlaygroundEnabled=%t", n.Server.Address(), http.PlaygroundEnabled))
		log.InfoContext(ctx, fmt.Sprintf("Providing GraphQL endpoint at %s/v0/graphql", n.Server.Address()))
		go func() {
			if err := n.Server.Serve(); err != nil && !errors.Is(err, gohttp.ErrServerClosed) {
				log.ErrorContextE(ctx, "HTTP server stopped", err)
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
	if n.Peer != nil {
		n.Peer.Close()
	}
	if n.DB != nil {
		n.DB.Close()
	}
	return err
}
