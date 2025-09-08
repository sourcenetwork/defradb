// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package net

import (
	"context"
	"time"

	"github.com/ipfs/boxo/blockservice"
	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	dualdht "github.com/libp2p/go-libp2p-kad-dht/dual"
	record "github.com/libp2p/go-libp2p-record"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	ma "github.com/multiformats/go-multiaddr"
	rpc "github.com/sourcenetwork/go-libp2p-pubsub-rpc"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/net/config"
)

// setupHost returns a host and router configured with the given options.
func setupHost(ctx context.Context, options *config.Options) (host.Host, *dualdht.DHT, error) {
	connManager, err := connmgr.NewConnManager(100, 400, connmgr.WithGracePeriod(time.Second*20))
	if err != nil {
		return nil, nil, err
	}

	dhtOpts := []dualdht.Option{
		dualdht.DHTOption(dht.NamespacedValidator("pk", record.PublicKeyValidator{})),
		dualdht.DHTOption(dht.Concurrency(10)),
		dualdht.DHTOption(dht.Mode(dht.ModeAuto)),
	}

	var ddht *dualdht.DHT
	routing := func(h host.Host) (routing.PeerRouting, error) {
		ddht, err = dualdht.New(ctx, h, dhtOpts...)
		return ddht, err
	}

	libp2pOpts := []libp2p.Option{
		libp2p.ConnectionManager(connManager),
		libp2p.DefaultTransports,
		libp2p.ListenAddrStrings(options.ListenAddresses...),
		libp2p.Routing(routing),
	}

	// relay is enabled by default unless explicitly disabled
	if !options.EnableRelay {
		libp2pOpts = append(libp2pOpts, libp2p.DisableRelay())
	}

	// use the private key from options or generate a random one
	if options.PrivateKey != nil {
		privateKey, err := crypto.UnmarshalEd25519PrivateKey(options.PrivateKey)
		if err != nil {
			return nil, nil, err
		}
		libp2pOpts = append(libp2pOpts, libp2p.Identity(privateKey))
	}

	h, err := libp2p.New(libp2pOpts...)
	if err != nil {
		return nil, nil, err
	}
	return h, ddht, nil
}

var _ client.Host = (*Peer)(nil)

func (p *Peer) ID() string {
	return p.host.ID().String()
}

func (p *Peer) Addrs() []string {
	addrs := []string{}
	for _, addr := range p.host.Addrs() {
		addrs = append(addrs, addr.String())
	}
	return addrs
}

func (p *Peer) PeerInfo() client.PeerInfo {
	return client.PeerInfo{
		ID:        p.ID(),
		Addresses: p.Addrs(),
	}
}

func (p *Peer) Pubkey() ([]byte, error) {
	return crypto.MarshalPublicKey(p.host.Peerstore().PubKey(p.host.ID()))
}

func (p *Peer) Connect(ctx context.Context, info client.PeerInfo) error {
	peerID, err := peer.Decode(info.ID)
	if err != nil {
		return err
	}
	addrs := []ma.Multiaddr{}
	for _, addr := range info.Addresses {
		maddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			return err
		}
		addrs = append(addrs, maddr)
	}

	addrInfo := peer.AddrInfo{
		ID:    peerID,
		Addrs: addrs,
	}
	p.host.Peerstore().AddAddrs(peerID, addrs, peerstore.PermanentAddrTTL)
	return p.host.Connect(ctx, addrInfo)
}

func (p *Peer) Disconnect(ctx context.Context, peerID string) error {
	pid, err := peer.Decode(peerID)
	if err != nil {
		return err
	}
	p.host.Peerstore().ClearAddrs(pid)
	return nil
}

func (p *Peer) Send(ctx context.Context, data []byte, peerID string, protocolID string) error {
	pid, err := peer.Decode(peerID)
	if err != nil {
		return err
	}
	s, err := p.host.NewStream(ctx, pid, protocol.ID(protocolID))
	if err != nil {
		return err
	}
	defer func() {
		closeErr := s.Close()
		err = errors.Join(err, closeErr)
	}()

	_, err = s.Write(data)
	if err != nil {
		resetErr := s.Reset()
		return errors.Join(err, resetErr)
	}
	return s.Close()
}

func (p *Peer) Sign(data []byte) ([]byte, error) {
	return p.host.Peerstore().PrivKey(p.host.ID()).Sign(data)
}

func (p *Peer) SetStreamHandler(protocolID string, handler client.StreamHandler) {
	p.host.SetStreamHandler(protocol.ID(protocolID), func(stream network.Stream) {
		handler(stream, stream.Conn().RemotePeer().String())
	})
}

func (p *Peer) AddPubSubTopic(topicName string, subscribe bool, handler client.PubsubMessageHandler) error {
	messageHandler := func(from peer.ID, topic string, msg []byte) ([]byte, error) {
		return handler(from.String(), topic, msg)
	}
	_, err := p.addPubSubTopic(topicName, subscribe, messageHandler)
	return err
}

func (p *Peer) RemovePubSubTopic(topic string) error {
	return p.removePubSubTopic(topic)
}

// PublishToTopicAsync publishes the given data on the PubSub network via the
// corresponding topic asynchronously.
//
// This is a non blocking operation.
func (p *Peer) PublishToTopicAsync(ctx context.Context, topic string, data []byte) error {
	_, err := p.publishToTopic(ctx, topic, data, rpc.WithIgnoreResponse(true))
	return err
}

// PublishToTopic publishes the given data on the PubSub network via the
// corresponding topic.
//
// It will block until a response is received
func (p *Peer) PublishToTopic(
	ctx context.Context,
	topic string,
	data []byte,
	withMultiResponse bool,
) (<-chan client.PubsubResponse, error) {
	if withMultiResponse {
		return p.publishToTopic(ctx, topic, data, rpc.WithMultiResponse(true))
	}
	return p.publishToTopic(ctx, topic, data)
}

func (p *Peer) publishToTopic(
	ctx context.Context,
	topic string,
	data []byte,
	options ...rpc.PublishOption,
) (<-chan client.PubsubResponse, error) {
	if p.ps == nil { // skip if we aren't running with a pubsub net
		return nil, nil
	}

	p.topicMu.Lock()
	t, ok := p.topics[topic]
	p.topicMu.Unlock()
	if ok {
		resp, err := t.Publish(ctx, data, options...)
		if err != nil {
			return nil, NewErrPushLog(err, errors.NewKV("Topic", topic))
		}
		if resp != nil {
			respChan := make(chan client.PubsubResponse)
			go func() {
				for {
					select {
					case <-ctx.Done():
						close(respChan)
						return
					case r, ok := <-resp:
						if !ok {
							close(respChan)
							return
						}
						respChan <- client.PubsubResponse{
							ID:   r.ID,
							From: r.From.String(),
							Data: r.Data,
							Err:  r.Err,
						}
					}
				}
			}()
			return respChan, nil
		}
		return nil, nil
	}

	// If the topic hasn't been explicitly subscribed to, we temporarily join it
	// to publish the log.
	return nil, p.publishDirectToTopic(ctx, topic, data, false)
}

func (p *Peer) BlockService() blockservice.BlockService {
	return p.blockService
}

func (p *Peer) SetBlockAccessFunc(accessFunc client.BlockAccessFunc) {
	p.accessFuncMu.Lock()
	defer p.accessFuncMu.Unlock()
	p.blockAccessFunc = immutable.Some(accessFunc)
}
