// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build !js

package tests

import (
	"net"

	"github.com/sourcenetwork/go-p2p"

	"github.com/sourcenetwork/defradb/node"
)

func RandomNetworkingConfig() ConfigureNode {
	return func() []node.Option {
		return []node.Option{
			p2p.WithListenAddresses("/ip4/" + getIPString() + "/tcp/0"),
			p2p.WithEnableRelay(true),
		}
	}
}

func getIPString() string {
	loopbackIP := "127.0.0.1"

	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		// If getting the local address fails, we simply return the loopback address.
		// This would occur if the machine running the tests has no network connection.
		// This will cause the integration tests that depend on DHT relaying of messages to fail.
		return loopbackIP
	}
	defer func() {
		// The test doesn't care about an error on close so we can ignore it.
		_ = conn.Close()
	}()

	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return loopbackIP
	}

	return localAddr.IP.String()
}

func getP2POptions(opts []node.Option) []node.Option {
	netOpts := make([]node.Option, 0)
	for _, opt := range opts {
		if _, ok := opt.(p2p.NodeOpt); ok {
			netOpts = append(netOpts, opt)
		}
	}
	return netOpts
}

func withPrivateKey(opts []node.Option, key []byte) []node.Option {
	return append(opts, p2p.WithPrivateKey(key))
}

func withWithListenAddresses(opts []node.Option, addresses ...string) []node.Option {
	return append(opts, p2p.WithListenAddresses(addresses...))
}
