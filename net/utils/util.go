package utils

import (
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
