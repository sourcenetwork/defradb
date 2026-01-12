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
)

func (n *Node) startP2P(ctx context.Context, store corekv.ReaderWriter, chunkSize immutable.Option[int]) error {
	if n.config.disableP2P {
		return nil
	}

	p2pConfig := &n.opts.P2P
	var p2pOpts []p2p.NodeOpt

	if len(p2pConfig.ListenAddresses) > 0 {
		p2pOpts = append(p2pOpts, p2p.WithListenAddresses(p2pConfig.ListenAddresses...))
	}
	if len(p2pConfig.BootstrapPeers) > 0 {
		p2pOpts = append(p2pOpts, p2p.WithBootstrapPeers(p2pConfig.BootstrapPeers...))
	}
	if p2pConfig.EnablePubSub {
		p2pOpts = append(p2pOpts, p2p.WithEnablePubSub(true))
	}
	if p2pConfig.EnableRelay {
		p2pOpts = append(p2pOpts, p2p.WithEnableRelay(true))
	}
	if p2pConfig.EnableClearBackoffOnRetry {
		p2pOpts = append(p2pOpts, p2p.WithClearBackoffOnRetry(true))
	}
	if len(p2pConfig.PrivateKey) > 0 {
		p2pOpts = append(p2pOpts, p2p.WithPrivateKey(p2pConfig.PrivateKey))
	}
	p2pOpts = append(p2pOpts, p2p.WithBlockstore(datastore.P2PBlockstoreFrom(store, chunkSize)))

	peer, err := p2p.NewPeer(ctx, p2pOpts...)
	if err != nil {
		return err
	}
	n.peer = peer
	return nil
}
