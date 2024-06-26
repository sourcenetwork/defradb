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
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ipfs/boxo/bitswap"
	"github.com/ipfs/boxo/bitswap/network"
	"github.com/ipfs/boxo/blockservice"
	exchange "github.com/ipfs/boxo/exchange"
	"github.com/ipfs/boxo/ipns"
	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	libp2p "github.com/libp2p/go-libp2p"
	gostream "github.com/libp2p/go-libp2p-gostream"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	dualdht "github.com/libp2p/go-libp2p-kad-dht/dual"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	record "github.com/libp2p/go-libp2p-record"
	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
	libp2pEvent "github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"

	// @TODO: https://github.com/sourcenetwork/defradb/issues/1902
	//nolint:staticcheck
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoreds"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/multiformats/go-multiaddr"
	"github.com/sourcenetwork/corelog"
	"google.golang.org/grpc"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	corenet "github.com/sourcenetwork/defradb/internal/core/net"
	pb "github.com/sourcenetwork/defradb/net/pb"
)

// Peer is a DefraDB Peer node which exposes all the LibP2P host/peer functionality
// to the underlying DefraDB instance.
type Peer struct {
	blockstore datastore.Blockstore

	bus       *event.Bus
	updateSub *event.Subscription

	host host.Host
	dht  routing.Routing
	ps   *pubsub.PubSub

	server *server
	p2pRPC *grpc.Server // rpc server over the P2P network

	// peer DAG service
	exch  exchange.Interface
	bserv blockservice.BlockService

	ctx      context.Context
	cancel   context.CancelFunc
	dhtClose func() error
}

// NewPeer creates a new instance of the DefraDB server as a peer-to-peer node.
func NewPeer(
	ctx context.Context,
	rootstore datastore.Rootstore,
	blockstore datastore.Blockstore,
	bus *event.Bus,
	opts ...NodeOpt,
) (p *Peer, err error) {
	if rootstore == nil || blockstore == nil {
		return nil, ErrNilDB
	}

	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	connManager, err := connmgr.NewConnManager(100, 400, connmgr.WithGracePeriod(time.Second*20))
	if err != nil {
		return nil, err
	}

	var listenAddresses []multiaddr.Multiaddr
	for _, addr := range options.ListenAddresses {
		listenAddress, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
		listenAddresses = append(listenAddresses, listenAddress)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		if p == nil {
			cancel()
		}
	}()

	peerstore, err := pstoreds.NewPeerstore(ctx, rootstore, pstoreds.DefaultOpts())
	if err != nil {
		return nil, err
	}

	if options.PrivateKey == nil {
		// generate an ephemeral private key
		key, err := crypto.GenerateEd25519()
		if err != nil {
			return nil, err
		}
		options.PrivateKey = key
	}

	// unmarshal the private key bytes
	privateKey, err := libp2pCrypto.UnmarshalEd25519PrivateKey(options.PrivateKey)
	if err != nil {
		return nil, err
	}

	var ddht *dualdht.DHT

	libp2pOpts := []libp2p.Option{
		libp2p.ConnectionManager(connManager),
		libp2p.DefaultTransports,
		libp2p.Identity(privateKey),
		libp2p.ListenAddrs(listenAddresses...),
		libp2p.Peerstore(peerstore),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			// Delete this line and uncomment the next 6 lines once we remove batchable datastore support.
			// var store ds.Batching
			// // If `rootstore` doesn't implement `Batching`, `nil` will be passed
			// // to newDHT which will cause the DHT to be stored in memory.
			// if dsb, isBatching := rootstore.(ds.Batching); isBatching {
			// 	store = dsb
			// }
			ddht, err = newDHT(ctx, h, rootstore)
			return ddht, err
		}),
	}
	if !options.EnableRelay {
		libp2pOpts = append(libp2pOpts, libp2p.DisableRelay())
	}

	h, err := libp2p.New(libp2pOpts...)
	if err != nil {
		return nil, err
	}
	log.InfoContext(
		ctx,
		"Created LibP2P host",
		corelog.Any("PeerId", h.ID()),
		corelog.Any("Address", options.ListenAddresses),
	)

	var ps *pubsub.PubSub
	if options.EnablePubSub {
		ps, err = pubsub.NewGossipSub(
			ctx,
			h,
			pubsub.WithPeerExchange(true),
			pubsub.WithFloodPublish(true),
		)
		if err != nil {
			return nil, err
		}
	}

	if err != nil {
		return nil, err
	}

	sub, err := h.EventBus().Subscribe(&libp2pEvent.EvtPeerConnectednessChanged{})
	if err != nil {
		return nil, err
	}
	// publish subscribed events to the event bus
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case val, isOpen := <-sub.Out():
				if !isOpen {
					return
				}
				bus.Publish(event.NewMessage(event.PeerName, val))
			}
		}
	}()

	p = &Peer{
		host:       h,
		dht:        ddht,
		ps:         ps,
		blockstore: blockstore,
		bus:        bus,
		p2pRPC:     grpc.NewServer(options.GRPCServerOptions...),
		ctx:        ctx,
		cancel:     cancel,
	}

	p.bus.Publish(event.NewMessage(event.PeerInfoName, event.PeerInfo{Info: p.PeerInfo()}))

	p.server, err = newServer(p, options.GRPCDialOptions...)
	if err != nil {
		return nil, err
	}

	p.setupBlockService()

	return p, nil
}

// Start all the internal workers/goroutines/loops that manage the P2P state.
func (p *Peer) Start() error {
	// reconnect to known peers
	var wg sync.WaitGroup
	for _, id := range p.host.Peerstore().PeersWithAddrs() {
		if id == p.host.ID() {
			continue
		}
		wg.Add(1)
		go func(id peer.ID) {
			defer wg.Done()
			addr := p.host.Peerstore().PeerInfo(id)
			err := p.host.Connect(p.ctx, addr)
			if err != nil {
				log.InfoContext(
					p.ctx,
					"Failure while reconnecting to a known peer",
					corelog.Any("peer", id))
			}
		}(id)
	}
	wg.Wait()

	p2plistener, err := gostream.Listen(p.host, corenet.Protocol)
	if err != nil {
		return err
	}

	if p.ps != nil {
		sub, err := p.bus.Subscribe(event.UpdateName, event.P2PTopicName, event.ReplicatorName)
		if err != nil {
			return err
		}
		p.updateSub = sub
		log.InfoContext(p.ctx, "Starting internal broadcaster for pubsub network")
		go p.handleMessageLoop()
	}

	log.InfoContext(
		p.ctx,
		"Starting P2P node",
		corelog.Any("P2P addresses", p.host.Addrs()))
	// register the P2P gRPC server
	go func() {
		pb.RegisterServiceServer(p.p2pRPC, p.server)
		if err := p.p2pRPC.Serve(p2plistener); err != nil &&
			!errors.Is(err, grpc.ErrServerStopped) {
			log.ErrorContextE(p.ctx, "Fatal P2P RPC server error", err)
		}
	}()

	return nil
}

// Close the peer node and all its internal workers/goroutines/loops.
func (p *Peer) Close() {
	// close topics
	if err := p.server.removeAllPubsubTopics(); err != nil {
		log.ErrorContextE(p.ctx, "Error closing pubsub topics", err)
	}

	// stop gRPC server
	for _, c := range p.server.conns {
		if err := c.Close(); err != nil {
			log.ErrorContextE(p.ctx, "Failed closing server RPC connections", err)
		}
	}

	if p.updateSub != nil {
		p.bus.Unsubscribe(p.updateSub)
	}

	if err := p.bserv.Close(); err != nil {
		log.ErrorContextE(p.ctx, "Error closing block service", err)
	}

	if err := p.host.Close(); err != nil {
		log.ErrorContextE(p.ctx, "Error closing host", err)
	}

	if p.dhtClose != nil {
		err := p.dhtClose()
		if err != nil {
			log.ErrorContextE(p.ctx, "Failed to close DHT", err)
		}
	}

	stopGRPCServer(p.ctx, p.p2pRPC)

	if p.cancel != nil {
		p.cancel()
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
				log.ErrorContextE(p.ctx, "Error while handling broadcast log", err)
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
	if err := p.server.addPubSubTopic(docID.String(), !p.server.hasPubSubTopic(schemaRoot)); err != nil {
		log.ErrorContextE(
			p.ctx,
			"Failed to create new pubsub topic",
			err,
			corelog.String("DocID", docID.String()),
		)
		return err
	}

	// publish log
	req := &pb.PushLogRequest{
		Body: &pb.PushLogRequest_Body{
			DocID:      []byte(docID.String()),
			Cid:        c.Bytes(),
			SchemaRoot: []byte(schemaRoot),
			Creator:    p.host.ID().String(),
			Log: &pb.Document_Log{
				Block: rawBlock,
			},
		},
	}

	return p.server.publishLog(p.ctx, schemaRoot, req)
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
	docID, err := client.NewDocIDFromString(evt.DocID)
	if err != nil {
		return NewErrFailedToGetDocID(err)
	}

	body := &pb.PushLogRequest_Body{
		DocID:      []byte(docID.String()),
		Cid:        evt.Cid.Bytes(),
		SchemaRoot: []byte(evt.SchemaRoot),
		Creator:    p.host.ID().String(),
		Log: &pb.Document_Log{
			Block: evt.Block,
		},
	}
	req := &pb.PushLogRequest{
		Body: body,
	}

	// push to each peer (replicator)
	p.pushLogToReplicators(evt)

	if err := p.server.publishLog(p.ctx, evt.DocID, req); err != nil {
		return NewErrPublishingToDocIDTopic(err, evt.Cid.String(), evt.DocID)
	}

	if err := p.server.publishLog(p.ctx, evt.SchemaRoot, req); err != nil {
		return NewErrPublishingToSchemaTopic(err, evt.Cid.String(), evt.SchemaRoot)
	}

	return nil
}

func (p *Peer) pushLogToReplicators(lg event.Update) {
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
				if err := p.server.pushLog(p.ctx, lg, peerID); err != nil {
					log.ErrorContextE(
						p.ctx,
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

func (p *Peer) setupBlockService() {
	bswapnet := network.NewFromIpfsHost(p.host, p.dht)
	bswap := bitswap.New(p.ctx, bswapnet, p.blockstore)
	p.bserv = blockservice.New(p.blockstore, bswap)
	p.exch = bswap
}

func stopGRPCServer(ctx context.Context, server *grpc.Server) {
	stopped := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(stopped)
	}()
	timer := time.NewTimer(10 * time.Second)
	select {
	case <-timer.C:
		server.Stop()
		log.InfoContext(ctx, "Peer gRPC server was shutdown ungracefully")
	case <-stopped:
		timer.Stop()
	}
}

// Bootstrap connects to the given peers.
func (p *Peer) Bootstrap(addrs []peer.AddrInfo) {
	var connected uint64

	var wg sync.WaitGroup
	for _, pinfo := range addrs {
		wg.Add(1)
		go func(pinfo peer.AddrInfo) {
			defer wg.Done()
			err := p.host.Connect(p.ctx, pinfo)
			if err != nil {
				log.InfoContext(p.ctx, "Cannot connect to peer", corelog.Any("Error", err))
				return
			}
			log.InfoContext(p.ctx, "Connected", corelog.Any("PeerID", pinfo.ID))
			atomic.AddUint64(&connected, 1)
		}(pinfo)
	}

	wg.Wait()

	if nPeers := len(addrs); int(connected) < nPeers/2 {
		log.InfoContext(p.ctx, fmt.Sprintf("Only connected to %d bootstrap peers out of %d", connected, nPeers))
	}

	err := p.dht.Bootstrap(p.ctx)
	if err != nil {
		log.ErrorContextE(p.ctx, "Problem bootstraping using DHT", err)
		return
	}
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

func newDHT(ctx context.Context, h host.Host, dsb ds.Batching) (*dualdht.DHT, error) {
	dhtOpts := []dualdht.Option{
		dualdht.DHTOption(dht.NamespacedValidator("pk", record.PublicKeyValidator{})),
		dualdht.DHTOption(dht.NamespacedValidator("ipns", ipns.Validator{KeyBook: h.Peerstore()})),
		dualdht.DHTOption(dht.Concurrency(10)),
		dualdht.DHTOption(dht.Mode(dht.ModeAuto)),
	}
	if dsb != nil {
		dhtOpts = append(dhtOpts, dualdht.DHTOption(dht.Datastore(dsb)))
	}

	return dualdht.New(ctx, h, dhtOpts...)
}
