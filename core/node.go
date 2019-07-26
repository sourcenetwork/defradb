package core

import (
	"context"

	"github.com/ipld/go-ipld"
)

// NodeDeltaPair is a Node with its underlying delta
// already extracted. Used in a channel response for streaming
type NodeDeltaPair interface {
	GetNode() ipld.Node
	GetDelta() Delta
	Error() error
}

// A NodeGetter extended from ipld.NodeGetter with delta related
// functions
type NodeGetter interface {
	ipld.NodeGetter
	GetDelta(context.Context, cid.Cid) (ipld.Node, Delta, error)
	GetDeltas(context.Context, []cid.Cid) <-chan NodeDeltaPair
	GetPriority(context.Context, cid.Cid) (uint64, error)
}
