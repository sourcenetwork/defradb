package core

import (
	cid "github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
)

// MerkleClock is the core logical clock implementation that manages
// writing to and from the MerkleDAG structure, ensuring a casual
// ordering of
type MerkleClock interface {
	AddDAGNode(delta Delta) (cid.Cid, error)
	ProcessNode(NodeGetter, cid.Cid, uint64, Delta, ipld.Node) ([]cid.Cid, error)
}
