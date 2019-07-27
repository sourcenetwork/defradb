package core

import (
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld"
	"github.com/sourcenetwork/defradb/core"
)

// MerkleClock is the core logical clock implementation that manages
// writing to and from the MerkleDAG structure, ensuring a casual
// ordering of
type MerkleClock interface {
	AddDAGNode(delta core.Delta) (cid.Cid, error)
	ProcessNode(NodeGetter, cid.Cid, uint64, Delta, ipld.Node) ([]cid.Cid, error)
}
