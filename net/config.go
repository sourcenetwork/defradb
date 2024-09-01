// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/* Node configuration, in which NodeOpt functions are applied on Options. */

package net

import (
	"google.golang.org/grpc"
)

// Options is the node options.
type Options struct {
	ListenAddresses   []string
	PrivateKey        []byte
	EnablePubSub      bool
	EnableRelay       bool
	GRPCServerOptions []grpc.ServerOption
	GRPCDialOptions   []grpc.DialOption
	BootstrapPeers    []string
}

// DefaultOptions returns the default net options.
func DefaultOptions() *Options {
	return &Options{
		ListenAddresses: []string{"/ip4/0.0.0.0/tcp/9171"},
		EnablePubSub:    true,
		EnableRelay:     false,
	}
}

type NodeOpt func(*Options)

// WithPrivateKey sets the p2p host private key.
func WithPrivateKey(priv []byte) NodeOpt {
	return func(opt *Options) {
		opt.PrivateKey = priv
	}
}

// WithEnablePubSub enables the pubsub feature.
func WithEnablePubSub(enable bool) NodeOpt {
	return func(opt *Options) {
		opt.EnablePubSub = enable
	}
}

// WithEnableRelay enables the relay feature.
func WithEnableRelay(enable bool) NodeOpt {
	return func(opt *Options) {
		opt.EnableRelay = enable
	}
}

// WithListenAddress sets the address to listen on given as a multiaddress string.
func WithListenAddresses(addresses ...string) NodeOpt {
	return func(opt *Options) {
		opt.ListenAddresses = addresses
	}
}

// WithBootstrapPeers sets the bootstrap peer addresses to attempt to connect to.
func WithBootstrapPeers(peers ...string) NodeOpt {
	return func(opt *Options) {
		opt.BootstrapPeers = peers
	}
}
