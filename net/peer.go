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
	blocks "github.com/ipfs/go-block-format"
	gostream "github.com/libp2p/go-libp2p-gostream"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	libp2pevent "github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"

	"github.com/multiformats/go-multiaddr"
	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"
	"google.golang.org/grpc"

	"github.com/sourcenetwork/defradb/acp/dac"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	corenet "github.com/sourcenetwork/defradb/internal/core/net"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/se"
	"github.com/sourcenetwork/defradb/internal/telemetry"
	"github.com/sourcenetwork/defradb/net/config"
)

var tracer = telemetry.NewTracer()

// DB hold the database related methods that are required by Peer.
type DB interface {
	NewTxn(ctx context.Context, readOnly bool) (client.Txn, error)
	// Blockstore returns the blockstore, within which all blocks (commits) managed by DefraDB are held.
	Blockstore() datastore.Blockstore
	// GetCollections returns the list of collections according to the given options.
	GetCollections(ctx context.Context, opts client.CollectionFetchOptions) ([]client.Collection, error)
	// GetNodeIndentityToken returns an identity token for the given audience.
	GetNodeIdentityToken(ctx context.Context, audience immutable.Option[string]) ([]byte, error)
	// Rootstore returns the instance's root store.
	Rootstore() corekv.TxnStore
}

// Peer is a DefraDB Peer node which exposes all the LibP2P host/peer functionality
// to the underlying DefraDB instance.
type Peer struct {
	bus       event.Bus
	updateSub event.Subscription

	ctx    context.Context
	cancel context.CancelFunc

	host host.Host
	dht  routing.Routing
	ps   *pubsub.PubSub

	server *server
	p2pRPC *grpc.Server // rpc server over the P2P network

	// peer DAG service
	blockService blockservice.BlockService

	documentACP immutable.Option[dac.DocumentACP]
	db          DB

	bootCloser io.Closer

	// The intervals at which to retry replicator failures.
	// For example, this can define an exponential backoff strategy.
	retryIntervals   []time.Duration
	handleRetryMutex *sync.Mutex
}

var _ client.P2P = (*Peer)(nil)

// NewPeer creates a new instance of the DefraDB server as a peer-to-peer node.
func NewPeer(
	ctx context.Context,
	bus event.Bus,
	documentACP immutable.Option[dac.DocumentACP],
	db DB,
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

	if db == nil {
		return nil, ErrNilDB
	}

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
		host:             h,
		dht:              ddht,
		ctx:              ctx,
		cancel:           cancel,
		bus:              bus,
		documentACP:      documentACP,
		db:               db,
		p2pRPC:           grpc.NewServer(options.GRPCServerOptions...),
		retryIntervals:   options.RetryIntervals,
		handleRetryMutex: &sync.Mutex{},
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
		p.updateSub, err = p.bus.Subscribe(event.UpdateName, event.P2PTopicName, event.ReplicatorName, se.UpdateEventName)
		if err != nil {
			return nil, err
		}
		log.Info("Starting internal broadcaster for pubsub network")
		go p.handleMessageLoop()
	}

	p.server, err = newServer(p, options.GRPCDialOptions...)
	if err != nil {
		return nil, err
	}

	bs := datastore.BlockstoreFrom(db.Rootstore())
	bswapnet := bsnet.NewFromIpfsHost(h)
	bswap := bitswap.New(ctx, bswapnet, ddht, bs, bitswap.WithPeerBlockRequestFilter(p.server.hasAccess))
	p.blockService = blockservice.New(bs, bswap)

	p2pListener, err := gostream.Listen(h, corenet.Protocol)
	if err != nil {
		return nil, err
	}

	p.bootCloser, err = bootstrap.Bootstrap(p.PeerID(), h, ddht, bootstrap.BootstrapConfigWithPeers(peers))
	if err != nil {
		return nil, err
	}

	// register the P2P gRPC server
	go func() {
		registerServiceServer(p.p2pRPC, p.server)
		if err := p.p2pRPC.Serve(p2pListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.ErrorE("Fatal P2P RPC server error", err)
		}
	}()

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

	bus.Publish(event.NewMessage(event.PeerInfoName, event.PeerInfo{Info: p.PeerInfo()}))

	go p.handleReplicatorRetries(ctx)

	err = p.loadAndPublishReplicators(ctx)
	if err != nil {
		return nil, err
	}

	go func() {
		// This can be a long running operation so running it in a goroutine
		// ensures calling `NewPeer` won't block.
		err := p.loadAndPublishP2PCollections(ctx)
		if err != nil {
			log.ErrorE("Error loading P2P collections", err)
		}
	}()

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

	if p.server != nil {
		// close topics
		if err := p.server.removeAllPubsubTopics(); err != nil {
			log.ErrorE("Error closing pubsub topics", err)
		}

		// stop gRPC server
		for _, c := range p.server.conns {
			if err := c.Close(); err != nil {
				log.ErrorE("Failed closing server RPC connections", err)
			}
		}
	}

	if p.updateSub != nil {
		p.bus.Unsubscribe(p.updateSub)
	}

	if err := p.blockService.Close(); err != nil {
		log.ErrorE("Error closing block service", err)
	}

	if err := p.host.Close(); err != nil {
		log.ErrorE("Error closing host", err)
	}

	stopped := make(chan struct{})
	go func() {
		p.p2pRPC.GracefulStop()
		close(stopped)
	}()
	timer := time.NewTimer(10 * time.Second)
	select {
	case <-timer.C:
		p.p2pRPC.Stop()
		log.Info("Peer gRPC server was shutdown ungracefully")
	case <-stopped:
		timer.Stop()
	}
}

// handleMessage loop manages the transition of messages
// from the internal broadcaster to the external pubsub network
func (p *Peer) handleMessageLoop() {
	for {
		msg, isOpen := <-p.updateSub.Message()
		if !isOpen {
			return
		}

		switch evt := msg.Data.(type) {
		case event.Update:
			err := p.handleLog(evt)
			if err != nil {
				log.ErrorE("Error while handling broadcast log", err)
			}

		case event.P2PTopic:
			p.server.updatePubSubTopics(evt)

		case se.ReplicateEvent:
			err := p.handleSELog(evt)
			if err != nil {
				log.ErrorE("Error while handling SE log", err)
			}

		default:
			// ignore other events
			continue
		}
	}
}

func (p *Peer) handleLog(evt event.Update) error {
	if evt.DocID != "" {
		_, err := client.NewDocIDFromString(evt.DocID)
		if err != nil {
			return NewErrFailedToGetDocID(err)
		}
	}

	// push to each peer (replicator)
	p.pushLogToReplicators(evt)

	// Retries are for replicators only and should not pollute the pubsub network.
	if !evt.IsRetry {
		req := &pushLogRequest{
			DocID:        evt.DocID,
			CID:          evt.Cid.Bytes(),
			CollectionID: evt.CollectionID,
			Creator:      p.host.ID().String(),
			Block:        evt.Block,
		}

		if evt.DocID != "" {
			if err := p.server.publishLog(p.ctx, evt.DocID, req); err != nil {
				return NewErrPublishingToDocIDTopic(err, evt.Cid.String(), evt.DocID)
			}
		}

		if err := p.server.publishLog(p.ctx, evt.CollectionID, req); err != nil {
			return NewErrPublishingToSchemaTopic(err, evt.Cid.String(), evt.CollectionID)
		}
	}

	return nil
}

func (p *Peer) pushLogToReplicators(lg event.Update) {
	// let the exchange know we have this block
	// this should speed up the dag sync process
	err := p.blockService.Exchange().NotifyNewBlocks(context.Background(), blocks.NewBlock(lg.Block))
	if err != nil {
		log.ErrorE("Failed to notify new blocks", err)
	}

	p.server.mu.Lock()
	reps, exists := p.server.replicators[lg.CollectionID]
	p.server.mu.Unlock()

	if exists {
		for pid := range reps {
			go func(peerID peer.ID) {
				if err := p.server.pushLog(lg, peerID); err != nil {
					log.ErrorE(
						"Failed pushing log",
						err,
						corelog.String("DocID", lg.DocID),
						corelog.Any("CID", lg.Cid),
						corelog.Any("PeerID", peerID))
				}
			}(pid)
		}
	}
}

func (p *Peer) handleSELog(evt se.ReplicateEvent) error {
	// Push to each replicator (similar to handleLog)
	p.pushSEArtifactsToReplicators(evt)
	return nil
}

func (p *Peer) pushSEArtifactsToReplicators(evt se.ReplicateEvent) {
	p.server.mu.Lock()
	reps, exists := p.server.replicators[evt.CollectionID]
	p.server.mu.Unlock()

	if exists {
		for pid := range reps {
			go func(peerID peer.ID) {
				if err := p.server.pushSEArtifacts(evt, peerID); err != nil {
					log.ErrorE(
						"Failed pushing SE artifacts",
						err,
						corelog.String("DocID", evt.DocID),
						corelog.String("CollectionID", evt.CollectionID),
						corelog.Any("PeerID", peerID))
				}
			}(pid)
		}
	}
}

// Connect initiates a connection to the peer with the given address.
func (p *Peer) Connect(ctx context.Context, addr peer.AddrInfo) error {
	return p.host.Connect(ctx, addr)
}

func (p *Peer) PeerID() peer.ID {
	return p.host.ID()
}

func (p *Peer) ListenAddrs() []multiaddr.Multiaddr {
	return p.host.Network().ListenAddresses()
}

func (p *Peer) PeerInfo() peer.AddrInfo {
	return peer.AddrInfo{
		ID:    p.host.ID(),
		Addrs: p.host.Network().ListenAddresses(),
	}
}

func (p *Peer) Server() *server {
	return p.server
}
