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
	// config values after applying options
	config *Config
	// opts contains the unified node options
	opts *options.NodeOptions
	// the URL the API is served at.
	APIURL string
}

// New returns a new node instance configured with the given options.
// If no options are provided, default options are used.
func New(ctx context.Context, opts ...*options.NodeOptions) (*Node, error) {
	var nodeOpts *options.NodeOptions
	if len(opts) > 0 && opts[0] != nil {
		nodeOpts = opts[0]
	} else {
		nodeOpts = options.Node()
	}
	n := Node{
		config: DefaultConfig(),
		opts:   nodeOpts,
	}
	n.config.applyNodeOptions(nodeOpts)
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

	var chunkSize immutable.Option[int]
	var lensOpts []lensNode.Option
	if isValueSizeLimited {
		chunkSize = immutable.Some(defaultChunkSize)
		lensOpts = append(lensOpts, lensNode.WithBlockstoreChunkSize(defaultChunkSize))
		n.opts.DB.ChunkSize = immutable.Some(defaultChunkSize)
	}

	err = n.startP2P(ctx, rootstore, chunkSize)
	if err != nil {
		return err
	}

	dbOpts := n.buildDBOptions(lensOpts)

	n.DB, err = db.NewDB(ctx, rootstore, nodeACP, documentACP, dbOpts...)
	if err != nil {
		return err
	}

	return n.startAPI(ctx)
}

// buildDBOptions converts NodeDBOptions to db.Option slice.
func (n *Node) buildDBOptions(lensOpts []lensNode.Option) []db.Option {
	dbConfig := &n.opts.DB
	var opts []db.Option

	if dbConfig.MaxTxnRetries.HasValue() {
		opts = append(opts, db.WithMaxRetries(dbConfig.MaxTxnRetries.Value()))
	}
	if dbConfig.Identity.HasValue() {
		opts = append(opts, db.WithNodeIdentity(dbConfig.Identity.Value()))
	}
	opts = append(opts, db.WithEnabledSigning(dbConfig.EnableSigning))
	if len(dbConfig.SearchableEncryptionKey) > 0 {
		opts = append(opts, db.WithSearchableEncryptionKey(dbConfig.SearchableEncryptionKey))
	}
	if len(dbConfig.RetryIntervals) > 0 {
		opts = append(opts, db.WithRetryInterval(dbConfig.RetryIntervals))
	}
	if dbConfig.P2PBlockSyncTimeout > 0 {
		opts = append(opts, db.WithP2PBlockSyncTimeout(dbConfig.P2PBlockSyncTimeout))
	}
	if dbConfig.LensRuntime != "" {
		opts = append(opts, db.WithLensRuntime(db.LensRuntimeType(dbConfig.LensRuntime)))
	}
	if len(lensOpts) > 0 {
		opts = append(opts, db.WithLensOpts(lensOpts...))
	}
	if dbConfig.ChunkSize.HasValue() {
		opts = append(opts, db.WithBlockStoreChunkSize(dbConfig.ChunkSize.Value()))
	}
	if n.peer != nil {
		opts = append(opts, db.WithP2P(n.peer))
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
	if !n.config.enableDevelopment {
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
