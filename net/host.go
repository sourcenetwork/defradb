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

	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	dualdht "github.com/libp2p/go-libp2p-kad-dht/dual"
	record "github.com/libp2p/go-libp2p-record"
	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
)

// setupHost returns a host and router configured with the given options.
func setupHost(ctx context.Context, options *Options) (host.Host, *dualdht.DHT, error) {
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
		privateKey, err := libp2pCrypto.UnmarshalEd25519PrivateKey(options.PrivateKey)
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
