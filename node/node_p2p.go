// Copyright 2025 Democratized Data Foundation
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

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/go-p2p"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

// defaultResourceFileDescriptors is the file descriptor budget used when a memory budget is set
// without one. It matches the order of magnitude libp2p autoscales to on a small host.
const defaultResourceFileDescriptors = 512

// mib converts a MiB count to bytes.
const mib = 1 << 20

func (n *Node) startP2P(ctx context.Context, store corekv.TxnReaderWriter, chunkSize immutable.Option[int]) error {
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
	p2pOpts = append(p2pOpts, p2p.WithEnablePubSub(n.opts.P2P.EnablePubSub))
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

	rm, err := p2pResourceManager(n.opts.P2P)
	if err != nil {
		return err
	}
	if rm != nil {
		p2pOpts = append(p2pOpts, p2p.WithResourceManager(rm))
	}

	peer, err := p2p.NewPeer(ctx, p2pOpts...)
	if err != nil {
		return err
	}
	n.peer = peer
	return nil
}

// p2pResourceManager builds the libp2p resource manager from the configured limits, or returns nil
// to leave libp2p its autoscaled defaults.
//
// The defaults scale with the memory the process can see, which inside a container is the host's
// memory and not the container's limit, so a node can end up with limits its cgroup cannot honour.
// Exceeding a limit resets the stream with StreamResourceLimitExceeded, which the peer that opened
// it sees as a failed request.
func p2pResourceManager(opts options.NodeP2POptions) (network.ResourceManager, error) {
	if opts.ResourceMemoryMiB <= 0 && opts.MaxStreamsPerPeer <= 0 {
		return nil, nil
	}

	limits := rcmgr.DefaultLimits
	// Gives the protocols libp2p runs itself, such as identify and ping, their own headroom rather
	// than letting them compete with DefraDB's traffic for the peer budget.
	libp2p.SetDefaultServiceLimits(&limits)

	if opts.MaxStreamsPerPeer > 0 {
		limits.PeerBaseLimit.StreamsInbound = opts.MaxStreamsPerPeer
		limits.PeerBaseLimit.StreamsOutbound = opts.MaxStreamsPerPeer
		limits.PeerBaseLimit.Streams = 2 * opts.MaxStreamsPerPeer
	}

	scaled := limits.AutoScale()
	if opts.ResourceMemoryMiB > 0 {
		fds := opts.ResourceFileDescriptors
		if fds <= 0 {
			fds = defaultResourceFileDescriptors
		}
		scaled = limits.Scale(int64(opts.ResourceMemoryMiB)*mib, fds)
	}
	return rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(scaled))
}
