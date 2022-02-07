package net

import (
	"github.com/libp2p/go-libp2p-core/protocol"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	// Name is the protocol slug.
	Name = "defra"
	// Code is the protocol code.
	Code = 961 // @TODO: Register code with Multicodec https://github.com/multiformats/multicodec. 961 is arbitrary at the moment
	// Version is the current protocol version.
	Version = "0.0.1"
	// Protocol is the threads protocol tag.
	Protocol protocol.ID = "/" + Name + "/" + Version
)

var addrProtocol = ma.Protocol{
	Name:  Name,
	Code:  Code,
	VCode: ma.CodeToVarint(Code),
	// Size:  ma.LengthPrefixedVarSize,
}

func init() {
	if err := ma.AddProtocol(addrProtocol); err != nil {
		panic(err)
	}
}
