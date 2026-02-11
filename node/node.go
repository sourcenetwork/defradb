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

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"
	lensNode "github.com/sourcenetwork/lens/host-go/node"

	"github.com/sourcenetwork/defradb/acp/dac"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/internal/db"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
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

// New returns a new node instance configured with the given options.
func New(ctx context.Context, opts ...options.Lister[options.NodeOptions]) (*Node, error) {
	nodeOpts := utils.NewOptions(opts...)
	n := Node{
		opts: nodeOpts,
	}
	return &n, nil
}

// Start starts the node sub-systems.
func (n *Node) Start(ctx context.Context) error {
	rootstore, isValueSizeLimited, err := NewStore(ctx, &n.opts.Store)
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

	dbOpts := n.buildDBOptions()

	if isValueSizeLimited {
		dbOpts = append(dbOpts,
			db.WithLensOpts(
				lensNode.WithBlockstoreChunkSize(defaultChunkSize),
			),
		)
		dbOpts = append(dbOpts,
			db.WithBlockStoreChunkSize(defaultChunkSize),
		)
		n.opts.DB.ChunkSize = immutable.Some(defaultChunkSize)
	}

	err = n.startP2P(ctx, rootstore, n.opts.DB.ChunkSize)
	if err != nil {
		return err
	}

	dbOpts = append(dbOpts, n.buildP2PDBOption()...)

	n.DB, err = db.NewDB(ctx, rootstore, nodeACP, documentACP, dbOpts...)
	if err != nil {
		return err
	}

	return n.startAPI(ctx)
}

// buildDBOptions converts NodeOptions into db.Option slice.
func (n *Node) buildDBOptions() []db.Option {
	var opts []db.Option

	if n.opts.DB.MaxTxnRetries.HasValue() {
		opts = append(opts, db.WithMaxRetries(n.opts.DB.MaxTxnRetries.Value()))
	}
	if n.opts.DB.Identity.HasValue() {
		opts = append(opts, db.WithNodeIdentity(n.opts.DB.Identity.Value()))
	}
	opts = append(opts, db.WithEnabledSigning(n.opts.DB.EnableSigning))
	if len(n.opts.DB.SearchableEncryptionKey) > 0 {
		opts = append(opts, db.WithSearchableEncryptionKey(n.opts.DB.SearchableEncryptionKey))
	}
	if len(n.opts.DB.RetryIntervals) > 0 {
		opts = append(opts, db.WithRetryInterval(n.opts.DB.RetryIntervals))
	}
	if n.opts.DB.P2PBlockSyncTimeout > 0 {
		opts = append(opts, db.WithP2PBlockSyncTimeout(n.opts.DB.P2PBlockSyncTimeout))
	}
	if n.opts.DB.LensRuntime != "" {
		opts = append(opts, db.WithLensRuntime(db.LensRuntimeType(n.opts.DB.LensRuntime)))
	}
	if n.opts.DB.ChunkSize.HasValue() {
		opts = append(opts, db.WithBlockStoreChunkSize(n.opts.DB.ChunkSize.Value()))
	}

	return opts
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
