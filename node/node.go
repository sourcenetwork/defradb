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
	"time"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/event"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-peerstore/pstoreds"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/textileio/go-libp2p-pubsub-rpc/finalizer"
	"github.com/textileio/go-threads/broadcast"

	"github.com/sourcenetwork/defradb/client"
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

	host     host.Host
	pubsub   *pubsub.PubSub
	litepeer *ipfslite.Peer

	peerEvent    chan event.EvtPeerConnectednessChanged
	pubSubEvent  chan net.EvtPubSub
	pushLogEvent chan net.EvtReceivedPushLog

	ctx context.Context
}

// NewNode creates a new network node instance of DefraDB, wired into libp2p.
func NewNode(
	ctx context.Context,
	db client.DB,
	bs *broadcast.Broadcaster,
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

	libp2pOpts := []libp2p.Option{
		libp2p.Peerstore(peerstore),
		libp2p.ConnectionManager(options.ConnManager),
	}
	if options.EnableRelay {
		libp2pOpts = append(libp2pOpts, libp2p.EnableRelay())
	}

	h, d, err := ipfslite.SetupLibp2p(
		ctx,
		hostKey,
		nil,
		options.ListenAddrs,
		rootstore,
		libp2pOpts...,
	)
	log.Info(
		ctx,
		"Created LibP2P host",
		logging.NewKV("PeerId", h.ID()),
		logging.NewKV("Address", options.ListenAddrs),
	)
	if err != nil {
		return nil, fin.Cleanup(err)
	}

	bstore := db.Blockstore()
	lite, err := ipfslite.New(ctx, rootstore, bstore, h, d, nil)
	if err != nil {
		return nil, fin.Cleanup(err)
	}

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
		ps,
		bs,
		lite,
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
		pubsub:       ps,
		DB:           db,
		litepeer:     lite,
		ctx:          ctx,
	}

	n.subscribeToPeerConnectionEvents()
	n.subscribeToPubSubEvents()
	n.subscribeToPushLogEvents()

	return n, nil
}

func (n *Node) Boostrap(addrs []peer.AddrInfo) {
	n.litepeer.Bootstrap(addrs)
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
			return fmt.Errorf("waiting for peer connection timed out")
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
			return fmt.Errorf("waiting for pushlog timed out")
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
			return fmt.Errorf("waiting for pushlog timed out")
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

func (n Node) Close() error {
	return n.Peer.Close()
}
