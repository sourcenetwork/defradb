package core

import (
	ipld "github.com/ipfs/go-ipld-format"
)

// Delta represents a delta-state update to delta-CRDT
// They are serialized to and from Protobuf (or CBOR)
type Delta interface {
	GetPriority() uint64
	SetPriority(uint64)
	Marshal() ([]byte, error)
}

type CompositeDelta interface {
	Delta
	Links() map[string]*ipld.Link
}
