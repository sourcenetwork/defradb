// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package node

import (
	"time"

	connmgr "github.com/libp2p/go-libp2p-connmgr"
	cconnmgr "github.com/libp2p/go-libp2p-core/connmgr"
	ma "github.com/multiformats/go-multiaddr"
	"google.golang.org/grpc"
)

type Options struct {
	ListenAddrs       []ma.Multiaddr
	TCPAddr           ma.Multiaddr
	DataPath          string
	ConnManager       cconnmgr.ConnManager
	EnablePubSub      bool
	GRPCServerOptions []grpc.ServerOption
	GRPCDialOptions   []grpc.DialOption
}

type NodeOpt func(*Options) error

func DataPath(path string) NodeOpt {
	return func(opt *Options) error {
		opt.DataPath = path
		return nil
	}
}

func WithPubSub(enable bool) NodeOpt {
	return func(opt *Options) error {
		opt.EnablePubSub = enable
		return nil
	}
}

// ListenP2PAddrStrings sets the address to listen on given as strings
func ListenP2PAddrStrings(addrs ...string) NodeOpt {
	return func(opt *Options) error {
		for _, addrstr := range addrs {
			a, err := ma.NewMultiaddr(addrstr)
			if err != nil {
				return err
			}
			opt.ListenAddrs = append(opt.ListenAddrs, a)
		}
		return nil
	}
}

// DefaultOpts returns a set of sane defaults for a Node
func DefaultOpts() NodeOpt {
	return func(opt *Options) error {
		if opt.ListenAddrs == nil {
			addr, err := ma.NewMultiaddr("/ip4/0.0.0.0/tcp/9171")
			if err != nil {
				return err
			}
			opt.ListenAddrs = []ma.Multiaddr{addr}
		}
		if opt.TCPAddr == nil {
			addr, err := ma.NewMultiaddr("/ip4/0.0.0.0/tcp/9161")
			if err != nil {
				return err
			}
			opt.TCPAddr = addr
		}
		if opt.ConnManager == nil {
			opt.ConnManager = connmgr.NewConnManager(100, 400, time.Second*20)
		}
		return nil
	}
}
