package core

import (
	"github.com/ipfs/go-cid"
)

type MerkleClock interface {
	AddDAGNode(delta core.Delta) (cid.Cid, error)
	ProcessNode(NodeGetter, cid.Cid, uint64, Delta, ipld.Node) ([]cid.Cid, error)
}
