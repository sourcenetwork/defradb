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
	"time"

	"github.com/ipfs/boxo/bitswap"
	"github.com/ipfs/boxo/bitswap/network"
	"github.com/ipfs/boxo/blockservice"
	"github.com/ipfs/boxo/bootstrap"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	gostream "github.com/libp2p/go-libp2p-gostream"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"

	"github.com/multiformats/go-multiaddr"
	"github.com/sourcenetwork/corelog"
	"google.golang.org/grpc"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	corenet "github.com/sourcenetwork/defradb/internal/core/net"
)

// Peer is a DefraDB Peer node which exposes all the LibP2P host/peer functionality
// to the underlying DefraDB instance.
type Peer struct {
	blockstore datastore.Blockstore
	encstore   datastore.Blockstore

	bus       *event.Bus
	updateSub *event.Subscription

	ctx    context.Context
	cancel context.CancelFunc

	host host.Host
	dht  routing.Routing
	ps   *pubsub.PubSub

	server *server
	p2pRPC *grpc.Server // rpc server over the P2P network

	// peer DAG service
	bserv blockservice.BlockService

	bootCloser io.Closer
}

// NewPeer creates a new instance of the DefraDB server as a peer-to-peer node.
func NewPeer(
	ctx context.Context,
	blockstore datastore.Blockstore,
	encstore datastore.Blockstore,
	bus *event.Bus,
	opts ...NodeOpt,
) (p *Peer, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		if p == nil {
			cancel()
		} else if err != nil {
			p.Close()
		}
	}()

	if blockstore == nil || encstore == nil {
		return nil, ErrNilDB
	}

	options := DefaultOptions()
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

	bswapnet := network.NewFromIpfsHost(h, ddht)
	bswap := bitswap.New(ctx, bswapnet, blockstore)

	p = &Peer{
		host:       h,
		dht:        ddht,
		blockstore: blockstore,
		encstore:   encstore,
		ctx:        ctx,
		cancel:     cancel,
		bus:        bus,
		p2pRPC:     grpc.NewServer(options.GRPCServerOptions...),
		bserv:      blockservice.New(blockstore, bswap),
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
		p.updateSub, err = p.bus.Subscribe(event.UpdateName, event.P2PTopicName, event.ReplicatorName)
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
		if err := p.p2pRPC.Serve(p2pListener); err != nil &&
			!errors.Is(err, grpc.ErrServerStopped) {
			log.ErrorE("Fatal P2P RPC server error", err)
		}
	}()

	bus.Publish(event.NewMessage(event.PeerInfoName, event.PeerInfo{Info: p.PeerInfo()}))

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

	if err := p.bserv.Close(); err != nil {
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
			var err error
			if evt.IsCreate {
				err = p.handleDocCreateLog(evt)
			} else {
				err = p.handleDocUpdateLog(evt)
			}

			if err != nil {
				log.ErrorE("Error while handling broadcast log", err)
			}

		case event.P2PTopic:
			p.server.updatePubSubTopics(evt)

		case event.Replicator:
			p.server.updateReplicators(evt)
		default:
			// ignore other events
			continue
		}
	}
}

// RegisterNewDocument registers a new document with the peer node.
func (p *Peer) RegisterNewDocument(
	ctx context.Context,
	docID client.DocID,
	c cid.Cid,
	rawBlock []byte,
	schemaRoot string,
) error {
	// register topic
	err := p.server.addPubSubTopic(docID.String(), !p.server.hasPubSubTopic(schemaRoot), nil)
	if err != nil {
		log.ErrorE(
			"Failed to create new pubsub topic",
			err,
			corelog.String("DocID", docID.String()),
		)
		return err
	}

	req := &pushLogRequest{
		DocID:      docID.String(),
		CID:        c.Bytes(),
		SchemaRoot: schemaRoot,
		Creator:    p.host.ID().String(),
		Block:      rawBlock,
	}

	return p.server.publishLog(ctx, schemaRoot, req)
}

func (p *Peer) handleDocCreateLog(evt event.Update) error {
	docID, err := client.NewDocIDFromString(evt.DocID)
	if err != nil {
		return NewErrFailedToGetDocID(err)
	}

	// We need to register the document before pushing to the replicators if we want to
	// ensure that we have subscribed to the topic.
	err = p.RegisterNewDocument(p.ctx, docID, evt.Cid, evt.Block, evt.SchemaRoot)
	if err != nil {
		return err
	}
	// push to each peer (replicator)
	p.pushLogToReplicators(evt)

	return nil
}

func (p *Peer) handleDocUpdateLog(evt event.Update) error {
	// push to each peer (replicator)
	p.pushLogToReplicators(evt)

	_, err := client.NewDocIDFromString(evt.DocID)
	if err != nil {
		return NewErrFailedToGetDocID(err)
	}

	req := &pushLogRequest{
		DocID:      evt.DocID,
		CID:        evt.Cid.Bytes(),
		SchemaRoot: evt.SchemaRoot,
		Creator:    p.host.ID().String(),
		Block:      evt.Block,
	}

	if err := p.server.publishLog(p.ctx, evt.DocID, req); err != nil {
		return NewErrPublishingToDocIDTopic(err, evt.Cid.String(), evt.DocID)
	}

	if err := p.server.publishLog(p.ctx, evt.SchemaRoot, req); err != nil {
		return NewErrPublishingToSchemaTopic(err, evt.Cid.String(), evt.SchemaRoot)
	}

	return nil
}

func (p *Peer) pushLogToReplicators(lg event.Update) {
	// let the exchange know we have this block
	// this should speed up the dag sync process
	err := p.bserv.Exchange().NotifyNewBlocks(context.Background(), blocks.NewBlock(lg.Block))
	if err != nil {
		log.ErrorE("Failed to notify new blocks", err)
	}

	// push to each peer (replicator)
	peers := make(map[string]struct{})
	for _, peer := range p.ps.ListPeers(lg.DocID) {
		peers[peer.String()] = struct{}{}
	}
	for _, peer := range p.ps.ListPeers(lg.SchemaRoot) {
		peers[peer.String()] = struct{}{}
	}

	p.server.mu.Lock()
	reps, exists := p.server.replicators[lg.SchemaRoot]
	p.server.mu.Unlock()

	if exists {
		for pid := range reps {
			// Don't push if pid is in the list of peers for the topic.
			// It will be handled by the pubsub system.
			if _, ok := peers[pid.String()]; ok {
				continue
			}
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
