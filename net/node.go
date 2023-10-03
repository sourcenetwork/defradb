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
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ipfs/boxo/ipns"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	dualdht "github.com/libp2p/go-libp2p-kad-dht/dual"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	record "github.com/libp2p/go-libp2p-record"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoreds"
	"github.com/multiformats/go-multiaddr"
	"github.com/textileio/go-libp2p-pubsub-rpc/finalizer"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/logging"
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

	ctx    context.Context
	cancel context.CancelFunc
}

// NewNode creates a new network node instance of DefraDB, wired into libp2p.
func NewNode(
	ctx context.Context,
	db client.DB,
	opts ...NodeOpt,
) (*Node, error) {
	options, err := NewMergedOptions(opts...)
	if err != nil {
		return nil, err
	}

	fin := finalizer.NewFinalizer()

	// create our peerstore from the underlying defra rootstore
	// prefixed with "p2p"
	rootstore := db.Root()
	pstore := namespace.Wrap(rootstore, ds.NewKey("/db"))
	peerstore, err := pstoreds.NewPeerstore(ctx, pstore, pstoreds.DefaultOpts())
	if err != nil {
		return nil, fin.Cleanup(err)
	}
	fin.Add(peerstore)

	hostKey, err := getHostKey(options.DataPath)
	if err != nil {
		return nil, fin.Cleanup(err)
	}

	var ddht *dualdht.DHT

	libp2pOpts := []libp2p.Option{
		libp2p.ConnectionManager(options.ConnManager),
		libp2p.DefaultTransports,
		libp2p.Identity(hostKey),
		libp2p.ListenAddrs(options.ListenAddrs...),
		libp2p.Peerstore(peerstore),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			// Delete this line and uncomment the next 6 lines once we remove batchable datastore support.
			// var store ds.Batching
			// // If `rootstore` doesn't implement `Batching`, `nil` will be passed
			// // to newDHT which will cause the DHT to be stored in memory.
			// if dsb, isBatching := rootstore.(ds.Batching); isBatching {
			// 	store = dsb
			// }
			store := rootstore // Delete this line once we remove batchable datastore support.
			ddht, err = newDHT(ctx, h, store)
			return ddht, err
		}),
	}
	if options.EnableRelay {
		libp2pOpts = append(libp2pOpts, libp2p.EnableRelay())
	}

	h, err := libp2p.New(libp2pOpts...)
	if err != nil {
		return nil, fin.Cleanup(err)
	}
	log.Info(
		ctx,
		"Created LibP2P host",
		logging.NewKV("PeerId", h.ID()),
		logging.NewKV("Address", options.ListenAddrs),
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

	ctx, cancel := context.WithCancel(ctx)

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
		cancel()
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
				log.Info(n.ctx, "Cannot connect to peer", logging.NewKV("Error", err))
				return
			}
			log.Info(n.ctx, "Connected", logging.NewKV("PeerID", pinfo.ID))
			atomic.AddUint64(&connected, 1)
		}(pinfo)
	}

	wg.Wait()

	if nPeers := len(addrs); int(connected) < nPeers/2 {
		log.Info(n.ctx, fmt.Sprintf("Only connected to %d bootstrap peers out of %d", connected, nPeers))
	}

	err := n.dht.Bootstrap(n.ctx)
	if err != nil {
		log.ErrorE(n.ctx, "Problem bootstraping using DHT", err)
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
		log.Info(
			n.ctx,
			fmt.Sprintf("failed to subscribe to peer connectedness changed event: %v", err),
		)
		return
	}
	go func() {
		for e := range sub.Out() {
			select {
			case n.peerEvent <- e.(event.EvtPeerConnectednessChanged):
			default:
				<-n.peerEvent
				n.peerEvent <- e.(event.EvtPeerConnectednessChanged)
			}
		}
	}()
}

// subscribeToPubSubEvents subscribes the node to the event bus for a pubsub.
func (n *Node) subscribeToPubSubEvents() {
	sub, err := n.host.EventBus().Subscribe(new(EvtPubSub))
	if err != nil {
		log.Info(
			n.ctx,
			fmt.Sprintf("failed to subscribe to pubsub event: %v", err),
		)
		return
	}
	go func() {
		for e := range sub.Out() {
			select {
			case n.pubSubEvent <- e.(EvtPubSub):
			default:
				<-n.pubSubEvent
				n.pubSubEvent <- e.(EvtPubSub)
			}
		}
	}()
}

// subscribeToPushLogEvents subscribes the node to the event bus for a push log request completion.
func (n *Node) subscribeToPushLogEvents() {
	sub, err := n.host.EventBus().Subscribe(new(EvtReceivedPushLog))
	if err != nil {
		log.Info(
			n.ctx,
			fmt.Sprintf("failed to subscribe to push log event: %v", err),
		)
		return
	}
	go func() {
		for e := range sub.Out() {
			select {
			case n.pushLogEvent <- e.(EvtReceivedPushLog):
			default:
				<-n.pushLogEvent
				n.pushLogEvent <- e.(EvtReceivedPushLog)
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

// replace with proper keystore
func getHostKey(keypath string) (crypto.PrivKey, error) {
	// If a local datastore is used, the key is written to a file
	pth := filepath.Join(keypath, "key")
	_, err := os.Stat(pth)
	if os.IsNotExist(err) {
		key, bytes, err := newHostKey()
		if err != nil {
			return nil, err
		}
		if err := os.MkdirAll(keypath, os.ModePerm); err != nil {
			return nil, err
		}
		if err = os.WriteFile(pth, bytes, 0400); err != nil {
			return nil, err
		}
		return key, nil
	} else if err != nil {
		return nil, err
	} else {
		bytes, err := os.ReadFile(pth)
		if err != nil {
			return nil, err
		}
		return crypto.UnmarshalPrivateKey(bytes)
	}
}

func newHostKey() (crypto.PrivKey, []byte, error) {
	priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 0)
	if err != nil {
		return nil, nil, err
	}
	key, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return nil, nil, err
	}
	return priv, key, nil
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
func (n Node) Close() error {
	n.cancel()
	if n.Peer != nil {
		return n.Peer.Close()
	}
	return nil
}
