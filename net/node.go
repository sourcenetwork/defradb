// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
Package node is responsible for interfacing a given DefraDB instance with a networked peer instance
and GRPC server.

Basically it combines db/DB, net/Peer, and net/Server into a single Node object.
*/
package net

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ipfs/boxo/ipns"
	ds "github.com/ipfs/go-datastore"
	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	dualdht "github.com/libp2p/go-libp2p-kad-dht/dual"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	record "github.com/libp2p/go-libp2p-record"
	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/multiformats/go-multiaddr"

	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/go-libp2p-pubsub-rpc/finalizer"

	// @TODO: https://github.com/sourcenetwork/defradb/issues/1902
	//nolint:staticcheck
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoreds"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
)

var evtWaitTimeout = 10 * time.Second

var _ client.P2P = (*Node)(nil)

// Node is a networked peer instance of DefraDB.
type Node struct {
	// embed the DB interface into the node
	client.DB

	*Peer

	// receives an event when the status of a peer connection changes.
	peerEvent chan event.EvtPeerConnectednessChanged

	// receives an event when a pubsub topic is added.
	pubSubEvent chan EvtPubSub

	// receives an event when a pushLog request has been processed.
	pushLogEvent chan EvtReceivedPushLog

	ctx      context.Context
	cancel   context.CancelFunc
	dhtClose func() error
}

// NewNode creates a new network node instance of DefraDB, wired into libp2p.
func NewNode(
	ctx context.Context,
	db client.DB,
	opts ...NodeOpt,
) (node *Node, err error) {
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

	fin := finalizer.NewFinalizer()

	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		if node == nil {
			cancel()
		}
	}()

	peerstore, err := pstoreds.NewPeerstore(ctx, db.Peerstore(), pstoreds.DefaultOpts())
	if err != nil {
		return nil, fin.Cleanup(err)
	}
	fin.Add(peerstore)

	if options.PrivateKey == nil {
		// generate an ephemeral private key
		key, err := crypto.GenerateEd25519()
		if err != nil {
			return nil, fin.Cleanup(err)
		}
		options.PrivateKey = key
	}

	// unmarshal the private key bytes
	privateKey, err := libp2pCrypto.UnmarshalEd25519PrivateKey(options.PrivateKey)
	if err != nil {
		return nil, fin.Cleanup(err)
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
			store := db.Root() // Delete this line once we remove batchable datastore support.
			ddht, err = newDHT(ctx, h, store)
			return ddht, err
		}),
	}
	if !options.EnableRelay {
		libp2pOpts = append(libp2pOpts, libp2p.DisableRelay())
	}

	h, err := libp2p.New(libp2pOpts...)
	if err != nil {
		return nil, fin.Cleanup(err)
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
			return nil, fin.Cleanup(err)
		}
	}
	peer, err := NewPeer(
		ctx,
		db,
		h,
		ddht,
		ps,
		options.GRPCServerOptions,
		options.GRPCDialOptions,
	)
	if err != nil {
		return nil, fin.Cleanup(err)
	}

	n := &Node{
		// WARNING: The current usage of these channels means that consumers of them
		// (the WaitForFoo funcs) can recieve events that occured before the WaitForFoo
		// function call.  This is tolerable at the moment as they are only used for
		// test, but we should resolve this when we can (e.g. via using subscribe-like
		// mechanics, potentially via use of a ring-buffer based [events.Channel]
		// implementation): https://github.com/sourcenetwork/defradb/issues/1358.
		pubSubEvent:  make(chan EvtPubSub, 20),
		pushLogEvent: make(chan EvtReceivedPushLog, 20),
		peerEvent:    make(chan event.EvtPeerConnectednessChanged, 20),
		Peer:         peer,
		DB:           db,
		ctx:          ctx,
		cancel:       cancel,
		dhtClose:     ddht.Close,
	}

	n.subscribeToPeerConnectionEvents()
	n.subscribeToPubSubEvents()
	n.subscribeToPushLogEvents()

	return n, nil
}

// Bootstrap connects to the given peers.
func (n *Node) Bootstrap(addrs []peer.AddrInfo) {
	var connected uint64

	var wg sync.WaitGroup
	for _, pinfo := range addrs {
		wg.Add(1)
		go func(pinfo peer.AddrInfo) {
			defer wg.Done()
			err := n.host.Connect(n.ctx, pinfo)
			if err != nil {
				log.InfoContext(n.ctx, "Cannot connect to peer", corelog.Any("Error", err))
				return
			}
			log.InfoContext(n.ctx, "Connected", corelog.Any("PeerID", pinfo.ID))
			atomic.AddUint64(&connected, 1)
		}(pinfo)
	}

	wg.Wait()

	if nPeers := len(addrs); int(connected) < nPeers/2 {
		log.InfoContext(n.ctx, fmt.Sprintf("Only connected to %d bootstrap peers out of %d", connected, nPeers))
	}

	err := n.dht.Bootstrap(n.ctx)
	if err != nil {
		log.ErrorContextE(n.ctx, "Problem bootstraping using DHT", err)
		return
	}
}

func (n *Node) PeerID() peer.ID {
	return n.host.ID()
}

func (n *Node) ListenAddrs() []multiaddr.Multiaddr {
	return n.host.Network().ListenAddresses()
}

func (n *Node) PeerInfo() peer.AddrInfo {
	return peer.AddrInfo{
		ID:    n.host.ID(),
		Addrs: n.host.Network().ListenAddresses(),
	}
}

// subscribeToPeerConnectionEvents subscribes the node to the event bus for a peer connection change.
func (n *Node) subscribeToPeerConnectionEvents() {
	sub, err := n.host.EventBus().Subscribe(new(event.EvtPeerConnectednessChanged))
	if err != nil {
		log.InfoContext(
			n.ctx,
			fmt.Sprintf("failed to subscribe to peer connectedness changed event: %v", err),
		)
		return
	}
	go func() {
		for {
			select {
			case <-n.ctx.Done():
				err := sub.Close()
				if err != nil {
					log.ErrorContextE(
						n.ctx,
						"Failed to close peer connectedness changed event subscription",
						err,
					)
				}
				return
			case e, ok := <-sub.Out():
				if !ok {
					return
				}
				select {
				case n.peerEvent <- e.(event.EvtPeerConnectednessChanged):
				default:
					<-n.peerEvent
					n.peerEvent <- e.(event.EvtPeerConnectednessChanged)
				}
			}
		}
	}()
}

// subscribeToPubSubEvents subscribes the node to the event bus for a pubsub.
func (n *Node) subscribeToPubSubEvents() {
	sub, err := n.host.EventBus().Subscribe(new(EvtPubSub))
	if err != nil {
		log.InfoContext(
			n.ctx,
			fmt.Sprintf("failed to subscribe to pubsub event: %v", err),
		)
		return
	}
	go func() {
		for {
			select {
			case <-n.ctx.Done():
				err := sub.Close()
				if err != nil {
					log.ErrorContextE(
						n.ctx,
						"Failed to close pubsub event subscription",
						err,
					)
				}
				return
			case e, ok := <-sub.Out():
				if !ok {
					return
				}
				select {
				case n.pubSubEvent <- e.(EvtPubSub):
				default:
					<-n.pubSubEvent
					n.pubSubEvent <- e.(EvtPubSub)
				}
			}
		}
	}()
}

// subscribeToPushLogEvents subscribes the node to the event bus for a push log request completion.
func (n *Node) subscribeToPushLogEvents() {
	sub, err := n.host.EventBus().Subscribe(new(EvtReceivedPushLog))
	if err != nil {
		log.InfoContext(
			n.ctx,
			fmt.Sprintf("failed to subscribe to push log event: %v", err),
		)
		return
	}
	go func() {
		for {
			select {
			case <-n.ctx.Done():
				err := sub.Close()
				if err != nil {
					log.ErrorContextE(
						n.ctx,
						"Failed to close push log event subscription",
						err,
					)
				}
				return
			case e, ok := <-sub.Out():
				if !ok {
					return
				}
				select {
				case n.pushLogEvent <- e.(EvtReceivedPushLog):
				default:
					<-n.pushLogEvent
					n.pushLogEvent <- e.(EvtReceivedPushLog)
				}
			}
		}
	}()
}

// WaitForPeerConnectionEvent listens to the event channel for a connection event from a given peer.
func (n *Node) WaitForPeerConnectionEvent(id peer.ID) error {
	if n.host.Network().Connectedness(id) == network.Connected {
		return nil
	}
	for {
		select {
		case evt := <-n.peerEvent:
			if evt.Peer != id {
				continue
			}
			return nil
		case <-time.After(evtWaitTimeout):
			return ErrPeerConnectionWaitTimout
		case <-n.ctx.Done():
			return nil
		}
	}
}

// WaitForPubSubEvent listens to the event channel for pub sub event from a given peer.
func (n *Node) WaitForPubSubEvent(id peer.ID) error {
	for {
		select {
		case evt := <-n.pubSubEvent:
			if evt.Peer != id {
				continue
			}
			return nil
		case <-time.After(evtWaitTimeout):
			return ErrPubSubWaitTimeout
		case <-n.ctx.Done():
			return nil
		}
	}
}

// WaitForPushLogByPeerEvent listens to the event channel for a push log event by a given peer.
//
// By refers to the log creator. It can be different than the log sender.
//
// It will block the calling thread until an event is yielded to an internal channel. This
// event is not necessarily the next event and is dependent on the number of concurrent callers
// (each event will only notify a single caller, not all of them).
func (n *Node) WaitForPushLogByPeerEvent(id peer.ID) error {
	for {
		select {
		case evt := <-n.pushLogEvent:
			if evt.ByPeer != id {
				continue
			}
			return nil
		case <-time.After(evtWaitTimeout):
			return ErrPushLogWaitTimeout
		case <-n.ctx.Done():
			return nil
		}
	}
}

// WaitForPushLogFromPeerEvent listens to the event channel for a push log event from a given peer.
//
// From refers to the log sender. It can be different that the log creator.
//
// It will block the calling thread until an event is yielded to an internal channel. This
// event is not necessarily the next event and is dependent on the number of concurrent callers
// (each event will only notify a single caller, not all of them).
func (n *Node) WaitForPushLogFromPeerEvent(id peer.ID) error {
	for {
		select {
		case evt := <-n.pushLogEvent:
			if evt.FromPeer != id {
				continue
			}
			return nil
		case <-time.After(evtWaitTimeout):
			return ErrPushLogWaitTimeout
		case <-n.ctx.Done():
			return nil
		}
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

// Close closes the node and all its services.
func (n Node) Close() {
	if n.cancel != nil {
		n.cancel()
	}
	if n.Peer != nil {
		n.Peer.Close()
	}
	if n.dhtClose != nil {
		err := n.dhtClose()
		if err != nil {
			log.ErrorContextE(n.ctx, "Failed to close DHT", err)
		}
	}
	n.DB.Close()
}
