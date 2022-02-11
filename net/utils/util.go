// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package utils

import (
	"fmt"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

var (
	bootstrapPeers = []string{}
)

func DefaultBoostrapPeers() []peer.AddrInfo {
	ipfspeers := ipfslite.DefaultBootstrapPeers()
	textilepeers, err := ParsePeers(bootstrapPeers)
	if err != nil {
		panic("coudn't parse default bootstrap peers")
	}
	return append(textilepeers, ipfspeers...)
}

func ParsePeers(addrs []string) ([]peer.AddrInfo, error) {
	maddrs := make([]ma.Multiaddr, len(addrs))
	for i, addr := range addrs {
		var err error
		maddrs[i], err = ma.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
	}
	return peer.AddrInfosFromP2pAddrs(maddrs...)
}

func TCPAddrFromMultiAddr(maddr ma.Multiaddr) (string, error) {
	var addr string
	if maddr == nil {
		return addr, fmt.Errorf("address can't be empty")
	}
	ip4, err := maddr.ValueForProtocol(ma.P_IP4)
	if err != nil {
		return addr, err
	}
	tcp, err := maddr.ValueForProtocol(ma.P_TCP)
	if err != nil {
		return addr, err
	}
	return fmt.Sprintf("%s:%s", ip4, tcp), nil
}
