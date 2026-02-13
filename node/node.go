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
	"time"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/dac"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/internal/db"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
	intOpts "github.com/sourcenetwork/defradb/internal/options"
	"github.com/sourcenetwork/defradb/internal/utils"
)

var log = corelog.NewLogger("node")

// Peer defines the minimal p2p network interface.
type Peer interface {
	client.Host
	Close()
}

type DB interface {
	client.TxnStore
	MaxTxnRetries() int
	Rootstore() corekv.TxnStore
	Events() event.Bus
	NodeACP() acpDB.NACInfo
	DocumentACP() immutable.Option[dac.DocumentACP]
	PurgeDACState(ctx context.Context) error
	PurgeNACState(ctx context.Context) error
	GetNodeIdentityToken(ctx context.Context, audience immutable.Option[string]) ([]byte, error)
	Close()
}

// Node is a DefraDB instance with optional sub-systems.
type Node struct {
	// DB is the database instance
	DB DB
	// Peer is the p2p networking subsystem instance
	peer Peer
	// api http server instance
	server *http.Server
	// opts is the resolved options
	opts *options.NodeOptions
	// the URL the API is served at.
	APIURL string
}

// DefaultNodeOptions returns default NodeOptions values.
func DefaultNodeOptions() options.NodeOptions {
	return options.NodeOptions{
		DisableP2P:        false,
		DisableAPI:        false,
		EnableDevelopment: false,
		Store: options.NodeStoreOptions{
			Store:          options.NodeDefaultStore,
			BadgerInMemory: false,
			BadgerFileSize: 1 << 30, // 1GB
		},
		DocumentACP: options.NodeDocumentACPOptions{
			DocumentACPType: options.NodeLocalDocumentACPType,
		},
		NodeACP: options.NodeACPOptions{
			IsEnabled: false,
		},
		DB: options.NodeDBOptions{
			MaxTxnRetries: immutable.Some(5),
			EnableSigning: true,
			RetryIntervals: []time.Duration{
				time.Second * 30,
				time.Minute,
				time.Minute * 2,
				time.Minute * 4,
				time.Minute * 8,
				time.Minute * 16,
				time.Minute * 32,
			},
			P2PBlockSyncTimeout: time.Second * 5,
			LensRuntime:         options.NodeDefaultLensRuntime,
		},
		P2P:  options.NodeP2POptions{},
		HTTP: options.NodeHTTPOptions{},
	}
}

// New returns a new node instance configured with the given options.
func New(ctx context.Context, opts ...options.Lister[options.NodeOptions]) (*Node, error) {
	nodeOpts := DefaultNodeOptions()
	utils.ApplyOptions(&nodeOpts, opts...)
	n := Node{
		opts: &nodeOpts,
	}
	return &n, nil
}

// Start starts the node sub-systems.
func (n *Node) Start(ctx context.Context) error {
	rootstore, isValueSizeLimited, err := NewStore(ctx, options.NodeStore().SetAll(n.opts.Store))
	if err != nil {
		return err
	}
	documentACP, err := NewDocumentACP(ctx, &n.opts.DocumentACP)
	if err != nil {
		return err
	}

	nodeACP, err := NewNodeACP(ctx, &n.opts.NodeACP)
	if err != nil {
		return err
	}

	if isValueSizeLimited {
		n.opts.DB.ChunkSize = immutable.Some(defaultChunkSize)
	}

	err = n.startP2P(ctx, rootstore, n.opts.DB.ChunkSize)
	if err != nil {
		return err
	}

	dbBuilder := intOpts.DB().SetNodeDBOptions(n.opts.DB)
	if documentACP.HasValue() {
		dbBuilder.SetDocumentACP(documentACP.Value())
	}
	if n.peer != nil {
		dbBuilder.SetP2P(n.peer)
	}

	n.DB, err = db.NewDB(ctx, rootstore, nodeACP, dbBuilder)
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
	if n.peer != nil {
		n.peer.Close()
	}
	if n.DB != nil {
		n.DB.Close()
	}
	return err
}

// PurgeAndRestart causes the node to shutdown, purge all data from
// its datastore, and restart.
func (n *Node) PurgeAndRestart(ctx context.Context) error {
	if !n.opts.EnableDevelopment {
		return ErrPurgeWithDevModeDisabled
	}

	// This will purge document acp state.
	err := n.DB.PurgeDACState(ctx)
	if err != nil {
		return err
	}

	// This will purge node acp state.
	err = n.DB.PurgeNACState(ctx)
	if err != nil {
		return err
	}

	// This will close db and all acp instances along with it.
	err = n.Close(ctx)
	if err != nil {
		return err
	}

	err = purgeStore(ctx, &n.opts.Store)
	if err != nil {
		return err
	}

	// The node is being started again. This restarts the above closed acp states too.
	return n.Start(ctx)
}
