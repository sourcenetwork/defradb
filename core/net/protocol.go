// Copyright 2022 Democratized Data Foundation
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
	"github.com/libp2p/go-libp2p/core/protocol"
	ma "github.com/multiformats/go-multiaddr"
)

// DefraDB's P2P protocol information (https://docs.libp2p.io/concepts/protocols/).

const (
	// Name is the protocol slug, the codename representing it.
	Name = "defra"
	// Code is DefraDB's multicodec code.
	Code = 961 // arbitrary
	// Version is the current protocol version.
	Version = "0.0.1"
	// Protocol is the complete libp2p protocol tag.
	Protocol protocol.ID = "/" + Name + "/" + Version
)

func init() {
	var addrProtocol = ma.Protocol{
		Name:  Name,
		Code:  Code,
		VCode: ma.CodeToVarint(Code),
		// Size:  ma.LengthPrefixedVarSize,
	}
	if err := ma.AddProtocol(addrProtocol); err != nil {
		panic(err)
	}
}
