package core

import (
	cid "github.com/ipfs/go-cid"
)

// Delta represents a delta-state update to delta-CRDT
// They are serialized to and from Protobuf (or CBOR)
type Delta interface {
	GetPriority() uint64
	SetPriority(uint64)
	Marshal() ([]byte, error)
	Value() interface{}
}

type CompositeDelta interface {
	Delta
	Links() []DAGLink
}

type DAGLink struct {
	Name string
	Cid  cid.Cid
}
