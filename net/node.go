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
	event1 "github.com/sourcenetwork/defradb/event"
)

var _ client.P2P = (*Node)(nil)

// Node is a networked peer instance of DefraDB.
type Node struct {
	// embed the DB interface into the node
	client.DB

	*Peer

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

	sub, err := h.EventBus().Subscribe(&event.EvtPeerConnectednessChanged{})
	if err != nil {
		return nil, fin.Cleanup(err)
	}
	// publish subscribed events to the event bus
	go func() {
		for val := range sub.Out() {
			db.Events().Publish(event1.NewMessage(event1.PeerName, val))
		}
	}()

	return &Node{
		Peer:     peer,
		DB:       db,
		ctx:      ctx,
		cancel:   cancel,
		dhtClose: ddht.Close,
	}, nil
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
