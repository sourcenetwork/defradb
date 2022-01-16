package node

import (
	"time"

	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/connmgr"
	ma "github.com/multiformats/go-multiaddr"
)

type Options struct {
	ListenAddrs  []ma.Multiaddr
	RepoPath     string
	ConnManager  connmgr.ConnManager
	EnablePubSub bool
}

type NodeOpt func(*Options) error

// ListenAddrStrings sets the address to listen on given as strings
func ListenAddrStrings(addrs ...string) NodeOpt {
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

// ListenAddrs sets the address to listen on given as MultiAddr(s)
func ListenAddrs(addrs ...ma.Multiaddr) NodeOpt {
	return func(opt *Options) error {
		opt.ListenAddrs = addrs
		return nil
	}
}

// DefaultOpts returns a set of sane defaults for a Node
func DefaultOpts() NodeOpt {
	return func(opt *Options) error {
		if opt.ListenAddrs == nil {
			addr, err := ma.NewMultiaddr("/ip4/0.0.0.0/tcp/0")
			if err != nil {
				return err
			}
			opt.ListenAddrs = []ma.Multiaddr{addr}
		}
		if opt.ConnManager == nil {
			opt.ConnManager = connmgr.NewConnManager(100, 400, time.Second*20)
		}
		return nil
	}
}
