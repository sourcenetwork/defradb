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
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/internal/kms"
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
	disableP2P        bool
	disableAPI        bool
	enableDevelopment bool
	kmsType           immutable.Option[kms.ServiceType]
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

func WithKMS(kms kms.ServiceType) NodeOpt {
	return func(o *Options) {
		o.kmsType = immutable.Some(kms)
	}
}

// WithEnableDevelopment sets the enable development mode flag.
func WithEnableDevelopment(enable bool) NodeOpt {
	return func(o *Options) {
		o.enableDevelopment = enable
	}
}

// Node is a DefraDB instance with optional sub-systems.
type Node struct {
	DB         client.DB
	Peer       *net.Peer
	Server     *http.Server
	kmsService kms.Service

	options    *Options
	dbOpts     []db.Option
	acpOpts    []ACPOpt
	netOpts    []net.NodeOpt
	storeOpts  []StoreOpt
	serverOpts []http.ServerOpt
	lensOpts   []LenOpt
}

// New returns a new node instance configured with the given options.
func New(ctx context.Context, opts ...Option) (*Node, error) {
	n := Node{
		options: DefaultOptions(),
	}
	for _, opt := range opts {
		switch t := opt.(type) {
		case NodeOpt:
			t(n.options)

		case ACPOpt:
			n.acpOpts = append(n.acpOpts, t)

		case StoreOpt:
			n.storeOpts = append(n.storeOpts, t)

		case db.Option:
			n.dbOpts = append(n.dbOpts, t)

		case http.ServerOpt:
			n.serverOpts = append(n.serverOpts, t)

		case net.NodeOpt:
			n.netOpts = append(n.netOpts, t)

		case LenOpt:
			n.lensOpts = append(n.lensOpts, t)
		}
	}
	return &n, nil
}

// Start starts the node sub-systems.
func (n *Node) Start(ctx context.Context) error {
	rootstore, err := NewStore(ctx, n.storeOpts...)
	if err != nil {
		return err
	}
	acp, err := NewACP(ctx, n.acpOpts...)
	if err != nil {
		return err
	}
	lens, err := NewLens(ctx, n.lensOpts...)
	if err != nil {
		return err
	}
	n.DB, err = db.NewDB(ctx, rootstore, acp, lens, n.dbOpts...)
	if err != nil {
		return err
	}

	if !n.options.disableP2P {
		// setup net node
		n.Peer, err = net.NewPeer(ctx, n.DB.Blockstore(), n.DB.Encstore(), n.DB.Events(), n.netOpts...)
		if err != nil {
			return err
		}
		if n.options.kmsType.HasValue() {
			switch n.options.kmsType.Value() {
			case kms.PubSubServiceType:
				n.kmsService, err = kms.NewPubSubService(
					ctx,
					n.Peer.PeerID(),
					n.Peer.Server(),
					n.DB.Events(),
					n.DB.Encstore(),
				)
			}
			if err != nil {
				return err
			}
		}
	}

	if !n.options.disableAPI {
		// setup http server
		handler, err := http.NewHandler(n.DB)
		if err != nil {
			return err
		}
		n.Server, err = http.NewServer(handler, n.serverOpts...)
		if err != nil {
			return err
		}
		err = n.Server.SetListener()
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

// PurgeAndRestart causes the node to shutdown, purge all data from
// its datastore, and restart.
func (n *Node) PurgeAndRestart(ctx context.Context) error {
	if !n.options.enableDevelopment {
		return ErrPurgeWithDevModeDisabled
	}
	err := n.Close(ctx)
	if err != nil {
		return err
	}
	err = purgeStore(ctx, n.storeOpts...)
	if err != nil {
		return err
	}
	return n.Start(ctx)
}
