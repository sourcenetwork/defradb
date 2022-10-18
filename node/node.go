// Copyright 2022 Democratized Data Foundation
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
package node

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	"github.com/ipfs/go-ipns"
	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/event"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	dualdht "github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/libp2p/go-libp2p-peerstore/pstoreds"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	record "github.com/libp2p/go-libp2p-record"
	"github.com/textileio/go-libp2p-pubsub-rpc/finalizer"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/sourcenetwork/defradb/net"
)

var (
	log = logging.MustNewLogger("defra.node")
)

const evtWaitTimeout = 10 * time.Second

type Node struct {
	// embed the DB interface into the node
	client.DB

	*net.Peer

	host   host.Host
	dht    routing.Routing
	pubsub *pubsub.PubSub

	// receives an event when the status of a peer connection changes.
	peerEvent chan event.EvtPeerConnectednessChanged

	// receives an event when a pubsub topic is added.
	pubSubEvent chan net.EvtPubSub

	// receives an event when a pushLog request has been processed.
	pushLogEvent chan net.EvtReceivedPushLog

	ctx context.Context
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
	pstore := namespace.Wrap(rootstore, ds.NewKey("peers"))
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

	peer, err := net.NewPeer(
		ctx,
		db,
		h,
		ddht,
		ps,
		options.TCPAddr,
		options.GRPCServerOptions,
		options.GRPCDialOptions,
	)
	if err != nil {
		return nil, fin.Cleanup(err)
	}

	n := &Node{
		pubSubEvent:  make(chan net.EvtPubSub),
		pushLogEvent: make(chan net.EvtReceivedPushLog),
		peerEvent:    make(chan event.EvtPeerConnectednessChanged),
		Peer:         peer,
		host:         h,
		dht:          ddht,
		pubsub:       ps,
		DB:           db,
		ctx:          ctx,
	}

	n.subscribeToPeerConnectionEvents()
	n.subscribeToPubSubEvents()
	n.subscribeToPushLogEvents()

	return n, nil
}

func (n *Node) Boostrap(addrs []peer.AddrInfo) {
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
			log.Info(n.ctx, "Connected", logging.NewKV("Peer ID", pinfo.ID))
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

// PeerID returns the node's peer ID.
func (n *Node) PeerID() peer.ID {
	return n.host.ID()
}

// subscribeToPeerConnectionEvents subscribes the node to the event bus for a peer connection change.
func (n *Node) subscribeToPeerConnectionEvents() {
	sub, err := n.host.EventBus().Subscribe(new(event.EvtPeerConnectednessChanged))
	if err != nil {
		log.Info(
			n.ctx,
			fmt.Sprintf("failed to subscribe to peer connectedness changed event: %v", err),
		)
	}
	go func() {
		for e := range sub.Out() {
			n.peerEvent <- e.(event.EvtPeerConnectednessChanged)
		}
	}()
}

// subscribeToPubSubEvents subscribes the node to the event bus for a pubsub.
func (n *Node) subscribeToPubSubEvents() {
	sub, err := n.host.EventBus().Subscribe(new(net.EvtPubSub))
	if err != nil {
		log.Info(
			n.ctx,
			fmt.Sprintf("failed to subscribe to pubsub event: %v", err),
		)
	}
	go func() {
		for e := range sub.Out() {
			n.pubSubEvent <- e.(net.EvtPubSub)
		}
	}()
}

// subscribeToPushLogEvents subscribes the node to the event bus for a push log request completion.
func (n *Node) subscribeToPushLogEvents() {
	sub, err := n.host.EventBus().Subscribe(new(net.EvtReceivedPushLog))
	if err != nil {
		log.Info(
			n.ctx,
			fmt.Sprintf("failed to subscribe to push log event: %v", err),
		)
	}
	go func() {
		for e := range sub.Out() {
			n.pushLogEvent <- e.(net.EvtReceivedPushLog)
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
			return errors.New("waiting for peer connection timed out")
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
			return errors.New("waiting for pubsub timed out")
		}
	}
}

// WaitForPushLogEvent listens to the event channel for a push log event from a given peer.
func (n *Node) WaitForPushLogEvent(id peer.ID) error {
	for {
		select {
		case evt := <-n.pushLogEvent:
			if evt.Peer != id {
				continue
			}
			return nil
		case <-time.After(evtWaitTimeout):
			return errors.New("waiting for pushlog timed out")
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

func (n Node) Close() error {
	return n.Peer.Close()
}
