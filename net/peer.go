// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package net

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/ipfs/boxo/bitswap"
	"github.com/ipfs/boxo/bitswap/network/bsnet"
	"github.com/ipfs/boxo/blockservice"
	"github.com/ipfs/boxo/bootstrap"
	"github.com/ipfs/go-cid"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	libp2pevent "github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/net/config"
)

// Peer is a DefraDB Peer node which exposes all the LibP2P host/peer functionality
// to the underlying DefraDB instance.
type Peer struct {
	ctx    context.Context
	cancel context.CancelFunc

	host host.Host
	dht  routing.Routing
	ps   *pubsub.PubSub

	topics  map[string]pubsubTopic
	topicMu sync.Mutex

	// peer DAG service
	blockService blockservice.BlockService

	bootCloser io.Closer

	blockAccessFunc immutable.Option[client.BlockAccessFunc]
	accessFuncMu    sync.Mutex
}

// NewPeer creates a new instance of the DefraDB server as a peer-to-peer node.
func NewPeer(
	ctx context.Context,
	blockstore datastore.Blockstore,
	opts ...config.NodeOpt,
) (p *Peer, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		if p == nil {
			cancel()
		} else if err != nil {
			p.Close()
		}
	}()

	options := config.DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	peers := make([]peer.AddrInfo, len(options.BootstrapPeers))
	for i, p := range options.BootstrapPeers {
		addr, err := peer.AddrInfoFromString(p)
		if err != nil {
			return nil, err
		}
		peers[i] = *addr
	}

	h, ddht, err := setupHost(ctx, options)
	if err != nil {
		return nil, err
	}

	log.InfoContext(
		ctx,
		"Created LibP2P host",
		corelog.Any("PeerId", h.ID()),
		corelog.Any("Address", options.ListenAddresses),
	)

	p = &Peer{
		host:   h,
		dht:    ddht,
		ctx:    ctx,
		cancel: cancel,
		topics: make(map[string]pubsubTopic),
	}

	if options.EnablePubSub {
		p.ps, err = pubsub.NewGossipSub(
			ctx,
			h,
			pubsub.WithPeerExchange(true),
			pubsub.WithFloodPublish(true),
		)
		if err != nil {
			return nil, err
		}
	}

	bswapnet := bsnet.NewFromIpfsHost(h)
	bswap := bitswap.New(ctx, bswapnet, ddht, blockstore, bitswap.WithPeerBlockRequestFilter(p.hasAccess))
	p.blockService = blockservice.New(blockstore, bswap)

	p.bootCloser, err = bootstrap.Bootstrap(h.ID(), h, ddht, bootstrap.BootstrapConfigWithPeers(peers))
	if err != nil {
		return nil, err
	}

	// There is a possibility for the PeerInfo event to trigger before the PeerInfo has been set for the host.
	// To avoid this, we wait for the host to indicate that its local address has been updated.
	sub, err := h.EventBus().Subscribe(&libp2pevent.EvtLocalAddressesUpdated{})
	if err != nil {
		return nil, err
	}
	select {
	case <-sub.Out():
		break
	case <-time.After(5 * time.Second):
		// This can only happen if the listening address has been mistakenly set to a zero value.
		return nil, ErrTimeoutWaitingForPeerInfo
	}

	return p, nil
}

// Close the peer node and all its internal workers/goroutines/loops.
func (p *Peer) Close() {
	defer p.cancel()

	if p.bootCloser != nil {
		// close bootstrap service
		if err := p.bootCloser.Close(); err != nil {
			log.ErrorE("Error closing bootstrap", err)
		}
	}

	if err := p.removeAllPubsubTopics(); err != nil {
		log.ErrorE("Error closing pubsub topics", err)
	}

	if err := p.blockService.Close(); err != nil {
		log.ErrorE("Error closing block service", err)
	}

	if err := p.host.Close(); err != nil {
		log.ErrorE("Error closing host", err)
	}
}

// hasAccess checks if the requesting peer has access to the given cid.
//
// This is used as a filter in bitswap to determine if we should send the block to the requesting peer.
func (p *Peer) hasAccess(pid peer.ID, c cid.Cid) bool {
	p.accessFuncMu.Lock()
	defer p.accessFuncMu.Unlock()
	if p.blockAccessFunc.HasValue() {
		return p.blockAccessFunc.Value()(p.ctx, pid.String(), c)
	}
	// if no blockAccessFunc has been defined we allow all block exchanges.
	return true
}
