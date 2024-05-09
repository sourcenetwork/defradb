// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cid

import (
	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
)

func NewDefaultSHA256PrefixV1() cid.Prefix {
	return cid.Prefix{
		Version:  1,
		Codec:    cid.Raw,
		MhType:   mh.SHA2_256,
		MhLength: -1, // default length
	}
}

// NewSHA256CidV1 returns a new CIDv1 with the SHA256 multihash.
func NewSHA256CidV1(data []byte) (cid.Cid, error) {
	// And then feed it some data
	return NewDefaultSHA256PrefixV1().Sum(data)
}
