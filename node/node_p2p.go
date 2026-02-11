// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// P2P networking stack does not work in JS builds.
//
//go:build !js

package node

import (
	"context"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/go-p2p"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db"
)

func (n *Node) startP2P(ctx context.Context, store corekv.ReaderWriter, chunkSize immutable.Option[int]) error {
	if n.opts.DisableP2P {
		return nil
	}

	var p2pOpts []p2p.NodeOpt
	if len(n.opts.P2P.ListenAddresses) > 0 {
		p2pOpts = append(p2pOpts, p2p.WithListenAddresses(n.opts.P2P.ListenAddresses...))
	}
	if len(n.opts.P2P.BootstrapPeers) > 0 {
		p2pOpts = append(p2pOpts, p2p.WithBootstrapPeers(n.opts.P2P.BootstrapPeers...))
	}
	if n.opts.P2P.EnablePubSub {
		p2pOpts = append(p2pOpts, p2p.WithEnablePubSub(true))
	}
	if n.opts.P2P.EnableRelay {
		p2pOpts = append(p2pOpts, p2p.WithEnableRelay(true))
	}
	if n.opts.P2P.EnableClearBackoffOnRetry {
		p2pOpts = append(p2pOpts, p2p.WithClearBackoffOnRetry(true))
	}
	if len(n.opts.P2P.PrivateKey) > 0 {
		p2pOpts = append(p2pOpts, p2p.WithPrivateKey(n.opts.P2P.PrivateKey))
	}
	p2pOpts = append(p2pOpts, p2p.WithBlockstore(datastore.P2PBlockstoreFrom(store, chunkSize)))

	peer, err := p2p.NewPeer(ctx, p2pOpts...)
	if err != nil {
		return err
	}
	n.peer = peer
	return nil
}

// buildP2PDBOption returns the db.Option for P2P if the peer is set.
func (n *Node) buildP2PDBOption() []db.Option {
	if n.peer != nil {
		return []db.Option{db.WithP2P(n.peer)}
	}
	return nil
}
