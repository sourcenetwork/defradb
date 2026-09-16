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
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corelog"
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
// maxProtocolPeers is how many peers one protocol is budgeted to serve at the per-peer limit at
// once. It only sizes the protocol-wide stream ceiling, which exists to stop a single protocol
// starving the others, not to limit any one peer.
const maxProtocolPeers = 8

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

		// The peer scope is not the binding one. Each protocol also has its own budget per peer,
		// which defaults to far less: a node serving CARs to three peers was refused at 68 inbound
		// streams on protocol:/defradb/car_req/0.0.1.peer:..., resetting the rest with
		// StreamResourceLimitExceeded while the peer scope sat far below its limit. A protocol may
		// use the whole per-peer budget, and across all peers a multiple of it.
		limits.ProtocolPeerBaseLimit.StreamsInbound = opts.MaxStreamsPerPeer
		limits.ProtocolPeerBaseLimit.StreamsOutbound = opts.MaxStreamsPerPeer
		limits.ProtocolPeerBaseLimit.Streams = 2 * opts.MaxStreamsPerPeer
		limits.ProtocolBaseLimit.StreamsInbound = maxProtocolPeers * opts.MaxStreamsPerPeer
		limits.ProtocolBaseLimit.StreamsOutbound = maxProtocolPeers * opts.MaxStreamsPerPeer
		limits.ProtocolBaseLimit.Streams = 2 * maxProtocolPeers * opts.MaxStreamsPerPeer
	}

	scaled := limits.AutoScale()
	if opts.ResourceMemoryMiB > 0 {
		fds := opts.ResourceFileDescriptors
		if fds <= 0 {
			fds = defaultResourceFileDescriptors
		}
		scaled = limits.Scale(int64(opts.ResourceMemoryMiB)*mib, fds)
	}
	return rcmgr.NewResourceManager(
		rcmgr.NewFixedLimiter(scaled),
		rcmgr.WithTraceReporter(newBlockedScopeReporter()),
	)
}

// blockedScopeReporter names the resource-manager scope that refused a stream, connection or
// memory reservation. A peer whose request fails only sees a reset stream, so without this the
// limit that caused it — system, transient, peer, protocol or service — is invisible.
type blockedScopeReporter struct {
	mu   sync.Mutex
	seen map[string]int64
}

func newBlockedScopeReporter() *blockedScopeReporter {
	return &blockedScopeReporter{seen: make(map[string]int64)}
}

// ConsumeEvent is called synchronously by the resource manager, so it only counts, and logs the
// first of each scope and event, then every thousandth, which names the limit without flooding.
func (r *blockedScopeReporter) ConsumeEvent(evt rcmgr.TraceEvt) {
	switch evt.Type {
	case rcmgr.TraceBlockAddStreamEvt, rcmgr.TraceBlockAddConnEvt, rcmgr.TraceBlockReserveMemoryEvt:
	default:
		return
	}

	key := string(evt.Type) + " " + evt.Name
	r.mu.Lock()
	r.seen[key]++
	count := r.seen[key]
	r.mu.Unlock()

	if count != 1 && count%1000 != 0 {
		return
	}
	log.Info("resource manager blocked",
		corelog.String("event", string(evt.Type)),
		corelog.String("scope", evt.Name),
		corelog.String("limit", fmt.Sprintf("%v", evt.Limit)),
		corelog.Int64("blocked", count),
		corelog.Int("streamsIn", evt.StreamsIn),
		corelog.Int("streamsOut", evt.StreamsOut),
		corelog.Int64("memory", evt.Memory))
}
