// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package clock

import (
	"github.com/sourcenetwork/defradb/core"

	cid "github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	dag "github.com/ipfs/go-merkledag"
	mh "github.com/multiformats/go-multihash"
)

// Credit: This file is from github.com/ipfs/go-ds-crdt

// IPLD related things

var _ core.NodeGetter = (*CrdtNodeGetter)(nil)

func init() {
	ipld.Register(cid.DagProtobuf, dag.DecodeProtobufBlock)
}

type DeltaExtractorFn func(ipld.Node) (core.Delta, error)

// crdtNodeGetter wraps an ipld.NodeGetter with some additional utility methods
type CrdtNodeGetter struct {
	ipld.NodeGetter
	DeltaExtractor DeltaExtractorFn
}

func makeNode(delta core.Delta, heads []cid.Cid) (ipld.Node, error) {
	var data []byte
	var err error
	if delta != nil {
		data, err = delta.Marshal()
		if err != nil {
			return nil, err
		}
	}

	nd := dag.NodeWithData(data)
	// The cid builder defaults to v0, we want to be using v1 Cids
	nd.SetCidBuilder(
		cid.V1Builder{
			Codec:    cid.DagProtobuf,
			MhType:   mh.SHA2_256,
			MhLength: -1,
		})

	// add heads
	for _, h := range heads {
		if err = nd.AddRawLink("_head", &ipld.Link{Cid: h}); err != nil {
			return nil, err
		}
	}

	// add delta specific links
	if comp, ok := delta.(core.CompositeDelta); ok {
		for _, dlink := range comp.Links() {
			if err = nd.AddRawLink(dlink.Name, &ipld.Link{Cid: dlink.Cid}); err != nil {
				return nil, err
			}
		}
	}
	return nd, nil
}

// type LocalNodeGetter
