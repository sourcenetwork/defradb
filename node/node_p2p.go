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
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/libp2p/go-libp2p/x/rate"
	ma "github.com/multiformats/go-multiaddr"

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
	if opts.ResourceMemoryMiB <= 0 && opts.MaxStreamsPerPeer <= 0 &&
		len(opts.AllowlistCIDRs) == 0 && opts.MaxConnsPerIP <= 0 {
		return nil, nil
	}

	rmOpts := []rcmgr.Option{rcmgr.WithTraceReporter(newBlockedScopeReporter())}
	var allowedPrefixes []netip.Prefix
	if len(opts.AllowlistCIDRs) > 0 {
		allowlist, err := cidrsToMultiaddrs(opts.AllowlistCIDRs)
		if err != nil {
			return nil, err
		}
		rmOpts = append(rmOpts, rcmgr.WithAllowlistedMultiaddrs(allowlist))
		for _, cidr := range opts.AllowlistCIDRs {
			prefix, err := netip.ParsePrefix(cidr)
			if err != nil {
				return nil, NewErrInvalidAllowlistCIDR(cidr, err)
			}
			allowedPrefixes = append(allowedPrefixes, prefix.Masked())
		}
	}
	if opts.MaxConnsPerIP > 0 {
		rmOpts = append(rmOpts, rcmgr.WithLimitPerSubnet(connLimitsPerSubnet(opts.MaxConnsPerIP)))
	}
	if len(allowedPrefixes) > 0 || opts.MaxConnsPerIP > 0 {
		rmOpts = append(rmOpts, rcmgr.WithConnRateLimiters(connRateLimiter(allowedPrefixes, opts.MaxConnsPerIP)))
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
	return rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(scaled), rmOpts...)
}

// connLimitsPerSubnet returns per-subnet connection caps that mirror libp2p's defaults with the
// per-address count replaced: one IPv4 address (/32), one IPv6 /56, and eight times that per /48.
func connLimitsPerSubnet(maxConnsPerIP int) (ipv4, ipv6 []rcmgr.ConnLimitPerSubnet) {
	ipv4 = []rcmgr.ConnLimitPerSubnet{{PrefixLength: 32, ConnCount: maxConnsPerIP}}
	ipv6 = []rcmgr.ConnLimitPerSubnet{
		{PrefixLength: 56, ConnCount: maxConnsPerIP},
		{PrefixLength: 48, ConnCount: 8 * maxConnsPerIP},
	}
	return ipv4, ipv6
}

// libp2pDefaultMaxConnsPerIP is go-libp2p's per-address connection cap, from which its default
// connection rate-limit bursts are derived.
const libp2pDefaultMaxConnsPerIP = 8

// connRateLimiter rebuilds libp2p's default new-connection rate limiter with two changes.
//
// Allowlisted prefixes are exempt. libp2p checks this limiter before the allowlist, and
// WithAllowlistedMultiaddrs does not touch it, so without this an allowlisted NAT address would
// still be refused after a burst of 16 new connections and then admitted at one every 5s.
//
// The bursts scale with maxConnsPerIP in libp2p's own ratio (2x per /32 and /56, 10x per /48), so
// raising the concurrent cap does not leave the rate limit as the tighter bound. Zero keeps the
// libp2p defaults. Loopback stays exempt, as in libp2p.
func connRateLimiter(allowlisted []netip.Prefix, maxConnsPerIP int) *rate.Limiter {
	n := maxConnsPerIP
	if n <= 0 {
		n = libp2pDefaultMaxConnsPerIP
	}
	prefixes := []rate.PrefixLimit{
		{Prefix: netip.MustParsePrefix("127.0.0.0/8")},
		{Prefix: netip.MustParsePrefix("::1/128")},
	}
	for _, p := range allowlisted {
		prefixes = append(prefixes, rate.PrefixLimit{Prefix: p}) // zero Limit: unlimited
	}
	return &rate.Limiter{
		NetworkPrefixLimits: prefixes,
		SubnetRateLimiter: rate.SubnetLimiter{
			IPv4SubnetLimits: []rate.SubnetLimit{
				{PrefixLength: 32, Limit: rate.Limit{RPS: 0.2, Burst: 2 * n}},
			},
			IPv6SubnetLimits: []rate.SubnetLimit{
				{PrefixLength: 56, Limit: rate.Limit{RPS: 0.2, Burst: 2 * n}},
				{PrefixLength: 48, Limit: rate.Limit{RPS: 0.5, Burst: 10 * n}},
			},
			GracePeriod: time.Minute,
		},
	}
}

// cidrsToMultiaddrs converts CIDRs such as "10.244.0.0/16" or "fd00::/8" into the
// /ip4/<net>/ipcidr/<bits> or /ip6/<net>/ipcidr/<bits> multiaddrs the resource manager's allowlist
// takes. A bare address is not accepted, so a missing prefix length fails here rather than
// silently allowlisting nothing.
func cidrsToMultiaddrs(cidrs []string) ([]ma.Multiaddr, error) {
	out := make([]ma.Multiaddr, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, NewErrInvalidAllowlistCIDR(cidr, err)
		}
		bits, _ := ipNet.Mask.Size()
		proto := "ip6"
		if ip4 := ipNet.IP.To4(); ip4 != nil {
			proto = "ip4"
		}
		addr, err := ma.NewMultiaddr(fmt.Sprintf("/%s/%s/ipcidr/%d", proto, ipNet.IP, bits))
		if err != nil {
			return nil, NewErrInvalidAllowlistCIDR(cidr, err)
		}
		out = append(out, addr)
	}
	return out, nil
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
