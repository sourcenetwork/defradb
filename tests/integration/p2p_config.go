// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

//go:build !js

package tests

import (
	"net"

	"github.com/sourcenetwork/defradb/client/options"
)

func RandomNetworkingConfig() ConfigureNode {
	return func() options.NodeP2POptions {
		return options.NodeP2POptions{
			ListenAddresses:           []string{"/ip4/" + getIPString() + "/tcp/0"},
			EnablePubSub:              true,
			EnableRelay:               true,
			EnableClearBackoffOnRetry: true,
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

func withPrivateKey(p2pOpts *options.NodeP2POptions, key []byte) {
	p2pOpts.PrivateKey = key
}

func withListenAddresses(p2pOpts *options.NodeP2POptions, addresses ...string) {
	p2pOpts.ListenAddresses = addresses
}
