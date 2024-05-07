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
	"context"

	dag "github.com/ipfs/boxo/ipld/merkledag"
	cid "github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	mh "github.com/multiformats/go-multihash"

	"github.com/sourcenetwork/defradb/internal/core"
)

// Credit: This file is from github.com/ipfs/go-ds-crdt

// IPLD related things

var _ core.NodeGetter = (*CrdtNodeGetter)(nil)

type DeltaExtractorFn func(ipld.Node) (core.Delta, error)

// crdtNodeGetter wraps an ipld.NodeGetter with some additional utility methods
type CrdtNodeGetter struct {
	ipld.NodeGetter
	DeltaExtractor DeltaExtractorFn
}

func (ng *CrdtNodeGetter) GetDelta(ctx context.Context, c cid.Cid) (ipld.Node, core.Delta, error) {
	nd, err := ng.Get(ctx, c)
	if err != nil {
		return nil, nil, err
	}
	delta, err := ng.DeltaExtractor(nd)
	return nd, delta, err
}

// GetHeight returns the height of a block
func (ng *CrdtNodeGetter) GetPriority(ctx context.Context, c cid.Cid) (uint64, error) {
	_, delta, err := ng.GetDelta(ctx, c)
	if err != nil {
		return 0, err
	}
	return delta.GetPriority(), nil
}

type DeltaEntry struct {
	Delta core.Delta
	Node  ipld.Node
	Err   error
}

func (de DeltaEntry) GetNode() ipld.Node {
	return de.Node
}

func (de DeltaEntry) GetDelta() core.Delta {
	return de.Delta
}

func (de DeltaEntry) Error() error {
	return de.Err
}

// GetDeltas uses GetMany to obtain many deltas.
func (ng *CrdtNodeGetter) GetDeltas(ctx context.Context, cids []cid.Cid) <-chan core.NodeDeltaPair {
	deltaOpts := make(chan core.NodeDeltaPair, 1)
	go func() {
		defer close(deltaOpts)
		nodeOpts := ng.GetMany(ctx, cids)
		for nodeOpt := range nodeOpts {
			if nodeOpt.Err != nil {
				deltaOpts <- &DeltaEntry{Err: nodeOpt.Err}
				continue
			}
			delta, err := ng.DeltaExtractor(nodeOpt.Node)
			if err != nil {
				deltaOpts <- &DeltaEntry{Err: err}
				continue
			}
			deltaOpts <- &DeltaEntry{
				Delta: delta,
				Node:  nodeOpt.Node,
			}
		}
	}()
	return deltaOpts
}

// add this as a field to a NodeGetter so it can be typed to a specific
// delta type (ie. LWWRegisterDelta)
// func extractDelta(nd ipld.Node) (core.Delta, error) {
// 	protonode, ok := nd.(*dag.ProtoNode)
// 	if !ok {
// 		return nil, errors.New("node is not a ProtoNode")
// 	}
// 	d := &pb.Delta{}
// 	err := proto.Unmarshal(protonode.Data(), d)
// 	return d, err
// }

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
	err = nd.SetCidBuilder(cid.V1Builder{
		Codec:    cid.DagProtobuf,
		MhType:   mh.SHA2_256,
		MhLength: -1,
	})
	if err != nil {
		return nil, err
	}

	// add heads
	for _, h := range heads {
		if err = nd.AddRawLink("_head", &ipld.Link{Cid: h}); err != nil {
			return nil, err
		}
	}

	// add delta specific links
	if comp, ok := delta.(core.CompositeDelta); ok {
		for _, dagLink := range comp.Links() {
			if err = nd.AddRawLink(dagLink.Name, &ipld.Link{Cid: dagLink.Cid}); err != nil {
				return nil, err
			}
		}
	}
	return nd, nil
}

// type LocalNodeGetter
