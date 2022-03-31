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
	"github.com/libp2p/go-libp2p-core/protocol"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	// Name is the protocol slug.
	Name = "defra"

	// @TODO: Register code with Multicodec https://github.com/multiformats/multicodec.
	// 961 is arbitrary at the moment
	// Code is the protocol code.
	Code = 961

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
