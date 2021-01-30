package clock

import (
	"context"

	"github.com/sourcenetwork/defradb/core"

	// pb "github.com/ipfs/go-ds-crdt/pb"
	cid "github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	dag "github.com/ipfs/go-merkledag"
)

// Credit: This file is from github.com/ipfs/go-ds-crdt

// IPLD related things

var _ core.NodeGetter = (*crdtNodeGetter)(nil)

func init() {
	ipld.Register(cid.DagProtobuf, dag.DecodeProtobufBlock)
}

// crdtNodeGetter wraps an ipld.NodeGetter with some additional utility methods
type crdtNodeGetter struct {
	ipld.NodeGetter
	deltaExtractor func(ipld.Node) (core.Delta, error)
}

func (ng *crdtNodeGetter) GetDelta(ctx context.Context, c cid.Cid) (ipld.Node, core.Delta, error) {
	nd, err := ng.Get(ctx, c)
	if err != nil {
		return nil, nil, err
	}
	delta, err := ng.deltaExtractor(nd)
	return nd, delta, err
}

// GetHeight returns the height of a block
func (ng *crdtNodeGetter) GetPriority(ctx context.Context, c cid.Cid) (uint64, error) {
	_, delta, err := ng.GetDelta(ctx, c)
	if err != nil {
		return 0, err
	}
	return delta.GetPriority(), nil
}

type deltaEntry struct {
	delta core.Delta
	node  ipld.Node
	err   error
}

func (de deltaEntry) GetNode() ipld.Node {
	return de.node
}

func (de deltaEntry) GetDelta() core.Delta {
	return de.delta
}

func (de deltaEntry) Error() error {
	return de.err
}

// GetDeltas uses GetMany to obtain many deltas.
func (ng *crdtNodeGetter) GetDeltas(ctx context.Context, cids []cid.Cid) <-chan core.NodeDeltaPair {
	deltaOpts := make(chan core.NodeDeltaPair, 1)
	go func() {
		defer close(deltaOpts)
		nodeOpts := ng.GetMany(ctx, cids)
		for nodeOpt := range nodeOpts {
			if nodeOpt.Err != nil {
				deltaOpts <- &deltaEntry{err: nodeOpt.Err}
				continue
			}
			delta, err := ng.deltaExtractor(nodeOpt.Node)
			if err != nil {
				deltaOpts <- &deltaEntry{err: err}
				continue
			}
			deltaOpts <- &deltaEntry{
				delta: delta,
				node:  nodeOpt.Node,
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
	// add heads
	for _, h := range heads {
		if err = nd.AddRawLink("head", &ipld.Link{Cid: h}); err != nil {
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
